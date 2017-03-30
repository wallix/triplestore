package triplestore

import "testing"

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

	tri = Subject("subject").Predicate("predicate").StringLiteral("rdf")

	lit, ok := tri.Object().Literal()
	if got, want := ok, true; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}
	if got, want := lit.Value(), "rdf"; got != want {
		t.Fatalf("got %d, want %d", got, want)
	}
	if got, want := lit.Type(), XsdString; got != want {
		t.Fatalf("got %d, want %d", got, want)
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
		t.Fatalf("got %d, want %d", got, want)
	}
	_, ok = tri.Object().Literal()
	if got, want := ok, false; got != want {
		t.Fatalf("got %t, want %t", got, want)
	}
}
