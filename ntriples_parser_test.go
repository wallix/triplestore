package triplestore

import (
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
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
			input: "<sub> <pred> \"2\"^^myinteger .\n<sub2> <pred2> <lol2> .",
			expected: []Triple{
				SubjPred("sub", "pred").Object(object{isLit: true, lit: literal{typ: "myinteger", val: "2"}}),
				SubjPred("sub2", "pred2").Resource("lol2"),
			},
		},
	}

	for j, tcase := range tcases {
		p := newNTParser(tcase.input)
		tris := p.parse()
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
func TestLexer(t *testing.T) {
	tcases := []struct {
		input    string
		expected []ntToken
	}{
		{"<node>", []ntToken{ntToken{t: NODE_TOK, lit: "node"}}},
		{"# comment", []ntToken{ntToken{t: COMMENT_TOK, lit: " comment"}}},
		{"\"lit\"", []ntToken{ntToken{t: LIT_TOK, lit: "lit"}}},
		{"^^float", []ntToken{ntToken{t: DATATYPE_TOK, lit: "float"}}},
		{" ", []ntToken{ntToken{t: WHITESPACE_TOK, lit: " "}}},
		{".", []ntToken{ntToken{t: FULLSTOP_TOK, lit: "."}}},

		{"<sub> <pred> \"3\"^^integer .", []ntToken{
			ntToken{t: NODE_TOK, lit: "sub"},
			ntToken{t: WHITESPACE_TOK, lit: " "},
			ntToken{t: NODE_TOK, lit: "pred"},
			ntToken{t: WHITESPACE_TOK, lit: " "},
			ntToken{t: LIT_TOK, lit: "3"},
			ntToken{t: DATATYPE_TOK, lit: "integer"},
			ntToken{t: FULLSTOP_TOK, lit: "."},
		}},
		{"<sub> <pred> \"lit\" . # commenting", []ntToken{
			ntToken{t: NODE_TOK, lit: "sub"},
			ntToken{t: WHITESPACE_TOK, lit: " "},
			ntToken{t: NODE_TOK, lit: "pred"},
			ntToken{t: WHITESPACE_TOK, lit: " "},
			ntToken{t: LIT_TOK, lit: "lit"},
			ntToken{t: WHITESPACE_TOK, lit: " "},
			ntToken{t: FULLSTOP_TOK, lit: "."},
			ntToken{t: WHITESPACE_TOK, lit: " "},
			ntToken{t: COMMENT_TOK, lit: " commenting"},
		}},
	}

	for _, tcase := range tcases {
		l := newLexer(tcase.input)
		var toks []ntToken
		for tok := l.nextToken(); tok.t != EOF_TOK; tok = l.nextToken() {
			toks = append(toks, tok)
		}
		if got, want := toks, tcase.expected; !reflect.DeepEqual(got, want) {
			t.Fatalf("\ngot %#v\n\nwant %#v", got, want)
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
