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
}

func NewParser(s *lexer.Scanner, src []byte) *Parser {
	p := &Parser{
		scanner: s,
		src:     src,
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
		}

		// The "INTO" Enforcer: Check for Dangling Stack after each statement
		if p.depth != 0 {
			return nil, fmt.Errorf("Floating State Detected at line %d. Stack depth is %d. All data must be assigned using 'INTO'.", p.curTok.Line, p.depth)
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

func (p *Parser) parseBlock(terminators []lexer.Kind) ([]ast.Statement, error) {
	var stmts []ast.Statement
	for !p.isTerminator(p.curTok.Kind, terminators) && p.curTok.Kind != lexer.KindEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if s, ok := stmt.(ast.Statement); ok {
			stmts = append(stmts, s)
		} else if stmt != nil {
			// Handle Node that is not a Statement (e.g. Definition if allowed in blocks)
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
			// The entire expression before IF is part of the setup/condition
			return p.parseIfStmt(exprs)
		}

		if p.curTok.Kind == lexer.KindInto {
			p.nextToken() // move to target identifier

			if p.curTok.Kind != lexer.KindIdentifier {
				return nil, fmt.Errorf("expected identifier after INTO, got %v", p.curTok.Kind)
			}

			target := p.curTok
			p.nextToken()

			// INTO consumes the top of the stack
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
		// If depth is 0, it means the expressions consumed themselves (e.g. 1 2 WRITE)
		// and it is a VoidOperation.
		if p.depth == 0 {
			// Find the last operation token for the VoidOperation
			var lastTok lexer.Token
			if len(exprs) > 0 {
				lastTok = exprs[len(exprs)-1].Pos()
			}
			return &ast.VoidOperation{Token: lastTok, Args: exprs}, nil
		}
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
	// Minimal implementation for now to satisfy Parse() loop
	p.nextToken()
	return nil, nil
}
