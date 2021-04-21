// Copyright (c) 2018-2019,2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains the XpathNode interface along with various
// helper functions that operate on nodes and nodesets.

package xutils

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"
)

type SortSpec bool

const (
	Unsorted SortSpec = false
	Sorted            = true
)

// To isolate us from node types we may want to work with, we have our own
// interface.  To avoid any namespace collisions with other interfaces, all
// methods are prefixed with 'X'.
type XpathNode interface {
	// Return parent node
	XParent() XpathNode

	// Return all children, including list keys.
	// Specify 'Sorted' to get returned nodes in deterministic order.
	// Xpath uses 'document' order, so for YANG, our system sorts in natural
	// sorting order, unless ordered-by-user is specified.
	// 'Unsorted' should be used when order doesn't matter as it is much faster.
	XChildren(filter XFilter, sortSpec SortSpec) []XpathNode

	// Should return {"/"} for root node.  Returns XPATH-compliant path
	// where tagnodes and other list elements are treated as siblings.
	XPath() PathType

	XRoot() XpathNode

	// Node name.  For list entries, eg interfaces/dataplane entries, all list
	// entries would share the 'dataplane' name.  The 'tagnode' children
	// would have 'tagnode' as name etc.
	XName() string

	// Node value, if a text node (ie leaf / leaf list element).
	XValue() string

	// Type check functions to make sure we are operating on the expected
	// type.
	XIsLeaf() bool
	XIsLeafList() bool
	XIsNonPresCont() bool

	// Ephemeral nodes are created for the purposes of evaluating must
	// statements on non-presence, unconfigured, containers.
	XIsEphemeral() bool

	// Return true if node is a ListEntry with a key that has the given value.
	XListKeyMatches(key xml.Name, val string) bool

	// If node is a list entry, return keys.  Otherwise return nil.
	XListKeys() []NodeRefKey
}

// If 2 nodes have the same NodeString then they are identical.  Two
// separate list elements may have the same path, but add in the key values
// and they differ again.
func NodesEqual(n1, n2 XpathNode) error {
	if NodeString(n1) != NodeString(n2) {
		return fmt.Errorf("Nodes have different index strings: %s vs %s",
			NodeString(n1), NodeString(n2))
	}

	return nil
}

func NodesetsEqual(ns1, ns2 []XpathNode) error {
	if len(ns1) != len(ns2) {
		var ns1Names, ns2Names string
		for _, n1 := range ns1 {
			ns1Names = ns1Names + " " + n1.XName()
		}
		for _, n2 := range ns2 {
			ns2Names = ns2Names + " " + n2.XName()
		}
		return fmt.Errorf("Nodesets have different length: %d (%s) vs %d (%s)",
			len(ns1), ns1Names, len(ns2), ns2Names)
	}

	for index, n1 := range ns1 {
		if err := NodesEqual(n1, ns2[index]); err != nil {
			return err
		}
	}

	return nil
}

// This is the code that actually does the work for leafref predicates.
// We need to find the nodes (if any) in the nodeset that match the
// given key/value tuple.
func FilterNodeset(
	ns []XpathNode,
	key xml.Name,
	leafValue string,
) (retNs []XpathNode, debugLog string) {
	var b bytes.Buffer

	if len(ns) == 0 {
		return nil, ""
	}
	retNs = make([]XpathNode, 0, len(ns))
	for _, node := range ns {
		if node.XListKeyMatches(key, leafValue) {
			retNs = append(retNs, node)
		}
	}
	if len(retNs) == 0 {
		return nil, ""
	}

	return retNs, b.String()
}

// Generate a unique string representation of a node, in 'leafref' format,
// including all key values for list entries, and ending with a leaf or
// leaf-list value if the node is of those types.
//
// NB: This function is used to determine if 2 nodes are identical, not
//     just for pretty-printing.  If in any doubt that changes may impact
//     performance, try out TestFWPerformance in configd/session directory
//     with a large number of rules and compare old and new times.
//
// NB: This is NOT string-value, which is not unique to a specific node,
//     and which can vary for a single node depending on what is configured
//     under it!
//
func NodeString(xNode XpathNode) string {
	if xNode.XIsLeaf() || xNode.XIsLeafList() {
		return getNodeRef(xNode).String() + fmt.Sprintf(" (%s)",
			xNode.XValue())
	}

	return getNodeRef(xNode).String()
}

// Return a slice of strings containing the string-value for each node in
// <nodes>.  If <addEmptyStr> is true, return a slice with a single empty
// string instead of an empty slice when <nodes> is empty.  This is useful
// when using the strings with equality or relational operators.
func GetStringValues(nodes []XpathNode, addEmptyStr bool) []string {
	var retStrs []string

	for _, node := range nodes {
		retStrs = append(retStrs, constructStringValue(node, ""))
	}

	if addEmptyStr && (len(retStrs) == 0) {
		retStrs = append(retStrs, "")
	}

	return retStrs
}

