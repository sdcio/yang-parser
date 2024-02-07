// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Implements datum interface and types, representing the 4 base data
// types in XPath (nodesets, literals, numbers and booleans).

package xpath

import (
	"bytes"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/sdcio/yang-parser/xpath/xutils"
)

// Interpreter stack type used when executing a Machine object to store
// intermediate values.

type Datum interface {
	name() string
	isSameType(d Datum) bool // Are 2 datum objects of same type?
	equalTo(d Datum) error

	// Return datum value as the type requested, if possible.  All 4
	// base types (bool, literal, number, nodeset) may be converted
	// to bool / literal / number.  Only a nodeset datum can take
	// nodeset form - you cannot convert to it.
	//
	// 'context' is a string that is added to the panic error message if
	// the conversion fails for any reason.
	Boolean(context string) bool
	Literal(context string) string
	Nodeset(context string) []xutils.XpathNode
	Number(context string) float64

	stackable
}

// Helper functions to make code elsewhere a little cleaner.
func isBool(d Datum) bool    { _, ok := d.(boolDatum); return ok }
func isLiteral(d Datum) bool { _, ok := d.(litDatum); return ok }
func isNodeset(d Datum) bool { _, ok := d.(nodesetDatum); return ok }
func isNum(d Datum) bool     { _, ok := d.(numDatum); return ok }

// For type checking the likes of return values for built-in functions.
// string provides a name to identify the type that has or hasn't been
// matched.  Purely for debug
type DatumTypeChecker func(d Datum) (bool, string)

func TypeIsBool(d Datum) (bool, string) {
	return isBool(d), "BOOL"
}

func TypeIsLiteral(d Datum) (bool, string) {
	return isLiteral(d), "LITERAL"
}

func TypeIsNodeset(d Datum) (bool, string) {
	return isNodeset(d), "NODESET"
}

func TypeIsNumber(d Datum) (bool, string) {
	return isNum(d), "NUMBER"
}

// Allow for invalidDatum here hence default case.
func TypeIsObject(d Datum) (bool, string) {
	switch d.(type) {
	case boolDatum, litDatum, nodesetDatum, numDatum:
		return true, "OBJECT"
	default:
		return false, "OBJECT"
	}
}

// Used to convert nodesets as well as strings.  Returns NaN in cases of
// error.
func numberFromString(numStr string) float64 {
	num, err := strconv.ParseFloat(strings.TrimSpace(numStr), 0)
	if err != nil {
		return math.NaN()
	}
	return num
}

// Purely for testing - allows us to exercise error handling code.
type invalidDatum struct {
}

func (i invalidDatum) name() string { return "INVALID" }

func (i invalidDatum) isSameType(d Datum) bool { return false } // Never equal!

func (b1 invalidDatum) equalTo(b2 Datum) error {
	// Invalid so never equal to anything else!
	return fmt.Errorf("Cannot compare invalid datum to anything")
}

func (i invalidDatum) Boolean(context string) bool {
	panic(fmt.Errorf("%s: Unable to convert datum to a boolean.", context))
}

func (i invalidDatum) Literal(context string) string {
	panic(fmt.Errorf("%s: Unable to convert datum to a string.", context))
}

func (i invalidDatum) Nodeset(context string) []xutils.XpathNode {
	panic(fmt.Errorf("%s: Unable to convert datum to a nodeset.", context))
}

func (i invalidDatum) Number(context string) float64 {
	panic(fmt.Errorf("%s: Unable to convert datum to a number.", context))
}

func NewInvalidDatum() Datum {
	return invalidDatum{}
}

// boolDatum
type boolDatum struct {
	boolVal bool
}

func NewBoolDatum(boolVal bool) Datum {
	return boolDatum{boolVal}
}

func (b boolDatum) name() string { return "BOOL" }

func (b boolDatum) isSameType(d Datum) bool {
	return isBool(d)
}

func (b boolDatum) String() string {
	return fmt.Sprintf("%s\t\t%t", b.name(), b.boolVal)
}

func (b1 boolDatum) equalTo(b2 Datum) error {
	if !b1.isSameType(b2) {
		return fmt.Errorf("Cannot compare boolean with %s", b2.name())
	}
	if b1.boolVal != b2.(boolDatum).boolVal {
		return fmt.Errorf("Boolean values don't match.")
	}

	return nil
}

func (b boolDatum) Boolean(context string) bool {
	return b.boolVal
}

func (b boolDatum) Literal(context string) string {
	if b.boolVal == true {
		return "true"
	}

	return "false"
}

func (b boolDatum) Nodeset(context string) []xutils.XpathNode {
	panic(fmt.Errorf("%s: Unable to convert boolean to a nodeset.", context))
}

func (b boolDatum) Number(context string) float64 {
	if b.boolVal == true {
		return 1
	}

	return 0
}

// litDatum
type litDatum struct {
	lit string
}

func NewLiteralDatum(lit string) Datum {
	return litDatum{lit}
}

func (l litDatum) name() string { return "LITERAL" }

func (l litDatum) isSameType(d Datum) bool {
	return isLiteral(d)
}

func (l litDatum) String() string {
	return fmt.Sprintf("%s\t\t%s", l.name(), l.lit)
}

