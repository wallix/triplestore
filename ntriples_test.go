package triplestore

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"
)

func TestEncodeDecodeNTriples(t *testing.T) {
	path := filepath.Join("testdata", "ntriples", "*.nt")
	filenames, _ := filepath.Glob(path)

	for _, f := range filenames {
		b, err := ioutil.ReadFile(f)

		tris, err := NewNTriplesDecoder(bytes.NewReader(b)).Decode()
		if err != nil {
			t.Fatal(err)
		}

		var buff bytes.Buffer
		err = NewNTriplesEncoder(&buff).Encode(tris...)
		if err != nil {
			t.Fatal(err)
		}

		compareMultiline(t, buff.String(), string(b))
	}
}

func compareMultiline(t *testing.T, actual, expected string) {
	actuals := strings.Split(actual, "\n")
	expecteds := strings.Split(expected, "\n")

	for _, a := range actuals {
		if !contains(expecteds, a) {
			fmt.Printf("\texpected content\n%q\n", expected)
			fmt.Printf("\tactual content\n%q\n", actual)
			t.Fatalf("extra line not in expected content\n%s\n", a)
		}
	}

	for _, e := range expecteds {
		if !contains(actuals, e) {
			t.Fatalf("expected line is missing from actual content\n%s\n", e)
		}
	}
}

func contains(arr []string, s string) bool {
	for _, a := range arr {
		if s == a {
			return true
		}
	}
	return false
}
