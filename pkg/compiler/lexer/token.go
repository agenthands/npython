package lexer

// Kind represents the type of token identified by the scanner.
type Kind uint8

const (
	KindEOF Kind = iota
	KindError
	KindIdentifier
	KindNumber
	KindString
	KindInto       // INTO or ->
	KindColon      // :
	KindSemicolon  // ;
	KindLBrace     // {
	KindRBrace     // }
	KindAddress    // ADDRESS
	KindSugarGate  // <ENV-GATE>
	KindExit       // <EXIT>
	KindNoise      // THE, WITH, USING, etc.
	KindIf
	KindElse
	KindThen
)

// Token represents a lexical unit pointing back to the source.
// 12-byte struct to minimize stack overhead and avoid allocations.
type Token struct {
	Kind   Kind
	Offset uint32
	Length uint32
	Line   uint32
}
