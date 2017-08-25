package triplestore

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"unicode/utf8"
)

type lenientNTParser struct {
	lex       *ntLexer
	scanner   *bufio.Scanner
	lineCount int
}

func newLenientNTParser(r io.Reader) *lenientNTParser {
	return &lenientNTParser{
		lex:     new(ntLexer),
		scanner: bufio.NewScanner(r),
	}
}

var (
	commentLine = regexp.MustCompile(`^\s*#.*$`)
	emptyLine   = regexp.MustCompile(`^\s*$`)
)

func (p *lenientNTParser) parse() (triples []Triple, err error) {
	for p.scanner.Scan() {
		if err = p.scanner.Err(); err != nil {
			return
		}
		p.lineCount++
		line := p.scanner.Bytes()
		if emptyLine.Match(line) || commentLine.Match(line) {
			continue
		}
		t, err := p.parseTriple(line)
		if err != nil {
			return triples, fmt.Errorf("ntriples parser: %s", err)
		}
		triples = append(triples, t)
	}
	err = p.scanner.Err()
	return
}

func (p *lenientNTParser) parseTriple(line []byte) (Triple, error) {
	var tok ntToken
	var nodeCount int
	var sub, pred, lit, datatype, langtag string
	var isLit, isResource, isSubBnode, isObjBnode, hasLangtag, hasDatatype, fullStopped bool
	var obj object

	reset := func() {
		sub, pred, lit, datatype, langtag = "", "", "", "", ""
		obj = object{}
		isLit, isResource, isSubBnode, isObjBnode, hasDatatype, hasLangtag, fullStopped = false, false, false, false, false, false, false
		nodeCount = 0
	}

	p.lex.reset(line)

	for tok.kind != EOF_TOK {
		var err error
		tok, err = p.lex.nextToken()
		if err != nil {
			return nil, err
		}
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
		case BNODE_TOK:
			nodeCount++
			switch nodeCount {
			case 1:
				sub = tok.lit
				isSubBnode = true
			case 2:
				return nil, errors.New("blank node can only be subject or object")
			case 3:
				isObjBnode = true
				lit = tok.lit
			}
		case LANGTAG_TOK:
			if nodeCount != 3 {
				return nil, errors.New("langtag misplaced")
			}
			hasLangtag = true
			langtag = tok.lit
		case LIT_TOK:
			if nodeCount != 2 {
				return nil, fmt.Errorf("tok '%s':reaching literate but missing element (node count %d)", tok.lit, nodeCount)
			}
			nodeCount++
			isLit = true
			lit = tok.lit
		case DATATYPE_TOK:
			hasDatatype = true
			datatype = tok.lit
		case FULLSTOP_TOK:
			if nodeCount != 3 {
				return nil, fmt.Errorf("reaching full stop but missing element (node count %d)", nodeCount)
			}
			fullStopped = true
			var tBuilder *tripleBuilder
			if isSubBnode {
				tBuilder = BnodePred(sub, pred)
			} else {
				tBuilder = SubjPred(sub, pred)
			}

			if isResource {
				return tBuilder.Resource(lit), nil
			} else if isObjBnode {
				return tBuilder.Bnode(lit), nil
			} else if isLit {
				if hasDatatype {
					obj = object{
						isLit: true,
						lit: literal{
							typ: XsdType(datatype),
							val: lit,
						},
					}
					return tBuilder.Object(obj), nil
				}
				if hasLangtag {
					return tBuilder.StringLiteralWithLang(lit, langtag), nil
				}
				return tBuilder.StringLiteral(lit), nil
			}
			reset()
		case UNKNOWN_TOK:
			continue
		case LINEFEED_TOK:
			continue
		}
	}

	if nodeCount > 0 {
		return nil, fmt.Errorf("line %d: cannot parse at token '%s' (node count: %d)", p.lineCount, tok.lit, nodeCount)
	}

	if nodeCount != 0 && !fullStopped {
		return nil, errors.New("wrong number of elements")
	}

	return nil, fmt.Errorf("line %d: reached end with no triple", p.lineCount)
}

