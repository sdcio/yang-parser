// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains the NodeRef object

package xutils

import (
	"bytes"
	"fmt"
)

// NodeRef: These allow Leafref-like objects be constructed and manipulated.
// These objects are references to nodes and so stop short of containing the
// actual value of leaf / leaf-list elements.
//
// NB:
//   (1) Root node is represented by a NodeRef of length 0.
//
//   (2)All NodeRefs are absolute.
//
type NodeRef struct {
	elems []NodeRefElem
}

type NodeRefElem struct {
	name string
	keys []NodeRefKey
}

type NodeRefKey struct {
	keyName string
	value   string
}

func NewNodeRef(entries int) NodeRef {
	var retNodeRef NodeRef
	retNodeRef.elems = make([]NodeRefElem, entries)
	return retNodeRef
}

func (yp *NodeRef) AddElem(name string, keys []NodeRefKey) {
	yp.elems = append(yp.elems, NodeRefElem{
		name: name, keys: keys})
}

func NewNodeRefKey(name, value string) (retKey NodeRefKey) {
	return NodeRefKey{name, value}
}

func (yp NodeRef) String() string {
	var b bytes.Buffer
	for _, elem := range yp.elems {
		b.WriteString("/" + elem.name)
		for _, key := range elem.keys {
			b.WriteString(fmt.Sprintf("[%s='%s']", key.keyName, key.value))
		}
	}
	return b.String()
}

func (yp1 NodeRef) EqualTo(yp2 NodeRef) bool {
	if len(yp1.elems) != len(yp2.elems) {
		return false
	}

	for index, elem := range yp1.elems {
		if !elem.EqualTo(yp2.elems[index]) {
			return false
		}
	}
	return true
}

func (ype1 NodeRefElem) EqualTo(ype2 NodeRefElem) bool {
	if ype1.name != ype2.name {
		return false
	}
	if len(ype1.keys) != len(ype2.keys) {
		return false
	}

	for index, key := range ype1.keys {
		if !key.EqualTo(ype2.keys[index]) {
			return false
		}
	}
	return true
}

func (ypk1 NodeRefKey) EqualTo(ypk2 NodeRefKey) bool {
	if ypk1.keyName != ypk2.keyName {
		return false
	}
	if ypk1.value != ypk2.value {
		return false
	}
	return true
}

func getNodeRef(node XpathNode) NodeRef {
	// Find out number of nodes.
	var count = 0
	for curNode := node; curNode != nil; curNode = curNode.XParent() {
		count++
	}
	if count <= 1 {
		// Either no path or root node - return empty in either case.
		return NodeRef{}
	}

	// Root node is not added so we remove one from initial count, and
	// test against parent for nil not current node to avoid adding it.
	count--
	retPath := NewNodeRef(count)
	for curNode := node; curNode.XParent() != nil; curNode = curNode.XParent() {
		count--
		retPath.elems[count].name = curNode.XName()
		retPath.elems[count].keys = curNode.XListKeys() // May be nil!
	}

	return retPath
}

// Dumb function that uses brute force to find a node with the given path.
// Designed for use with leafrefs and tab-completion.
func FindNode(startNode XpathNode, pathToFind NodeRef) XpathNode {
	findFn := func(node XpathNode, index int) (bool, error) {
		if getNodeRef(node).EqualTo(pathToFind) {
			return true, nil
		}
		return false, nil
	}

	retNode, _, err := WalkTree(startNode, findFn, 0)
	if err != nil {
		return nil
	}
	return retNode
}