func (l1 litDatum) equalTo(l2 Datum) error {
	if !l1.isSameType(l2) {
		return fmt.Errorf("Cannot compare literal with %s", l2.name())
	}
	if l1.lit != l2.(litDatum).lit {
		return fmt.Errorf("Literal values don't match: '%s' vs '%s'.",
			l1.lit, l2.(litDatum).lit)
	}

	return nil
}

func (l litDatum) Boolean(context string) bool {
	return (len(l.lit) > 0)
}

func (l litDatum) Literal(context string) string {
	return l.lit
}

func (l litDatum) Nodeset(context string) []xutils.XpathNode {
	panic(fmt.Errorf("%s: Unable to convert literal to a nodeset.", context))
}

func (l litDatum) Number(context string) float64 {
	return numberFromString(l.lit)
}

// nodesetDatum
type nodesetDatum struct {
	nodes []xutils.XpathNode
}

func NewNodesetDatum(nodes []xutils.XpathNode) Datum {
	return nodesetDatum{nodes}
}

func (ns nodesetDatum) name() string { return "NODESET" }

func (ns nodesetDatum) isSameType(d Datum) bool {
	return isNodeset(d)
}

func (ns nodesetDatum) String() (retStr string) {
	if len(ns.nodes) == 0 {
		return fmt.Sprintf("%s\t(empty)", ns.name())
	}

	var retBuf bytes.Buffer
	for index, node := range ns.nodes {
		if index == 0 {
			retBuf.WriteString(ns.name())
			retBuf.WriteString("\t\t")
		} else {
			retBuf.WriteString("\n\t\t\t")
		}
		retBuf.WriteString(xutils.NodeString(node))
	}
	return retBuf.String()
}

// Pretty-print the nodeset.  <pfx> is the indent for the (sub) machine
// as a whole, then we add 3 extra tab stops for the nodeset listing here.
func (ns nodesetDatum) Print(pfx string) (retStr string) {
	if len(ns.nodes) == 0 {
		return pfx + "\t\t\t(empty)\n"
	}

	var retBuf bytes.Buffer
	for _, node := range ns.nodes {
		retBuf.WriteString(pfx + "\t\t\t")
		if node.XRoot() == node {
			retBuf.WriteString("(root)")
		} else {
			retBuf.WriteString(xutils.NodeString(node))
		}
		retBuf.WriteString("\n")
	}
	return retBuf.String()
}

func (ns1 nodesetDatum) equalTo(ns2 Datum) error {
	if !ns1.isSameType(ns2) {
		return fmt.Errorf("Cannot compare nodeset with %s", ns2.name())
	}

	return xutils.NodesetsEqual(ns1.nodes, ns2.(nodesetDatum).nodes)
}

func (ns nodesetDatum) Boolean(context string) bool {
	return (len(ns.nodes) != 0)
}

func (ns nodesetDatum) Literal(context string) string {
	return xutils.GetStringValue(ns.nodes)
}

func (ns nodesetDatum) Nodeset(context string) []xutils.XpathNode {
	return ns.nodes
}

func (ns nodesetDatum) Number(context string) float64 {
	// Returns string-value of FIRST node in nodeset
	return numberFromString(xutils.GetStringValue(ns.nodes))
}

func (ns nodesetDatum) literalSlice() []Datum {
	var litSlice []Datum

	// If ns is empty, litVals will contain single empty string (true param).
	litVals := xutils.GetStringValues(ns.nodes, true)
	for _, lit := range litVals {
		litSlice = append(litSlice, NewLiteralDatum(lit))
	}
	return litSlice
}

// numDatum
type numDatum struct {
	num float64
}

func NewNumDatum(val float64) Datum {
	return numDatum{val}
}

func (n numDatum) name() string { return "NUMBER" }

func (n numDatum) isSameType(d Datum) bool {
	return isNum(d)
}

func (n numDatum) String() string {
	// Double tab as we 'know' we need this to align it in output.
	// TBD: better alignment using specified width print format specifier
	return fmt.Sprintf("%s\t\t%v", n.name(), n.num)
}

func (n1 numDatum) equalTo(n2 Datum) error {
	if !n1.isSameType(n2) {
		return fmt.Errorf("Cannot compare number with %s", n2.name())
	}

	if n1.num != n2.(numDatum).num {
		return fmt.Errorf("Number values don't match: %v vs %v",
			n1.num, n2.(numDatum).num)
	}

	return nil
}

func (n numDatum) Boolean(context string) bool {
	if n.num != 0 {
		return true
	}

	return false
}

func (n numDatum) Literal(context string) string {
	// A few special cases ...
	switch {
	case n.num == 0:
		// Comparison here ignores sign bit, but %v below will print it.
		// XPATH spec for string() states both +ve and -ve zero should be
		// printed as '0'.
		return "0"
	case math.IsInf(n.num, 1):
		return "Infinity"
	case math.IsInf(n.num, -1):
		return "-Infinity"
	}

	// ... then the easy ones.
	return fmt.Sprintf("%v", n.num)
}

func (n numDatum) Nodeset(context string) []xutils.XpathNode {
	panic(fmt.Errorf("%s: Unable to convert number to a nodeset.", context))
}

func (n numDatum) Number(context string) float64 {
	return n.num
}