type ntTokenType int

const (
	UNKNOWN_TOK ntTokenType = iota
	IRI_TOK
	BNODE_TOK
	EOF_TOK
	WHITESPACE_TOK
	FULLSTOP_TOK
	LIT_TOK
	DATATYPE_TOK
	LANGTAG_TOK
	COMMENT_TOK
	LINEFEED_TOK
)

type ntToken struct {
	kind ntTokenType
	lit  string
}

func nodeTok(s string) ntToken     { return ntToken{kind: IRI_TOK, lit: s} }
func bnodeTok(s string) ntToken    { return ntToken{kind: BNODE_TOK, lit: s} }
func litTok(s string) ntToken      { return ntToken{kind: LIT_TOK, lit: s} }
func datatypeTok(s string) ntToken { return ntToken{kind: DATATYPE_TOK, lit: s} }
func langtagTok(s string) ntToken  { return ntToken{kind: LANGTAG_TOK, lit: s} }
func commentTok(s string) ntToken  { return ntToken{kind: COMMENT_TOK, lit: s} }
func unknownTok(s string) ntToken  { return ntToken{kind: UNKNOWN_TOK, lit: s} }

var (
	wspaceTok   = ntToken{kind: WHITESPACE_TOK, lit: " "}
	fullstopTok = ntToken{kind: FULLSTOP_TOK, lit: "."}
	lineFeedTok = ntToken{kind: LINEFEED_TOK, lit: "\n"}
	eofTok      = ntToken{kind: EOF_TOK}
)

type ntLexer struct {
	input        []byte
	current      rune
	width, index int
}

func (l *ntLexer) reset(input []byte) {
	l.input = input
	l.current, l.width, l.index = 0, 0, 0
}

func (l *ntLexer) nextToken() (ntToken, error) {
	if err := l.readRune(); err != nil {
		return ntToken{}, err
	}

	switch l.current {
	case '<':
		n, err := l.readNode()
		return nodeTok(n), err
	case '_':
		if err := l.readRune(); err != nil {
			return ntToken{}, err
		}
		if l.current != ':' {
			return ntToken{}, fmt.Errorf("invalid blank node: expecting ':', got '%c': input [%s]", l.current, l.input)
		}
		n, err := l.readBnode()
		return bnodeTok(n), err
	case ' ':
		return wspaceTok, nil
	case '.':
		return fullstopTok, nil
	case '\n':
		return lineFeedTok, nil
	case '"':
		n, err := l.readStringLiteral()
		return litTok(n), err
	case '@':
		n, err := l.readBnode()
		return langtagTok(n), err
	case '^':
		if err := l.readRune(); err != nil {
			return ntToken{}, err
		}
		if l.current == 0 {
			return eofTok, nil
		}
		if l.current != '^' {
			return ntToken{}, fmt.Errorf("invalid datatype: expecting '^', got '%c': input [%s]", l.current, l.input)
		}
		if err := l.readRune(); err != nil {
			return ntToken{}, err
		}
		if l.current == 0 {
			return eofTok, nil
		}
		if l.current != '<' {
			return ntToken{}, fmt.Errorf("invalid datatype: expecting '<', got '%c'. Input: [%s]", l.current, l.input)
		}
		n, err := l.readNode()
		return datatypeTok(n), err
	case '#':
		n, err := l.readComment()
		return commentTok(n), err
	case 0:
		return eofTok, nil
	default:
		return unknownTok(string(l.current)), nil
	}
}

