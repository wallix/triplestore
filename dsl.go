package triplestore

import (
	"fmt"
	"time"
)

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

func SubjPred(s, p string) *tripleBuilder {
	return &tripleBuilder{sub: s, pred: p}
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

func (b *tripleBuilder) DateTimeLiteral(tm time.Time) *triple {
	t := &triple{sub: subject(b.sub), pred: predicate(b.pred)}

	text, err := tm.UTC().MarshalText()
	if err != nil {
		panic(fmt.Errorf("date time literal: %s", err))
	}
	t.obj = object{
		isLit: true,
		lit:   literal{typ: XsdDateTime, val: string(text)},
	}

	return t
}
