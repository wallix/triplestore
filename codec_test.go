package triplestore

import (
	"bytes"
	"math"
	"strings"
	"testing"
	"time"
)

func TestEncodeAndDecodeAllTripleTypes(t *testing.T) {
	tcases := []struct {
		in Triple
	}{
		{SubjPred("", "").Resource("")},
		{SubjPred("one", "two").Resource("three")},

		{SubjPred("", "").StringLiteral("")},
		{SubjPred("one", "two").StringLiteral("three")},

		{SubjPred("one", "two").IntegerLiteral(math.MaxInt64)},
		{SubjPred("one", "two").IntegerLiteral(284765293570)},
		{SubjPred("one", "two").IntegerLiteral(-345293239432)},

		{SubjPred("one", "two").BooleanLiteral(true)},
		{SubjPred("one", "two").BooleanLiteral(false)},

		{SubjPred("one", "two").DateTimeLiteral(time.Now())},

		//large data
		{SubjPred(strings.Repeat("s", 65100), "two").Resource("three")},
		{SubjPred(strings.Repeat("s", 65540), "two").Resource("three")},
		{SubjPred("one", strings.Repeat("t", 65100)).Resource("three")},
		{SubjPred("one", strings.Repeat("t", 65540)).Resource("three")},
		{SubjPred("one", "two").Resource(strings.Repeat("t", 66000))},
		{SubjPred("one", "two").StringLiteral(strings.Repeat("t", 66000))},
	}

	for _, tcase := range tcases {
		var buff bytes.Buffer
		enc := NewBinaryEncoder(&buff)

		if err := enc.Encode(tcase.in); err != nil {
			t.Fatal(err)
		}

		dec := NewBinaryDecoder(&buff)
		all, err := dec.Decode()
		if err != nil {
			t.Fatal(err)
		}

		if got, want := len(all), 1; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}

		if got, want := tcase.in, all[0]; !got.Equal(want) {
			t.Fatalf("case %v: \ngot\n%v\nwant\n%v\n", tcase.in, got, want)
		}
	}
}

func TestEncodeNTriples(t *testing.T) {
	triples := []Triple{
		SubjPred("http://test.url#one", "http://www.w3.org/1999/02/22-rdf-syntax-ns#type").Resource("http://test.url#onetype"),
		SubjPred("http://test.url#one", "http://test.url#prop1").StringLiteral("two"),
		SubjPred("http://test.url#one", "http://test.url#prop2").IntegerLiteral(284765293570),
		SubjPred("http://test.url#one", "http://test.url#prop3").BooleanLiteral(true),
		SubjPred("http://test.url#one", "http://test.url#prop4").DateTimeLiteral(time.Unix(1233456789, 0).UTC()),
	}

	var buff bytes.Buffer
	enc := NewNTriplesEncoder(&buff)
	if err := enc.Encode(triples...); err != nil {
		t.Fatal(err)
	}

	expect := `<http://test.url#one> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://test.url#onetype> .
<http://test.url#one> <http://test.url#prop1> "two" .
<http://test.url#one> <http://test.url#prop2> "284765293570"^^<http://www.w3.org/2001/XMLSchema#integer> .
<http://test.url#one> <http://test.url#prop3> "true"^^<http://www.w3.org/2001/XMLSchema#boolean> .
<http://test.url#one> <http://test.url#prop4> "2009-02-01T02:53:09Z"^^<http://www.w3.org/2001/XMLSchema#dateTime> .
`
	if got, want := buff.String(), expect; got != want {
		t.Fatalf("got \n%s\nwant \n%s\n", got, want)
	}
}
