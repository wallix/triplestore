package triplestore

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type Store interface {
	Add(...Triple)
	Remove(...Triple)
	Snapshot() RDFGraph
}

type RDFGraph interface {
	Contains(Triple) bool
	Triples() []Triple
	Count() int
	WithSubject(s string) []Triple
	WithPredicate(p string) []Triple
	WithObject(o Object) []Triple
	WithSubjObj(s string, o Object) []Triple
	WithSubjPred(s, p string) []Triple
	WithPredObj(p string, o Object) []Triple
}

type Triples []Triple

func (ts Triples) Equal(others Triples) bool {
	if len(ts) != len(others) {
		return false
	}

	this := make(map[string]struct{})
	for _, tri := range ts {
		this[tri.(*triple).key()] = struct{}{}
	}

	other := make(map[string]struct{})
	for _, tri := range others {
		other[tri.(*triple).key()] = struct{}{}
	}

	return reflect.DeepEqual(this, other)
}

func (ts Triples) Map(fn func(Triple) string) (out []string) {
	for _, t := range ts {
		out = append(out, fn(t))
	}
	return
}

func (ts Triples) String() string {
	joined := strings.Join(ts.Map(
		func(t Triple) string { return fmt.Sprint(t) },
	), "\n")
	return fmt.Sprintf("[%s]", joined)
}

type store struct {
	mu      sync.RWMutex
	triples map[string]Triple
}

func New() *store {
	return &store{triples: make(map[string]Triple)}
}

func (s *store) Add(ts ...Triple) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range ts {
		tr := t.(*triple)
		s.triples[tr.key()] = t
	}
}
func (s *store) Remove(ts ...Triple) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range ts {
		tr := t.(*triple)
		delete(s.triples, tr.key())
	}
}

func (s *store) Snapshot() RDFGraph {
	gph := &graph{
		s:   make(map[string][]Triple),
		p:   make(map[string][]Triple),
		o:   make(map[string][]Triple),
		sp:  make(map[string][]Triple),
		so:  make(map[string][]Triple),
		po:  make(map[string][]Triple),
		spo: make(map[string]Triple),
	}

	s.mu.RLock()
	for k, t := range s.triples {
		obj := t.Object().(object)
		gph.s[t.Subject()] = append(gph.s[t.Subject()], t)
		gph.p[t.Predicate()] = append(gph.p[t.Predicate()], t)
		gph.o[obj.key()] = append(gph.o[obj.key()], t)

		sp := fmt.Sprintf("%s%s", t.Subject(), t.Predicate())
		gph.sp[sp] = append(gph.sp[sp], t)

		so := fmt.Sprintf("%s%s", t.Subject(), obj.key())
		gph.so[so] = append(gph.so[so], t)

		po := fmt.Sprintf("%s%s", t.Predicate(), obj.key())
		gph.po[po] = append(gph.po[po], t)

		gph.spo[k] = t
	}
	s.mu.RUnlock()

	for _, t := range gph.spo {
		gph.unique = append(gph.unique, t)
	}

	return gph
}

type graph struct {
	unique []Triple
	s      map[string][]Triple
	p      map[string][]Triple
	o      map[string][]Triple
	sp     map[string][]Triple
	so     map[string][]Triple
	po     map[string][]Triple
	spo    map[string]Triple
}

func (g *graph) Contains(t Triple) bool {
	_, ok := g.spo[t.(*triple).key()]
	return ok
}
func (g *graph) Triples() []Triple {
	return g.unique
}
func (g *graph) Count() int {
	return len(g.spo)
}

func (g *graph) WithSubject(s string) []Triple {
	return g.s[s]
}
func (g *graph) WithPredicate(p string) []Triple {
	return g.p[p]
}
func (g *graph) WithObject(o Object) []Triple {
	return g.o[o.(object).key()]
}
func (g *graph) WithSubjObj(s string, o Object) []Triple {
	return g.so[fmt.Sprintf("%s%s", s, o.(object).key())]
}
func (g *graph) WithSubjPred(s, p string) []Triple {
	return g.sp[fmt.Sprintf("%s%s", s, p)]
}
func (g *graph) WithPredObj(p string, o Object) []Triple {
	return g.po[fmt.Sprintf("%s%s", p, o.(object).key())]
}
