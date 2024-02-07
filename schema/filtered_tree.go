// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/sdcio/yang-parser/data/datanode"
)

type Filter func(Node, datanode.DataNode, []datanode.DataNode) bool
type filteredTree struct {
	datanode.DataNode                     // The underlying data tree
	children          []datanode.DataNode // Cached children
	sch               Node                // The schema for the tree
}

// Create a wrapper around a datanode that filters out unwanted nodes
func FilterTree(schema Node, under datanode.DataNode, keep_it Filter) datanode.DataNode {

	var children []datanode.DataNode

	// Recursively filter in required decorated children
	for _, cn := range under.YangDataChildren() {

		name := cn.YangDataName()
		csn := schema.Child(name)

		child := FilterTree(csn, cn, keep_it)
		if child == nil {
			continue
		}

		children = append(children, child)
	}

	if _, isRoot := schema.(Tree); isRoot {
		return &filteredTree{under, children, schema}
	}

	if !keep_it(schema, under, children) {
		return nil
	}

	switch y := schema.(type) {
	case ListEntry:
		// If it's a list entry, then ensure keys are included
		var missing_keys []datanode.DataNode

	KeyLoop:
		for _, key := range y.Keys() {
			for _, ch := range children {
				if ch.YangDataName() == key {
					// Already have key in list
					continue KeyLoop
				}
			}

			for _, cn := range under.YangDataChildren() {
				if cn.YangDataName() == key {
					missing_keys = append(missing_keys, cn)
					break
				}
			}
		}
		children = append(missing_keys, children...)

	case Container:
		// Don't include empty non-presence containers
		if !schema.HasPresence() {
			if len(children) == 0 {
				return nil
			}
		}
	}
	return &filteredTree{under, children, schema}
}

// Override the underlying datanode's implementation
func (n *filteredTree) YangDataChildren() []datanode.DataNode {
	return n.children
}

func (n *filteredTree) YangDataChildrenNoSorting() []datanode.DataNode {
	return n.children
}
