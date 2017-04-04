package triplestore

import "fmt"

type Store interface {
	Add(...Triple)
	Remove(...Triple)
	Snapshot() RDFGraph
}

type RDFGraph interface {
	Contains(Triple) bool
	Triples() []Triple
	Count() int
	WithSubject() []Triple
	WithPredicate() []Triple
	WithObject() []Triple
	WithSubjObj() []Triple
	WithSubjPred() []Triple
	WithPredObj() []Triple
}

type store struct {
	triples map[string]Triple
}

func New() *store {
	return &store{triples: make(map[string]Triple)}
}

func (s *store) Add(ts ...Triple) {
	for _, t := range ts {
		tr := t.(*triple)
		s.triples[tr.key()] = t
	}
}
func (s *store) Remove(ts ...Triple) {
	for _, t := range ts {
		tr := t.(*triple)
		delete(s.triples, tr.key())
	}
}

func (s *store) Snapshot() RDFGraph {
	gph := &graph{
		s:   make(map[string]Triple),
		p:   make(map[string]Triple),
		o:   make(map[string]Triple),
		sp:  make(map[string]Triple),
		so:  make(map[string]Triple),
		po:  make(map[string]Triple),
		spo: make(map[string]Triple),
	}
	for k, t := range s.triples {
		obj := t.Object().(object)
		gph.s[t.Subject()] = t
		gph.p[t.Predicate()] = t
		gph.p[obj.key()] = t
		gph.sp[fmt.Sprintf("%s%s", t.Subject(), t.Predicate())] = t
		gph.so[fmt.Sprintf("%s%s", t.Subject(), obj.key())] = t
		gph.po[fmt.Sprintf("%s%s", t.Predicate(), obj.key())] = t
		gph.spo[k] = t
	}
	return gph
}

func (g *graph) Contains(t Triple) bool {
	_, ok := g.spo[t.(*triple).key()]
	return ok
}
func (g *graph) Triples() []Triple {
	return []Triple{}
}
func (g *graph) Count() int {
	return len(g.spo)
}
func (g *graph) WithSubject() []Triple {
	return []Triple{}
}
func (g *graph) WithPredicate() []Triple {
	return []Triple{}
}
func (g *graph) WithObject() []Triple {
	return []Triple{}
}
func (g *graph) WithSubjObj() []Triple {
	return []Triple{}
}
func (g *graph) WithSubjPred() []Triple {
	return []Triple{}
}
func (g *graph) WithPredObj() []Triple {
	return []Triple{}
}

type graph struct {
	s   map[string]Triple
	p   map[string]Triple
	o   map[string]Triple
	sp  map[string]Triple
	so  map[string]Triple
	po  map[string]Triple
	spo map[string]Triple
}
