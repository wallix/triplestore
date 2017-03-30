package triplestore

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

func (enc *Encoder) Encode(tris ...Triple) error {
	for _, t := range tris {
		b, err := encodeTriple(t)
		if err != nil {
			return err
		}

		if _, err := enc.w.Write(b); err != nil {
			return err
		}
	}

	return nil
}

func encodeTriple(t Triple) ([]byte, error) {
	sub, pred := t.Subject(), t.Predicate()

	var buff bytes.Buffer

	binary.Write(&buff, binary.BigEndian, uint8(len(sub)))
	buff.WriteString(sub)
	binary.Write(&buff, binary.BigEndian, uint8(len(pred)))
	buff.WriteString(pred)

	obj := t.Object()
	if lit, isLit := obj.Literal(); isLit {
		binary.Write(&buff, binary.BigEndian, uint8(1))
		binary.Write(&buff, binary.BigEndian, lit.Type())
		litVal := lit.Value()
		binary.Write(&buff, binary.BigEndian, uint8(len(litVal)))
		buff.WriteString(litVal)
	} else {
		binary.Write(&buff, binary.BigEndian, uint8(0))
		resID, _ := obj.ResourceID()
		binary.Write(&buff, binary.BigEndian, uint8(len(resID)))
		buff.WriteString(resID)
	}

	return buff.Bytes(), nil
}

type Decoder struct {
	r       io.Reader
	triples []Triple
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r: r}
}

func (dec *Decoder) Decode() ([]Triple, error) {
	for {
		done, err := dec.decodeTriple()
		if done {
			break
		} else if err != nil {
			return nil, err
		}
	}

	return dec.triples, nil
}

func (dec *Decoder) decodeTriple() (bool, error) {
	sub, err := dec.readLengthFirstWord()
	if err == io.EOF {
		return true, nil
	} else if err != nil {
		return false, fmt.Errorf("subject: %s", err)
	}

	pred, err := dec.readLengthFirstWord()
	if err != nil {
		return false, fmt.Errorf("predicate: %s", err)
	}

	var objType uint8
	if err := binary.Read(dec.r, binary.BigEndian, &objType); err != nil {
		return false, fmt.Errorf("object type: %s", err)
	}

	var decodedObj object
	if objType == uint8(0) {
		resource, err := dec.readLengthFirstWord()
		if err != nil {
			return false, fmt.Errorf("resource: %s", err)
		}
		decodedObj.resourceID = string(resource)

	} else {
		decodedObj.isLit = true
		var decodedLiteral literal

		var litType uint8
		if err := binary.Read(dec.r, binary.BigEndian, &litType); err != nil {
			return false, fmt.Errorf("literate type: %s", err)
		}
		decodedLiteral.typ = XsdType(litType)

		val, err := dec.readLengthFirstWord()
		if err != nil {
			return false, fmt.Errorf("literate: %s", err)
		}

		decodedLiteral.val = string(val)
		decodedObj.lit = decodedLiteral
	}

	dec.triples = append(dec.triples, &triple{
		sub:  subject(string(sub)),
		pred: predicate(string(pred)),
		obj:  decodedObj,
	})

	return true, nil
}

func (dec *Decoder) readLengthFirstWord() ([]byte, error) {
	var len uint8
	if err := binary.Read(dec.r, binary.BigEndian, &len); err != nil {
		return nil, err
	}

	word := make([]byte, len)
	if _, err := io.ReadFull(dec.r, word); err != nil {
		return nil, errors.New("cannot decode length first word")
	}

	return word, nil
}
