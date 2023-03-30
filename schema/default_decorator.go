// Copyright (c) 2019-2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"github.com/steiler/yang-parser/data/datanode"
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

func (n *addDefaults) isAChoice(name string) bool {
	for _, chs := range n.sch.Choices() {
		if chs.Child(name) != nil {
			return true
		}
	}
	return false
}

type ConfigChecker func(sch Node) bool

func hasCfg(seen map[string]struct{}, sch Node) bool {
	for _, chs := range sch.Children() {
		if _, ok := seen[chs.Name()]; ok {
			return true
		}
	}
	return false
}

// IsActiveDefault - Check default nodes and determine if they are active
// defaults that should be instantiated.
//
// # A node is an active default in one of the given circumstances
//
//   - The node is under a choice but there is no active configuration under the choice,
//     any nodes under the default case, if it is defined, become active defaults.
//   - There is config under one of the cases in a choice, any other node in that case
//     will be an active default.
//   - The node is not under a choice and has a default
//
// isActiveDefault() and isActiveDefaultCase() are called recursively to traverse
// down a choice/case
func IsActiveDefault(sch Node, name string, cfgChkr ConfigChecker) bool {
	return isActiveDefault(sch, name, false, cfgChkr)
}

func isActiveDefault(sch Node, name string, defCase bool, cfgChkr ConfigChecker) bool {
	for _, cd := range sch.Choices() {
		switch choice := cd.(type) {
		case Choice:
			if cd.Child(name) != nil {
				cfg := cfgChkr(cd)
				switch {
				case cfg == true:
					return isActiveDefaultCase(cd, name, "", cfgChkr)
				case cd.HasDefault() == true:
					return isActiveDefaultCase(cd, name, choice.DefaultCase(), cfgChkr)
				}
			}
		default:
			if cd.Name() == name {
				if defCase {
					return true
				}
				if cfgChkr(sch) {
					return true
				}
			}
		}
	}
	return false

}

// Check nodes that are immediately under a case, implicit and explicit, and test
// if it is an active default
// It will be an active default if the case has other active configuration
// or the case is the choice default case and no other config exists for the choice
func isActiveDefaultCase(sch Node, name string, def string, cfg ConfigChecker) bool {
	for _, cd := range sch.Choices() {
		switch cd.(type) {
		case Case:
			ch := cd.Child(name)
			if ch != nil {
				hcfg := cfg(cd)
				if def == "" {
					if !hcfg {
						return false
					}
				} else if def != cd.Name() {
					return false
				}

				return isActiveDefault(cd, name, cd.Name() == def, cfg)
			}
		default:
			if cd.Name() == def && def == name {
				return true
			}
		}
	}
	return false

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
		if n.isAChoice(name) {
			active := IsActiveDefault(n.sch, name,
				func(schNode Node) bool {
					return hasCfg(seen, schNode)
				})
			if !active {
				continue
			}
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
