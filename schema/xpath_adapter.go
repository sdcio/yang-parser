// Copyright 2024 Nokia
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Copyright (c) 2018-2019,2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains the XpathNode interface along with various
// helper functions that operate on nodes and nodesets.

package schema

import (
	"encoding/xml"
	"sort"

	"github.com/danos/utils/natsort"

	"github.com/sdcio/yang-parser/data/datanode"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

// XPath Support
//
// XPath requires node hierarchy to conform to a specific model such that
// child / parent operations return the expected set of nodes.  If we take
// the following configuration:
//
//	interface {
//		dataplane dp0s1 {
//			address 1234
//			address 1235
//			address 4444
//		}
//		dataplane dp0s2
//		loopback lo2
//		serial s1 {
//			address 1234
//		}
//	}
//	protocols {
//		mpls {
//			min-label 16
//		}
//	}
//
// ... then the diffNode structure looks like this (number on LHS is 'depth'):
//
//	0: 'root'                     schema.Tree
//	1:     'interface'                *schema.Container
//	2:         'dataplane'                *schema.List
//	3:             'dp0s1'                    schema.ListEntry
//	4:                 'address'                  schema.LeafList
//	5: 				    '1234'                     schema.LeafValue
//	5: 				    '1235'                     schema.LeafValue
//	5: 				    '4444'                     schema.LeafValue
//	3:             'dp0s2'                    schema.ListEntry
//	2:         'loopback'                 *schema.List
//	3:             'lo2'                      schema.ListEntry
//	2:         'serial'                   *schema.List
//	3:             's1'                       schema.ListEntry
//	4:                 'address'                  schema.LeafList
//	5:                     '1234'                     schema.LeafValue
//	1:     'protocols'                *schema.Container
//	2:         'mpls'                     *schema.Container
//	3:             'min-label'                schema.Leaf
//	4:                 '16'                       schema.LeafValue
//
// This has two problems:
//
// (a) Lists exist at 2 levels (list, and key/tagnode) whereas
//
//	we want a single set of list entry nodes that contain the keys, and then
//	keys also appear as distinct children.
//
// (b) Leaves / Leaf-lists exist at 2 levels (name, value(s)) whereas we want
//
//	a single level of nodes representing each {name, value} pair.
//
// Making these transforms converts above structure into that below.  We
// now go from depth of 0 - 3 below, versus 0 - 5 above.  One might note the
// similarity to the configuration with the sole exception that the list key
// is now explicitly shown as a child of the list.  One might also note that
// going from single to multiple key support will be interesting here ...
//
//	0: 'root'                           schema.Tree
//	1:     'interface'                      *schema.Container
//	2:         'dataplane' {name, dp0s1}        schema.ListEntry
//	3:             name: 'dp0s1'                    Diff.XpathListKeyNode
//	3:             address: '1234'                  schema.LeafValue
//	3: 		    address: '1235'                  schema.LeafValue
//	3: 			address: '4444'                  schema.LeafValue
//	2:         'dataplane' {name, dp0s2}        schema.ListEntry
//	3:             name: 'dp0s2'                    Diff.XpathListKeyNode
//	2:         'loopback' {name, lo2}           schema.ListEntry
//	3:             name: 'lo2'                      Diff.XpathListKeyNode
//	2:         'serial', {name, s1}             schema.ListEntry
//	3:             name: 's1'                       Diff.XpathListKeyNode
//	3:             address: '1234'                  schema.LeafValue
//	1:     'protocols'                          *schema.Container
//	2:         'mpls'                           *schema.Container
//	3:             min-label: '16'                  schema.LeafValue
//
// To implement the transforms, we need to do the following when navigating
// the diff / schema tree:
//
// (a) When we get a List as a child, we replace with ALL ListEntry children
//
//	as we need one child per ListEntry.  A 'List' node doesn't represent
//	an XPath-addressable node.
//
// (b) ListEntry has Name() of parent node. Value() is not relevant as we
//
//	only need to provide a Value() for leaves.  It should however be able
//	to interpret key values when we introduce predicates.
//
// (c) Parent of a ListEntry is actually its grandparent, as we skip the List
//
//	node.  Taking (a) and (c) together means you can never get a List node
//	returned by XChildren() or Parent().
//
// (d) When we generate the children of a ListEntry, we must generate a
//
//	node representing each key.  To handle all these we create a virtual
//	node, a DiffXpathListKeyNode, that is essentially a Diff Node but which
//	overrides various functions so the node appears at the child level
//	in the tree relative to the Diff.Node it is derived from.
//
// (e) LeafLists are similar to lists in that we have to return the children
//
//	instead of the LeafList.  It is simplest to treat Leaves as single
//	element LeafLists.
type xnode interface {
	datanode.DataNode
	xutils.XpathNode

	children(sortSpec xutils.SortSpec) []xnode
	schema() Node
	path() []string
}

type xdatanode struct {
	datanode.DataNode
	sch       Node
	parent    *xdatanode
	ephemeral bool
}

func ConvertToXpathNode(c datanode.DataNode, s Node) xutils.XpathNode {
	return &xdatanode{c, s, nil, false}
}

func createXNode(c datanode.DataNode, s Node, p *xdatanode) *xdatanode {
	return &xdatanode{c, s, p, false}
}

func createEphemeralXNode(
	c datanode.DataNode,
	s Node,
	p *xdatanode,
) *xdatanode {
	return &xdatanode{c, s, p, true}
}

func (n *xdatanode) XIsLeaf() bool {
	_, ok := n.sch.(Leaf)
	return ok
}

func (n *xdatanode) XIsLeafList() bool {
	_, ok := n.sch.(LeafList)
	return ok
}

func (n *xdatanode) XIsNonPresCont() bool {
	node, ok := n.sch.(Container)
	if !ok {
		return false
	}
	return !node.HasPresence()
}

func (n *xdatanode) XIsEphemeral() bool { return n.ephemeral }

// If node has a key with the given value, return true.
func (n *xdatanode) XListKeyMatches(testKey xml.Name, val string) bool {
	// Technically we should probably be checking the namespace of the key
	// not the ListEntry, but as you can't augment a listentry with a key,
	// it comes to the same thing.
	if n.schema().Namespace() != testKey.Space {
		return false
	}

	if lesn, ok := n.schema().(ListEntry); ok {
		for _, key := range lesn.Keys() {
			if key == testKey.Local {
				if n.XValue() == val {
					return true
				}
			}
		}
	}

	return false
}

func (n *xdatanode) XListKeys() []xutils.NodeRefKey {
	if lesn, ok := n.schema().(ListEntry); ok {
		var keys []xutils.NodeRefKey
		for _, key := range lesn.Keys() {
			keys = append(keys, xutils.NewNodeRefKey(key, n.XValue()))
		}
		return keys
	}
	return nil
}

func (n *xdatanode) XParent() xutils.XpathNode {
	parent := n.parent
	if parent == nil {
		// Must return explicit nil or we'll get 'interface' nil which is
		// not the same.
		return nil
	}

	// Return grandparent if LeafValue (skipping Leaf / LeafList)
	// Also return grandparent if ListEntry (skipping List)
	switch n.sch.(type) {
	case ListEntry, LeafValue:
		return parent.XParent()
	}

	return n.parent
}

func (n *xdatanode) XRoot() xutils.XpathNode {
	retNode := n
	for retNode.parent != nil {
		retNode = retNode.parent
	}
	return retNode
}

func (n *xdatanode) XName() string {
	return n.sch.Name()
}

func (n *xdatanode) XValue() string {
	switch n.sch.(type) {
	case ListEntry, LeafValue:
		// When we have key support this will need to change, but for now
		// we have a single key/value pair with value stored in Name() and
		// key name in parent (List element) Key array ([0] element!).
		return n.YangDataName()
	default:
		return ""
	}
}

func (n *xdatanode) XPath() xutils.PathType {
	if n == nil {
		return xutils.PathType([]string{"/"})
	}
	switch n.sch.(type) {
	case Tree:
		return xutils.PathType([]string{"/"})
	case ListEntry, LeafValue:
		return n.parent.XPath()
	default:
		// For RPCs, the 'top-level' node is a Container not a Tree, so we need
		// to avoid recursing upwards!
		if n.parent == nil {
			return xutils.PathType([]string{n.XName()})
		}
		return append(n.parent.XPath(), n.XName())
	}
}
func (n *xdatanode) isKey() bool {
	if n.parent == nil {
		return false
	}

	if lesn, ok := n.parent.schema().(ListEntry); ok {
		key := lesn.Keys()[0]
		if n.schema().Name() == key {
			return true
		}
	}

	return false
}

func (n *xdatanode) path() []string {
	if n.parent == nil {
		// This is the root node
		return []string{}
	}

	if n.isKey() {
		// We're the key, configd path is up one
		return n.parent.path()
	}

	return append(n.parent.path(), n.YangDataName())
}

func (n *xdatanode) XChildren(
	filter xutils.XFilter,
	sortSpec xutils.SortSpec,
) []xutils.XpathNode {

	// Return early for nodes we know need no processing.
	switch n.sch.(type) {
	case List, Leaf, LeafList, LeafValue:
		// We either cannot be called with one of these (as they are skipped),
		// or they cannot have children that we would return.
		return nil
	}

	xChildren := make([]xutils.XpathNode, 0)

	// For valid nodes, add all children that match the filter to the list
	// of returned nodes.
	//
	// Each child gets a unique index value, even if not returned by this
	// filter.  That is because we need to get the same index for the same
	// child with different filters to be able to remove duplicate nodes from
	// a nodeset.
	children := n.children(sortSpec)
	for _, child := range children {
		targetType := xutils.NotConfigOrOpdTarget
		if child.schema().Config() {
			targetType = xutils.ConfigTarget
		}
		switch child.schema().(type) {
		case Tree, Container:
			if xutils.MatchFilter(filter,
				xutils.NewXTarget(
					xml.Name{Space: child.schema().Namespace(),
						Local: child.XName()},
					targetType)) {
				xChildren = append(xChildren, child)
			}
		case Leaf, LeafList, List:
			// Treat Leaf as a degenerate (single entry) (Leaf)List.
			if xutils.MatchFilter(filter,
				xutils.NewXTarget(
					xml.Name{Space: child.schema().Namespace(),
						Local: child.XName()},
					targetType)) {
				innerChildren := child.children(sortSpec)
				for _, innerChild := range innerChildren {
					if !filter.MatchConfigOnly() ||
						innerChild.schema().Config() {
						xChildren = append(xChildren, innerChild)
					}
				}
			}
		}
	}

	if len(xChildren) == 0 {
		// Return nil rather than empty array because xpath code
		// treats them differently
		return nil
	}

	return xChildren
}

type bySystem []xnode

func (b bySystem) Len() int      { return len(b) }
func (b bySystem) Swap(i, j int) { b[i], b[j] = b[j], b[i] }

func (b bySystem) Less(i, j int) bool {
	return natsort.Less(b[i].XName(), b[j].XName())
}

type bySystemValue []xnode

func (b bySystemValue) Len() int      { return len(b) }
func (b bySystemValue) Swap(i, j int) { b[i], b[j] = b[j], b[i] }
func (b bySystemValue) Less(i, j int) bool {
	return natsort.Less(b[i].XValue(), b[j].XValue())
}

func (n *xdatanode) children(sortSpec xutils.SortSpec) []xnode {

	children := []xnode{}

	switch n.sch.(type) {
	case Leaf, LeafList:
		switch n.sch.Type().(type) {
		case Empty:
			// Even if leaf-list, can't have more than one entry as they
			// must be unique.  Empty-type leaf-lists are of dubious validity
			// but here isn't the place to complain!
			children = append(children, createEmptyLeafNode(n))
		default:
			for _, leafValue := range n.YangDataValuesNoSorting() {
				leafNode := createValueNode(
					leafValue, n.sch.Child(leafValue), n)
				children = append(children, leafNode)
			}
		}

	default:
		// We sort at the end - no need to do it twice.
		for _, child := range n.YangDataChildrenNoSorting() {
			csn := n.sch.Child(child.YangDataName())
			children = append(children, createXNode(child, csn, n))
		}
	}

	if sortSpec == xutils.Unsorted {
		return children
	}

	switch n.sch.(type) {
	case Leaf:
	case LeafList, List:
		if n.sch.OrdBy() != "user" {
			sort.Sort(bySystemValue(children))
		}
	default:
		sort.Sort(bySystem(children))
	}

	return children
}

func (n *xdatanode) schema() Node {
	return n.sch
}

// Value nodes
//
// In XPath the values are treated like nodes in the document graph. This structure
// is used to convert string values into xpath compatible nodes. LeafLists will
// produce multiple nodes for the same schema, but with different values.
type xvaluenode struct {
	value  string
	sch    Node
	parent *xdatanode
}

func createValueNode(value string, sch Node, parent *xdatanode) *xvaluenode {
	return &xvaluenode{value, sch, parent}
}

func (n *xvaluenode) children(sortSpec xutils.SortSpec) []xnode { return nil }
func (n *xvaluenode) schema() Node                              { return n.sch }

func (n *xvaluenode) path() []string {
	if n.parent != nil && n.parent.isKey() {
		// We're the key, configd path is up two
		return n.parent.parent.path()
	}

	return append(n.parent.path(), n.YangDataName())
}

func (n *xvaluenode) YangDataName() string                  { return n.value }
func (n *xvaluenode) YangDataChildren() []datanode.DataNode { return nil }
func (n *xvaluenode) YangDataValues() []string              { return nil }

func (n *xvaluenode) YangDataChildrenNoSorting() []datanode.DataNode {
	return nil
}
func (n *xvaluenode) YangDataValuesNoSorting() []string {
	return nil
}

func (n *xvaluenode) XChildren(
	filter xutils.XFilter,
	sortSpec xutils.SortSpec) []xutils.XpathNode {
	return nil
}

func (n *xvaluenode) XListKeyMatches(key xml.Name, val string) bool {
	return false
}
func (n *xvaluenode) XListKeys() []xutils.NodeRefKey { return nil }

func (n *xvaluenode) XIsLeaf() bool             { return true }
func (n *xvaluenode) XIsLeafList() bool         { return false }
func (n *xvaluenode) XIsNonPresCont() bool      { return false }
func (n *xvaluenode) XIsEphemeral() bool        { return false }
func (n *xvaluenode) XName() string             { return n.sch.Name() }
func (n *xvaluenode) XValue() string            { return n.value }
func (n *xvaluenode) XRoot() xutils.XpathNode   { return n.parent.XRoot() }
func (n *xvaluenode) XParent() xutils.XpathNode { return n.parent.XParent() }
func (n *xvaluenode) XPath() xutils.PathType    { return n.parent.XPath() }

// Empty leaves are a specialisation of the xvaluenode type.  Specifically,
// value is always the empty string, and we need to override the path as
// our value is actually really nothing, not the empty string.
type xemptyleafnode struct {
	xvaluenode
}

// Make sure we don't end up with a trailing space on the path name.
func (n *xemptyleafnode) path() []string {
	return n.parent.path()
}

func createEmptyLeafNode(n *xdatanode) xnode {
	switch n.sch.(type) {
	case Leaf, LeafList:
		return &xemptyleafnode{xvaluenode{"", n.sch.Child(""), n}}
	default:
		return nil
	}
}
