// Copyright (c) 2018-2019, AT&T Intellectual Property.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile

import (
	"fmt"

	"github.com/steiler/yang-parser/parse"
)

type deviateProcessor interface {
	isAllowed(target, property parse.Node, ec extCard) error
	propertyAction(target, property parse.Node) error
	finalAction(target, property parse.Node) error
}

type deviateBase struct{}

func (d *deviateBase) isAllowed(target, property parse.Node, ec extCard) error {
	return nil
}
func (d *deviateBase) propertyAction(target, property parse.Node) error {
	return nil
}
func (d *deviateBase) finalAction(target, property parse.Node) error {
	return nil
}

type extCard func(target, property parse.Node) rune

func (c *Compiler) getExtCardinality() extCard {
	return func(target, property parse.Node) rune {
		switch {
		case property.Type().IsExtensionNode():
			if c.extensions == nil {
				// Extensions not usually available
				// in Unit Tests
				return 'n'
			}
			nc := c.extensions.NodeCardinality
			if nc == nil {
				return 0
			}
			ec := nc(target.Type())
			if ec == nil {
				return 0
			}
			return ec[property.Type()].End

		case property.Type() == parse.NodeUnknown:
			return 'n'
		}

		return 0
	}
}

// deviateNotSupported
//
// Nothing allowed as a sub-statement except unknown extensions
type deviateNotSupported struct {
	deviateBase
}

func (n *deviateNotSupported) isAllowed(target, property parse.Node, ec extCard) error {
	// Allow unknown extensions
	if property.Type() == parse.NodeUnknown {
		// Ignore unknown extensions
		return nil
	}
	return fmt.Errorf("Property not allowed in deviate not-supported '%s'", property.Type())

}

func (n *deviateNotSupported) finalAction(target, property parse.Node) error {
	target.MarkNotSupported()
	return nil
}

// deviateDelete
//
// Properties which can be deleted are:
//
//	units
//	must
//	unique
//	default
//	known extensions
//
// A property can only be deleted from a node using deviate
// if the property appears exactly as specified
type deviateDelete struct {
	deviateBase
}

func (n *deviateDelete) isAllowed(target, property parse.Node, ec extCard) error {
	switch property.Type() {
	case parse.NodeUnits, parse.NodeDefault,
		parse.NodeMust, parse.NodeUnique:
		return nil
	default:
		if property.Type().IsExtensionNode() || property.Type() == parse.NodeUnknown {
			return nil
		}
	}
	return fmt.Errorf("Property not allowed in deviate delete '%s'", property.Type())

}

func (n *deviateDelete) propertyAction(target, property parse.Node) error {
	if property.Type() == parse.NodeUnknown {
		return nil
	}
	ch := target.LookupChild(property.Type(), property.Name())
	if ch == nil {
		return fmt.Errorf("Property being deleted by deviation must exist [%s]", property.String())
	}
	target.ReplaceChild(ch)
	return nil
}

// deviateAdd
//
// Properties which can be added are:
//
//	must
//	unique
//
// Only if not already present:
//
//	units
//	default
//	config
//	mandatory
//	min-elements
//	max-elements
//
// Additionally:
//
//	Known extensions if cardinality allows
//
// A property can only be added to a node if the property does not already exist
// or has a cardinality greater than 1
type deviateAdd struct {
	deviateBase
}

func (n *deviateAdd) isAllowed(target, property parse.Node, ec extCard) error {
	var card rune
	switch property.Type() {
	case parse.NodeUnits, parse.NodeDefault,
		parse.NodeConfig, parse.NodeMandatory,
		parse.NodeMinElements, parse.NodeMaxElements,
		parse.NodeMust, parse.NodeUnique:

		card = target.GetCardinalityEnd(property.Type())

	default:
		card = ec(target, property)
	}

	switch card {
	case '0':
		return fmt.Errorf("Property '%s' not allowed on node of type %s\n", property.Type(), target.Type())
	case '1':
		if len(target.ChildrenByType(property.Type())) != 0 {
			return fmt.Errorf("Property being added to node already exists: %s", property.Type())
		}
	case 'n':
		return nil
	default:
		return fmt.Errorf("Property '%s' not allowed on node of type %s\n", property.Type(), target.Type())
	}
	return nil

}

func (n *deviateAdd) propertyAction(target, property parse.Node) error {
	if property.Type() != parse.NodeUnknown {
		target.AddChildren(property)
	}
	return nil
}

// deviateReplace
//
// Properties which can be replaced are:
//
//	type
//	units
//	default
//	config
//	mandatory
//	min-elements
//	max-elements
//
// A property being replaced must already be present on the node
type deviateReplace struct {
	deviateBase
}

func (n *deviateReplace) isAllowed(target, property parse.Node, ec extCard) error {
	switch property.Type() {
	case parse.NodeTyp, parse.NodeUnits, parse.NodeDefault,
		parse.NodeConfig, parse.NodeMandatory,
		parse.NodeMinElements, parse.NodeMaxElements:
		return nil

	default:
		if ec(target, property) == '1' {
			// Known extensions with cardinality '1' are allowed
			return nil
		}
		return fmt.Errorf("Property not allowed in deviate replace")
	}
}

func (n *deviateReplace) propertyAction(target, property parse.Node) error {
	if property.Type() == parse.NodeUnknown {
		// Ignore unknown extensions
		return nil
	}
	ch := target.ChildrenByType(property.Type())
	if len(ch) == 0 {
		return fmt.Errorf("Only existing proprties can be replaced by deviation")
	}
	target.ReplaceChildByType(property.Type(), property)
	return nil
}

func (c *Compiler) processDeviations(module *parse.Module) {

	nod := module.GetModule()

	children := nod.ChildrenByType(parse.NodeDeviation)
	for _, a := range children {
		applyToPath := a.ArgSchema()
		applyToPfx := applyToPath[0].Space
		applyToMod, err := nod.GetModuleByPrefix(
			applyToPfx, c.modules, c.skipUnknown)
		if err != nil {
			c.error(nod, err)
		}

		allowedNodes := getAugmentableNodesForModule(applyToMod)
		applyToNode := c.getDataDescendant(
			a, allowedNodes, applyToPath, func(dst parse.Node) {})

		if applyToNode == nil {
			c.error(a, fmt.Errorf("Invalid path: %s",
				xmlPathString(applyToPath)))
		}

		devs := a.ChildrenByType(parse.NodeDeviate)

		for _, d := range devs {
			switch d.Type() {
			case parse.NodeDeviateNotSupported:
				if len(devs) > 1 {
					c.error(a, fmt.Errorf("No other deviate statements allowed with not-supported"))
				}
				c.doDeviate(applyToNode, d, &deviateNotSupported{})

			case parse.NodeDeviateDelete:
				c.doDeviate(applyToNode, d, &deviateDelete{})

			case parse.NodeDeviateAdd:
				c.doDeviate(applyToNode, d, &deviateAdd{})

			case parse.NodeDeviateReplace:
				c.doDeviate(applyToNode, d, &deviateReplace{})
			}
		}
		if len(devs) > 0 {
			c.addDeviation(applyToNode.GetNodeModulename(applyToMod), nod.Name())
		}
	}
}

func (c *Compiler) doDeviate(target, deviate parse.Node, dp deviateProcessor) {

	for _, property := range deviate.Children() {
		err := dp.isAllowed(target, property, c.getExtCardinality())
		if err != nil {
			c.error(deviate, err)
			continue
		}
		err = dp.propertyAction(target, property)
		if err != nil {
			c.error(deviate, err)
		}
	}

	dp.finalAction(target, deviate)
}
