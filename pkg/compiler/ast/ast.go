package ast

import "github.com/agenthands/nforth/pkg/compiler/lexer"

// Node represents any node in the Abstract Syntax Tree.
type Node interface {
	Pos() lexer.Token
}

// Expr represents an expression that yields a value.
type Expr interface {
	Node
	exprNode()
}

// Statement represents a standalone unit of execution.
type Statement interface {
	Node
	stmtNode()
}

// Program is the root node.
type Program struct {
	Nodes []Node
}

// Definition: : NAME { ARGS } BODY ;
type Definition struct {
	Token  lexer.Token
	Name   lexer.Token
	Args   []lexer.Token
	Body   []Statement
}

func (d *Definition) Pos() lexer.Token { return d.Token }

// Assignment: EXPR INTO VAR
type Assignment struct {
	Expression []Expr
	Target     lexer.Token
}

func (a *Assignment) Pos() lexer.Token { return a.Target }
func (a *Assignment) stmtNode()        {}

// VoidOperation: NAME { ARGS }
type VoidOperation struct {
	Token lexer.Token
	Args  []Expr
}

func (v *VoidOperation) Pos() lexer.Token { return v.Token }
func (v *VoidOperation) stmtNode()        {}

// Literal values
type NumberLiteral struct {
	Token lexer.Token
}

func (n *NumberLiteral) Pos() lexer.Token { return n.Token }
func (n *NumberLiteral) exprNode()        {}

type StringLiteral struct {
	Token lexer.Token
}

func (s *StringLiteral) Pos() lexer.Token { return s.Token }
func (s *StringLiteral) exprNode()        {}

type BoolLiteral struct {
	Token lexer.Token
}

func (b *BoolLiteral) Pos() lexer.Token { return b.Token }
func (b *BoolLiteral) exprNode()        {}

type Identifier struct {
	Token lexer.Token
}

func (i *Identifier) Pos() lexer.Token { return i.Token }
func (i *Identifier) exprNode()        {}

// SecurityGate: ADDRESS ENV TOKEN
type SecurityGate struct {
	Token       lexer.Token
	Env         lexer.Token
	CapToken    lexer.Token
	IsSugarGate bool
}

func (s *SecurityGate) Pos() lexer.Token { return s.Token }
func (s *SecurityGate) stmtNode()        {}

// IfStmt: IF block ELSE block THEN
type IfStmt struct {
	Token      lexer.Token
	Condition  Expr
	ThenBranch []Statement
	ElseBranch []Statement
}

func (i *IfStmt) Pos() lexer.Token { return i.Token }
func (i *IfStmt) stmtNode()        {}
