package parser

import (
	"bytes"
	"fmt"
	"github.com/agenthands/nforth/pkg/compiler/ast"
	"github.com/agenthands/nforth/pkg/compiler/lexer"
)

// OpSignature defines the stack effect and security requirements of a word.
type OpSignature struct {
	In, Out       int
	RequiredScope string
}

// StandardWords defines the stack effects for built-in operations.
var StandardWords = map[string]OpSignature{
	"ADD":        {2, 1, ""},
	"SUB":        {2, 1, ""},
	"MUL":        {2, 1, ""},
	"EQ":         {2, 1, ""},
	"GT":         {2, 1, ""},
	"LT":         {2, 1, ""},
	"FETCH":      {1, 1, "HTTP-ENV"},
	"WRITE-FILE": {2, 0, "FS-ENV"},
	"PRINT":      {1, 0, ""},
	"CONTAINS":   {2, 1, ""},
	"ERROR":      {1, 0, ""},
}

type Parser struct {
	scanner *lexer.Scanner
	curTok  lexer.Token
	peekTok lexer.Token

	depth  int      // Virtual Stack Depth
	scopes []string // Active capability scopes
	src    []byte

	functions map[string]int // Name -> Arg Count
}

func NewParser(s *lexer.Scanner, src []byte) *Parser {
	p := &Parser{
		scanner:   s,
		src:       src,
		functions: make(map[string]int),
	}
	// Read two tokens, so curTok and peekTok are both set
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curTok = p.peekTok
	p.peekTok = p.scanner.Next()

	// Skip noise words at the parser level
	if p.peekTok.Kind == lexer.KindNoise {
		p.nextToken()
	}
}

func (p *Parser) Parse() (*ast.Program, error) {
	program := &ast.Program{}

	for p.curTok.Kind != lexer.KindEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			program.Nodes = append(program.Nodes, stmt)
		} else if p.curTok.Kind != lexer.KindEOF {
			// If statement returned nil but not EOF, something is wrong
			return nil, fmt.Errorf("unexpected token at line %d: %v", p.curTok.Line, p.curTok.Kind)
		}

		// The "INTO" Enforcer: Check for Dangling Stack after each top-level statement
		if p.depth != 0 {
			return nil, fmt.Errorf("Syntactic Hallucination Error at line %d: Floating State Detected. Stack depth is %d. All data must be assigned using 'INTO'.", p.curTok.Line, p.depth)
		}
	}

	return program, nil
}

func (p *Parser) parseStatement() (ast.Node, error) {
	switch p.curTok.Kind {
	case lexer.KindIdentifier, lexer.KindNumber, lexer.KindString:
		return p.parseAssignmentOrExpr()
	case lexer.KindAddress, lexer.KindSugarGate:
		return p.parseSecurityGate()
	case lexer.KindExit:
		tok := p.curTok
		p.nextToken()
		if len(p.scopes) > 0 {
			p.scopes = p.scopes[:len(p.scopes)-1]
		}
		return &ast.SecurityGate{Token: tok, IsExit: true}, nil
	case lexer.KindIf:
		return p.parseIfStmt(nil)
	case lexer.KindBegin:
		return p.parseWhileStmt()
	case lexer.KindColon:
		return p.parseDefinition()
	default:
		return nil, fmt.Errorf("unexpected token at line %d: %v", p.curTok.Line, p.curTok.Kind)
	}
}

func (p *Parser) parseIfStmt(setup []ast.Expr) (ast.Node, error) {
	ifTok := p.curTok
	p.nextToken() // skip IF

	if len(setup) == 0 {
		return nil, fmt.Errorf("IF statement missing condition at line %d", ifTok.Line)
	}

	// The condition is the last term in the setup
	condition := setup[len(setup)-1]
	actualSetup := setup[:len(setup)-1]

	// IF consumes the condition (bool)
	p.depth--

	thenBranch, err := p.parseBlock([]lexer.Kind{lexer.KindElse, lexer.KindThen})
	if err != nil {
		return nil, err
	}

	var elseBranch []ast.Statement
	if p.curTok.Kind == lexer.KindElse {
		p.nextToken() // skip ELSE
		elseBranch, err = p.parseBlock([]lexer.Kind{lexer.KindThen})
		if err != nil {
			return nil, err
		}
	}

	if p.curTok.Kind != lexer.KindThen {
		return nil, fmt.Errorf("expected THEN after IF block, got %v", p.curTok.Kind)
	}
	p.nextToken() // skip THEN

	return &ast.IfStmt{
		Token:      ifTok,
		Setup:      actualSetup,
		Condition:  condition,
		ThenBranch: thenBranch,
		ElseBranch: elseBranch,
	}, nil
}

