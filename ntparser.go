package triplestore

import (
	"fmt"
	"unicode/utf8"
)

type ntParser struct {
	lex *lexer
}

func newNTParser(s string) *ntParser {
	return &ntParser{
		lex: newLexer(s),
	}
}

func (p *ntParser) parse() []Triple {
	var tris []Triple
	var tok ntToken
	var nodeCount int
	var sub, pred, lit, datatype string
	var isLit, isResource, hasDatatype bool
	var obj object

	reset := func() {
		sub, pred, lit, datatype = "", "", "", ""
		obj = object{}
		isLit, isResource, hasDatatype = false, false, false
		nodeCount = 0
	}

	for tok.kind != EOF_TOK {
		tok = p.lex.nextToken()
		switch tok.kind {
		case COMMENT_TOK:
			continue
		case IRI_TOK:
			nodeCount++
			switch nodeCount {
			case 1:
				sub = tok.lit
			case 2:
				pred = tok.lit
			case 3:
				isResource = true
				lit = tok.lit
			}
		case LIT_TOK:
			isLit = true
			lit = tok.lit
		case DATATYPE_TOK:
			hasDatatype = true
			datatype = tok.lit
		case FULLSTOP_TOK:
			if isResource {
				tris = append(tris, SubjPred(sub, pred).Resource(lit))
			} else if isLit {
				if hasDatatype {
					obj = object{
						isLit: true,
						lit: literal{
							typ: XsdType(datatype),
							val: lit,
						},
					}
					tris = append(tris, SubjPred(sub, pred).Object(obj))
				} else {
					tris = append(tris, SubjPred(sub, pred).StringLiteral(lit))
				}
			}
			reset()
		case UNKNOWN_TOK:
			// noop
		}
	}
	return tris
}

type ntTokenType int

const (
	UNKNOWN_TOK ntTokenType = iota
	IRI_TOK
	EOF_TOK
	WHITESPACE_TOK
	FULLSTOP_TOK
	LIT_TOK
	DATATYPE_TOK
	COMMENT_TOK
)

type ntToken struct {
	kind ntTokenType
	lit  string
}

func iriTok(s string) ntToken      { return ntToken{kind: IRI_TOK, lit: s} }
func litTok(s string) ntToken      { return ntToken{kind: LIT_TOK, lit: s} }
func datatypeTok(s string) ntToken { return ntToken{kind: DATATYPE_TOK, lit: s} }
func commentTok(s string) ntToken  { return ntToken{kind: COMMENT_TOK, lit: s} }
func unknownTok(s string) ntToken  { return ntToken{kind: UNKNOWN_TOK, lit: s} }

var (
	wspaceTok   = ntToken{kind: WHITESPACE_TOK, lit: " "}
	fullstopTok = ntToken{kind: FULLSTOP_TOK, lit: "."}
	eofTok      = ntToken{kind: EOF_TOK}
)

type lexer struct {
	input                  string
	position, readPosition int
	char                   rune
}

func newLexer(s string) *lexer {
	return &lexer{
		input: s,
	}
}

func (l *lexer) nextToken() ntToken {
	l.readChar()
	switch l.char {
	case '<':
		n := l.readIRI()
		return iriTok(n)
	case ' ':
		return wspaceTok
	case '.':
		return fullstopTok
	case '"':
		n := l.readStringLiteral()
		return litTok(n)
	case '^':
		l.readChar()
		if l.char == 0 {
			return eofTok
		}
		if l.char != '^' {
			panic(fmt.Sprintf("invalid datatype: expecting '^', got '%c': input [%s]", l.char, l.input))
		}
		l.readChar()
		if l.char == 0 {
			return eofTok
		}
		if l.char != '<' {
			panic(fmt.Sprintf("invalid datatype: expecting '<', got '%c'. Input: [%s]", l.char, l.input))
		}
		n := l.readIRI()
		return datatypeTok(n)
	case '#':
		l.readChar()
		n := l.readComment()
		return commentTok(n)
	case 0:
		return eofTok
	default:
		return unknownTok(string(l.char))
	}
}

func (l *lexer) readChar() {
	var width int
	if l.readPosition >= len(l.input) {
		l.char = 0
	} else {
		l.char, width = utf8.DecodeRuneInString(l.input[l.readPosition:])
	}
	l.position = l.readPosition
	l.readPosition += width
}

func (l *lexer) peekNextNonWithespaceChar() (found rune, count int) {
	pos := l.readPosition
	if pos >= len(l.input) {
		return 0, 0
	}
	var width int
	for {
		found, width = utf8.DecodeRuneInString(l.input[pos:])
		if width == 0 {
			found = 0
			return
		}
		count++
		if found == ' ' {
			pos = pos + width
			continue
		} else {
			return
		}
	}
}

func (l *lexer) readIRI() string {
	start := l.readPosition
	for {
		l.readChar()
		if l.char == '>' {
			peek, _ := l.peekNextNonWithespaceChar()
			if peek == 0 || peek == '<' || peek == '"' || peek == '.' {
				return l.input[start:l.position]
			}
		}
		if l.char == 0 {
			return ""
		}
	}
}

func (l *lexer) readStringLiteral() string {
	start := l.readPosition
	for {
		l.readChar()
		if l.char == '"' {
			peek, _ := l.peekNextNonWithespaceChar()
			if peek == 0 || peek == '.' || peek == '^' {
				return l.input[start:l.position]
			}
		}
		if l.char == 0 {
			return ""
		}
	}
}

func (l *lexer) readComment() string {
	pos := l.position
	for untilLineEnd(l.char) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func untilLineEnd(c rune) bool {
	return c != '\n' && c != 0
}
