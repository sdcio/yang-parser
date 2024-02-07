// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains helper functions for testing parsing of XPATH
// expressions from the different grammars used.

package lexertest

import (
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/sdcio/yang-parser/testutils/assert"
	"github.com/sdcio/yang-parser/xpath"
	"github.com/sdcio/yang-parser/xpath/xpathtest"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

func CheckNumResult(t *testing.T, mach *xpath.Machine, expResult float64) {
	checkNumResultInternal(t, mach, expResult, nil)
}

func CheckNumResultWithContext(
	t *testing.T,
	mach *xpath.Machine,
	expResult float64,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
) {
	xNode := configTree.FindFirstNode(startPath)
	if xNode == nil {
		t.Fatalf("Unable to find node for path: %s\n", startPath)
		return
	}
	checkNumResultInternal(t, mach, expResult, xNode)
}

func checkNumResultInternal(
	t *testing.T,
	mach *xpath.Machine,
	expResult float64,
	currentNode xutils.XpathNode,
) {
	res := xpath.NewCtxFromMach(mach, currentNode).EnableValidation().Run()
	actResult, err := res.GetNumResult()
	if err != nil {
		t.Log(mach.PrintMachine())
		t.Fatalf("Unexpected error getting number result for %s: %s\n",
			mach.GetExpr(), err.Error())
		return
	}

	// Round off 'value' to 5dp.  Otherwise we get silly problems due to
	// rounding when using complex maths.
	actResult = math.Trunc(actResult*100000) / 100000
	if actResult != expResult {
		// Double check in case both are NaN ...
		if math.IsNaN(actResult) && math.IsNaN(expResult) {
			return
		}
		t.Fatalf("Wrong result for '%s': exp '%v', got '%v'",
			mach.GetExpr(), expResult, actResult)
		return
	}
}

func CheckBoolResult(t *testing.T, mach *xpath.Machine, expResult bool) {
	checkBoolResultInternal(t, mach, expResult, nil, false, "")
}

func CheckBoolResultWithContextDebug(
	t *testing.T,
	mach *xpath.Machine,
	expResult bool,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
	expOut string,
) {
	xNode := configTree.FindFirstNode(startPath)
	if xNode == nil {
		t.Fatalf("Unable to find node for path: %s\n", startPath)
		return
	}
	checkBoolResultInternal(t, mach, expResult, xNode, true, expOut)
}

func CheckBoolResultWithContext(
	t *testing.T,
	mach *xpath.Machine,
	expResult bool,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
) {
	xNode := configTree.FindFirstNode(startPath)
	if xNode == nil {
		t.Fatalf("Unable to find node for path: %s\n", startPath)
		return
	}
	checkBoolResultInternal(t, mach, expResult, xNode, false, "")
}

func checkBoolResultInternal(
	t *testing.T,
	mach *xpath.Machine,
	expResult bool,
	currentNode xutils.XpathNode,
	debug bool,
	expOut string,
) {
	res := xpath.NewCtxFromMach(mach, currentNode).
		EnableValidation().SetDebug(debug).Run()
	actResult, err := res.GetBoolResult()
	if err != nil {
		t.Logf(mach.PrintMachine())
		t.Fatalf("Unexpected error getting boolean result for %s: %s\n",
			mach.GetExpr(), err.Error())
		return
	}

	if actResult != expResult {
		t.Fatalf("Wrong bool result for '%s': exp '%v', got '%v'",
			mach.GetExpr(), expResult, actResult)
		return
	}

	checkDebugDivergence(t, debug, expOut, res)
}

func CheckLiteralResult(t *testing.T, mach *xpath.Machine, expResult string) {
	checkLiteralResultInternal(t, mach, expResult, nil)
}

func CheckLiteralResultWithContext(
	t *testing.T,
	mach *xpath.Machine,
	expResult string,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
) {
	xNode := configTree.FindFirstNode(startPath)
	if xNode == nil {
		t.Fatalf("Unable to find node for path: %s\n", startPath)
		return
	}
	checkLiteralResultInternal(t, mach, expResult, xNode)
}

func checkLiteralResultInternal(
	t *testing.T,
	mach *xpath.Machine,
	expResult string,
	currentNode xutils.XpathNode,
) {
	res := xpath.NewCtxFromMach(mach, currentNode).EnableValidation().Run()
	actResult, err := res.GetLiteralResult()
	if err != nil {
		t.Logf(mach.PrintMachine())
		t.Fatalf("Unexpected error getting literal result for %s: %s\n",
			mach.GetExpr(), err.Error())
		return
	}

	if actResult != expResult {
		t.Fatalf("Wrong literal result for '%s': exp '%v', got '%v'",
			mach.GetExpr(), expResult, actResult)
		return
	}
}

func CheckParseError(
	t *testing.T,
	expr string,
	err error,
	expErrMsgs []string,
) {
	if len(expErrMsgs) == 0 {
		t.Fatalf("Must specify at least one expected error message!")
		return
	}
	if err == nil {
		t.Fatalf("Parsing '%s' succeeded, but should have failed with:\n%s\n",
			expr, expErrMsgs)
		return
	}
	for _, msg := range expErrMsgs {
		if len(msg) == 0 {
			t.Fatalf("Cannot have empty expected error message!")
			return
		}
		if !strings.Contains(err.Error(), msg) {
			t.Fatalf("Wrong result for '%s':\nExp '%s'\nGot '%s'",
				expr, msg, err.Error())
			return
		}
	}
}

func CheckExecuteError(t *testing.T, mach *xpath.Machine, expErrMsgs []string) {
	res := xpath.NewCtxFromMach(mach, nil).EnableValidation().Run()
	err := res.GetError()
	if err == nil {
		t.Fatalf("Unexpected success getting result for %s: %s\n",
			mach.GetExpr(), expErrMsgs)
		return
	}

	if len(expErrMsgs) == 0 {
		t.Fatalf("Must specify at least one expected error message!")
		return
	}
	for _, msg := range expErrMsgs {
		if len(msg) == 0 {
			t.Fatalf("Cannot have empty expected error message!")
			return
		}
		if !strings.Contains(err.Error(), msg) {
			t.Fatalf("Wrong result for '%s': exp '%s', got '%s'",
				mach.GetExpr(), msg, err.Error())
			return
		}
	}
}

// If paths are the same, then the nodes are deemed identical as you cannot
// have 2 different nodes with the same path.
func testNodesetsEqual(ns1, ns2 []xutils.XpathNode) error {
	// First check we have same number of nodes in each set ...
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

	// ... then check each pair of nodes matches.
	for index, n1 := range ns1 {
		if err := n1.(*xpathtest.TNode).EqualTo(ns2[index]); err != nil {
			return err
		}
	}

	return nil
}

func CheckNodeSetResultWithDebug(
	t *testing.T,
	mach *xpath.Machine,
	configTree *xpathtest.TNode,
	absStartPath xutils.PathType,
	expResult xpathtest.TNodeSet,
	expOut string,
) {
	checkNodeSetResultInternal(t, mach, configTree, absStartPath, expResult,
		true, expOut)
}

func CheckNodeSetResult(
	t *testing.T,
	mach *xpath.Machine,
	configTree *xpathtest.TNode,
	absStartPath xutils.PathType,
	expResult xpathtest.TNodeSet,
) {
	checkNodeSetResultInternal(t, mach, configTree, absStartPath, expResult,
		false, "")
}

func checkNodeSetResultInternal(
	t *testing.T,
	mach *xpath.Machine,
	configTree *xpathtest.TNode,
	absStartPath xutils.PathType,
	expResult xpathtest.TNodeSet,
	debug bool,
	expOut string,
) {
	xNode := configTree.FindFirstNode(absStartPath)
	if xNode == nil {
		t.Fatalf("Unable to find testnode for path: %s\n", absStartPath)
		return
	}

	res := xpath.NewCtxFromMach(mach, xNode).
		EnableValidation().SetDebug(debug).Run()
	actResult, err := res.GetNodeSetResult()
	if err != nil {
		t.Logf(mach.PrintMachine())
		t.Fatalf("Unexpected error getting nodeset result for %s: %s\n",
			mach.GetExpr(), err.Error())
		return
	}

	// As xpathNodesetsEqual takes slices of XpathNodes, we have to 'convert'
	// our slice first to match.
	expSlice := make([]xutils.XpathNode, len(expResult))
	for index, exp := range expResult {
		expSlice[index] = exp
	}

	if err := testNodesetsEqual(expSlice, actResult); err != nil {
		t.Logf(mach.PrintMachine())
		t.Fatalf("Wrong nodeset result for '%s': %s",
			mach.GetExpr(), err.Error())
		return
	}

	checkDebugDivergence(t, debug, expOut, res)
}

// We may have a large debug string.  So, rather than just dump all
// of expected and actual, we dump actual up to the point of divergence,
// then flag the precise point of divergence and then show the next
// 10 characters of expected and actual.
//
// If your expected output now diverges, you'll get something
// like the following, making it way easier to see what's up!
//
//		--- FAIL: TestStackPrint (0.00s)
//		parsertest.go:350: Unexpected output.
//			Got:
//			Run	'substring('3456',1,2) + ../serial/name' on:
//				/interface/dataplane###[name=dp0s
//			Exp at ###:
//			' [name=dp0 ...'
//	FAIL
//	FAIL	yang/xpath/grammars/expr	0.099s
func checkDebugDivergence(
	t *testing.T,
	dbg bool,
	expOut string,
	res *xpath.Result,
) {
	if !dbg {
		return
	}

	assert.CheckStringDivergence(t, expOut, res.GetDebugOutput())
}
