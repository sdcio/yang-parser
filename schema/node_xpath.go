// Copyright (c) 2018-2019,2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains the XNode object used for walking schema nodes.

package schema

import (
	"encoding/xml"
	"reflect"

	"github.com/iptecharch/yang-parser/xpath/xutils"
)

type XNode struct {
	Node
	parent *XNode
}

func NewXNode(sn Node, parent *XNode) *XNode {
	return &XNode{Node: sn, parent: parent}
}

func (xn *XNode) XParent() xutils.XpathNode {
	if xn.parent == nil {
		// Must return explicit nil or we'll get 'interface' nil which is
		// not the same.
		return nil
	}

	return xn.parent
}

func (xn *XNode) XChildren(
	filter xutils.XFilter,
	sortSpec xutils.SortSpec,
) []xutils.XpathNode {
	xChildren := make([]xutils.XpathNode, 0, len(xn.Children()))

	// For valid nodes, add all children that match the filter to the list
	// of returned nodes.
	//
	// Each child gets a unique index value, even if not returned by this
	// filter.  That is because we need to get the same index for the same
	// child with different filters to be able to remove duplicate nodes from
	// a nodeset.
	children := xn.Children()
	for _, child := range children {
		if xutils.MatchFilter(filter,
			xutils.NewXTarget(
				xml.Name{Space: child.Namespace(),
					Local: child.Name()},
				xutils.ConfigTarget)) {
			xChildren = append(xChildren, NewXNode(child, xn))
		}
	}

	if len(xChildren) == 0 {
		// Return nil rather than empty array because xpath code
		// treats them differently
		return nil
	}

	return xChildren
}

func (xn *XNode) XPath() xutils.PathType {
	if xn.parent == nil {
		return xutils.PathType([]string{xn.XName()})
	}
	return append(xn.parent.XPath(), xn.XName())
}

func (xn *XNode) XRoot() xutils.XpathNode {
	retNode := xn
	for retNode.parent != nil {
		retNode = retNode.parent
	}
	return retNode
}

func (xn *XNode) XName() string {
	return xn.Name()
}

func (xn *XNode) XValue() string {
	panic("node_xpath: XValue() not implemented")
}

func (xn *XNode) XIsLeaf() bool {
	if reflect.TypeOf(xn.Node).Name() == "*schema.leaf" {
		return true
	}
	return false
}

func (xn *XNode) XIsLeafList() bool {
	if reflect.TypeOf(xn.Node).Name() == "*schema.leafList" {
		return true
	}
	return false
}

func (xn *XNode) XIsNonPresCont() bool {
	if n, ok := (xn.Node).(*container); ok {
		if !n.HasPresence() {
			return true
		}
	}
	return false
}

// The way XNodes are created means they must refer to real configured nodes,
// so cannot be ephemeral
func (xn *XNode) XIsEphemeral() bool { return false }

func (xn *XNode) XListKeyMatches(key xml.Name, val string) bool {
	panic("node_xpath: XListKeyMatches() not implemented")
}

func (xn *XNode) XListKeys() []xutils.NodeRefKey {
	return nil // Used to remove duplicate nodes which we don't care about.
}
