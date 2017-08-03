package triplestore

import (
	"bytes"
	"context"
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

// BenchmarkAllEncoding/binary-4         	            2000	    662725 ns/op
// BenchmarkAllEncoding/binary_streaming-4         	    1000	   1195385 ns/op
// BenchmarkAllEncoding/ntriples-4                 	    5000	    346044 ns/op
// BenchmarkAllEncoding/ntriples_with_context-4    	    2000	   1134855 ns/op
func BenchmarkAllEncoding(b *testing.B) {
	var triples []Triple

	for i := 0; i < 1000; i++ {
		triples = append(triples, SubjPred(fmt.Sprint(i), "digit").IntegerLiteral(i))
	}

	b.ResetTimer()

	b.Run("binary", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buff bytes.Buffer
			if err := NewBinaryEncoder(&buff).Encode(triples...); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("binary streaming", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			b.StopTimer()
			triC := make(chan Triple)
			go tripleChan(triples, triC)
			b.StartTimer()
			var buff bytes.Buffer
			if err := NewBinaryStreamEncoder(&buff).Encode(context.Background(), triC); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ntriples", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buff bytes.Buffer
			if err := NewNTriplesEncoder(&buff).Encode(triples...); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("ntriples with context", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			var buff bytes.Buffer
			if err := NewNTriplesEncoderWithContext(&buff, RDFContext).Encode(triples...); err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkAllDecoding/binary-4                   	 3000000	       414 ns/op
// BenchmarkAllDecoding/binary_streaming-4         	  200000	      5636 ns/op
// BenchmarkAllDecoding/ntriples-4                 	 3000000	       577 ns/op
func BenchmarkAllDecoding(b *testing.B) {
	binaryFile, err := os.Open(filepath.Join("testdata", "bench", "decode_1.bin"))
	if err != nil {
		b.Fatal(err)
	}
	defer binaryFile.Close()

	b.ResetTimer()

	b.Run("binary", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			if _, err := NewBinaryDecoder(binaryFile).Decode(); err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("binary streaming", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			results := NewBinaryStreamDecoder(binaryFile).Decode(context.Background())
			for r := range results {
				if r.Err != nil {
					b.Fatal(r.Err)
				}
			}
		}
	})

	b.Run("ntriples", func(b *testing.B) {
		ntFile, err := os.Open(filepath.Join("testdata", "bench", "decode_1.nt"))
		if err != nil {
			b.Fatal(err)
		}
		defer ntFile.Close()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			if _, err := NewNTriplesDecoder(ntFile).Decode(); err != nil {
				b.Fatal(err)
			}
		}
	})
}

func tripleChan(triples []Triple, triC chan<- Triple) {
	for _, t := range triples {
		triC <- t
	}
	close(triC)
}
