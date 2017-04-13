package triplestore

import (
	"bytes"
	"io/ioutil"
	"math"
	"os"
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

func TestEncodeDecodeOnFile(t *testing.T) {
	one := SubjPred("one", "pred1").StringLiteral("lit1")
	two := SubjPred("two", "pred2").StringLiteral("lit2")

	file, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	if err = NewBinaryEncoder(file).Encode(one, two); err != nil {
		t.Fatal(err)
	}

	file.Seek(0, 0)
	tris, err := NewBinaryDecoder(file).Decode()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(tris), 2; got != want {
		t.Fatalf("got %d, want %d", got, want)
	}

	s := NewSource()
	s.Add(tris...)
	snap := s.Snapshot()

	if !snap.Contains(one) {
		t.Fatalf("decoded file should contains %v", one)
	}
	if !snap.Contains(two) {
		t.Fatalf("decoded file should contains %v", two)
	}
}

func TestDecodeDataset(t *testing.T) {
	one := SubjPred("one", "pred1").StringLiteral("lit1")
	two := SubjPred("two", "pred2").StringLiteral("lit2")

	firstFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(firstFile.Name())

	secondFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(secondFile.Name())

	if err = NewBinaryEncoder(firstFile).Encode(one); err != nil {
		t.Fatal(err)
	}

	if err = NewBinaryEncoder(secondFile).Encode(two); err != nil {
		t.Fatal(err)
	}

	firstFile.Seek(0, 0)
	secondFile.Seek(0, 0)

	dec := NewDatasetDecoder(NewBinaryDecoder, firstFile, secondFile)
	tris, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}

	s := NewSource()
	s.Add(tris...)
	snap := s.Snapshot()
	if got, want := snap.Count(), 2; got != want {
		t.Fatalf("got %d, want %d", got, want)
	}

	if !snap.Contains(one) {
		t.Fatalf("decoded dataset should contains %v", one)
	}

	if !snap.Contains(two) {
		t.Fatalf("decoded dataset should contains %v", two)
	}
}

func TestEncodeNTriples(t *testing.T) {
	triples := []Triple{
		SubjPred("one", "rdf:type").Resource("onetype"),
		SubjPred("one", "prop1").StringLiteral("two"),
		SubjPred("http://my-url-to.test/#one", "prop2").IntegerLiteral(284765293570),
		SubjPred("one", "prop3").BooleanLiteral(true),
		SubjPred("one", "cloud:launched").DateTimeLiteral(time.Unix(1233456789, 0).UTC()),
		SubjPred("co<mplex", "\"with>").StringLiteral("with\"special<chars."),
		SubjPred("one", "with spaces").Resource("10 inbound-smtp.eu-west-1.amazonaws.com."),
	}

	var buff bytes.Buffer
	enc := NewNTriplesEncoderWithContext(&buff, &Context{Base: "http://test.url#",
		Prefixes: map[string]string{
			"rdf":   "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
			"cloud": "http://awless.io/rdf/cloud#",
		}})
	if err := enc.Encode(triples...); err != nil {
		t.Fatal(err)
	}

	expect := `<http://test.url#one> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://test.url#onetype> .
<http://test.url#one> <http://test.url#prop1> "two" .
<http://my-url-to.test/#one> <http://test.url#prop2> "284765293570"^^<http://www.w3.org/2001/XMLSchema#integer> .
<http://test.url#one> <http://test.url#prop3> "true"^^<http://www.w3.org/2001/XMLSchema#boolean> .
<http://test.url#one> <http://awless.io/rdf/cloud#launched> "2009-02-01T02:53:09Z"^^<http://www.w3.org/2001/XMLSchema#dateTime> .
<http://test.url#co%3Cmplex> <http://test.url#%22with%3E> "with\"special<chars." .
<http://test.url#one> <http://test.url#with+spaces> <http://test.url#10+inbound-smtp.eu-west-1.amazonaws.com.> .`
	if got, want := buff.String(), expect; got != want {
		t.Fatalf("got \n%s\nwant \n%s\n", got, want)
	}
}

func TestEncodeDotGraph(t *testing.T) {
	tris := []Triple{
		SubjPredRes("me", "rel", "you"),
		SubjPredRes("me", "rdf:type", "person"),
		SubjPredRes("you", "rel", "other"),
		SubjPredRes("you", "rdf:type", "child"),
		SubjPredRes("other", "any", "john"),
	}

	var buff bytes.Buffer
	if err := NewDotGraphEncoder(&buff, "rel").Encode(tris...); err != nil {
		t.Fatal(err)
	}

	exp := strings.Split(`digraph "rel" {
"me" -> "you";
"me" [label="me<person>"];
"you" -> "other";
"you" [label="you<child>"];
}`, "\n")

	result := buff.String()
	if splits := strings.Split(result, "\n"); len(splits) != 6 {
		t.Fatalf("expected 6 lines count in \n%s\n", result)
	}

	for _, line := range exp {
		if got, want := buff.String(), line; !strings.Contains(got, want) {
			t.Fatalf("\n%q\n should contain \n%q\n", got, want)
		}
	}
}
