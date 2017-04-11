package triplestore

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func SubjPredRes(s, p, r string) *triple {
	return &triple{
		sub:  subject(s),
		pred: predicate(p),
		obj:  Resource(r).(object),
	}
}

func SubjPredLit(s, p string, l interface{}) (*triple, error) {
	o, err := ObjectLiteral(l)
	if err != nil {
		return nil, err
	}
	return &triple{
		sub:  subject(s),
		pred: predicate(p),
		obj:  o.(object),
	}, nil
}

type tripleBuilder struct {
	sub, pred string
}

func SubjPred(s, p string) *tripleBuilder {
	return &tripleBuilder{sub: s, pred: p}
}

func Resource(s string) Object {
	return object{resource: s}
}

func (b *tripleBuilder) Resource(s string) *triple {
	return &triple{
		sub:  subject(b.sub),
		pred: predicate(b.pred),
		obj:  Resource(s).(object),
	}
}

func (b *tripleBuilder) Object(o Object) *triple {
	return &triple{
		sub:  subject(b.sub),
		pred: predicate(b.pred),
		obj:  o.(object),
	}
}

func ObjectLiteral(i interface{}) (Object, error) {
	switch ii := i.(type) {
	case string:
		return StringLiteral(ii), nil
	case bool:
		return BooleanLiteral(ii), nil
	case int:
		return IntegerLiteral(ii), nil
	case int64:
		return IntegerLiteral(int(ii)), nil
	case time.Time:
		return DateTimeLiteral(ii), nil
	case *time.Time:
		return DateTimeLiteral(*ii), nil
	default:
		return nil, fmt.Errorf("unsupported literal type %T", i)
	}
}

func ParseLiteral(obj Object) (interface{}, error) {
	if lit, ok := obj.Literal(); ok {
		switch lit.Type() {
		case XsdBoolean:
			return ParseBoolean(obj)
		case XsdDateTime:
			return ParseDateTime(obj)
		case XsdInteger:
			return ParseInteger(obj)
		case XsdString:
			return ParseString(obj)
		default:
			return nil, fmt.Errorf("unknown literal type: %s", lit.Type())
		}
	}
	return nil, errors.New("cannot parse literal: object is not literal")
}

func BooleanLiteral(bl bool) Object {
	return object{
		isLit: true,
		lit:   literal{typ: XsdBoolean, val: fmt.Sprint(bl)},
	}
}

func (b *tripleBuilder) BooleanLiteral(bl bool) *triple {
	return &triple{
		sub:  subject(b.sub),
		pred: predicate(b.pred),
		obj:  BooleanLiteral(bl).(object),
	}
}

func ParseBoolean(obj Object) (bool, error) {
	if lit, ok := obj.Literal(); ok {
		if lit.Type() != XsdBoolean {
			return false, fmt.Errorf("literal is not an boolean but %s", lit.Type())
		}

		return strconv.ParseBool(lit.Value())
	}

	return false, errors.New("cannot parse boolean: object is not literal")
}

func IntegerLiteral(i int) Object {
	return object{
		isLit: true,
		lit:   literal{typ: XsdInteger, val: fmt.Sprint(i)},
	}
}

func (b *tripleBuilder) IntegerLiteral(i int) *triple {
	return &triple{
		sub:  subject(b.sub),
		pred: predicate(b.pred),
		obj:  IntegerLiteral(i).(object),
	}
}

func ParseInteger(obj Object) (int, error) {
	if lit, ok := obj.Literal(); ok {
		if lit.Type() != XsdInteger {
			return 0, fmt.Errorf("literal is not an integer but %s", lit.Type())
		}

		return strconv.Atoi(lit.Value())
	}

	return 0, errors.New("cannot parse integer: object is not literal")
}

func StringLiteral(s string) Object {
	return object{
		isLit: true,
		lit:   literal{typ: XsdString, val: s},
	}
}

func (b *tripleBuilder) StringLiteral(s string) *triple {
	return &triple{
		sub:  subject(b.sub),
		pred: predicate(b.pred),
		obj:  StringLiteral(s).(object),
	}
}

func ParseString(obj Object) (string, error) {
	if lit, ok := obj.Literal(); ok {
		if lit.Type() != XsdString {
			return "", fmt.Errorf("literal is not a string but %s", lit.Type())
		}

		return lit.Value(), nil
	}

	return "", errors.New("cannot parse string: object is not literal")
}

func DateTimeLiteral(tm time.Time) Object {
	text, err := tm.UTC().MarshalText()
	if err != nil {
		panic(fmt.Errorf("date time literal: %s", err))
	}

	return object{
		isLit: true,
		lit:   literal{typ: XsdDateTime, val: string(text)},
	}
}

func (b *tripleBuilder) DateTimeLiteral(tm time.Time) *triple {
	return &triple{
		sub:  subject(b.sub),
		pred: predicate(b.pred),
		obj:  DateTimeLiteral(tm).(object),
	}
}

func ParseDateTime(obj Object) (time.Time, error) {
	var t time.Time
	if lit, ok := obj.Literal(); ok {
		if lit.Type() != XsdDateTime {
			return t, fmt.Errorf("literal is not an dateTime but %s", lit.Type())
		}

		err := t.UnmarshalText([]byte(lit.Value()))
		if err != nil {
			return t, err
		}
		return t, nil
	}

	return t, errors.New("cannot parse dateTime: object is not literal")
}
