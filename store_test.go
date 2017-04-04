package triplestore_test

import (
	"testing"

	"github.com/wallix/triplestore"
)

func TestStore(t *testing.T) {
	s := triplestore.New()
	s.Add(
		triplestore.SubjPred("one", "two").StringLiteral("three"),
		triplestore.SubjPred("one", "two").Resource("four"),
		triplestore.SubjPred("four", "two").IntegerLiteral(42),
		triplestore.SubjPred("one", "two").Resource("four"),
	)
	g := s.Snapshot()
	expected := []triplestore.Triple{
		triplestore.SubjPred("one", "two").StringLiteral("three"),
		triplestore.SubjPred("one", "two").Resource("four"),
		triplestore.SubjPred("four", "two").IntegerLiteral(42),
	}
	if got, want := g.Count(), len(expected); got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
	for _, tr := range expected {
		if got, want := g.Contains(tr), true; got != want {
			t.Fatalf("%v: got %t, want %t", tr, got, want)
		}
	}
	s.Remove(triplestore.SubjPred("one", "two").Resource("four"))
	newG := s.Snapshot()

	t.Run("old snapshot unmodified", func(t *testing.T) {
		if got, want := g.Count(), len(expected); got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		for _, tr := range expected {
			if got, want := g.Contains(tr), true; got != want {
				t.Fatalf("%v: got %t, want %t", tr, got, want)
			}
		}
	})

	t.Run("triple 1 removed in new snapshot", func(t *testing.T) {
		if got, want := newG.Count(), 2; got != want {
			t.Fatalf("got %d, want %d", got, want)
		}
		if got, want := newG.Contains(expected[0]), true; got != want {
			t.Fatalf("%v: got %t, want %t", expected[0], got, want)
		}
		if got, want := newG.Contains(expected[1]), false; got != want {
			t.Fatalf("%v: got %t, want %t", expected[1], got, want)
		}
		if got, want := newG.Contains(expected[2]), true; got != want {
			t.Fatalf("%v: got %t, want %t", expected[2], got, want)
		}
	})

}
