// Copyright (c) 2019-2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2016-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile

import (
	"github.com/steiler/yang-parser/parse"
	"github.com/steiler/yang-parser/schema"
)

/*
 * To add extensions to the schema tree a simple pattern is followed.
 * The compiler turns a set of parse nodes into a schema node. As each
 * node is compiled, it is passed to the extensions which are given an
 * opportunity to wrap (decorate) the node with any extensions. If
 * the extensions are misused then an error may be returned instead.
 */
type Extensions interface {

	// Returns a function used to get the cardinality of any extension
	NodeCardinality(parse.NodeType) map[parse.NodeType]parse.Cardinality

	// Extend the complete model set, including the combined tree of
	// all the underlying models
	ExtendModelSet(schema.ModelSet) (schema.ModelSet, error)

	// Extend the schema tree used for RPC parameters and schema Models
	ExtendModel(parse.Node, schema.Model, schema.Tree) (schema.Model, error)

	// Extend the RPC node which includes input and output parameters
	// as trees
	ExtendRpc(parse.Node, schema.Rpc) (schema.Rpc, error)

	// Extend the Notification node
	ExtendNotification(parse.Node, schema.Notification) (schema.Notification, error)

	// Extend the schema tree used for RPC parameters and schema Models
	ExtendTree(parse.Node, schema.Tree) (schema.Tree, error)

	// Extend the various nodes within a schema Tree
	ExtendContainer(parse.Node, schema.Container) (schema.Container, error)
	ExtendList(parse.Node, schema.List) (schema.List, error)
	ExtendLeaf(parse.Node, schema.Leaf) (schema.Leaf, error)
	ExtendLeafList(parse.Node, schema.LeafList) (schema.LeafList, error)

	// Extend choice and case nodes
	ExtendChoice(parse.Node, schema.Choice) (schema.Choice, error)
	ExtendCase(parse.Node, schema.Case) (schema.Case, error)

	// Extend the type, given the parse node and the base type that this
	// type is derived from. The base time may be nil, when the type is
	// a built in type.
	ExtendType(p parse.Node, base schema.Type, t schema.Type) (schema.Type, error)
	// Extend must to allow for custom must optimisations
	ExtendMust(p parse.Node, m parse.Node) (string, error)

	ExtendOpdCommand(parse.Node, schema.OpdCommand) (schema.OpdCommand, error)
	ExtendOpdOption(parse.Node, schema.OpdOption) (schema.OpdOption, error)
	ExtendOpdArgument(parse.Node, schema.OpdArgument) (schema.OpdArgument, error)
}

func (comp *Compiler) extendModelSet(m schema.ModelSet) (schema.ModelSet, error) {

	if comp.extensions == nil {
		return m, nil
	}

	return comp.extensions.ExtendModelSet(m)
}

func (comp *Compiler) extendModel(p parse.Node, m schema.Model, t schema.Tree) schema.Model {

	if comp.extensions == nil {
		return m
	}

	m2, e := comp.extensions.ExtendModel(p, m, t)
	if e != nil {
		comp.error(p, e)
		return m
	}

	return m2
}

func (comp *Compiler) extendRpc(p parse.Node, r schema.Rpc) schema.Rpc {

	if comp.extensions == nil {
		return r
	}

	r2, e := comp.extensions.ExtendRpc(p, r)
	if e != nil {
		comp.error(p, e)
		return r
	}

	return r2
}

func (comp *Compiler) extendNotification(p parse.Node, n schema.Notification) schema.Notification {
	if comp.extensions == nil {
		return n
	}

	n2, e := comp.extensions.ExtendNotification(p, n)
	if e != nil {
		comp.error(p, e)
		return n
	}

	return n2
}

func (comp *Compiler) extendTree(p parse.Node, t schema.Tree) schema.Tree {

	if comp.extensions == nil {
		return t
	}

	t2, e := comp.extensions.ExtendTree(p, t)
	if e != nil {
		comp.error(p, e)
		return t
	}

	return t2
}

func (comp *Compiler) extendContainer(
	p parse.Node, c schema.Container,
) schema.Container {

	if comp.extensions == nil {
		return c
	}

	c2, e := comp.extensions.ExtendContainer(p, c)
	if e != nil {
		comp.error(p, e)
		return c
	}

	return c2
}

func (comp *Compiler) extendList(p parse.Node, l schema.List) schema.List {

	if comp.extensions == nil {
		return l
	}

	l2, e := comp.extensions.ExtendList(p, l)
	if e != nil {
		comp.error(p, e)
		return l
	}

	return l2
}

func (comp *Compiler) extendLeaf(p parse.Node, l schema.Leaf) schema.Leaf {

	if comp.extensions == nil {
		return l
	}

	l2, e := comp.extensions.ExtendLeaf(p, l)
	if e != nil {
		comp.error(p, e)
		return l
	}

	return l2
}

func (comp *Compiler) extendLeafList(p parse.Node, l schema.LeafList) schema.LeafList {

	if comp.extensions == nil {
		return l
	}

	l2, e := comp.extensions.ExtendLeafList(p, l)
	if e != nil {
		comp.error(p, e)
		return l
	}

	return l2
}

func (comp *Compiler) extendChoice(p parse.Node, c schema.Choice) schema.Choice {

	if comp.extensions == nil {
		return c
	}

	choiceExt, e := comp.extensions.ExtendChoice(p, c)
	if e != nil {
		comp.error(p, e)
		return c
	}

	return choiceExt
}

func (comp *Compiler) extendCase(p parse.Node, c schema.Case) schema.Case {

	if comp.extensions == nil {
		return c
	}

	caseExt, e := comp.extensions.ExtendCase(p, c)
	if e != nil {
		comp.error(p, e)
		return c
	}

	return caseExt
}

func (comp *Compiler) extendType(
	p parse.Node, base schema.Type, t schema.Type,
) schema.Type {

	if comp.extensions == nil {
		return t
	}

	t2, e := comp.extensions.ExtendType(p, base, t)
	if e != nil {
		comp.error(p, e)
		return t
	}

	return t2
}

func (comp *Compiler) extendMust(
	p parse.Node, m parse.Node,
) string {

	if comp.extensions == nil {
		return ""
	}

	mustExt, e := comp.extensions.ExtendMust(p, m)
	if e != nil {
		comp.error(p, e)
		return ""
	}

	return mustExt
}

func (comp *Compiler) extendOpdCommand(
	p parse.Node, c schema.OpdCommand,
) schema.OpdCommand {
	if comp.extensions == nil {
		return c
	}

	c2, e := comp.extensions.ExtendOpdCommand(p, c)
	if e != nil {
		comp.error(p, e)
		return c
	}
	return c2
}

func (comp *Compiler) extendOpdOption(p parse.Node, o schema.OpdOption) schema.OpdOption {

	if comp.extensions == nil {
		return o
	}

	o2, e := comp.extensions.ExtendOpdOption(p, o)
	if e != nil {
		comp.error(p, e)
		return o
	}

	return o2
}

func (comp *Compiler) extendOpdArgument(p parse.Node, a schema.OpdArgument) schema.OpdArgument {

	if comp.extensions == nil {
		return a
	}

	a2, e := comp.extensions.ExtendOpdArgument(p, a)
	if e != nil {
		comp.error(p, e)
		return a
	}

	return a2
}
