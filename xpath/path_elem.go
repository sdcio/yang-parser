// Copyright (c) 2019,2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015,2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Implements pathElem interface and types.  These types represent path
// operations and nameTest objects used for navigating the Xpath tree.

package xpath

import (
	"encoding/xml"
	"fmt"

	"github.com/sdcio/yang-parser/xpath/xutils"
)

// pathElem - constituent parts of a path in Xpath
type pathElem interface {
	name() string
	baseString() string
	applyToNode(
		xNode xutils.XpathNode,
		matchPrefix bool,
		filter xutils.MatchType,
	) ([]xutils.XpathNode, string)
	stackable
}

func pathElemString(elems []pathElem) string {
	if len(elems) == 0 {
		return "(empty path)"
	}
	var retStr string
	var slashAdded = false
	for _, elem := range elems {
		elemStr := elem.baseString()
		retStr = retStr + elemStr
		if elemStr != "/" {
			retStr = retStr + "/"
			slashAdded = true
		}
	}

	// Need to remove trailing slash at end.  We can only have an element
	// equal to '/' at the start of a string, never part way through, so
	// if we added at least one slash we can safely just remove last char.
	if slashAdded {
		retStr = retStr[:len(retStr)-1]
	}

	return retStr
}

// nameTestElem
type nameTestElem struct {
	nameTest xml.Name
}

func newNameTestElem(nameTest xml.Name) pathElem {
	return nameTestElem{nameTest}
}

func (nt nameTestElem) name() string    { return "NAMETEST" }
func (nt nameTestElem) value() xml.Name { return nt.nameTest }

func (nt nameTestElem) baseString() string {
	return fmt.Sprintf("%s:%s", nt.nameTest.Space, nt.nameTest.Local)
}

func (nt nameTestElem) applyToNode(
	xNode xutils.XpathNode, matchPrefix bool, filter xutils.MatchType,
) ([]xutils.XpathNode, string) {

	revisedNT := nt.nameTest
	if !matchPrefix {
		revisedNT.Space = ""
	}
	return xNode.XChildren(
		xutils.NewXFilter(revisedNT, filter), xutils.Sorted), ""
}

func (nt nameTestElem) String() string {
	return fmt.Sprintf("%s\t%s", nt.name(), nt.nameTest)
}

// pathOperElem
type pathOperElem struct {
	pathOper int
}

func newPathOperElem(pathOper int) pathElem {
	return pathOperElem{pathOper}
}

func (po pathOperElem) name() string { return "PATHOPER" }

func (po pathOperElem) baseString() string {
	switch po.pathOper {
	case '.':
		return "."
	case xutils.DOTDOT:
		return ".."
	case xutils.DBLSLASH:
		return "//"
	case '/':
		return "/"
	default:
		return "(unknown)"
	}
}

func (po pathOperElem) applyToNode(
	xNode xutils.XpathNode, matchPrefix bool, filter xutils.MatchType,
) ([]xutils.XpathNode, string) {

	switch po.pathOper {
	case '.':
		if xNode.XIsEphemeral() {
			return nil, ""
		}
		return []xutils.XpathNode{xNode}, ""
	case xutils.DOTDOT:
		if newNode := xNode.XParent(); newNode == nil {
			return nil, ""
		} else {
			// Line below would return empty node not nil if newNode
			// was nil, which is not what we want!
			return []xutils.XpathNode{newNode}, ""
		}
	case xutils.DBLSLASH:
		return nil, "'//' operator not supported yet."
	case '/':
		return []xutils.XpathNode{xNode.XRoot()}, ""
	}

	return nil, "Unrecognised path operation"
}

func (po pathOperElem) String() string {
	return fmt.Sprintf("%s\t%s", po.name(), xutils.GetTokenName(po.pathOper))
}
