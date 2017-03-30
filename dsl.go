package triplestore

import "fmt"

type subject string
type predicate string

type triple struct {
	sub  subject
	pred predicate
	obj  object
}

func (t *triple) Object() object {
	return t.obj
}

func (t *triple) Subject() string {
	return string(t.sub)
}

func (t *triple) Predicate() string {
	return string(t.pred)
}

type object struct {
	isLit      bool
	resourceID string
	lit        literal
}

func (o object) Literal() (Literal, bool) {
	return o.lit, o.isLit
}

func (o object) ResourceID() (string, bool) {
	return o.resourceID, !o.isLit
}

type literal struct {
	typ, val string
}

func (l literal) Type() string {
	return l.typ
}

func (l literal) Value() string {
	return l.val
}

type tripleBuilder struct {
	sub, pred string
}

func Subject(s string) *tripleBuilder {
	return &tripleBuilder{sub: s}
}

func (b *tripleBuilder) Predicate(s string) *tripleBuilder {
	b.pred = s
	return b
}

func (b *tripleBuilder) Resource(s string) *triple {
	t := &triple{
		sub:  subject(b.sub),
		pred: predicate(b.pred),
		obj:  object{resourceID: s},
	}

	return t
}

func (b *tripleBuilder) BooleanLiteral(bl bool) *triple {
	t := &triple{sub: subject(b.sub), pred: predicate(b.pred)}

	t.obj = object{
		isLit: true,
		lit:   literal{typ: XsdBoolean, val: fmt.Sprint(bl)},
	}

	return t
}

func (b *tripleBuilder) IntegerLiteral(i int) *triple {
	t := &triple{sub: subject(b.sub), pred: predicate(b.pred)}

	t.obj = object{
		isLit: true,
		lit:   literal{typ: XsdInteger, val: fmt.Sprint(i)},
	}

	return t
}

func (b *tripleBuilder) StringLiteral(s string) *triple {
	t := &triple{sub: subject(b.sub), pred: predicate(b.pred)}

	t.obj = object{
		isLit: true,
		lit:   literal{typ: XsdString, val: s},
	}

	return t
}
