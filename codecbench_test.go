package triplestore

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkEncodingMemallocation-4   	   20000	     71052 ns/op	   27488 B/op	    1209 allocs/op
func BenchmarkEncodingMemallocation(b *testing.B) {
	var triples []Triple

	for i := 0; i < 100; i++ {
		triples = append(triples, SubjPred(fmt.Sprint(i), "digit").IntegerLiteral(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var buff bytes.Buffer
		err := NewBinaryEncoder(&buff).Encode(triples...)
		if err != nil {
			b.Fatal(err)
		}
	}

}

func BenchmarkAllEncoding(b *testing.B) {
	var triples []Triple

	for i := 0; i < 1000; i++ {
		triples = append(triples, SubjPred(fmt.Sprint(i), "digit").IntegerLiteral(i))
	}

	b.Run("binary", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buff bytes.Buffer
			if err := NewBinaryEncoder(&buff).Encode(triples...); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("ntriples", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var buff bytes.Buffer
			if err := NewNTriplesEncoder(&buff).Encode(triples...); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func BenchmarkAllDecoding(b *testing.B) {
	binaryFile, err := os.Open(filepath.Join("testdata", "bench", "decode_1.bin"))
	if err != nil {
		b.Fatal(err)
	}
	defer binaryFile.Close()

	b.Run("binary", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := NewBinaryDecoder(binaryFile).Decode(); err != nil {
				b.Fatal(err)
			}
		}
	})

	ntFile, err := os.Open(filepath.Join("testdata", "bench", "decode_1.nt"))
	if err != nil {
		b.Fatal(err)
	}
	defer ntFile.Close()

	b.Run("ntriples", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			if _, err := NewNTriplesDecoder(ntFile).Decode(); err != nil {
				b.Fatal(err)
			}
		}
	})
}
