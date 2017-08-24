package triplestore

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestNTParser(t *testing.T) {
	tcases := []struct {
		input    string
		expected []Triple
	}{
		{
			input: "<sub> <pred> <lol> .\n<sub2> <pred2> \"lol2\" .",
			expected: []Triple{
				SubjPred("sub", "pred").Resource("lol"),
				mustTriple("sub2", "pred2", "lol2"),
			},
		},
		{
			input: "<sub> <pred> \"2\"^^<myinteger> .\n<sub2> <pred2> <lol2> .",
			expected: []Triple{
				SubjPred("sub", "pred").Object(object{isLit: true, lit: literal{typ: "myinteger", val: "2"}}),
				SubjPred("sub2", "pred2").Resource("lol2"),
			},
		},
		{
			input:    `<sub> <pred> "dquote:\"" .\n`,
			expected: []Triple{SubjPred("sub", "pred").StringLiteral(`dquote:\"`)},
		},
		{
			input:    "<sub> <pred> _:anon.\n",
			expected: []Triple{SubjPred("sub", "pred").Bnode("anon")},
		},
		{
			input:    `<sub> <pred> _:anon.`,
			expected: []Triple{SubjPred("sub", "pred").Bnode("anon")},
		},
		{
			input:    "<sub> <pred> \"\u00E9\".\n",
			expected: []Triple{SubjPred("sub", "pred").StringLiteral("é")},
		},
		{
			input:    "<sub> <pred> \"\u00E9\".",
			expected: []Triple{SubjPred("sub", "pred").StringLiteral("é")},
		},
		{
			input:    "<sub> <pred> \"\032\".",
			expected: []Triple{SubjPred("sub", "pred").StringLiteral("\032")},
		},
		{
			input:    "<sub> <pred> \"\x1A\".",
			expected: []Triple{SubjPred("sub", "pred").StringLiteral("\x1A")},
		},
	}

	for j, tcase := range tcases {
		p := newLenientNTParser(tcase.input)
		tris, err := p.parse()
		if err != nil {
			t.Fatalf("input=[%s]: %s", tcase.input, err)
		}
		if got, want := len(tris), len(tcase.expected); got != want {
			t.Fatalf("triples size (case %d): got %d, want %d", j+1, got, want)
		}
		for i, tri := range tris {
			if got, want := tri, tcase.expected[i]; !got.Equal(want) {
				t.Fatalf("triple (%d)\ngot %v\n\nwant %v", i+1, got, want)
			}
		}
	}
}

func TestParserErrorHandling(t *testing.T) {
	tcases := []struct {
		input       string
		errContains string
	}{
		{input: "<sub> <pred> 1 ."},
		//{input: "<one> <two> <three>, <four> ."}, passes
	}

	for _, tcase := range tcases {
		tris, err := newLenientNTParser(tcase.input).parse()
		if err == nil {
			t.Fatalf("expected err, got none. Triples parsed:\n%#v", Triples(tris).Map(func(tr Triple) string { return fmt.Sprint(tr) }))
		}
		if msg := tcase.errContains; msg != "" {
			if !strings.Contains(err.Error(), msg) {
				t.Fatalf("expected '%s' to contains '%s'", err.Error(), tcase.errContains)
			}
		}
	}
}

func TestLexer(t *testing.T) {
	tcases := []struct {
		input    string
		expected []ntToken
	}{
		// simple
		{"<node>", []ntToken{iriTok("node")}},
		{"_:bnode .", []ntToken{bnodeTok("bnode"), fullstopTok}},
		{"_:bnode <pred>", []ntToken{bnodeTok("bnode"), iriTok("pred")}},
		{"#comment", []ntToken{commentTok("comment")}},
		{"# comment", []ntToken{commentTok(" comment")}},
		{"\"lit\"", []ntToken{litTok("lit")}},
		{"^^<xsd:float>", []ntToken{datatypeTok("xsd:float")}},
		{" ", []ntToken{wspaceTok}},
		{".", []ntToken{fullstopTok}},
		{"\n", []ntToken{lineFeedTok}},
		{"# comment\n", []ntToken{commentTok(" comment"), lineFeedTok}},
		{"@en .", []ntToken{langtagTok("en"), fullstopTok}},
		{"@en.", []ntToken{langtagTok("en"), fullstopTok}},
		{"@en .\n", []ntToken{langtagTok("en"), fullstopTok, lineFeedTok}},

		{"#", []ntToken{commentTok("")}}, // fixed with go-fuzz

		// escaped
		{`<no>de>`, []ntToken{iriTok("no>de")}},
		{`<no\>de>`, []ntToken{iriTok("no\\>de")}},
		{`<node\\>`, []ntToken{iriTok("node\\\\")}},
		{`"\\"`, []ntToken{litTok(`\\`)}},
		{`"quot"ed"`, []ntToken{litTok(`quot"ed`)}},
		{`"quot\"ed"`, []ntToken{litTok("quot\\\"ed")}},

		// triples
		{"<sub> <pred> \"3\"^^<xsd:integer> .", []ntToken{
			iriTok("sub"), wspaceTok, iriTok("pred"), wspaceTok, litTok("3"),
			datatypeTok("xsd:integer"), wspaceTok, fullstopTok,
		}},
		{"<sub><pred>\"3\"^^<xsd:integer>.", []ntToken{
			iriTok("sub"), iriTok("pred"), litTok("3"), datatypeTok("xsd:integer"), fullstopTok,
		}},
		{"<sub> <pred> \"lit\" . # commenting", []ntToken{
			iriTok("sub"), wspaceTok, iriTok("pred"), wspaceTok, litTok("lit"),
			wspaceTok, fullstopTok, wspaceTok, commentTok(" commenting"),
		}},
		{"<sub><pred>\"lit\".#commenting", []ntToken{
			iriTok("sub"), iriTok("pred"), litTok("lit"), fullstopTok, commentTok("commenting"),
		}},

		// triple with bnodes
		{"_:sub <pred>\"lit\".#commenting", []ntToken{
			bnodeTok("sub"), iriTok("pred"), litTok("lit"), fullstopTok, commentTok("commenting"),
		}},
		{"<sub> <pred> _:lit . #commenting", []ntToken{
			iriTok("sub"), wspaceTok, iriTok("pred"), wspaceTok, bnodeTok("lit"), fullstopTok, wspaceTok, commentTok("commenting"),
		}},
		{"_:sub<pred>_:lit.#commenting", []ntToken{
			bnodeTok("sub"), iriTok("pred"), bnodeTok("lit"), fullstopTok, commentTok("commenting"),
		}},

		// triples with langtag
		{`<sub> <pred> "lit"@russ . # commenting`, []ntToken{
			iriTok("sub"), wspaceTok, iriTok("pred"), wspaceTok, litTok("lit"),
			langtagTok("russ"), fullstopTok, wspaceTok, commentTok(" commenting"),
		}},
	}

	for i, tcase := range tcases {
		l := newLexer(tcase.input)
		var toks []ntToken
		for tok, _ := l.nextToken(); tok.kind != EOF_TOK; tok, _ = l.nextToken() {
			toks = append(toks, tok)
		}
		if got, want := toks, tcase.expected; !reflect.DeepEqual(got, want) {
			t.Fatalf("case %d input=[%s]\ngot %#v\n\nwant %#v", i+1, tcase.input, got, want)
		}
	}
}

