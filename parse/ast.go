// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse

import (
	"fmt"
)

type Pos int

func (p Pos) position() Pos { return p }

type newfunc func(item, string, []Node, *Scope) Node

type Node interface {
	Type() NodeType
	Statement() string
	Name() string
	String() string
	Clone(m Node) Node

	HasArgument
	Namespace

	Root() Node
	UsesRoot() Node
	Children() []Node
	ChildrenByType(t NodeType) []Node
	ChildByType(t NodeType) Node
	LookupChild(t NodeType, name string) Node

	// Needed to merge in imports and includes
	AddChildren(new ...Node)
	ReplaceChild(old Node, new ...Node)
	ReplaceChildByType(typ NodeType, new Node)

	// 'when' statements added directly under 'augment' need special handling.
	// We store them on the children, but run 'as parent'.  As we add the
	// children before we parse the 'when' statement, we have to store the
	// knowledge on the node.  As we might want this for other purposes,
	// 'AddedByAugment()' seems a reasonable name.
	AddWhenChildren(fromAugment bool, new ...Node)
	AddedByAugment() bool
	NotSupported() bool
	MarkNotSupported()
	GetCardinalityEnd(t NodeType) rune

	//
	// (agj) TODO - Combine environment types
	//
	Tenv() *TEnv
	LookupType(string) (Node, bool)
	Genv() *GEnv
	LookupGrouping(string) (Node, bool)

	// This is where file and line numbers are stored
	ErrorContext() (location, context string)

	// Shortcuts for child values. Possibly not needed
	Min() uint
	Max() uint
	OrdBy() string
	Desc() string
	Ref() string
	Presence() bool
	Keys() []string
	Def() string
	HasDef() bool
	Config() bool
	HasConfig() bool
	Units() string
	Mandatory() bool
	FracDigit() int
	Msg() string
	Cmsg() string
	Value() int
	Revision() string
	Prefix() string
	Ns() string
	Path() string
	Status() string
	OnEnter() string
	Privileged() bool
	Local() bool
	Secret() bool
	PassOpcArgs() bool

	// Internal build help
	position() Pos
	checkCardinality() error
	check() error
	buildSymbols() (error, Pos)
}

type DataDef interface {
	Node
	dataDefNode()
}

type TypeDef interface {
	Node
	typeDefNode()
}

type node struct {
	NodeType
	hasArgument

	stmt string
	Pos
	card     map[NodeType]Cardinality
	children []Node
	tenv     *TEnv
	genv     *GEnv

	tree    *Tree
	useTree *Tree

	fromAugment bool

	// Not supported, from deviation
	notSupported bool
}

func (n *node) GetCardinalityEnd(t NodeType) rune {
	return n.card[t].End
}

func (n *node) Clone(m Node) Node {
	copy := *n

	copy.children = nil
	if m != nil {
		// replace the module we are in
		copy.useTree = m.(*node).tree
	}
	for _, ch := range n.children {
		copy.children = append(copy.children, ch.Clone(m))
	}

	return &copy
}

type TypeRestriction interface {
	Node
	typeRestrictionNode()
}

func (n *node) Statement() string { return n.stmt }
func (n *node) Genv() *GEnv       { return n.genv }
func (n *node) Tenv() *TEnv       { return n.tenv }
func (n *node) String() string    { return n.stmt + " " + n.arg.String() }
func (n *node) Children() []Node  { return n.children }
func (n *node) Root() Node {
	if n.tree == nil {
		return nil
	}
	return n.tree.Root
}
func (n *node) Name() string {
	if n == nil || n.arg == nil {
		return "(unknown)"
	}
	name := n.arg.String()
	if name == "" {
		return n.stmt
	}
	// Input and Output nodes don't have an argument
	return name
}

func typesMatch(match, nt NodeType) bool {
	if match == NodeDataDef {
		return nt.IsDataNode() || nt.IsOpdDefNode()
	}
	if match == NodeOpdDef {
		return nt.IsOpdDefNode()
	}
	if match == NodeDeviate {
		return nt.IsDeviateNode()
	}
	return match == nt
}

func (n *node) getImplicitRpcChildren(match NodeType) []Node {
	ch := []Node{}
	switch match {
	case NodeInput:
		ch = append(ch,
			newNodeByType(match,
				n.tree,
				item{pos: n.Pos, val: "input"},
				"",
				nil,
				&Scope{tenv: n.tenv, genv: n.genv}))
	case NodeOutput:
		ch = append(ch,
			newNodeByType(match,
				n.tree,
				item{pos: n.Pos, val: "output"},
				"",
				nil,
				&Scope{tenv: n.tenv, genv: n.genv}))
	}
	return ch
}

func (n *node) ChildrenByType(match NodeType) []Node {

	ch := []Node{}
	for _, v := range n.Children() {
		if typesMatch(match, v.Type()) {
			ch = append(ch, v)
		}
	}
	if len(ch) == 0 && n.NodeType == NodeRpc {
		ch = n.getImplicitRpcChildren(match)
		n.AddChildren(ch...)
	}
	return ch
}

