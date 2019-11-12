// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Lightweight(!) XPathNode implementation used merely for testing.
// Restricted to string-type for keys and values.

package xpathtest

import (
	"encoding/xml"
	"fmt"
	"strings"
	"testing"

	"github.com/danos/yang/xpath/xutils"
)

const (
	TestModule = "xpathNodeTestModule"
)

type xpathKey struct {
	key   string
	value string
}

type nodeType int

const (
	Container nodeType = iota
	ListEntry
	LeafList
	Leaf
)

type TNode struct {
	root      *TNode
	parent    *TNode
	ntype     nodeType
	children  []*TNode
	module    string
	name      string
	path      xutils.PathType
	keys      []xpathKey
	value     string
	empty     bool
	ephemeral bool
}

type TNodeSet []*TNode

// Pretty-print a testnode for debug.
func (testnode *TNode) String() string {
	retStr := fmt.Sprintf("%s (%s)", testnode.name, testnode.path)
	if testnode.value != "" {
		retStr = retStr + fmt.Sprintf(", value '%s'", testnode.value)
	} else {
		fmt.Printf(",")
		for _, key := range testnode.keys {
			retStr = retStr + fmt.Sprintf(" {%s:%s}", key.key, key.value)
		}
	}
	retStr = retStr + fmt.Sprintf(", children: ")
	for _, child := range testnode.children {
		retStr = retStr + child.name + " "
	}
	retStr = retStr + "\n"
	return retStr
}

// XpathNode interface implementation

func (testnode *TNode) XIsLeaf() bool     { return testnode.ntype == Leaf }
func (testnode *TNode) XIsLeafList() bool { return testnode.ntype == LeafList }
func (testnode *TNode) XIsNonPresCont() bool {
	panic("testnode XIsNonPresCont() not yet implemented.")
}
func (testnode *TNode) XIsEphemeral() bool {
	return testnode.ephemeral
}

func (testnode *TNode) XListKeyMatches(key xml.Name, val string) bool {
	if key.Space != testnode.module {
		return false
	}
	for _, testKey := range testnode.keys {
		if testKey.key == key.Local {
			if testKey.value == val {
				return true
			}
		}
	}
	return false
}

func (testnode *TNode) XListKeys() []xutils.NodeRefKey {
	if testnode.ntype == ListEntry {
		var keys []xutils.NodeRefKey
		for _, testkey := range testnode.keys {
			keys = append(
				keys, xutils.NewNodeRefKey(testkey.key, testkey.value))
		}
		return keys
	}
	return nil
}

func (testnode *TNode) XPath() xutils.PathType {
	return testnode.path
}

func (testnode *TNode) XParent() xutils.XpathNode {
	if testnode.parent == nil {
		// Root node.  Can't go any higher!
		return nil
	}

	return testnode.parent
}

func (testnode *TNode) XChildren(filter xutils.XFilter) []xutils.XpathNode {
	var filteredChildren []xutils.XpathNode

	// First add key entries
	for _, childKey := range testnode.keys {
		if xutils.MatchFilter(
			filter, xutils.NewXConfigTarget(
				xml.Name{Space: testnode.module, Local: childKey.key})) {
			path := append(testnode.path, childKey.key)
			childNode := NewTLeaf(testnode.root, path,
				testnode.module, childKey.key, childKey.value)
			childNode.parent = testnode
			filteredChildren = append(filteredChildren, childNode)
		}
	}

	// Now add non-key entries
	for _, child := range testnode.children {
		if xutils.MatchFilter(
			filter, xutils.NewXConfigTarget(
				xml.Name{Space: testnode.module, Local: child.name})) {
			filteredChildren = append(filteredChildren, child)
		}
	}

	return filteredChildren
}

func (testnode *TNode) XRoot() xutils.XpathNode { return testnode.root }

func (testnode *TNode) XName() string { return testnode.name }

func (testnode *TNode) XValue() string { return testnode.value }

func NewTContainer(
	root *TNode,
	startPath xutils.PathType,
	module string,
	name string,
) *TNode {
	// Be careful - we need a completely distinct copy of the startPath.
	pathCopy := make(xutils.PathType, len(startPath))
	copy(pathCopy, startPath)

	return &TNode{
		root:   root,
		path:   pathCopy,
		module: module,
		name:   name,
		ntype:  Container}
}

