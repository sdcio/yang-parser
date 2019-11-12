// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/danos/yang/data/datanode"
)

type addDefaults struct {
	datanode.DataNode      // The underlying data tree
	sch               Node // The schema for the tree
}

// Create a wrapper around a datanode that provides defaults
func AddDefaults(schema Node, under datanode.DataNode) datanode.DataNode {
	return &addDefaults{under, schema}
}

// Override the underlying datanode's implementation
func (n *addDefaults) YangDataChildren() []datanode.DataNode {
	children := n.DataNode.YangDataChildren()

	return n.yangDataChildren(children)
}

func (n *addDefaults) YangDataChildrenNoSorting() []datanode.DataNode {
	children := n.DataNode.YangDataChildrenNoSorting()

	return n.yangDataChildren(children)
}

func (n *addDefaults) yangDataChildren(
	children []datanode.DataNode,
) []datanode.DataNode {

	seen := make(map[string]struct{})
	new_children := make([]datanode.DataNode, len(children))

	// Wrap any existing children with addDefaults decorators
	for i, cn := range children {
		name := cn.YangDataName()
		csn := n.sch.Child(name)

		seen[name] = struct{}{}
		new_children[i] = AddDefaults(csn, cn)
	}

	// Add any missing default children
	for _, def := range n.sch.DefaultChildren() {
		name := def.Name()
		if _, ok := seen[name]; ok {
			continue
		}
		new_children = append(new_children, createDefault(def))
	}

	return new_children
}

// Potentially the schema could store the defaults as a DataNode interface
// we could just get them from there.
func createDefault(sch Node) datanode.DataNode {

	switch v := sch.(type) {

	// Note that LeafLists do not currently support defaults
	case Leaf:
		val, _ := v.Default()
		return datanode.CreateDataNode(v.Name(), nil, []string{val})
	}

	var children []datanode.DataNode
	for _, ch := range sch.DefaultChildren() {
		children = append(children, createDefault(ch))
	}

	return datanode.CreateDataNode(sch.Name(), children, nil)
}
