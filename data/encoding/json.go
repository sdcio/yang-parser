// Copyright (c) 2017, 2019, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package encoding

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/danos/encoding/rfc7951"
	"github.com/steiler/yang-parser/data/datanode"
	"github.com/steiler/yang-parser/schema"
)

type JSONReader struct {
	decodedName string
	decodedMsg  interface{}
}

func (jr *JSONReader) name() string {
	if idx := strings.Index(jr.decodedName, ":"); idx != -1 {
		return jr.decodedName[idx+1:]
	}
	return jr.decodedName
}

func decodeValue(val interface{}) (string, error) {
	switch typeValue := val.(type) {
	case string: // Non-empty Leaf containing string
		return typeValue, nil

	case bool: // Non-empty Leaf containing boolean
		if typeValue == true {
			return "true", nil
		} else {
			return "false", nil
		}
	case float64: // Non-empty Leaf containing number of any sort
		return fmt.Sprintf("%d", int(typeValue)), nil
	case nil: // Empty leaf
		return "", nil
	default:
		return "", schema.NewMissingValueError(nil)
	}
}

func (jr *JSONReader) values() ([]string, error) {

	switch typeValue := jr.decodedMsg.(type) {
	case []interface{}:
		list := make([]string, 0, len(typeValue))
		for _, v := range typeValue {
			if val, e := decodeValue(v); e != nil {
				return nil, e
			} else {
				list = append(list, val)
			}
		}
		return list, nil

	default:
		v, e := decodeValue(jr.decodedMsg)
		if e != nil {
			return nil, e
		}
		return []string{v}, nil
	}
}

func (jr *JSONReader) unserializedChildren(_ []string, sn schema.Node) ([]unserialized, error) {
	children := make([]unserialized, 0)

	switch typeValue := jr.decodedMsg.(type) {
	case map[string]interface{}: // Container
		for k, v := range typeValue {
			child := &JSONReader{decodedName: k, decodedMsg: v}
			children = append(children, child)
		}
	case []interface{}: // List or leaf-list
		for _, listV := range typeValue {
			child := &JSONReader{decodedName: jr.decodedName, decodedMsg: listV}
			children = append(children, child)
		}
	}
	return children, nil
}

type ConfigOrState bool

const (
	Config ConfigOrState = true
	State                = false
)

type jsonRFC7951Unmarshaller struct {
	valType schema.ValidationType
	enc     EncType
}

func newRFC7951Unmarshaller() *jsonRFC7951Unmarshaller {
	return &jsonRFC7951Unmarshaller{
		valType: schema.ValidateAll,
		enc:     RFC7951}
}

func newJSONUnmarshaller() *jsonRFC7951Unmarshaller {
	return &jsonRFC7951Unmarshaller{
		valType: schema.ValidateAll,
		enc:     JSON}
}

func (jru *jsonRFC7951Unmarshaller) SetValidation(
	valType schema.ValidationType,
) Unmarshaller {
	jru.valType = valType
	return jru
}

func (jru *jsonRFC7951Unmarshaller) Unmarshal(
	sn schema.Node,
	input []byte,
) (datanode.DataNode, error) {
	return unmarshalJSONInternal(
		sn,
		input,
		jru.valType,
		jru.enc)
}

func UnmarshalRFC7951(sn schema.Node, json_input []byte) (datanode.DataNode, error) {
	return unmarshalJSONInternal(sn, json_input, schema.ValidateAll, RFC7951)
}

func UnmarshalRFC7951WithoutValidation(sn schema.Node, json_input []byte) (datanode.DataNode, error) {
	return unmarshalJSONInternal(sn, json_input, schema.DontValidate, RFC7951)
}

func UnmarshalJSON(sn schema.Node, json_input []byte) (datanode.DataNode, error) {
	return unmarshalJSONInternal(sn, json_input, schema.ValidateAll, JSON)
}

func UnmarshalJSONWithoutValidation(
	sn schema.Node,
	cfgOrState ConfigOrState,
	json_input []byte,
) (datanode.DataNode, error) {
	return unmarshalJSONInternal(sn, json_input, schema.DontValidate, JSON)
}

func unmarshalJSONInternal(
	sn schema.Node,
	json_input []byte,
	valType schema.ValidationType,
	enc EncType,
) (datanode.DataNode, error) {

	jr := JSONReader{decodedName: sn.Name()}
	if enc == RFC7951 {
		if err := rfc7951.Unmarshal(json_input, &jr.decodedMsg); err != nil {
			return nil, err
		}
	} else {
		if err := json.Unmarshal(json_input, &jr.decodedMsg); err != nil {
			return nil, err
		}
	}

	datatree, err := convertToDataNode([]string{}, sn.Name(), &jr, sn)
	if err != nil {
		return nil, err
	}

	if valType != schema.DontValidate {
		err = validateDataNode(datatree, sn, valType)
		if err != nil {
			return nil, err
		}
	}

	return datatree, nil
}