func NewTEphemeralContainer(
	root *TNode,
	startPath xutils.PathType,
	module string,
	name string,
) *TNode {
	// Be careful - we need a completely distinct copy of the startPath.
	pathCopy := make(xutils.PathType, len(startPath))
	copy(pathCopy, startPath)

	return &TNode{
		root:      root,
		path:      pathCopy,
		module:    module,
		name:      name,
		ntype:     Container,
		ephemeral: true}
}

func NewTListEntry(
	root *TNode,
	startPath xutils.PathType,
	module string,
	name string,
	key, keyValue string,
) *TNode {
	// Be careful - we need a completely distinct copy of the startPath.
	pathCopy := make(xutils.PathType, len(startPath))
	copy(pathCopy, startPath)
	newTNode := &TNode{
		root:   root,
		path:   pathCopy,
		module: module,
		name:   name,
		ntype:  ListEntry,
		value:  keyValue}
	newTNode.addKeyValue(key, keyValue)
	return newTNode
}

func NewTLeaf(
	root *TNode,
	startPath xutils.PathType,
	module string,
	name string,
	value string,
) *TNode {
	return NewTLeafOrLeafList(root, startPath, module, name, value, Leaf)
}

func NewTLeafList(
	root *TNode,
	startPath xutils.PathType,
	module string,
	name string,
	value string,
) *TNode {
	return NewTLeafOrLeafList(root, startPath, module, name, value, LeafList)
}

func NewTLeafOrLeafList(
	root *TNode,
	startPath xutils.PathType,
	module string,
	name string,
	value string,
	ntype nodeType,
) *TNode {

	// Be careful - we need a completely distinct copy of the startPath.
	pathCopy := make(xutils.PathType, len(startPath))
	copy(pathCopy, startPath)

	return &TNode{
		root:   root,
		path:   pathCopy,
		module: module,
		name:   name,
		value:  value,
		ntype:  ntype}
}

func NewTEmptyLeaf(
	root *TNode,
	startPath xutils.PathType,
	module string,
	name string,
) *TNode {
	// Be careful - we need a completely distinct copy of the startPath.
	pathCopy := make(xutils.PathType, len(startPath))
	copy(pathCopy, startPath)

	return &TNode{
		root:   root,
		path:   pathCopy,
		module: module,
		name:   name,
		empty:  true}
}

// We don't fully construct our expected testnodes into a tree, so we can't
// use the standard equalTo() function for XpathNodes.  Instead we do it in
// rather long-winded fashion.
func (t1 *TNode) EqualTo(n2 xutils.XpathNode) error {
	t2 := n2.(*TNode)
	if t1.name != t2.name {
		return fmt.Errorf("Node names don't match: '%s' [%s] vs '%s' [%s]",
			t1.name, t1.path, t2.name, t2.path)
	}

	if !t1.path.EqualTo(t2.path) {
		return fmt.Errorf("Node paths don't match: '%s' [%s] vs '%s' [%s]",
			t1.name, t1.path, t2.name, t2.path)
	}

	if t1.value != t2.value {
		return fmt.Errorf("Node values don't match: '%s'/'%s' vs '%s'/'%s'",
			t1.name, t1.value, t2.name, t2.value)
	}

	if t1.empty != t2.empty {
		return fmt.Errorf(
			"Node emptiness doesn't match: '%s'/'%t' vs '%s'/'%t'",
			t1.name, t1.empty, t2.name, t2.empty)
	}

	if len(t1.keys) != len(t2.keys) {
		return fmt.Errorf(
			"Nodes have different number of keys: '%s' %d vs '%s' %d",
			t1.name, len(t1.keys), t2.name, len(t2.keys))
	}
	for index, xKey := range t1.keys {
		if xKey.key != t2.keys[index].key {
			return fmt.Errorf(
				"Node key names don't match: '%s' [%s] vs '%s' [%s]",
				t1.name, xKey.key, t2.name, t2.keys[index].key)
		}
		if xKey.value != t2.keys[index].value {
			return fmt.Errorf(
				"Node key values don't match: '%s' [%s] vs '%s' [%s]",
				t1.name, xKey.value, t2.name, t2.keys[index].value)
		}
	}

	return nil
}

// If key/value pair matches, return true
func (testnode *TNode) keyMatches(key, value string) bool {
	for _, item := range testnode.keys {
		if (item.key == key) && (item.value == value) {
			return true
		}
	}
	return false
}

func (testnode *TNode) addKeyValue(key, value string) {
	testnode.keys = append(testnode.keys, xpathKey{key, value})
}

func (testnode *TNode) addChild(child *TNode) {
	child.parent = testnode
	testnode.children = append(testnode.children, child)
}