func (l *ntLexer) readRune() error {
	if len(l.input) < 1 {
		return nil
	}
	l.current, l.width = utf8.DecodeRune(l.input[l.index:])
	if l.current == utf8.RuneError && l.width == 1 {
		return fmt.Errorf("lexer read: invalid utf8 encoding in '%q'", l.input)
	}
	l.index = l.index + l.width
	if l.width == 0 {
		l.current = 0
		return nil
	}
	return nil
}

func (l *ntLexer) unreadRune() error {
	if len(l.input) < 1 {
		return nil
	}
	if l.index < 1 {
		l.current = 0
		return nil
	}

	var last int
	if len(l.input) == 1 {
		last = 1
	} else {
		last = l.index - 1
	}
	if last > len(l.input) {
		l.current = 0
		l.width = 0
		return nil
	}

	l.current, l.width = utf8.DecodeLastRune(l.input[:last])
	if l.current == utf8.RuneError && l.width == 1 {
		return fmt.Errorf("lexer unread: invalid utf8 encoding in '%q'", l.input)
	}
	l.index = l.index - l.width
	if l.width == 0 {
		l.current = 0
		return nil
	}
	return nil
}

func (l *ntLexer) peekNextNonWithespaceRune() (found rune, err error) {
	var count int
	for {
		index := l.index + count
		if len(l.input) < index {
			return 0, nil
		}
		r, w := utf8.DecodeRune(l.input[index:])
		if r == utf8.RuneError && w == 1 {
			return r, fmt.Errorf("lexer read: invalid utf8 encoding in '%q'", l.input)
		}
		if w == 0 {
			return 0, nil
		}
		count++
		if r != ' ' && r != '\t' {
			found = r
			break
		}
	}
	return found, nil
}

func (l *ntLexer) readNode() (string, error) {
	start := l.index
	for {
		if err := l.readRune(); err != nil {
			return "", err
		}
		if l.current == '>' {
			peek, err := l.peekNextNonWithespaceRune()
			if err != nil {
				return "", err
			}
			if peek == 0 || peek == '<' || peek == '"' || peek == '.' || peek == '_' {
				return l.extractFrom(start), nil
			}
		}
		if l.current == 0 {
			return "", nil
		}
	}
}

func (l *ntLexer) readStringLiteral() (string, error) {
	start := l.index
	for {
		if err := l.readRune(); err != nil {
			return "", err
		}
		if l.current == '"' {
			peek, err := l.peekNextNonWithespaceRune()
			if err != nil {
				return "", err
			}
			if peek == 0 || peek == '.' || peek == '^' || peek == '@' {
				return l.extractFrom(start), nil
			}
		}
		if l.current == 0 {
			return "", nil
		}
	}
}

func (l *ntLexer) readBnode() (string, error) {
	start := l.index
	for {
		if err := l.readRune(); err != nil {
			return "", err
		}
		if l.current == ' ' {
			peek, err := l.peekNextNonWithespaceRune()
			if err != nil {
				return "", err
			}
			if peek == 0 || peek == '<' || peek == '.' {
				return l.extractFrom(start), nil
			}
		}
		if l.current == '.' {
			peek, err := l.peekNextNonWithespaceRune()
			if err != nil {
				return "", err
			}
			if peek == 0 || peek == '#' || peek == '\n' { // brittle: but handles <sub> <pred> _:bnode.#commenting
				s := l.extractFrom(start)
				l.unreadRune()
				return s, nil
			}
		}
		if l.current == 0 {
			return "", nil
		}
		if l.current == '<' {
			s := l.extractFrom(start)
			l.unreadRune()
			return s, nil
		}
	}
}

func (l *ntLexer) readComment() (string, error) {
	start := l.index
	for {
		if err := l.readRune(); err != nil {
			return "", err
		}
		if l.current == '\n' {
			s := l.extractFrom(start)
			l.unreadRune()
			return s, nil
		}
		if l.current == 0 {
			return l.extractFrom(start), nil
		}
	}
}

func (l *ntLexer) extractFrom(start int) string {
	return string(l.input[start : l.index-l.width])
}
