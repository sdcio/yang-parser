// Copyright (c) 2017, 2019, AT&T Intellectual Property. All rights reserved
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package encoding

import (
	"strings"

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

func matchIdentityref(path []string, typ schema.Type, val string) bool {
	switch t := typ.(type) {
	case schema.Union:
		for _, utyp := range t.Typs() {
			if matchIdentityref(path, utyp, val) {
				return true
			}
		}
	case schema.Identityref:
		if err := t.Validate(nil, path, val); err == nil {
			return true
		}
	default:
		return false

	}
	return false
}

// RFC 7951; sec 6.8 requires that an identityref should accept both
// the namespace-qualified and simple form of an identity value where the
// namespace of the identity value matches the leaf-(list) node.
// Internally, we reject the namespace-qualified value when the simple form
// is possible. When this occurs, we convert the identityref value to
// a simple form identity value where possible.
// Note: RFC7951 uses the module name as the namespace.
func isIdentityrefSimpleFormValid(path []string, sn schema.Node, val string) (string, bool) {
	// Try trimming the module prefix from value
	modPrfx := sn.Module() + ":"
	if !strings.HasPrefix(val, modPrfx) {
		return "", false
	}

	simpleform := strings.TrimPrefix(val, modPrfx)

	// check that possible simpleform value
	// is a valid identityref value
	if matchIdentityref(path, sn.Type(), simpleform) {
		return simpleform, true
	}
	return "", false

}

func convertToDataNode(path []string, name string, node unserialized, sn schema.Node) (datanode.DataNode, error) {

	children := []datanode.DataNode{}
	vals := []string{}

	switch sn.(type) {
	case schema.Leaf, schema.LeafList, schema.LeafValue:
		values, err := node.values()
		if err != nil {
			return nil, err
		}
		if _, ok := sn.(schema.Leaf); ok {
			if _, isEmpty := sn.Type().(schema.Empty); isEmpty {
				if len(values) > 0 && (len(values) != 1 || values[0] != "") {
					return nil, schema.NewEmptyLeafValueError(node.name(), path)
				}
			}
		}
		// Validate the values
		for _, v := range values {
			if err := sn.Validate(nil, path, []string{v}); err != nil {
				// Check for an identityref that is using RFC7951 namespace-qualified form
				// where the simple form is preferred
				// e.g. we have "module-name:value" when "value" is preferred

				simple, valid := isIdentityrefSimpleFormValid(path, sn, v)
				if !valid {
					// No valid value found
					return nil, err
				}
				vals = append(vals, simple)
			} else {
				vals = append(vals, v)
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