func (n *node) LookupChild(t NodeType, name string) Node {
	for _, ch := range n.ChildrenByType(t) {
		if ch.Name() == name {
			return ch
		}
	}
	return nil
}

func (n *node) ChildByType(t NodeType) Node {
	ch := n.ChildrenByType(t)
	if ch == nil || len(ch) < 1 {
		return nil
	}
	return ch[0]
}

func (n *node) exists(t NodeType) bool {
	ch := n.ChildByType(t)
	if ch == nil {
		return false
	}
	return true
}

func (n *node) optString(t NodeType) string {
	ch := n.ChildByType(t)
	if ch == nil {
		return ""
	}
	return ch.ArgString()
}

func (n *node) optStatus(t NodeType) string {
	ch := n.ChildByType(t)
	if ch == nil {
		return ""
	}
	return ch.ArgStatus()
}

func (n *node) optUri(t NodeType) string {
	ch := n.ChildByType(t)
	if ch == nil {
		return ""
	}
	return ch.ArgUri()
}

func (n *node) optPrefix(t NodeType) string {
	ch := n.ChildByType(t)
	if ch == nil {
		return ""
	}
	return ch.ArgPrefix()
}

func (n *node) optInt(t NodeType) int {
	ch := n.ChildByType(t)
	if ch == nil {
		return 0
	}
	return ch.ArgInt()
}

func (n *node) optBool(t NodeType, def bool) bool {
	ch := n.ChildByType(t)
	if ch == nil {
		return def
	}
	return ch.ArgBool()
}

func (n *node) optFracDig(t NodeType) int {
	ch := n.ChildByType(t)
	if ch == nil {
		return 0
	}
	return ch.ArgFractionDigits()
}

func (n *node) optOrdBy(t NodeType) string {
	ch := n.ChildByType(t)
	if ch == nil {
		return ""
	}
	return ch.ArgOrdBy()
}

func (n *node) optDate(t NodeType) string {
	ch := n.ChildByType(t)
	if ch == nil {
		return ""
	}
	return ch.ArgDate()
}

func (n *node) LookupFeature(name string) Node {
	ch := n.LookupChild(NodeFeature, name)
	if ch == nil {
		return nil
	}
	return ch
}

func (n *node) Presence() bool    { return n.exists(NodePresence) }
func (n *node) Revision() string  { return n.optDate(NodeRevision) }
func (n *node) Desc() string      { return n.optString(NodeDescription) }
func (n *node) Ns() string        { return n.optUri(NodeNamespace) }
func (n *node) Ref() string       { return n.optString(NodeReference) }
func (n *node) Msg() string       { return n.optString(NodeErrorMessage) }
func (n *node) Cmsg() string      { return n.optString(NodeConfigdErrMsg) }
func (n *node) Value() int        { return n.optInt(NodeValue) }
func (n *node) HasDef() bool      { return n.exists(NodeDefault) }
func (n *node) Def() string       { return n.optString(NodeDefault) }
func (n *node) HasConfig() bool   { return n.exists(NodeConfig) }
func (n *node) Config() bool      { return n.optBool(NodeConfig, true) }
func (n *node) Units() string     { return n.optString(NodeUnits) }
func (n *node) OrdBy() string     { return n.optOrdBy(NodeOrderedBy) }
func (n *node) Mandatory() bool   { return n.optBool(NodeMandatory, false) }
func (n *node) FracDigit() int    { return n.optFracDig(NodeFractionDigits) }
func (n *node) Prefix() string    { return n.optPrefix(NodePrefix) }
func (n *node) Path() string      { return n.optString(NodePath) }
func (n *node) Status() string    { return n.optStatus(NodeStatus) }
func (n *node) OnEnter() string   { return n.optString(NodeOpdOnEnter) }
func (n *node) Privileged() bool  { return n.optBool(NodeOpdPrivileged, false) }
func (n *node) Local() bool       { return n.optBool(NodeOpdLocal, false) }
func (n *node) Secret() bool      { return n.optBool(NodeOpdSecret, false) }
func (n *node) PassOpcArgs() bool { return n.optBool(NodeOpdPassOpcArgs, false) }

func (n *node) Min() uint {
	ch := n.ChildByType(NodeMinElements)
	if ch != nil {
		return ch.ArgUint()
	}
	return 0
}

func (n *node) Max() uint {
	ch := n.ChildByType(NodeMaxElements)
	if ch != nil {
		return ch.ArgMax()
	}
	return ^uint(0)
}

func (n *node) Keys() []string {
	ch := n.ChildByType(NodeKey)
	if ch == nil {
		return []string{}
	}
	return ch.ArgKey()
}

func (n *node) AddChildren(ch ...Node) {
	n.children = append(n.children, ch...)
}

