package triplestore

import (
	"testing"
	"time"
)

func TestBuildTriple(t *testing.T) {
	tri := Subject("subject").Predicate("predicate").StringLiteral("any")

	if got, want := tri.Subject(), "subject"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	if got, want := tri.Predicate(), "predicate"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
}

func TestParseObject(t *testing.T) {
	tri := Subject("subject").Predicate("predicate").IntegerLiteral(123)
	num, err := ParseInteger(tri.Object())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := num, 123; got != want {
		t.Fatalf("got %d, want %d", got, want)
	}

	tri = Subject("subject").Predicate("predicate").BooleanLiteral(true)
	b, err := ParseBoolean(tri.Object())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := b, true; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}

	now := time.Now()
	tri = Subject("subject").Predicate("predicate").DateTimeLiteral(now)
	date, err := ParseDateTime(tri.Object())
	if err != nil {
		t.Fatal(err)
	}
	if got, want := date, now.UTC(); got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	tri = Subject("subject").Predicate("predicate").StringLiteral("rdf")
	lit, ok := tri.Object().Literal()
	if got, want := ok, true; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}
	if got, want := lit.Value(), "rdf"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	if got, want := lit.Type(), XsdString; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}

	_, ok = tri.Object().ResourceID()
	if got, want := ok, false; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}
}

func TestParseResourceID(t *testing.T) {
	tri := Subject("subject").Predicate("predicate").Resource("dbpedia:Bonobo")

	rid, ok := tri.Object().ResourceID()
	if got, want := ok, true; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}
	if got, want := rid, "dbpedia:Bonobo"; got != want {
		t.Fatalf("got %s, want %s", got, want)
	}
	_, ok = tri.Object().Literal()
	if got, want := ok, false; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}
}

func TestEquality(t *testing.T) {
	emptyTriple := new(triple)
	tcases := []struct {
		one, other Triple
		exp        bool
	}{
		{one: SubjPred("", "").Resource(""), other: SubjPred("", "").Resource(""), exp: true},
		{one: SubjPred("sub", "pred").Resource("Bonobo"), other: SubjPred("sub", "pred").Resource("Bonobo"), exp: true},
		{one: SubjPred("sub", "pred").Resource("Bonobo"), other: SubjPred("sub", "pred").Resource("Banaba"), exp: false},
		{one: SubjPred("sub", "pred").Resource("Bonobo"), other: SubjPred("sub", "newpred").Resource("Bonobo"), exp: false},
		{one: SubjPred("sub", "pred").Resource("Bonobo"), other: SubjPred("newsub", "pred").Resource("Bonobo"), exp: false},

		{one: SubjPred("sub", "pred").StringLiteral("Bonobo"), other: SubjPred("sub", "pred").StringLiteral("Bonobo"), exp: true},
		{one: SubjPred("sub", "pred").BooleanLiteral(true), other: SubjPred("sub", "pred").BooleanLiteral(true), exp: true},
		{one: SubjPred("sub", "pred").IntegerLiteral(42), other: SubjPred("sub", "pred").IntegerLiteral(42), exp: true},

		{one: SubjPred("", "").StringLiteral(""), other: SubjPred("", "").StringLiteral(""), exp: true},

		{one: SubjPred("sub", "pred").Resource("Bonobo"), other: SubjPred("sub", "pred").StringLiteral("Bonobo"), exp: false},
		{one: SubjPred("sub", "pred").StringLiteral("true"), other: SubjPred("sub", "pred").BooleanLiteral(true), exp: false},
		{one: SubjPred("sub", "pred").StringLiteral("2"), other: SubjPred("sub", "pred").IntegerLiteral(2), exp: false},

		{one: SubjPred("sub", "pred").Resource("Bonobo"), other: emptyTriple, exp: false},
		{one: emptyTriple, other: emptyTriple, exp: true},
	}
	for i, tcase := range tcases {
		if got, want := tcase.one.Equal(tcase.other), tcase.exp; got != want {
			t.Errorf("%d: got %t, want %t", i+1, got, want)
		}
		if got, want := tcase.other.Equal(tcase.one), tcase.exp; got != want {
			t.Errorf("%d: got %t, want %t", i, got, want)
		}
	}
}
