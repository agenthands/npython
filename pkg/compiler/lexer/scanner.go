package lexer

import (
	"bytes"
)

// Scanner performs lexical analysis on nFORTH source.
type Scanner struct {
	source []byte
	cursor int
	line   int
}

// NewScanner creates a new scanner for the given source.
func NewScanner(source []byte) *Scanner {
	return &Scanner{
		source: source,
		line:   1,
	}
}

// Reset re-initializes the scanner with new source for pool reuse.
func (s *Scanner) Reset(source []byte) {
	s.source = source
	s.cursor = 0
	s.line = 1
}

// Next returns the next token from the source.
func (s *Scanner) Next() Token {
	s.skipWhitespace()

	if s.cursor >= len(s.source) {
		return Token{Kind: KindEOF, Line: uint32(s.line)}
	}

	start := s.cursor
	ch := s.source[s.cursor]

	// 1. Handle Comments (\ ...)
	if ch == '\\' {
		s.skipComment()
		return s.Next()
	}

	// 2. Handle Sugar Gates (<ENV-GATE>)
	if ch == '<' {
		return s.scanGateSugar()
	}

	// 3. Handle Assignment Aliases (->) and Comparison (!=)
	if ch == '-' && s.peek() == '>' {
		s.cursor += 2
		return Token{Kind: KindInto, Offset: uint32(start), Length: 2, Line: uint32(s.line)}
	}
	if ch == '!' && s.peek() == '=' {
		s.cursor += 2
		return Token{Kind: KindIdentifier, Offset: uint32(start), Length: 2, Line: uint32(s.line)}
	}

	// 4. Handle Strings
	if ch == '"' {
		return s.scanString()
	}

	// 5. Handle Numbers and Identifiers
	if isDigit(ch) || (ch == '-' && isDigit(s.peek())) {
		return s.scanNumber()
	}

	if isAlpha(ch) {
		return s.scanIdentifier()
	}

	// 6. Handle Punctuation
	s.cursor++
	kind := KindError
	switch ch {
	case ':':
		kind = KindColon
	case ';':
		kind = KindSemicolon
	case '{':
		kind = KindLBrace
	case '}':
		kind = KindRBrace
	}

	return Token{Kind: kind, Offset: uint32(start), Length: 1, Line: uint32(s.line)}
}

func (s *Scanner) skipWhitespace() {
	for s.cursor < len(s.source) {
		ch := s.source[s.cursor]
		if ch == ' ' || ch == '\t' || ch == '\r' {
			s.cursor++
		} else if ch == '\n' {
			s.line++
			s.cursor++
		} else {
			break
		}
	}
}

func (s *Scanner) skipComment() {
	for s.cursor < len(s.source) && s.source[s.cursor] != '\n' {
		s.cursor++
	}
}

func (s *Scanner) scanGateSugar() Token {
	start := s.cursor
	s.cursor++ // Skip '<'
	
	innerStart := s.cursor
	for s.cursor < len(s.source) && s.source[s.cursor] != '>' {
		s.cursor++
	}
	
	if s.cursor >= len(s.source) {
		return Token{Kind: KindError, Offset: uint32(start), Length: uint32(s.cursor - start), Line: uint32(s.line)}
	}
	
	s.cursor++ // Skip '>'
	
	literal := s.source[innerStart : s.cursor-1]
	if bytes.Equal(literal, []byte("EXIT")) {
		return Token{Kind: KindExit, Offset: uint32(innerStart), Length: uint32(len(literal)), Line: uint32(s.line)}
	}
	
	return Token{Kind: KindSugarGate, Offset: uint32(innerStart), Length: uint32(s.cursor - 1 - innerStart), Line: uint32(s.line)}
}

func (s *Scanner) scanString() Token {
	start := s.cursor
	s.cursor++ // Skip opening '"'
	for s.cursor < len(s.source) && s.source[s.cursor] != '"' {
		if s.source[s.cursor] == '\n' {
			s.line++
		}
		s.cursor++
	}
	
	if s.cursor >= len(s.source) {
		return Token{Kind: KindError, Offset: uint32(start), Length: uint32(s.cursor - start), Line: uint32(s.line)}
	}
	
	s.cursor++ // Skip closing '"'
	return Token{Kind: KindString, Offset: uint32(start), Length: uint32(s.cursor - start), Line: uint32(s.line)}
}

func (s *Scanner) scanNumber() Token {
	start := s.cursor
	if s.source[s.cursor] == '-' {
		s.cursor++
	}
	for s.cursor < len(s.source) && isDigit(s.source[s.cursor]) {
		s.cursor++
	}
	return Token{Kind: KindNumber, Offset: uint32(start), Length: uint32(s.cursor - start), Line: uint32(s.line)}
}

func (s *Scanner) scanIdentifier() Token {
	start := s.cursor
	for s.cursor < len(s.source) && (isAlpha(s.source[s.cursor]) || isDigit(s.source[s.cursor]) || s.source[s.cursor] == '-' || s.source[s.cursor] == '_') {
		s.cursor++
	}
	
	literal := s.source[start:s.cursor]
	kind := KindIdentifier
	
	// Map keywords and noise words
	if bytes.Equal(literal, []byte("INTO")) {
		kind = KindInto
	} else if bytes.Equal(literal, []byte("ADDRESS")) {
		kind = KindAddress
	} else if bytes.Equal(literal, []byte("IF")) {
		kind = KindIf
	} else if bytes.Equal(literal, []byte("ELSE")) {
		kind = KindElse
	} else if bytes.Equal(literal, []byte("THEN")) {
		kind = KindThen
	} else if bytes.Equal(literal, []byte("BEGIN")) {
		kind = KindBegin
	} else if bytes.Equal(literal, []byte("WHILE")) {
		kind = KindWhile
	} else if bytes.Equal(literal, []byte("REPEAT")) {
		kind = KindRepeat
	} else if isNoise(literal) {
		kind = KindNoise
	}
	
	return Token{Kind: kind, Offset: uint32(start), Length: uint32(s.cursor - start), Line: uint32(s.line)}
}

func (s *Scanner) peek() byte {
	if s.cursor+1 >= len(s.source) {
		return 0
	}
	return s.source[s.cursor+1]
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isNoise(lit []byte) bool {
	noise := [][]byte{
		[]byte("THE"), []byte("WITH"), []byte("USING"), []byte("FROM"),
	}
	for _, n := range noise {
		if bytes.Equal(lit, n) {
			return true
		}
	}
	return false
}
