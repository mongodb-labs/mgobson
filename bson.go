// Copyright (C) MongoDB, Inc. 2017-present.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
// Based on gopkg.in/mgo.v2/bson by Gustavo Niemeyer
// See THIRD-PARTY-NOTICES for original license terms.

package mgobson

import (
	"encoding/binary"

	"github.com/mongodb/mongo-go-driver/bson"
)

var (
	_ bson.Marshaler = (M)(nil)
	_ bson.Marshaler = (D)(nil)
	_ bson.Marshaler = (RawD)(nil)

	_ bson.Unmarshaler = (*M)(nil)
	_ bson.Unmarshaler = (*D)(nil)
	_ bson.Unmarshaler = (*RawD)(nil)
)

func DocsToArray(docs []interface{}) *bson.Array {
	array := bson.NewArray()

	for _, doc := range docs {
		d, err := bson.NewDocumentEncoder().EncodeDocument(doc)
		if err != nil {
			panic(err)
		}

		array.Append(bson.VC.Document(d))
	}

	return array
}

func appendToDoc(doc *bson.Document, key string, value interface{}) error {
	switch v := value.(type) {
	case D:
		d, err := v.MarshalBSONDocument()
		if err != nil {
			return err
		}

		doc.Append(bson.EC.SubDocument(key, d))
	case M:
		d, err := v.MarshalBSONDocument()
		if err != nil {
			return err
		}

		doc.Append(bson.EC.SubDocument(key, d))
	case RawD:
		d, err := v.MarshalBSONDocument()
		if err != nil {
			return err
		}

		doc.Append(bson.EC.SubDocument(key, d))
	default:
		doc.Append(bson.EC.Interface(key, v))
	}

	return nil
}

// M is a convenient alias for a map[string]interface{} map, useful for
// dealing with BSON in a native way.  For instance:
//
//     bson.M{"a": 1, "b": true}
//
// There's no special handling for this type in addition to what's done anyway
// for an equivalent map type.  Elements in the map will be dumped in an
// undefined ordered. See also the bson.D type for an ordered alternative.
type M map[string]interface{}

func (m M) MarshalBSONDocumentUnsafe() *bson.Document {
	doc, err := m.MarshalBSONDocument()
	if err != nil {
		panic(err)
	}

	return doc
}

func (m M) MarshalBSONDocument() (*bson.Document, error) {
	doc := bson.NewDocument()

	for k, v := range m {
		err := appendToDoc(doc, k, v)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

func (m M) MarshalBSON() ([]byte, error) {
	return bson.Marshal(map[string]interface{}(m))
}

func (m *M) UnmarshalBSON(b []byte) error {
	newM := make(map[string]interface{})

	err := bson.Unmarshal(b, newM)
	if err != nil {
		return err
	}

	for key, val := range newM {
		switch v := val.(type) {
		case map[string]interface{}:
			newM[key] = M(v)
		}
	}

	*m = newM

	return nil
}

// D represents a BSON document containing ordered elements. For example:
//
//     bson.D{{"a", 1}, {"b", true}}
//
// In some situations, such as when creating indexes for MongoDB, the order in
// which the elements are defined is important.  If the order is not important,
// using a map is generally more comfortable. See bson.M and bson.RawD.
type D []DocElem

func (d D) MarshalBSONDocumentUnsafe() *bson.Document {
	doc, err := d.MarshalBSONDocument()
	if err != nil {
		panic(err)
	}

	return doc
}

func (d D) MarshalBSONDocument() (*bson.Document, error) {
	doc := bson.NewDocument()

	for _, elem := range d {
		err := appendToDoc(doc, elem.Name, elem.Value)
		if err != nil {
			return nil, err
		}
	}

	return doc, nil
}

func (d D) MarshalBSON() ([]byte, error) {
	doc, err := d.MarshalBSONDocument()
	if err != nil {
		return nil, err
	}

	return doc.MarshalBSON()
}

func (d *D) UnmarshalBSON(b []byte) error {
	doc, err := bson.UnmarshalDocument(b)
	if err != nil {
		return err
	}

	newD := make(D, 0, doc.Len())

	itr := doc.Iterator()
	for itr.Next() {
		elem := itr.Element()

		var val interface{}
		switch elem.Value().Type() {
		case bson.TypeEmbeddedDocument:
			subD := make(D, 0)
			err := subD.UnmarshalBSON(elem.Value().ReaderDocument())
			if err != nil {
				return err
			}
			val = subD
		default:
			val = elem.Value().Interface()
		}

		newD = append(newD, DocElem{elem.Key(), val})
	}
	if err := itr.Err(); err != nil {
		return err
	}

	*d = newD
	return nil
}

// DocElem is an element of the bson.D document representation.
type DocElem struct {
	Name  string
	Value interface{}
}

// The Raw type represents raw unprocessed BSON documents and elements.
// Kind is the kind of element as defined per the BSON specification, and
// Data is the raw unprocessed data for the respective element.
// Using this type it is possible to unmarshal or marshal values partially:// Relevant documentation:
//
//     http://bsonspec.org/#/specification
//
type Raw struct {
	Kind byte
	Data []byte
}

// RawD represents a BSON document containing raw unprocessed elements.
// This low-level representation may be useful when lazily processing
// documents of uncertain content, or when manipulating the raw content
// documents in general.
type RawD []RawDocElem

// RawDocElem is documented by Raw.
type RawDocElem struct {
	Name  string
	Value Raw
}

func (r RawD) MarshalBSONDocumentUnsafe() *bson.Document {
	doc, err := r.MarshalBSONDocument()
	if err != nil {
		panic(err)
	}

	return doc
}

func (r RawD) MarshalBSONDocument() (*bson.Document, error) {
	b, err := r.MarshalBSON()
	if err != nil {
		return nil, err
	}

	return bson.UnmarshalDocument(b)
}

func (r RawD) MarshalBSON() ([]byte, error) {
	b := []byte{0, 0, 0, 0}

	for _, elem := range r {
		b = append(b, elem.Value.Kind)
		b = append(b, elem.Name...)
		b = append(b, 0)
		b = append(b, elem.Value.Data...)
	}

	b = append(b, 0)
	length := uint32(len(b))
	binary.LittleEndian.PutUint32(b, length)

	return b, nil
}

func (r *RawD) UnmarshalBSON(b []byte) error {
	itr, err := bson.NewReaderIterator(b)
	if err != nil {
		return err
	}

	newR := make(RawD, 0)

	for itr.Next() {
		readerElem := itr.Element()
		elemBytes, err := readerElem.MarshalBSON()
		if err != nil {
			return err
		}

		rawDocElem := RawDocElem{
			Name: readerElem.Key(),
			Value: Raw{
				Kind: byte(readerElem.Value().Type()),
				Data: elemBytes[1+len(readerElem.Key())+1:],
			},
		}

		newR = append(newR, rawDocElem)
	}
	if err := itr.Err(); err != nil {
		return err
	}

	*r = newR
	return nil
}