func (p *Parser) parseWhileStmt() (ast.Node, error) {
	beginTok := p.curTok
	p.nextToken() // skip BEGIN

	// 1. Parse setup/condition until WHILE
	var setup []ast.Expr
	for p.curTok.Kind != lexer.KindWhile && p.curTok.Kind != lexer.KindEOF {
		node, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		
		// If it's a VoidOperation, it might be "i limit LT" which is now a VoidOp 
		// because its depth was 0? Wait, LT should have depth 1.
		if vo, ok := node.(*ast.VoidOperation); ok {
			setup = append(setup, vo.Args...)
		} else if e, ok := node.(ast.Expr); ok {
			setup = append(setup, e)
		} else if a, ok := node.(*ast.Assignment); ok {
			// Assignments in setup push their result?
			setup = append(setup, a.Expression...)
		}
	}

	if p.curTok.Kind != lexer.KindWhile {
		return nil, fmt.Errorf("expected WHILE after BEGIN block, got %v", p.curTok.Kind)
	}
	p.nextToken() // skip WHILE
	
	// WHILE consumes the condition bool produced by setup
	p.depth-- 

	// 2. Parse body until REPEAT
	body, err := p.parseBlock([]lexer.Kind{lexer.KindRepeat})
	if err != nil {
		return nil, err
	}

	if p.curTok.Kind != lexer.KindRepeat {
		return nil, fmt.Errorf("expected REPEAT at end of loop, got %v", p.curTok.Kind)
	}
	p.nextToken() // skip REPEAT

	return &ast.WhileStmt{
		Token: beginTok,
		Setup: setup,
		Body:  body,
	}, nil
}

func (p *Parser) parseBlock(terminators []lexer.Kind) ([]ast.Statement, error) {
	var stmts []ast.Statement
	for !p.isTerminator(p.curTok.Kind, terminators) && p.curTok.Kind != lexer.KindEOF {
		node, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s, ok := node.(ast.Statement); ok {
			stmts = append(stmts, s)
		} else if node != nil {
			// Wrap Expr into VoidOperation if found standalone in a block?
			// For now, assume statements are Assignments or VoidOps
		}
	}
	return stmts, nil
}

func (p *Parser) isTerminator(k lexer.Kind, terminators []lexer.Kind) bool {
	for _, t := range terminators {
		if k == t {
			return true
		}
	}
	return false
}

func (p *Parser) parseAssignmentOrExpr() (ast.Node, error) {
	var exprs []ast.Expr

	for p.isTerm(p.curTok.Kind) {
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, expr)

		if p.curTok.Kind == lexer.KindIf {
			return p.parseIfStmt(exprs)
		}
		if p.curTok.Kind == lexer.KindWhile {
			break
		}
		if p.curTok.Kind == lexer.KindRepeat {
			break
		}
		if p.curTok.Kind == lexer.KindSemicolon {
			break
		}
		if p.curTok.Kind == lexer.KindElse || p.curTok.Kind == lexer.KindThen {
			break
		}

		if p.curTok.Kind == lexer.KindInto {
			p.nextToken() // move to target identifier
			if p.curTok.Kind != lexer.KindIdentifier {
				return nil, fmt.Errorf("expected identifier after INTO, got %v", p.curTok.Kind)
			}
			target := p.curTok
			p.nextToken()
			p.depth--
			return &ast.Assignment{
				Expression: exprs,
				Target:     target,
			}, nil
		}

		// End of statement check
		if p.curTok.Kind == lexer.KindEOF || p.curTok.Kind == lexer.KindAddress || p.curTok.Kind == lexer.KindSugarGate || p.curTok.Kind == lexer.KindExit || p.curTok.Kind == lexer.KindColon {
			break
		}
	}

	if len(exprs) > 0 {
		// Return a VoidOperation that contains all terms to be emitted
		lastTok := exprs[len(exprs)-1].Pos()
		return &ast.VoidOperation{Token: lastTok, Args: exprs}, nil
	}

	return nil, nil
}

