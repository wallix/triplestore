package triplestore

type Triple interface {
	Subject() string
	Predicate() string
	Object() Object
}

type Object interface {
	Literal() (Literal, bool)
	ResourceID() (string, bool)
}

type Literal interface {
	Type() string
	Value() string
}

const (
	XsdString  = "xsd:string"
	XsdBoolean = "xsd:boolean"
	XsdInteger = "xsd:integer"
)
