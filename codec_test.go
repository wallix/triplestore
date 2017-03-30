package triplestore

import (
	"bytes"
	"testing"
)

func TestEncodeAndDecodeTriples(t *testing.T) {
	var buff bytes.Buffer
	enc := NewEncoder(&buff)

	in := SubjPred("one", "two").StringLiteral("three")
	err := enc.Encode(in)

	dec := NewDecoder(&buff)
	all, err := dec.Decode()
	if err != nil {
		t.Fatal(err)
	}

	if got, want := len(all), 1; got != want {
		t.Fatalf("got %d, want %d", got, want)
	}

	if got, want := in, all[0]; !got.Equal(want) {
		t.Fatalf("\ngot\n%v\nwant\n%v\n", got, want)
	}
}