func (p *Parser) parseExpr() (ast.Expr, error) {
	switch p.curTok.Kind {
	case lexer.KindNumber:
		tok := p.curTok
		p.nextToken()
		p.depth++
		return &ast.NumberLiteral{Token: tok}, nil
	case lexer.KindString:
		tok := p.curTok
		p.nextToken()
		p.depth++
		return &ast.StringLiteral{Token: tok}, nil
	case lexer.KindIdentifier:
		tok := p.curTok
		literal := p.src[tok.Offset : tok.Offset+tok.Length]

		if sig, ok := isStandardWord(literal); ok {
			p.depth -= sig.In
			if p.depth < 0 {
				return nil, fmt.Errorf("Stack Underflow at line %d: word '%s' requires %d arguments", tok.Line, string(literal), sig.In)
			}
			p.depth += sig.Out
			
			// Scope Validation
			if sig.RequiredScope != "" {
				if !p.hasScope(sig.RequiredScope) {
					return nil, fmt.Errorf("Security Violation at line %d: Word '%s' requires scope '%s'. Active scopes: %v", tok.Line, string(literal), sig.RequiredScope, p.scopes)
				}
			}
		} else if argCount, isFunc := p.functions[string(literal)]; isFunc {
			p.depth -= argCount
			if p.depth < 0 {
				return nil, fmt.Errorf("Stack Underflow at line %d: function '%s' requires %d arguments", tok.Line, string(literal), argCount)
			}
			// Functions return void in current spec?
			// Spec says "SQUARE { n } n n MUL INTO res"
			// In Forth, SQUARE would return 1 value.
			// In our current spec, they seem to return void (pop internally).
			// If we want functions to return values, we'd add +1 here.
			// Based on the TestSuite_FunctionDefinitions, it looks like void.
		} else {
			// Assume it's a local variable push
			p.depth++
		}

		p.nextToken()
		return &ast.Identifier{Token: tok}, nil
	default:
		return nil, fmt.Errorf("unexpected expression token: %v", p.curTok.Kind)
	}
}

func (p *Parser) isTerm(k lexer.Kind) bool {
	return k == lexer.KindIdentifier || k == lexer.KindNumber || k == lexer.KindString || k == lexer.KindInto
}

func (p *Parser) hasScope(name string) bool {
	for _, s := range p.scopes {
		if s == name {
			return true
		}
	}
	return false
}

func isStandardWord(lit []byte) (OpSignature, bool) {
	for name, sig := range StandardWords {
		if bytes.Equal(lit, []byte(name)) {
			return sig, true
		}
	}
	return OpSignature{}, false
}

func (p *Parser) parseSecurityGate() (ast.Node, error) {
	tok := p.curTok
	gate := &ast.SecurityGate{Token: tok}

	if tok.Kind == lexer.KindSugarGate {
		gate.IsSugarGate = true
		gate.Env = tok // In scanGateSugar, Offset/Length points to the inner name
		p.scopes = append(p.scopes, string(p.src[tok.Offset:tok.Offset+tok.Length]))
		p.nextToken()
	} else {
		// ADDRESS ENV TOKEN
		p.nextToken() // move to ENV name
		if p.curTok.Kind != lexer.KindIdentifier {
			return nil, fmt.Errorf("expected environment name after ADDRESS")
		}
		gate.Env = p.curTok
		p.scopes = append(p.scopes, string(p.src[p.curTok.Offset:p.curTok.Offset+p.curTok.Length]))

		p.nextToken() // move to capability token
		if p.curTok.Kind != lexer.KindIdentifier && p.curTok.Kind != lexer.KindString {
			return nil, fmt.Errorf("expected capability token identifier or string after environment name")
		}
		gate.CapToken = p.curTok
		p.nextToken() // consume token
	}

	return gate, nil
}

func (p *Parser) parseDefinition() (ast.Node, error) {
	p.nextToken() // skip :
	if p.curTok.Kind != lexer.KindIdentifier {
		return nil, fmt.Errorf("expected function name after ':', got %v", p.curTok.Kind)
	}
	name := p.curTok
	p.nextToken()

	if p.curTok.Kind != lexer.KindLBrace {
		return nil, fmt.Errorf("expected '{' for local variable declaration")
	}
	p.nextToken()

	var args []lexer.Token
	for p.curTok.Kind == lexer.KindIdentifier {
		args = append(args, p.curTok)
		p.nextToken()
	}

	if p.curTok.Kind != lexer.KindRBrace {
		return nil, fmt.Errorf("expected '}' after local variables")
	}
	p.nextToken()

	// Register function for stack effect tracking (Name -> Arg Count)
	nameStr := string(p.src[name.Offset : name.Offset+name.Length])
	p.functions[nameStr] = len(args)

	// Save and reset depth for function scope
	oldDepth := p.depth
	p.depth = 0

	// Body continues until ;
	body, err := p.parseBlock([]lexer.Kind{lexer.KindSemicolon})
	if err != nil {
		return nil, err
	}

	if p.depth != 0 {
		return nil, fmt.Errorf("Syntactic Hallucination Error in function '%s': Floating State Detected. All data must be assigned using 'INTO'.", string(p.src[name.Offset:name.Offset+name.Length]))
	}

	// Restore depth
	p.depth = oldDepth

	if p.curTok.Kind != lexer.KindSemicolon {
		return nil, fmt.Errorf("expected ';' at end of definition")
	}
	p.nextToken()

	// Register function for stack effect tracking (Name -> Arg Count)
	p.functions[string(p.src[name.Offset:name.Offset+name.Length])] = len(args)

	return &ast.Definition{
		Token: name, // Using name as Pos
		Name:  name,
		Args:  args,
		Body:  body,
	}, nil
}