func (jw *JSONWriter) writeValue(sn schema.Node, value string) {
	switch tt := sn.Type().(type) {
	case schema.Empty:
		if jw.rfc7951 {
			jw.WriteString("[null]")
		} else {
			jw.WriteString("null")
		}
	case schema.Boolean:
		// Write the raw value out as a native JSON type
		jw.WriteString(value)
	case schema.Uinteger:
		// Write the raw value out as a native JSON type
		if jw.rfc7951 && tt.BitWidth() > 32 {
			buf, _ := json.Marshal(value)
			jw.Write(buf)
		} else {
			jw.WriteString(value)
		}
	case schema.Integer:
		// Write the raw value out as a native JSON type
		if jw.rfc7951 && tt.BitWidth() > 32 {
			buf, _ := json.Marshal(value)
			jw.Write(buf)
		} else {
			jw.WriteString(value)
		}
	default:
		// Treat as a string, with appropriate escaping and quotes.
		// Note that Decimal64 is our variable precision floating point and
		// must be encoded as a string, as per draft-ietf-netmod-yang-json-05

		// json.Marshal won't err on a string
		buf, _ := json.Marshal(value)
		jw.Write(buf)
	}
}

func (jw *JSONWriter) writeNullLeafValue(sn schema.Node) {
	if jw.rfc7951 {
		if _, ok := sn.Type().(schema.Empty); ok {
			jw.WriteString("[null]")
			return
		}
	}
	jw.WriteString("null")
}

func (jw *JSONWriter) PushName(sn schema.Node) string {
	if jw.moduleName == nil {
		jw.moduleName = make([]string, 0)
	}
	nm := sn.Module()
	newname := nm
	if len(jw.moduleName) > 0 {
		if nm == jw.moduleName[len(jw.moduleName)-1] {
			newname = ""
		}
	}
	jw.moduleName = append(jw.moduleName, nm)

	return newname
}
func (jw *JSONWriter) PopName() {
	if len(jw.moduleName) > 0 {
		jw.moduleName = jw.moduleName[:len(jw.moduleName)-1]
	}
}

func (jw *JSONWriter) CurrentModuleName() string {
	switch len(jw.moduleName) {
	case 0:
		return ""
	case 1:
		return jw.moduleName[0]
	default:
		if jw.moduleName[len(jw.moduleName)-1] !=
			jw.moduleName[len(jw.moduleName)-2] {
			return jw.moduleName[len(jw.moduleName)-1]
		}
	}
	return ""
}

func (jw *JSONWriter) writeJsonName(sn schema.Node, n datanode.DataNode) {
	if _, ok := sn.(schema.ListEntry); ok {
		return
	}

	jw.WriteByte('"')
	if jw.rfc7951 {
		if nm := jw.CurrentModuleName(); nm != "" {
			jw.WriteString(nm)
			jw.WriteString(":")
		}
	}
	jw.WriteString(n.YangDataName())
	jw.WriteString("\":")
}

func (jw *JSONWriter) encodeJsonChildren(sn schema.Node, n datanode.DataNode) {

	for i, cn := range n.YangDataChildren() {
		csn := sn.Child(cn.YangDataName())

		if i != 0 {
			jw.WriteByte(',')
		}

		if jw.rfc7951 {
			jw.PushName(csn)
		}
		jw.writeJsonName(csn, cn)
		switch csn.(type) {
		case schema.Container, schema.Tree:
			jw.WriteByte('{')
			jw.encodeJsonChildren(csn, cn)
			jw.WriteByte('}')

		case schema.List:
			jw.WriteByte('[')
			jw.encodeJsonChildren(csn, cn)
			jw.WriteByte(']')
		case schema.ListEntry:
			jw.WriteByte('{')
			jw.encodeJsonChildren(csn, cn)
			jw.WriteByte('}')

		case schema.Leaf:
			vals := cn.YangDataValuesNoSorting()
			if len(vals) == 0 {
				jw.writeNullLeafValue(csn)
			} else {
				jw.writeValue(csn, vals[0])
			}

		case schema.LeafList:
			vals := cn.YangDataValues()
			if len(vals) == 0 {
				jw.WriteString("null")
			} else {
				jw.WriteByte('[')
				for i, v := range vals {
					if i != 0 {
						jw.WriteByte(',')
					}
					jw.writeValue(csn, v)
				}
			}
			jw.WriteByte(']')
		}
		if jw.rfc7951 {
			jw.PopName()
		}
	}
}

type JSONWriter struct {
	bytes.Buffer
	rfc7951    bool
	moduleName []string
}

func toJSONInternal(sn schema.Node, node datanode.DataNode, jw *JSONWriter) []byte {
	jw.WriteByte('{')
	jw.encodeJsonChildren(sn, node)
	jw.WriteByte('}')
	return jw.Bytes()

}

func ToJSON(sn schema.Node, node datanode.DataNode) []byte {
	return toJSONInternal(sn, node, &JSONWriter{})
}

func ToRFC7951(sn schema.Node, node datanode.DataNode) []byte {
	return toJSONInternal(sn, node, &JSONWriter{rfc7951: true})
}
