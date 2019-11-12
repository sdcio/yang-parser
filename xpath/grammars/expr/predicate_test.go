// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// These tests verify predicate functionality for XPATH

package expr

import (
	"testing"

	"github.com/danos/yang/xpath"
	. "github.com/danos/yang/xpath/grammars/lexertest"
	"github.com/danos/yang/xpath/xpathtest"
	"github.com/danos/yang/xpath/xutils"
)

// Standard models examples for possible testing ...
//
//  must "address-family=/routing/ribs/rib[name=current()/"
//             + "rib-name]/address-family" {
//
//  must "boolean(../underlay-topology[*]/node[./supporting-nodes/node-ref])";
//
//  must "boolean(../underlay-topology/link[./supporting-link])";

// NB: I've mostly just used relative paths, starting at interface, for
//     these tests.  However, there are a couple of variations such as using
//     absolute paths, or straying out of 'dataplane[' as the relative path.
//     As the tests are primarily about what is inside the [], that seems
//     reasonable - what comes before is tested elsewhere.

// For many (all?) tests we have a standard config so encapsulate it here ...
func getConfigTree(t *testing.T) *xpathtest.TNode {
	return xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "dataplane/name+dp0s2", "address@2333"},
			{"interface", "dataplane/name+dp0s3"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
			{"protocols", "mpls", "max-label+1000000"},
			{"protocols", "mpls", "interface", "name+dp0s2"},
		})
}

func checkExprNodeSetResultWithDebug(
	t *testing.T,
	expr string,
	mapFn xpath.PfxMapFn,
	configTree *xpathtest.TNode,
	absStartPath xutils.PathType,
	expResult xpathtest.TNodeSet,
	expOut string,
) {
	mach, err := NewExprMachine(expr, mapFn)
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %s", expr, err.Error())
		return
	}
	CheckNodeSetResultWithDebug(t, mach, configTree, absStartPath, expResult,
		expOut)
}

// Basic predicate parsing - invalid, no evaluation.  Valid predicates are
// tested by rest of the tests here, with evaluation.

func TestParsePredicateUnclosed(t *testing.T) {
	errMsgs := []string{
		"Failed to compile 'foo['",
		"Parse Error: syntax error",
		"Got to approx [X] in 'foo[ [X] '",
	}
	checkParseError(t, "foo[", errMsgs)
}

// Path after ']' without intervening '/'
func TestParsePredicateMisplacedPath(t *testing.T) {
	errMsgs := []string{
		"Failed to compile 'foo[2[2]bar]'",
		"Parse Error: Nested predicates not yet supported.",
		"Got to approx [X] in 'foo[2[2]bar] [X] '",
	}
	checkParseError(t, "foo[2[2]bar]", errMsgs)
}

// Empty Predicate
func TestParseEmptyPredicate(t *testing.T) {
	errMsgs := []string{
		"Failed to compile 'foo[]'",
		"Parse Error: syntax error",
		"Got to approx [X] in 'foo[] [X] '",
	}
	checkParseError(t, "foo[]", errMsgs)
}

// Double open predicate without intervening value
func TestParsePredicateDoubleOpen(t *testing.T) {
	errMsgs := []string{
		"Failed to compile 'foo[[2]2]'",
		"Parse Error: syntax error",
		"Got to approx [X] in 'foo[[ [X] 2]2]'",
	}
	checkParseError(t, "foo[[2]2]", errMsgs)
}

// Empty nested predicate
func TestParsePredicateEmptyNested(t *testing.T) {
	t.Skipf("Needs nested predicate support.")
	errMsgs := []string{
		"Failed to compile 'foo[2[]]'",
		"Parse Error: syntax error",
		"Got to approx [X] in 'foo[2[] [X] ]'",
	}
	checkParseError(t, "foo[2[]]", errMsgs)
}

// BOOLEAN predicates

func TestPredicateBoolTrue(t *testing.T) {
	checkNodeSetResult(t, "dataplane[true()]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s3")}))
}

