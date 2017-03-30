package triplestore

type XsdType uint8

var (
	XsdString   = XsdType(0)
	XsdBoolean  = XsdType(1)
	XsdInteger  = XsdType(2)
	XsdDateTime = XsdType(3)
)

func (x XsdType) XsdString() string {
	return xsdTypes[x]
}

var xsdTypes = map[XsdType]string{
	XsdString:   "xsd:string",
	XsdBoolean:  "xsd:boolean",
	XsdInteger:  "xsd:integer",
	XsdDateTime: "xsd:dateTime",
}