// If no nodes, return empty string.  Otherwise return string-value of first
// node in nodeset.
//
// Just as a reminder, the string-value for anything other than a leaf node
// is (a) variable (depends what child nodes are configured) and (b) not
// unique to a node (imagine a chain of node -> child -> grandchild -> leaf
// which, in the absence of any defaults, or multiple nodes at any level,
// would all have the same value).
//
func GetStringValue(nodes []XpathNode) string {
	if len(nodes) == 0 {
		return ""
	}

	var stringValue string

	return constructStringValue(nodes[0], stringValue)
}

func constructStringValue(node XpathNode, stringValue string) string {
	children := node.XChildren(AllChildren, Sorted)
	if children == nil {
		return stringValue + node.XValue()
	}

	var childrenStr string
	for _, child := range children {
		childrenStr = childrenStr + constructStringValue(child, "")
	}

	return stringValue + childrenStr
}

// Naive implementation is to simply loop through range of nodes and if not
// already in our return list, add them.  However, this quickly multiplies up
// if we were to have a large number of nodes.
//
// Instead, we create a map with a unique representation of each node, being
// the NodeString (categorically NOT the string-value which varies for a
// single node and may be the same across multiple nodes)
//
// NodeString is guaranteed unique for different nodes as it combines both
// path and (where siblings with the same path exist) value.
//
func RemoveDuplicateNodes(nodes []XpathNode) []XpathNode {
	var retNodes = make([]XpathNode, 0)
	var nodeMap = make(map[string]struct{}, len(nodes))

	for _, node := range nodes {
		mapKey := NodeString(node)
		if _, inMap := nodeMap[mapKey]; !inMap {
			nodeMap[mapKey] = struct{}{}
			retNodes = append(retNodes, node)
		}
	}

	return retNodes
}

type worker func(node XpathNode, index int) (done bool, err error)

func PrintTree(startNode XpathNode) {
	printFn := func(node XpathNode, index int) (bool, error) {
		const indentStr = "                                                "
		indent := strings.Count(NodeString(node), "/")
		fmt.Printf("%s %s: {%s} %s\n",
			indentStr[:indent],
			node.XName(), node.XValue(), NodeString(node))
		return false, nil
	}

	WalkTree(startNode, printFn, 0)
}

// Takes root node and verifies at each level:
// - parent is correct
// - root is correct
// - children have correct parent
// - children have unique NodeString
// - no child is equal to another child
// - index is as expected
func ValidateTree(root XpathNode) error {
	// First, check we have been given root node.
	if err := NodesEqual(root, root.XRoot()); err != nil {
		return fmt.Errorf(
			"Either root's root pointer isn't set, or root isn't root!")
	}

	if root.XParent() != nil {
		return fmt.Errorf("Root cannot have a parent!")
	}

	validateFn := func(node XpathNode, index int) (bool, error) {
		// Root node check
		if err := NodesEqual(root, node.XRoot()); err != nil {
			return true, fmt.Errorf("Node %s has wrong root node", node.XName())
		}

		// Children checks
		children := node.XChildren(AllChildren, Sorted)
		pathIndexMap := make(map[string]bool, len(children))
		for _, child := range children {
			if err := NodesEqual(root, child.XRoot()); err != nil {
				return true, fmt.Errorf("Child %s of %s has wrong root.",
					child.XName(), node.XName())
			}

			if err := NodesEqual(node, child.XParent()); err != nil {
				return true, fmt.Errorf("Child %s of %s has wrong parent.",
					child.XName(), node.XName())
			}

			pathIndexStr := NodeString(child)
			if _, ok := pathIndexMap[pathIndexStr]; ok {
				return true,
					fmt.Errorf("Child doesn't have unique index string.")
			}
			pathIndexMap[pathIndexStr] = true
		}

		return false, nil
	}

	_, _, err := WalkTree(root, validateFn, 0)
	return err
}

// Walk tree, calling workFn for each node, then children, recursively.
// If workFn returns error, end walk immediately.
func WalkTree(
	node XpathNode,
	workFn worker,
	index int,
) (retNode XpathNode, finished bool, retErr error) {
	if done, err := workFn(node, index); done || err != nil {
		return node, done, err
	}

	for index, child := range node.XChildren(AllChildren, Sorted) {
		retNode, done, err := WalkTree(child, workFn, index)
		if done || err != nil {
			return retNode, done, err
		}
	}

	return node, false, nil
}