func TestLexerReadIRI(t *testing.T) {
	tcases := []struct {
		input string
		node  string
	}{
		{"<", ""},
		{">", ""},
		{" >", " "},
		{"", ""},
		{"z", ""},
		{`\z>`, "\\z"},
		{"\n>", "\n"},
		{"subject>", "subject"},
		{"s  ubject>", "s  ubject"},
		{"subject>   <", "subject"},
		{"    subject>   <", "    subject"},
		{"subject><", "subject"},
		{"subje   ct><", "subje   ct"},
		{"sub>ject>", "sub>ject"},
		{"sub > ject>", "sub > ject"},
		{"sub>ject>      ", "sub>ject"},
		{"subject", ""},

		{"pred>   \"", "pred"},
		{"pred>\"", "pred"},

		{"resource>.", "resource"},
		{"resource> .", "resource"},
		{"resource>> .", "resource>"},
		{"resource>  .   ", "resource"},
	}

	for i, tcase := range tcases {
		l := newLexer(tcase.input)
		n, err := l.readIRI()
		if err != nil {
			t.Fatalf("case %d: '%s': %s", i+1, tcase.input, err)
		}
		if got, want := n, tcase.node; got != want {
			t.Fatalf("case %d '%s': got '%s', want '%s'", i+1, tcase.input, got, want)
		}
	}
}

func TestLexerReadBnode(t *testing.T) {
	tcases := []struct {
		input string
		node  string
	}{
		{"a .", "a"},
		{"a<", "a"},
		{"a    <", "a"},
		{"a <", "a"},
		{"a .", "a"},
		{"a     .", "a"},
		{"a.\n", "a"},
	}

	for i, tcase := range tcases {
		l := newLexer(tcase.input)
		n, err := l.readBnode()
		if err != nil {
			t.Fatalf("case %d: '%s': %s", i+1, tcase.input, err)
		}
		if got, want := n, tcase.node; got != want {
			t.Fatalf("case %d '%s': got '%s', want '%s'", i+1, tcase.input, got, want)
		}
	}
}
func TestLexerReadStringLiteral(t *testing.T) {
	tcases := []struct {
		input string
		node  string
	}{
		{"", ""},
		{`"`, ""},
		{`  "`, "  "},
		{"z", ""},
		{"\u00E9\" .", "\u00E9"},
		{`\n"`, "\\n"},
		{`lit"`, "lit"},
		{`l it"`, "l it"},
		{"li\"t\"", "li\"t"},
		{"li \"t\"", "li \"t"},
		{"li\"t\" .", "li\"t"},
		{"li\"t\".", "li\"t"},
		{"li\"t\" .", "li\"t"},
		{"li\"t\"  .  ", "li\"t"},
		{"li\"t\"^", "li\"t"},
		{"li\"t\"^^", "li\"t"},
		{"li\"t\" ^", "li\"t"},
		{"li\"t\" ^^", "li\"t"},
		{"li\"t\"   ^", "li\"t"},
		{"li\"t\"     ^^", "li\"t"},
	}

	for i, tcase := range tcases {
		l := newLexer(tcase.input)
		n, err := l.readStringLiteral()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := n, tcase.node; got != want {
			t.Fatalf("case %d: got '%s', want '%s'", i+1, got, want)
		}
	}
}

func TestLexerReadComment(t *testing.T) {
	tcases := []struct {
		input string
		node  string
	}{
		//{"", ""},
		//{"comment  ", "comment "},
		//{"comment  \n", "comment "},
		{"a", "a"},
		//{"\n", ""},
	}

	for i, tcase := range tcases {
		l := newLexer(tcase.input)
		n, err := l.readComment()
		if err != nil {
			t.Fatal(err)
		}
		if got, want := n, tcase.node; got != want {
			t.Fatalf("case %d: got '%s', want '%s'", i+1, got, want)
		}
	}
}

func mustTriple(s, p string, i interface{}) Triple {
	t, err := SubjPredLit(s, p, i)
	if err != nil {
		panic(err)
	}
	return t
}
