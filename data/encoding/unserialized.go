// Copyright (c) 2017, 2019, AT&T Intellectual Property. All rights reserved
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package encoding

import (
	"github.com/danos/mgmterror"
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/schema"
)

type EncType int

const (
	JSON EncType = iota
	RFC7951
	XML
)

// TEMP - DELETE
type unserialized interface {
	name() string
	values() ([]string, error)
	unserializedChildren([]string, schema.Node) ([]unserialized, error)
}

func getChildName(path []string, node unserialized, sn schema.Node) (string, error) {

	name := sn.Name()

	// If it's a listEntry use the key!
	if l, ok := sn.(schema.ListEntry); ok {
		found := false

		key := l.Keys()[0]
		kids, err := node.unserializedChildren(path, sn)
		if err != nil {
			return "", err
		}
		for _, ch := range kids {
			if ch.name() != key {
				continue
			}
			found = true

			vals, _ := ch.values()
			name = vals[0]

			// Validate the value of the key
			if err := sn.Child(key).Validate(nil, path, []string{name}); err != nil {
				return "", err
			}
		}
		if found == false {
			return "", schema.NewMissingKeyError([]string{key})
		}
	}
	return name, nil
}

func convertToDataNode(path []string, name string, node unserialized, sn schema.Node) (datanode.DataNode, error) {

	var err error
	children := []datanode.DataNode{}
	vals := []string{}

	switch sn.(type) {
	case schema.Leaf, schema.LeafList, schema.LeafValue:
		vals, err = node.values()
		if err != nil {
			return nil, err
		}
		if _, ok := sn.(schema.Leaf); ok {
			if _, isEmpty := sn.Type().(schema.Empty); isEmpty {
				if len(vals) > 0 && (len(vals) != 1 || vals[0] != "") {
					return nil, schema.NewEmptyLeafValueError(node.name(), path)
				}
			}
		}
		// Validate the values
		for _, v := range vals {
			if err := sn.Validate(nil, path, []string{v}); err != nil {
				return nil, err
			}
		}

	default:
		// Setup keyName if this is a list
		var keyName string
		if n, ok := sn.(schema.ListEntry); ok {
			keyName = n.Keys()[0]
		}
		ukids, err := node.unserializedChildren(path, sn)
		if err != nil {
			return nil, err
		}
		children = make([]datanode.DataNode, len(ukids), len(ukids))
		for i, ch := range ukids {
			csn := sn.Child(ch.name())
			if csn == nil {
				return nil, schema.NewSchemaMismatchError(ch.name(), path)
			}

			childName, err := getChildName(path, ch, csn)
			if err != nil {
				return nil, err
			}

			// Construct child path correctly for list case
			childPath := path
			if keyName != childName {
				childPath = append(childPath, childName)
			}
			children[i], err = convertToDataNode(childPath, childName, ch, csn)
			if err != nil {
				return nil, err
			}
		}
	}

	return datanode.CreateDataNode(name, children, vals), nil
}

func validateDataNode(
	n datanode.DataNode,
	sn schema.Node,
	valType schema.ValidationType) error {

	n = schema.AddDefaults(sn, n)
	if _, errs, ok := schema.NewSchemaValidator(
		sn, n).SetValidation(valType).Validate(); !ok {

		var errList mgmterror.MgmtErrorList
		errList.MgmtErrorListAppend(errs...)
		return errList
	}

	return nil
}

type Unmarshaller interface {
	SetValidation(schema.ValidationType) Unmarshaller
	Unmarshal(sn schema.Node, input []byte) (datanode.DataNode, error)
}

func NewUnmarshaller(enc EncType) Unmarshaller {
	switch enc {
	case RFC7951:
		return newRFC7951Unmarshaller()
	case JSON:
		return newJSONUnmarshaller()
	case XML:
		return newXMLUnmarshaller()
	}

	return nil
}
