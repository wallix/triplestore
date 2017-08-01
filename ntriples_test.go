package triplestore

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"testing"
)

func TestEncodeDecodeW3CSuite(t *testing.T) {
	path := filepath.Join("testdata", "w3c_ntriples", "*.nt")
	filenames, _ := filepath.Glob(path)

	for _, filename := range filenames {
		b, err := ioutil.ReadFile(filename)
		if err != nil {
			t.Fatalf("cannot read file %s", filename)
		}

		tris, err := NewNTriplesDecoder(bytes.NewReader(b)).Decode()
		if err != nil {
			t.Fatal(err)
		}

		var buf bytes.Buffer
		if err := NewNTriplesEncoder(&buf).Encode(tris...); err != nil {
			t.Fatalf("file %s: re-encoding error: %s", filename, err)
		}

		if got, want := buf.Bytes(), removeNTriplesCommentsAndEmptyLines(b); !bytes.Equal(got, want) {
			t.Fatalf("file %s: original and re-encoded mismatch\n\ngot\n%s\n\nwant\n%s\n", filename, got, want)
		}
	}
}
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

		compareMultiline(t, buff.Bytes(), b)
	}
}

func compareMultiline(t *testing.T, actual, expected []byte) {
	expected = removeNTriplesCommentsAndEmptyLines(expected)
	actual = removeNTriplesCommentsAndEmptyLines(actual)
	actuals := bytes.Split(actual, []byte("\n"))
	expecteds := bytes.Split(expected, []byte("\n"))

	for _, a := range actuals {
		if !contains(expecteds, a) {
			fmt.Printf("\texpected content\n%q\n", expected)
			fmt.Printf("\tactual content\n%q\n", actual)
			t.Fatalf("extra line not in expected content\n%s\n", a)
		}
	}

	for _, e := range expecteds {
		if !contains(actuals, e) {
			t.Fatalf("expected line is missing from actual content\n'%s'\n", e)
		}
	}
}

func contains(arr [][]byte, s []byte) bool {
	for _, a := range arr {
		if bytes.Equal(s, a) {
			return true
		}
	}
	return false
}

var endOfLineComents = regexp.MustCompile(`(.*\.)\s+(#.*)`)

func removeNTriplesCommentsAndEmptyLines(b []byte) []byte {
	scn := bufio.NewScanner(bytes.NewReader(b))
	var cleaned bytes.Buffer
	for scn.Scan() {
		line := scn.Text()
		if empty, _ := regexp.MatchString(`^\s*$`, line); empty {
		}
		if comment, _ := regexp.MatchString(`^\s*#`, line); comment {
			continue
		}
		l := endOfLineComents.ReplaceAll([]byte(line), []byte("$1"))
		cleaned.Write(l)
		cleaned.WriteByte('\n')
	}

	return cleaned.Bytes()
}
