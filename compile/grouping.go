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

// Copyright (c) 2017-2021, AT&T Intellectual Property.  All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile

import (
	"encoding/xml"
	"fmt"

	"github.com/sdcio/yang-parser/parse"
	"github.com/sdcio/yang-parser/schema"
)

func (c *Compiler) validateModuleGroupings(m parse.Node) error {

	return c.validateGroupingsWalk(m, m)
}

func (c *Compiler) validateGroupingsWalk(m parse.Node, n parse.Node) error {

	if err := c.validateAllGroupings(m, n); err != nil {
		return err
	}

	for _, d := range n.Children() {
		if err := c.validateGroupingsWalk(m, d); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) validateAllGroupings(m parse.Node, n parse.Node) error {

	for _, g := range n.ChildrenByType(parse.NodeGrouping) {
		group_map := make(map[string]bool)
		if err := c.validateGrouping(m, g, group_map); err != nil {
			return err
		}
	}
	return nil
}

func (c *Compiler) validateGrouping(
	m parse.Node,
	g parse.Node,
	group_map map[string]bool) error {

	if _, present := group_map[g.Name()]; present {
		return fmt.Errorf("Grouping cycle detected in: grouping %s", g.Name())
	}

	group_map[g.Name()] = true
	for _, u := range g.ChildrenByType(parse.NodeUses) {
		gname := u.ArgIdRef()
		mod, err := u.GetModuleByPrefix(
			gname.Space, c.modules, c.skipUnknown)
		if err != nil {
			c.error(u, err)
		}
		if m != mod {
			// Not a local grouping so ignore it. We only have to check for
			// cycles within this module because a cross-module cycle is
			// prevented by protecting against import cycles.
			continue
		}

		ug, ok := g.LookupGrouping(gname.Local)
		if !ok {
			return fmt.Errorf(
				"Unknown grouping (grouping %s) referenced from grouping %s",
				gname.Local, g.Name())
		}

		if err := c.validateGrouping(m, ug, group_map); err != nil {
			return err
		}
	}

	return nil
}

func isMandatory(nod parse.Node) bool {
	switch nod.Type() {
	case parse.NodeLeaf, parse.NodeChoice:
		return nod.Mandatory()
	case parse.NodeLeafList, parse.NodeList:
		// List/Leaf-List is mandatory if min-elements > 0
		// A Lists children are ignored when determining if
		// it is mandatory
		return nod.Min() > 0
	case parse.NodeContainer:
		fallthrough
	default:
		// default catches such things as tree roots
		if nod.Presence() {
			// Presence on a container limits the scope
			// of mandatory nodes
			return false
		}
		for _, ch := range nod.Children() {
			if isMandatory(ch) {
				return true
			}
		}
	}

	return false
}

// Only some node types are augmentable - data (leaf, list, leaf-list and
// cont), Input and Output.  RPC is also needed here as we have to include
// nodes that may be parents of augmentable nodes.
func getAugmentableNodesForModule(applyToMod parse.Node) []parse.Node {
	allowedNodes := applyToMod.ChildrenByType(parse.NodeDataDef)
	allowedNodes = append(allowedNodes,
		applyToMod.ChildrenByType(parse.NodeCase)...)
	allowedNodes = append(allowedNodes,
		applyToMod.ChildrenByType(parse.NodeOpdDef)...)
	allowedNodes = append(allowedNodes,
		applyToMod.ChildrenByType(parse.NodeRpc)...)
	allowedNodes = append(allowedNodes,
		applyToMod.ChildrenByType(parse.NodeInput)...)
	allowedNodes = append(allowedNodes,
		applyToMod.ChildrenByType(parse.NodeOutput)...)
	return allowedNodes
}

func (c *Compiler) applyAugment(
	a parse.Node, allowedNodes []parse.Node, applyToPath []xml.Name, parentStatus schema.Status,
) {

	assertRef := func(dst parse.Node) {
		c.assertReferenceStatus(a, dst, parentStatus)
	}

	applyToNode := c.getDataDescendant(
		a, allowedNodes, applyToPath, assertRef)

	if applyToNode == nil {
		if !c.skipUnknown {
			c.error(a, fmt.Errorf("Invalid path: %s",
				xmlPathString(applyToPath)))
		}
		return
	}

	c.assertReferenceStatus(a, applyToNode, parentStatus)

	for _, ch := range a.Children() {
		if ch.Type().IsDataNode() || ch.Type().IsOpdDefNode() || ch.Type().IsExtensionNode() {
			inheritCommonProperties(a, ch, true)
			c.applyChange(a, applyToNode, ch)
		}
	}
	for _, kid := range a.Children() {
		if kid.Type() == parse.NodeUses {
			// Handle a uses within an augment which is augmenting
			// a node in a parent uses
			applyToPath := a.ArgSchema()
			applyToPfx := applyToPath[0].Space
			applyToMod, _ := kid.GetModuleByPrefix(
				applyToPfx, c.modules, c.skipUnknown)
			if err := c.expandGroupings(applyToMod, applyToNode, schema.Current); err != nil {
				c.error(applyToNode, err)
				return
			}
		}
	}
}

func (c *Compiler) expandModule(module *parse.Module) {

	nod := module.GetModule()

	// Expand Groupings
	if err := c.expandGroupings(nod, nod, schema.Current); err != nil {
		c.error(nod, err)
	}
	for _, sm := range module.GetSubmodules() {
		if err := c.expandGroupings(nod, sm, schema.Current); err != nil {
			c.error(sm, err)
		}
	}

	// Apply augments
	children := nod.ChildrenByType(parse.NodeAugment)
	children = append(children, nod.ChildrenByType(parse.NodeOpdAugment)...)
	for _, a := range children {

		if _, ok := a.Argument().(*parse.AbsoluteSchemaArg); !ok {
			c.error(a,
				fmt.Errorf("invalid argument %s expected absolute schema id",
					a.Argument().String()))
		}
		applyToPath := a.ArgSchema()
		applyToPfx := applyToPath[0].Space
		applyToMod, err := nod.GetModuleByPrefix(
			applyToPfx, c.modules, c.skipUnknown)
		if err != nil {
			c.error(nod, err)
		}
		if applyToMod != nod {
			if isMandatory(a) {
				c.error(a, fmt.Errorf("Cannot add mandatory nodes to another module: %s",
					applyToPfx))
			}
		}

		// In this mode we add paths we might need
		if c.skipUnknown {
			var nc parse.NodeCardinality
			if c.extensions != nil {
				nc = c.extensions.NodeCardinality
			}
			c.addFakePathToNode(nc, applyToMod, applyToPath)
		}
		allowedNodes := getAugmentableNodesForModule(applyToMod)
		c.applyAugment(a, allowedNodes, applyToPath, schema.Current) //AGJ
		nod.ReplaceChild(a)
	}
}

func (c *Compiler) expandGroupings(mod, nod parse.Node, parentStatus schema.Status) error {

	status := parentStatus

	if statusStatement := nod.ChildByType(parse.NodeStatus); statusStatement != nil {
		status = parseStatus(statusStatement)
	}

	// Expand any groupings found in any children before applying refines
	for _, kid := range nod.Children() {
		// If any expanded grouping contains a 'uses' at the top-level,
		// we need to expand this directly.  Otherwise we will pass the
		// 'uses' into expandGroupings (instead of as child of the node
		// passed in) and won't expand it.
		if kid.Type() == parse.NodeUses {
			if err := c.applyUsesToNode(mod, nod, kid, status); err != nil {
				return err
			}
		}
		if err := c.expandGroupings(mod, kid, status); err != nil {
			return err
		}
	}

	// Paranoia: generate error if we have failed to expand a uses statement.
	if len(nod.ChildrenByType(parse.NodeUses)) > 0 {
		panic(fmt.Errorf("Uses should be eliminated"))
	}
	return nil
}

func (c *Compiler) getNext(
	srcNode parse.Node, // See comment in function
	nods []parse.Node,
	name xml.Name,
) parse.Node {

	for _, next := range nods {
		// We need to get the namespace for 'next', and then see if our
		// augment path (name) matches up with it.
		nextNS := next.GetNodeNamespace(nil, c.modules)
		if nextNS == "" {
			continue
		}
		// Getting the namespace for 'name' (the path / node we are looking
		// for) requires a lookup using the prefix->namespace map for the
		// node that had the uses / augment statement on it, as that is the
		// correct lookup context.  This node is passed in as 'srcNode'.
		namespace, _ := srcNode.YangPrefixToNamespace(
			name.Space, c.modules, c.skipUnknown)

		if name.Local == next.Name() && (namespace == nextNS) {
			return next
		}
	}
	return nil
}

func (c *Compiler) getDataDescendant(
	srcNode parse.Node,
	nods []parse.Node, // allowed nodes that we could augment at current level
	path []xml.Name, // path we are trying to augment
	checker func(parse.Node),
) parse.Node {

	if len(path) == 0 {
		return nil
	}

	next := c.getNext(srcNode, nods, path[0])
	if next == nil {
		return nil
	}

	checker(next)

	if len(path) == 1 {
		return next
	}

	// Need to differentiate between augment and refine.
	return c.getDataDescendant(srcNode, getAugmentableNodesForModule(next),
		path[1:], checker)
}

func (c *Compiler) addFakePathToNode(
	extCard parse.NodeCardinality,
	n parse.Node,
	path []xml.Name,
) {
	if len(path) == 0 {
		return
	}

	next := c.getNext(n /* check! */, n.Children(), path[0])
	if next == nil {
		next = parse.NewFakeNodeByType(extCard, parse.NodeContainer, path[0].Local)
		n.AddChildren(next)
	}

	c.addFakePathToNode(extCard, next, path[1:])
}

func (c *Compiler) refinementIsValid(refine, applyToNode, refinement parse.Node) error {
	if refinement.Type().IsExtensionNode() {
		return nil
	}
	switch refinement.Type() {
	case parse.NodeDescription, parse.NodeReference, parse.NodeConfig,
		parse.NodeMandatory, parse.NodePresence, parse.NodeMust,
		parse.NodeDefault, parse.NodeMinElements, parse.NodeMaxElements:
		return nil
	}

	return fmt.Errorf("invalid refinement %s for statement %s",
		refinement.Type(), applyToNode.Statement())
}

// Check applyToNode is augmentable
// The target node MUST be either an opd command, option or argument
func (c *Compiler) augmentationOpdIsValid(node, ref parse.Node) error {
	switch node.Type() {
	case parse.NodeOpdCommand, parse.NodeOpdOption, parse.NodeOpdArgument:
		return nil
	default:
		return fmt.Errorf("Augment not permitted for target %s", node.Type())
	}
}

func (c *Compiler) augmentationIsValid(node, ref parse.Node) error {
	switch node.Type() {
	// Check applyToNode is augmentable
	// The target node MUST be either a container, list, choice, case, input,
	// output, or notification node.
	case parse.NodeContainer, parse.NodeList, parse.NodeChoice, parse.NodeCase,
		parse.NodeOpdCommand, parse.NodeOpdOption, parse.NodeOpdArgument,
		parse.NodeInput, parse.NodeOutput, parse.NodeNotification:

		return nil

	default:
		return fmt.Errorf("Augment not permitted for target %s", node.Type())
	}
}

func (c *Compiler) applyChange(modifier, applyToNode, refinement parse.Node) {
	switch modifier.Type() {
	case parse.NodeRefine:
		if err := c.refinementIsValid(modifier, applyToNode, refinement); err != nil {
			c.error(modifier, err)
			return
		}
	case parse.NodeAugment:
		if err := c.augmentationIsValid(applyToNode, refinement); err != nil {
			c.error(modifier, err)
			return
		}
	case parse.NodeOpdAugment:
		if err := c.augmentationIsValid(applyToNode, refinement); err != nil {
			c.error(modifier, err)
			return
		}
	default:
		c.error(modifier, fmt.Errorf("Unexpected modifier: %s", modifier.Type()))
	}

	switch applyToNode.GetCardinalityEnd(refinement.Type()) {
	case '0':
		// Skip unknown extensions
	case '1':
		applyToNode.ReplaceChildByType(refinement.Type(), refinement)
	case 'n':
		applyToNode.AddChildren(refinement)
	default:
		c.error(modifier,
			fmt.Errorf("invalid refinement %s for statement %s",
				refinement.Type(), applyToNode.Statement()))
	}
}

// Certain statements in the uses and augment apply to each child.
// We store the 'when' statements on the children (so these children can
// have 2 'when' statements despite the cardinality of 1, but they must
// be executed as if on the parent - hence the special AddWhenChildren()
// function.
func inheritCommonProperties(parent, child parse.Node, fromAugment bool) {
	// (agj) Should we use applyChange here?
	child.AddChildren(parent.ChildrenByType(parse.NodeIfFeature)...)
	child.AddWhenChildren(fromAugment, parent.ChildrenByType(parse.NodeWhen)...)
	child.AddChildren(parent.ChildrenByType(parse.NodeStatus)...)
}

func (c *Compiler) assertReferenceStatus(src, dst parse.Node, parentStatus schema.Status) {

	// Only check within the same module
	if src.Root() != dst.Root() {
		return
	}

	srcStatus := c.getStatus(src, parentStatus)
	dstStatus := c.getStatus(dst, schema.Current)

	if srcStatus < dstStatus {
		c.error(
			src,
			fmt.Errorf("%s node cannot reference %s node within same module",
				srcStatus, dstStatus))
	}
}

func (c *Compiler) refChecker(
	src parse.Node, parentStatus schema.Status,
) func(parse.Node) {
	return func(dst parse.Node) {
		c.assertReferenceStatus(src, dst, parentStatus)
	}
}

func (c *Compiler) applyUsesToNode(mod, nod, use parse.Node, parentStatus schema.Status) error {
	gname := use.ArgIdRef()

	var group parse.Node
	var ok bool

	gmod, err := use.GetModuleByPrefix(
		gname.Space, c.modules, c.skipUnknown)
	if err != nil {
		c.error(use, err)
	}
	if gmod == mod {
		// Local grouping. Search the grouping space of the local node,
		// not just the module globals. Also check for status conflicts
		group, ok = nod.LookupGrouping(gname.Local)
	} else {
		group, ok = gmod.LookupGrouping(gname.Local)
	}
	if !ok {
		if c.skipUnknown {
			// Skip unknown grouping
			nod.ReplaceChild(use)
			return nil
		}
		return fmt.Errorf(
			"Unknown grouping (grouping %s) referenced from %s",
			gname.Local, nod.Name())
	}

	var assertRef func(parse.Node)
	if use.Root() == group.Root() {
		c.assertReferenceStatus(use, group, parentStatus)
		assertRef = func(dst parse.Node) {
			c.assertReferenceStatus(use, dst, parentStatus)
		}
	} else {
		assertRef = func(parse.Node) {}
	}

	// To handle any groupings that have a uses statement as a direct
	// descendant (not grandchild) we must apply them here.  If we pass
	// them back to expandGroupings they will be ignored as that function
	// only deals with 'uses' on child nodes of the node passed in.
	for _, kid := range group.Children() {
		if kid.Type() == parse.NodeUses {
			if err := c.applyUsesToNode(gmod, group, kid, parentStatus); err != nil {
				return err
			}
		}
	}

	// Clone the children of the group, apply the refine statements and
	// then replace the uses node with the refined children
	// Replace once refines are done here. This preserves order.
	//
	// When dealing with 'uses' in a submodule, we need to ensure that the
	// cloned 'kid' is associated with the submodule rather than the parent
	// module.
	kidmod := mod
	if ur := use.Root(); ur != nil && ur.Type() == parse.NodeSubmodule {
		kidmod = c.submodules[ur.Name()].GetModule()
	}

	refinedNodes := []parse.Node{}
	for _, kid := range group.Children() {
		newKid := kid.Clone(kidmod)
		inheritCommonProperties(use, newKid, false)

		// Deal with 'double' forward reference of grouping where first
		// forward referenced grouping contains a second forward reference
		// that is not at top level of grouping (that scenario is dealt with
		// in expandGroupings())
		if err := c.expandGroupings(gmod, newKid, schema.Current); err != nil {
			c.error(newKid, err)
		}
		refinedNodes = append(refinedNodes, newKid)
	}

	for _, r := range use.ChildrenByType(parse.NodeRefine) {

		applyToPath := r.ArgDescendantSchema()
		applyToNode := c.getDataDescendant(
			use, refinedNodes, applyToPath, assertRef)
		if applyToNode == nil {
			c.error(r, fmt.Errorf("Invalid path: %s", xmlPathString(applyToPath)))
		}

		// Special case for begin and end. If we are applying a new configd:begin/end
		// then get rid of the old ones. We don't do this for create/update/delete.
		for _, ch := range r.Children() {
			if ch.Type().IsExtensionNode() {
				remove := applyToNode.ChildrenByType(ch.Type())
				for _, rm := range remove {
					applyToNode.ReplaceChild(rm)
				}
			}
		}
		for _, ch := range r.Children() {
			c.applyChange(r, applyToNode, ch)
		}
	}

	status := parentStatus
	if st := use.ChildByType(parse.NodeStatus); st != nil {
		status = parseStatus(st)
	}
	for _, a := range use.ChildrenByType(parse.NodeAugment) {

		applyToPath := a.ArgDescendantSchema()
		c.applyAugment(a, refinedNodes, applyToPath, status)
	}
	for _, a := range use.ChildrenByType(parse.NodeOpdAugment) {
		applyToPath := a.ArgDescendantSchema()
		c.applyAugment(a, refinedNodes, applyToPath, status)
	}

	nod.ReplaceChild(use, refinedNodes...)
	return nil
}