func (testnode *TNode) child(name string) *TNode {
	for _, child := range testnode.children {
		if (name == "*") || (name == child.name) {
			return child
		}
	}

	// Sort alphanumerically.  That's our defined order for everything.

	return nil
}

// Find first node that matches (ie for list entries, we're only interested
// in finding an entry with the path, not a specific one.
// Start at root
func (testnode *TNode) FindFirstNode(path xutils.PathType) *TNode {
	retNode := testnode

	if len(path) == 0 {
		return nil
	}
	if path[0] == "/" {
		path = path[1:]
	}

	for _, elem := range path {
		retNode = retNode.child(elem)
		if retNode == nil {
			return nil
		}
	}

	return retNode
}

// Takes a set of partially defined test nodes and stitches them together into
// a tree.
// Each element of a partialNode can be of different types, and is specified
// in partial nodes using the following syntax:
//
// - Container     : string
// - EphemeralCont : string$
// - List     	   : string/key+value
// - LeafList 	   : string@value
// - Leaf     	   : string+value
// - EmptyLeaf	   : string%
//
func CreateTree(t *testing.T, partialNodes []xutils.PathType) *TNode {
	tree := TNode{
		path:   xutils.PathType([]string{"/"}),
		module: TestModule,
		name:   "root",
		ntype:  Container,
	}
	tree.root = &tree

	curNode := &tree
	var newNode *TNode
	var path xutils.PathType
	var key, value string

	for _, testNode := range partialNodes {
		curNode = &tree
		for _, elem := range testNode {
			switch {
			case strings.Contains(elem, "/"):
				// Handle list entry
				elemSlashSplit := strings.Split(elem, "/")
				path = append(curNode.path, elemSlashSplit[0])
				key = strings.Split(elemSlashSplit[1], "+")[0]
				value = strings.Split(elemSlashSplit[1], "+")[1]

				// See if there's an existing entry that matches.
				newNode = nil
				for _, child := range curNode.children {
					if child.path.EqualTo(path) { // Same list
						if child.keyMatches(key, value) {
							newNode = child
							break
						}
					}
				}

				// If not, create it.
				if newNode == nil {
					newNode = NewTListEntry(
						&tree, path, TestModule, elemSlashSplit[0], key, value)
					curNode.addChild(newNode)
					// Dump tree in Test Code before processing.
				}

				// In either case, set curNode to point to the new node so
				// next step in path is dealt with at correct level.
				curNode = newNode

			case strings.Contains(elem, "+"):
				// Handle leaf
				elemToPlus := strings.Split(elem, "+")[0]
				path = append(curNode.path, elemToPlus)
				newNode = curNode.child(elemToPlus)
				if newNode == nil {
					newNode = NewTLeaf(
						&tree, path, TestModule, elemToPlus,
						strings.Split(elem, "+")[1])
					curNode.addChild(newNode)
				} else {
					// Update value.
					newNode.value = strings.Split(elem, "+")[1]
				}
				curNode = newNode

			case strings.Contains(elem, "%"):
				// Handle empty leaf
				elemToPercent := strings.Split(elem, "%")[0]
				path = append(curNode.path, elemToPercent)
				newNode = curNode.child(elemToPercent)
				if newNode == nil {
					newNode = NewTEmptyLeaf(
						&tree, path, TestModule, elemToPercent)
					curNode.addChild(newNode)
				}
				curNode = newNode

			case strings.Contains(elem, "@"):
				// Handle leaf list
				leafListName := strings.Split(elem, "@")[0]
				leafListValue := strings.Split(elem, "@")[1]
				path = append(curNode.path, leafListName)
				// Terminal element, assume no existing entry with this value.
				newNode = NewTLeafList(
					&tree, path, TestModule, leafListName, leafListValue)
				curNode.addChild(newNode)

			case strings.Contains(elem, "$"):
				// Handle ephemeral container
				elemToDollar := strings.Split(elem, "$")[0]
				path = append(curNode.path, elemToDollar)
				newNode = curNode.child(elemToDollar)
				if newNode == nil {
					newNode = NewTEphemeralContainer(
						&tree, path, TestModule, elemToDollar)
					curNode.addChild(newNode)
				}
				curNode = newNode

			default:
				// Handle other non-terminal elements
				path = append(curNode.path, elem)
				newNode = curNode.child(elem)
				if newNode == nil {
					newNode = NewTContainer(
						&tree, path, TestModule, elem)
					curNode.addChild(newNode)
				}
				curNode = newNode

			}
		}
	}

	return &tree
}
