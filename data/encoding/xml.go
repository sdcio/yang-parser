// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package encoding

import (
	"bytes"
	"encoding/xml"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/schema"
)

type unmarshaledXML struct {
	XMLName  xml.Name
	Chardata string            `xml:",chardata"`
	Children []*unmarshaledXML `xml:",any"`
}

func (xmlNode *unmarshaledXML) name() string {
	return xmlNode.XMLName.Local
}

func (xmlNode *unmarshaledXML) values() ([]string, error) {
	if len(xmlNode.Children) > 0 {
		if xmlNode.Chardata != "" {
			return nil, mgmterror.NewUnknownElementApplicationError(xmlNode.Chardata)
		}
		list := make([]string, len(xmlNode.Children))
		for i, v := range xmlNode.Children {
			if len(v.Children) > 0 {
				return nil, mgmterror.NewUnknownElementApplicationError(v.name())
			}
			list[i] = v.Chardata
		}
		return list, nil
	}
	return []string{xmlNode.Chardata}, nil
}

func (xmlNode *unmarshaledXML) unserializedChildren(path []string, sn schema.Node) ([]unserialized, error) {
	fields := make(map[string]*unmarshaledXML)
	list := make([]unserialized, 0)

	for _, c := range xmlNode.Children {
		name := c.name()
		cn := sn.Child(name)
		if cn == nil {
			err := mgmterror.NewUnknownElementApplicationError(name)
			err.Path = pathutil.Pathstr(path)
			return nil, err
		}

		v, ok := fields[name]
		switch cn.(type) {
		// Without the schema In XML we can't actually tell the
		// difference between leaf or container vs a leaf-list or
		// list with only one entry. We therefore need to convert
		// our flat list entries into a proper list here.
		case schema.List, schema.LeafList:
			if !ok {
				v = &unmarshaledXML{c.XMLName, "", make([]*unmarshaledXML, 0)}
				fields[name] = v
				list = append(list, v)
			}
			v.Children = append(v.Children, c)
		default:
			if ok {
				err := mgmterror.NewTooManyElementsError(name)
				err.Path = pathutil.Pathstr(path)
				return nil, err
			}
			fields[name] = c
			list = append(list, c)
		}
	}

	return list, nil
}

type xmlUnmarshaller struct {
	valType schema.ValidationType
}

func newXMLUnmarshaller() *xmlUnmarshaller {
	return &xmlUnmarshaller{valType: schema.ValidateAll}
}

func (xu *xmlUnmarshaller) SetValidation(
	valType schema.ValidationType,
) Unmarshaller {
	xu.valType = valType
	return xu
}

func (xu *xmlUnmarshaller) Unmarshal(
	sn schema.Node,
	input []byte,
) (datanode.DataNode, error) {
	return unmarshalXMLInternal(
		sn,
		input,
		xu.valType)
}

func UnmarshalXML(sn schema.Node, xml_input []byte) (datanode.DataNode, error) {
	return unmarshalXMLInternal(sn, xml_input, schema.ValidateAll)
}

func unmarshalXMLInternal(
	sn schema.Node,
	xml_input []byte,
	valType schema.ValidationType,
) (datanode.DataNode, error) {

	var xmlNode unmarshaledXML
	if err := xml.Unmarshal(xml_input, &xmlNode); err != nil {
		return nil, err
	}

	datatree, err := convertToDataNode([]string{}, sn.Name(), &xmlNode, sn)
	if err != nil {
		return nil, err
	}

	err = validateDataNode(datatree, sn, valType)
	if err != nil {
		return nil, err
	}

	return datatree, nil
}

func encodeXmlChildren(enc *xml.Encoder, sn schema.Node, n datanode.DataNode) {

	for _, cn := range n.YangDataChildren() {
		csn := sn.Child(cn.YangDataName())
		c_name := xml.Name{Space: csn.Namespace(), Local: csn.Name()}
		switch csn.(type) {
		case schema.Container, schema.ListEntry, schema.Tree:
			enc.EncodeToken(xml.StartElement{Name: c_name})
			encodeXmlChildren(enc, csn, cn)
			enc.EncodeToken(xml.EndElement{Name: c_name})

		case schema.List:
			encodeXmlChildren(enc, csn, cn)

		case schema.Leaf, schema.LeafList:
			for _, v := range cn.YangDataValues() {
				enc.EncodeToken(xml.StartElement{Name: c_name})
				enc.EncodeToken(xml.CharData([]byte(v)))
				enc.EncodeToken(xml.EndElement{Name: c_name})
			}
		}
	}
}

func ToXML(sn schema.Node, node datanode.DataNode) []byte {
	var b bytes.Buffer
	enc := xml.NewEncoder(&b)

	enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: node.YangDataName()}})
	encodeXmlChildren(enc, sn, node)
	enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: node.YangDataName()}})

	enc.Flush()
	return b.Bytes()
}
