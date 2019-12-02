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
	"strings"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/schema"
)

type unmarshaledXML struct {
	XMLName  xml.Name
	XMLAttr  []xml.Attr        `xml:",any,attr"`
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
			list[i] = (string)(v.Chardata)
		}
		return list, nil
	}
	return []string{xmlNode.Chardata}, nil
}

func locateIdentity(typ schema.Type, val, ns string) *schema.Identity {
	switch t := typ.(type) {
	case schema.Identityref:
		for _, id := range t.Identities() {
			if id.Namespace == ns && id.Value == val {
				return id
			}
		}
	case schema.Union:
		for _, utyp := range t.Typs() {
			i := locateIdentity(utyp, val, ns)
			if i != nil {
				return i
			}
		}
	default:
	}
	return nil
}

func (xmlNode *unmarshaledXML) convertPrefixedValue(sn schema.Node) {
	for _, atr := range xmlNode.XMLAttr {
		if atr.Name.Space == "xmlns" {
			if strings.HasPrefix(xmlNode.Chardata, atr.Name.Local+":") {
				// Assume that this is an identityref value
				id := locateIdentity(sn.Type(), strings.TrimPrefix(xmlNode.Chardata, atr.Name.Local+":"), atr.Value)
				if id != nil {
					xmlNode.Chardata = id.Val
				}
			}
		}
	}

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
		case schema.LeafList:
			c.convertPrefixedValue(cn)
			if !ok {
				v = &unmarshaledXML{c.XMLName, c.XMLAttr, "", make([]*unmarshaledXML, 0)}
				fields[name] = v
				list = append(list, v)
			}
			v.Children = append(v.Children, c)
		case schema.List:
			// We may validly have multiple list elements with the same
			// name so no need to check ok.  For each element we create a
			// List entry in <list>, with a single child for the listEntry.
			v = &unmarshaledXML{c.XMLName, c.XMLAttr, "", make([]*unmarshaledXML, 0)}
			fields[name] = v
			list = append(list, v)
			v.Children = append(v.Children, c)
		case schema.Leaf:
			if ok {
				err := mgmterror.NewTooManyElementsError(name)
				err.Path = pathutil.Pathstr(path)
				return nil, err
			}
			c.convertPrefixedValue(cn)
			fields[name] = c
			list = append(list, c)
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

func namespacePrefixes(sn schema.Node, value string) []xml.Attr {
	nsprefixes := make([]xml.Attr, 0)

	typ := sn.Type()
	if utyp, ok := typ.(schema.Union); ok {
		// For a union type, get the base type the value matches
		// so that it can be correctly encoded
		typ = utyp.MatchType(nil, []string{}, value)
	}

	if idref, ok := typ.(schema.Identityref); ok {
		// identityref requires local prefix for identity namepace
		for _, idn := range idref.Identities() {
			if value == idn.Val && sn.Namespace() != idn.Namespace {
				nsprefixes = append(nsprefixes,
					xml.Attr{Name: xml.Name{Local: "xmlns:" + idn.Module},
						Value: idn.Namespace})
			}
		}
	}

	return nsprefixes
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
				nsprefixes := namespacePrefixes(csn, v)
				enc.EncodeToken(xml.StartElement{Name: c_name, Attr: nsprefixes})
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
