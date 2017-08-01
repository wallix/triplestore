package triplestore

import (
	"bufio"
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

	for tok.t != EOF_TOK {
		tok = p.lex.nextToken()
		switch tok.t {
		case COMMENT_TOK:
			continue
		case NODE_TOK:
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
	NODE_TOK
	EOF_TOK
	WHITESPACE_TOK
	FULLSTOP_TOK
	LIT_TOK
	DATATYPE_TOK
	COMMENT_TOK
)

type ntToken struct {
	t   ntTokenType
	lit string
}

type lexer struct {
	input                  string
	position, readPosition int
	char                   rune

	reader *bufio.Reader
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
		l.readChar()
		n := l.readNode()
		return ntToken{t: NODE_TOK, lit: n}
	case ' ':
		return ntToken{t: WHITESPACE_TOK, lit: " "}
	case '.':
		return ntToken{t: FULLSTOP_TOK, lit: "."}
	case '"':
		l.readChar()
		n := l.readLit()
		return ntToken{t: LIT_TOK, lit: n}
	case '^':
		l.readChar()
		n := l.readDataType()
		if len(n) > 0 && n[0] != '^' {
			panic("invalid datatype, missing carret")
		}
		return ntToken{t: DATATYPE_TOK, lit: n[1:]}
	case '#':
		l.readChar()
		n := l.readComment()
		return ntToken{t: COMMENT_TOK, lit: n}
	case 0:
		return ntToken{t: EOF_TOK}
	default:
		return ntToken{t: UNKNOWN_TOK, lit: string(l.char)}
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

func (l *lexer) readNode() string {
	pos := l.position
	for untilNodeEnd(l.char) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *lexer) readDataType() string {
	pos := l.position
	for untilDataTypeEnd(l.char) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *lexer) readLit() string {
	pos := l.position
	for untilLitEnd(l.char) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *lexer) readComment() string {
	pos := l.position
	for untilLineEnd(l.char) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func untilNodeEnd(c rune) bool {
	return c != '>' && c != '\n' && c != 0
}

func untilLitEnd(c rune) bool {
	return c != '"' && c != '\n' && c != 0
}

func untilLineEnd(c rune) bool {
	return c != '\n' && c != 0
}

func untilDataTypeEnd(c rune) bool {
	return c != ' ' && c != '\n' && c != 0
}
