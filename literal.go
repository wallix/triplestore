package triplestore

import (
	"errors"
	"fmt"
	"strconv"
)

func ParseInteger(obj Object) (int, error) {
	if lit, ok := obj.Literal(); ok {
		if lit.Type() != XsdInteger {
			return 0, fmt.Errorf("literal is not an integer but %s", lit.Type())
		}

		return strconv.Atoi(lit.Value())
	}

	return 0, errors.New("cannot parse integer: object is not literal ")
}
