// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Wrapper functions so our test calls are a little more readable.  Some of
// the wrapped functions are currently not used outside the 'expr' grammar,
// but as they might be in future AND they take up a lot of space in this
// file, it makes sense to stick them in parsertest.go where they can be
// reused.

package expr

import (
	"testing"

	"github.com/iptecharch/yang-parser/xpath"
	. "github.com/iptecharch/yang-parser/xpath/grammars/lexertest"
	"github.com/iptecharch/yang-parser/xpath/xpathtest"
	"github.com/iptecharch/yang-parser/xpath/xutils"
)

func getMachine(
	t *testing.T,
	expr string,
	mapFn xpath.PfxMapFn,
) *xpath.Machine {

	mach, err := NewExprMachine(expr, mapFn)
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %s", expr, err.Error())
		return nil
	}

	return mach
}

func getMachineError(
	t *testing.T,
	expr string,
	mapFn xpath.PfxMapFn,
) error {
	_, err := NewExprMachine(expr, mapFn)
	return err
}

func checkNumResult(t *testing.T, expr string, expResult float64) {
	CheckNumResult(t, getMachine(t, expr, nil), expResult)
}

func checkNumResultWithContext(
	t *testing.T,
	expr string,
	expResult float64,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
) {
	CheckNumResultWithContext(t, getMachine(t, expr, nil), expResult,
		configTree, startPath)
}

func checkBoolResultWithContext(
	t *testing.T,
	expr string,
	expResult bool,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
) {
	CheckBoolResultWithContext(t, getMachine(t, expr, nil), expResult,
		configTree, startPath)
}

func checkBoolResult(t *testing.T, expr string, expResult bool) {
	CheckBoolResult(t, getMachine(t, expr, nil), expResult)
}

func checkBoolResultWithContextDebugAndMap(
	t *testing.T,
	expr string,
	expResult bool,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
	mapFn xpath.PfxMapFn,
	expOut string,
) {
	CheckBoolResultWithContextDebug(t, getMachine(t, expr, mapFn), expResult,
		configTree, startPath, expOut)
}

func checkLiteralResult(t *testing.T, expr, expResult string) {
	CheckLiteralResult(t, getMachine(t, expr, nil), expResult)
}

func checkLiteralResultWithContext(
	t *testing.T,
	expr string,
	expResult string,
	configTree *xpathtest.TNode,
	startPath xutils.PathType,
) {
	CheckLiteralResultWithContext(t, getMachine(t, expr, nil), expResult,
		configTree, startPath)
}

func checkNodeSetResult(
	t *testing.T,
	expr string,
	mapFn xpath.PfxMapFn,
	configTree *xpathtest.TNode,
	absStartPath xutils.PathType,
	expResult xpathtest.TNodeSet,
) {
	CheckNodeSetResult(t, getMachine(t, expr, mapFn), configTree,
		absStartPath, expResult)
}

func checkParseError(t *testing.T, expr string, errMsgs []string) {
	_, err := NewExprMachine(expr, nil)
	CheckParseError(t, expr, err, errMsgs)
}

func checkParseErrorWithMap(
	t *testing.T,
	expr string,
	errMsgs []string,
	mapFn xpath.PfxMapFn,
) {
	_, err := NewExprMachine(expr, mapFn)
	CheckParseError(t, expr, err, errMsgs)
}

func checkExecuteError(t *testing.T, expr string, errMsgs []string) {
	CheckExecuteError(t, getMachine(t, expr, nil), errMsgs)
}
