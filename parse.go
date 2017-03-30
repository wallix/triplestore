package triplestore

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

func ParseInteger(obj Object) (int, error) {
	if lit, ok := obj.Literal(); ok {
		if lit.Type() != XsdInteger {
			return 0, fmt.Errorf("literal is not an integer but %s", lit.Type())
		}

		return strconv.Atoi(lit.Value())
	}

	return 0, errors.New("cannot parse integer: object is not literal")
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