func (n *node) AddWhenChildren(fromAugment bool, ch ...Node) {
	for _, child := range ch {
		child.(*node).fromAugment = fromAugment
	}
	n.children = append(n.children, ch...)
}

func (n *node) AddedByAugment() bool { return n.fromAugment }

func (n *node) MarkNotSupported() { n.notSupported = true }

func (n *node) NotSupported() bool { return n.notSupported }

func (n *node) ReplaceChildByType(typ NodeType, new Node) {
	old := n.ChildByType(typ)
	if old == nil {
		n.AddChildren(new)
	}
	n.ReplaceChild(old, new)
}

func (n *node) ReplaceChild(old Node, new ...Node) {

	fresh := []Node{}
	for _, v := range n.children {
		if v == old {
			fresh = append(fresh, new...)
		} else {
			fresh = append(fresh, v)
		}
	}
	n.children = fresh
}

func (n *node) buildSymbols() (error, Pos) {
	for _, g := range n.ChildrenByType(NodeGrouping) {
		if err := n.genv.Put(g.Name(), g); err != nil {
			return err, g.position()
		}
	}
	for _, t := range n.ChildrenByType(NodeTypedef) {
		if err := n.tenv.Put(t.Name(), t); err != nil {
			return err, t.position()
		}
	}

	for _, c := range n.children {
		if err, pos := c.buildSymbols(); err != nil {
			return err, pos
		}
	}
	return nil, Pos(0)
}
func (n *node) check() error {

	switch n.Type() {
	case NodeModule, NodeSubmodule:
		e := checkModule(n)
		if e != nil {
			return e
		}
		if e := checkRevisionOrder(n); e != nil {
			return e
		}
	}

	e := n.checkArgument()
	if e != nil {
		return e
	}
	e = n.checkCardinality()
	if e != nil {
		return e
	}
	return nil
}

func (n *node) checkCardinality() error {
	// Refine deviate nodes have varying cardinality so we let the compiler sort them.
	if (n.Type() == NodeUnknown) || (n.Type() == NodeRefine) || n.IsDeviateNode() {
		return nil
	}
	cmap := make(map[NodeType]int)
	//Sum up the node types
	for _, c := range n.children {
		cmap[c.Type()] = cmap[c.Type()] + 1

		// We may want to check cardinality in terms of 'do we have at least
		// 'n' data nodes under this node / 'no more than n'.
		if c.Type().IsDataNode() {
			cmap[NodeDataDef] = cmap[NodeDataDef] + 1
		}
	}
	//Check against cardinality table
	for k, v := range n.card {
		nt := NodeType(k)
		switch {
		case v.Start == '1' && v.End == 'n' && cmap[k] < 1:
			return fmt.Errorf("%s: missing required '%s' statement", ErrCard, nt)
		case v.Start == '1' && v.End == '1' && cmap[k] < 1:
			return fmt.Errorf("%s: missing required '%s' statement", ErrCard, nt)
		case v.End == '1' && cmap[k] > 1:
			return fmt.Errorf("%s: only one '%s' statement is allowed", ErrCard, nt)
		}
	}
	//Ensure only valid nodes
	for k, _ := range cmap {
		if _, ok := n.card[k]; k != NodeUnknown && k != NodeDataDef && !ok {
			return fmt.Errorf("%s: invalid substatement '%s'", ErrCard, NodeType(k))
		}
	}
	return nil
}

func (n *node) LookupType(s string) (Node, bool) {
	return n.tenv.Get(s)
}
func (n *node) LookupGrouping(s string) (Node, bool) {
	return n.genv.Get(s)
}

func newNodeByType(ntype NodeType, tree *Tree, id item, a string, children []Node, s *Scope) *node {
	card := make(map[NodeType]Cardinality, len(cardinalities[ntype]))

	for k, v := range yangCardinality(ntype) {
		card[k] = v
	}

	for k, v := range tree.extCard(ntype) {
		card[k] = v
	}

	node := &node{
		NodeType: ntype,
		stmt:     id.val,
		Pos:      id.pos,
		card:     card,
		children: children,
		tree:     tree,
		tenv:     s.tenv,
		genv:     s.genv,
	}

	// Composed bits need added later
	node.arg = getArgByType(ntype, a)

	return node
}

type fakeNode struct {
	node
	name string
}

func NewFakeNodeByType(extCard NodeCardinality, ntype NodeType, name string) Node {
	node := &fakeNode{
		name: name,
	}

	node.node.NodeType = ntype
	node.node.card = cardinalities[ntype]
	if extCard == nil {
		return node
	}

	for k, v := range extCard(ntype) {
		node.node.card[k] = v
	}

	return node
}

func (n *fakeNode) Name() string {
	return n.name
}

func (n *fakeNode) Statement() string {
	return "container " + n.name
}

// ErrorContext returns a textual representation of the location of the node in the input text.
func (n *node) ErrorContext() (location, context string) {
	pos := int(n.position())
	context = n.String()
	return n.tree.ErrorContextPosition(pos, context)
}
