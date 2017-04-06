package triplestore

import "reflect"

const tagName = "predicate"

// Convert a Struct or ptr to Struct into triples
// using field tags.
// For each struct's field a triple is created:
// - Subject: function first argument
// - Predicate: tag value
// - Literal: actual field value according to field's type
// Unsupported types are ignored
func TriplesFromStruct(sub string, i interface{}) (out []Triple) {
	val := reflect.ValueOf(i)

	switch val.Kind() {
	case reflect.Struct:
		reflect.ValueOf(i)
	case reflect.Ptr:
		val = val.Elem()
		if val.Kind() != reflect.Struct {
			return
		}
	default:
		return
	}

	st := val.Type()

	for i := 0; i < st.NumField(); i++ {
		pred := st.Field(i).Tag.Get(tagName)
		objLit, err := ObjectLiteral(val.Field(i).Interface())
		if pred == "" || err != nil {
			continue
		}
		out = append(out, SubjPred(sub, pred).Object(objLit))
	}

	return
}