func TestPredicateBoolFalse(t *testing.T) {
	checkNodeSetResult(t, "dataplane[false()]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// LITERAL predicates

func TestPredicateLiteral(t *testing.T) {
	checkNodeSetResult(t, "dataplane['non-zero-length string is true']", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s3")}))
}

func TestPredicateEmptyLiteral(t *testing.T) {
	checkNodeSetResult(t, "dataplane['']", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// NUMERIC predicates (including position() and last())

func TestPredicateImplicitPosition(t *testing.T) {
	checkNodeSetResult(t, "dataplane[2]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

func TestPredicateImplicitPositionInvalid(t *testing.T) {
	checkNodeSetResult(t, "dataplane[4]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// Zero is not a valid position
func TestPredicateImplicitPositionZero(t *testing.T) {
	checkNodeSetResult(t, "dataplane[0]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

func TestPredicateImplicitPositionEmptyNodeset(t *testing.T) {
	checkNodeSetResult(t, "planedata[1]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// Now use position() explicitly.
func TestPredicatePosition(t *testing.T) {
	checkNodeSetResult(t, "dataplane[position() = 1]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1")}))
}

func TestPredicatePositionInvalid(t *testing.T) {
	checkNodeSetResult(t, "dataplane[position() = 10]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

func TestPredicatePositionEmptyNodeset(t *testing.T) {
	checkNodeSetResult(t, "planedata[position() = 1]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

func TestPredicateLast(t *testing.T) {
	checkNodeSetResult(t, "dataplane[last()]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s3")}))
}

// last() when nodeset is empty.
func TestPredicateLastEmpty(t *testing.T) {
	checkNodeSetResult(t, "planedata[last()]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// NODESET predicates

func TestPredicateNodeset(t *testing.T) {
	checkNodeSetResult(t, "*[name]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s3"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "serial"}),
				"", "serial", "name", "s1"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "loopback"}),
				"", "loopback", "name", "lo2")}))
}

func TestPredicateNodesetEmpty(t *testing.T) {
	checkNodeSetResult(t, "/interface/dataplane[unknownNodeName]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// Compound / more complex expressions
func TestPredicateChildValue(t *testing.T) {
	checkNodeSetResult(t, "../interface/dataplane[address='2222']", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

// Predicate is node name alone.  Note we only get first 2 dataplane
// interfaces as third doesn't have an address leaf.
func TestPredicateNodeName(t *testing.T) {
	checkNodeSetResult(t, "../interface/dataplane[address]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

// Check path expression can continue after predicate.
func TestPredicateMidway(t *testing.T) {
	checkNodeSetResult(t, "../interface/dataplane[name='dp0s2']/address", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2222"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2333")}))
}

// This tests a 'FilterExpr' inside the predicate
func TestPredicateChildValueContains(t *testing.T) {
	checkNodeSetResult(t,
		"/interface/dataplane[contains(name, current()/name)]",
		nil, getConfigTree(t),
		xutils.PathType([]string{"/", "protocols", "mpls", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

func TestPredicateConsecutive(t *testing.T) {
	checkNodeSetResult(t, "../interface/dataplane[2][address='2222']", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

func TestPredicateConsecutiveFirstEmpty(t *testing.T) {
	checkNodeSetResult(t, "../interface/dataplane[3][address='2222']", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

func TestPredicateConsecutiveSecondEmpty(t *testing.T) {
	checkNodeSetResult(t, "../interface/dataplane[2][address='1212']", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// Order of consecutive predicates matters
//
// dataplane[2][address = '2222']
//   -> selects dp0s2 (single node), address matches, so result is 1 node
// dataplane[address = '2222'][2]
//   -> selects dp0s2 (single node) via address match, only one node in set
//      so [2] doesn't match anything.
//
func TestPredicateConsecutiveOrderMatters(t *testing.T) {
	checkNodeSetResult(t, "../interface/dataplane[2][address='2222']", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))

	checkNodeSetResult(t, "../interface/dataplane[address='2222'][2]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

// Simple to put 2 consecutive predicates instead - unclear that nested
// predicates are essential yet ...
func TestPredicateNested(t *testing.T) {
	// For now we check we get an error, until this is supported.
	errMsgs := []string{
		"Failed to compile '../interface/dataplane[2[address='2222']]'",
		"Parse Error: Nested predicates not yet supported.",
		"Got to approx [X] in '../interface/dataplane[2[address='2222']] [X] '",
	}
	checkParseError(t, "../interface/dataplane[2[address='2222']]", errMsgs)

	// Skip test for working case until it does work!
	t.Skipf("Nested predicates not working.")
	checkNodeSetResult(t, "../interface/dataplane[2[address='2222']]", nil,
		getConfigTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

func TestPredicatePrintProgram(t *testing.T) {
	testMachine, _ := NewExprMachine("dataplane[2]", nil)

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"nameTestPush\t{ dataplane}\n" +
			"evalLocPath(PredStart)\n" +
			"evalSubMachine\n" +
			"\t--- machine start ---\n" +
			"\tnumpush		2\n" +
			"\tstore\n" +
			"\t---- machine end ----\n" +
			"evalLocPath\n" +
			"store\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

// These test predicates on a leaf list.  You can use a predicate on
// them, but you need to be careful!
func TestPredicateLeafListInvalidChild(t *testing.T) {
	checkNodeSetResult(t,
		"/interface/dataplane/address[address='1111']",
		nil, getConfigTree(t),
		xutils.PathType([]string{"/", "protocols", "mpls", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

func TestPredicateLeafListCurrent(t *testing.T) {
	checkNodeSetResult(t,
		"/interface/dataplane/address[current()='1111']",
		nil, getConfigTree(t),
		xutils.PathType([]string{"/", "protocols", "mpls", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{}))
}

func TestPredicateLeafListDot(t *testing.T) {
	checkNodeSetResult(t,
		"/interface/dataplane/address[. = '1111']",
		nil, getConfigTree(t),
		xutils.PathType([]string{"/", "protocols", "mpls", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1111")}))
}

func TestPredicateLeafListPosition(t *testing.T) {
	checkNodeSetResult(t,
		"/interface/dataplane/address[2]",
		nil, getConfigTree(t),
		xutils.PathType([]string{"/", "protocols", "mpls", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2111")}))
}

// Definitions local to this file, but common enough to extract.
const IntDP = "/interface/dataplane"
const NmD1_CR = "[name='dp0s1']\n"
const NmD2_CR = "[name='dp0s2']\n"
const NmD3_CR = "[name='dp0s3']\n"

func TestPredicatePrintExecution(t *testing.T) {
	expOut :=
		xpathtest.Run + "'dataplane[2]' on:\n" +
			"\t/interface\n" +
			xpathtest.Brk +
			xpathtest.InstNtPsh_B + " dataplane}\n" +
			xpathtest.Stack + "(empty)\n" +
			xpathtest.Brk +
			xpathtest.InstELPPS + "\n" +
			xpathtest.StNT_B + " dataplane}\n" +
			xpathtest.Brk +
			xpathtest.CrtNS + "Ctx: '/interface'\n" +
			xpathtest.T_ApNT_B + " dataplane}\n" +
			xpathtest.Tab3 + IntDP + NmD1_CR +
			xpathtest.Tab3 + IntDP + NmD2_CR +
			xpathtest.Tab3 + IntDP + NmD3_CR +
			xpathtest.Brk +
			xpathtest.InstESM + "\n" +
			xpathtest.StNS + IntDP + NmD1_CR +
			xpathtest.Tab3 + IntDP + NmD2_CR +
			xpathtest.Tab3 + IntDP + NmD3_CR +
			xpathtest.Indent(
				xpathtest.Brk+
					xpathtest.Run+"'[2]' on:\n"+
					"\t"+IntDP+NmD1_CR+
					xpathtest.Brk+
					xpathtest.InstNumPsh+"2\n"+
					xpathtest.Stack+"(empty)\n"+
					xpathtest.Brk+
					xpathtest.InstStore+"\n"+
					xpathtest.StNum+"2\n"+
					xpathtest.Brk+
					xpathtest.PredNoMatch+
					xpathtest.Brk+
					xpathtest.Run+"'[2]' on:\n"+
					"\t"+IntDP+NmD2_CR+
					xpathtest.Brk+
					xpathtest.InstNumPsh+"2\n"+
					xpathtest.Stack+"(empty)\n"+
					xpathtest.Brk+
					xpathtest.InstStore+"\n"+
					xpathtest.StNum+"2\n"+
					xpathtest.Brk+
					xpathtest.PredMatch+
					xpathtest.Brk+
					xpathtest.Run+"'[2]' on:\n"+
					"\t"+IntDP+NmD3_CR+
					xpathtest.Brk+
					xpathtest.InstNumPsh+"2\n"+
					xpathtest.Stack+"(empty)\n"+
					xpathtest.Brk+
					xpathtest.InstStore+"\n"+
					xpathtest.StNum+"2\n"+
					xpathtest.Brk+
					xpathtest.PredNoMatch+
					xpathtest.Brk) +
			xpathtest.Brk +
			xpathtest.InstELP + "\n" +
			xpathtest.StNS + IntDP + NmD2_CR +
			xpathtest.Brk +
			xpathtest.CrtNS + xpathtest.T_UNS +
			xpathtest.Tab3 + IntDP + NmD2_CR +
			xpathtest.Brk +
			xpathtest.InstStore + "\n" +
			xpathtest.StNS + IntDP + NmD2_CR +
			xpathtest.Brk

	checkExprNodeSetResultWithDebug(t, "dataplane[2]",
		nil, getConfigTree(t), xutils.PathType([]string{"/", "interface"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}),
		expOut)
}

func TestPredicatePanicInSubMachine(t *testing.T) {
	// Check error message and debug all as expected
	t.Skipf("Panic in submachine")
}
