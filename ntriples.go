package triplestore

import (
	"io"
	"io/ioutil"
)

func NewNTriplesDecoder(r io.Reader) Decoder {
	return &ntDecoder{r: r}
}

type ntDecoder struct {
	r io.Reader
}

func (d *ntDecoder) Decode() ([]Triple, error) {
	b, err := ioutil.ReadAll(d.r)
	if err != nil {
		return nil, err
	}
	return newNTParser(string(b)).parse(), err
}
