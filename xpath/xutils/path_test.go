// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Test for converting a XPATH path expression into an absolute path with
// no prefixes or predicates present.

package xutils_test

import (
	"testing"

	"github.com/steiler/yang-parser/xpath/xutils"
)

func checkPath(
	t *testing.T,
	expr string,
	curPath xutils.PathType,
	expPath xutils.PathType) {

	if !xutils.GetAbsPath(expr, curPath).EqualTo(expPath) {
		t.Fatalf("Wrong path.\nExp: '%s'\nGot: '%s'\n",
			expPath, xutils.GetAbsPath(expr, curPath))
	}
}

func TestAbsPath(t *testing.T) {
	expr := "/foo/bar"
	curPath := xutils.NewPathType("/some/other/leaf/value")
	expPath := xutils.NewPathType("/foo/bar")

	checkPath(t, expr, curPath, expPath)
}

func TestPredEndPath(t *testing.T) {
	expr := "/foo/bar/bar2[tagnode = current()/../foo]"
	curPath := xutils.NewPathType("/some/other/leaf/value")
	expPath := xutils.NewPathType("/foo/bar/bar2")

	checkPath(t, expr, curPath, expPath)
}

func TestPredPartWayPath(t *testing.T) {
	expr := "/foo/bar[tagnode = current()/../../foo]/bar2"
	curPath := xutils.NewPathType("/some/other/leaf/value")
	expPath := xutils.NewPathType("/foo/bar/bar2")

	checkPath(t, expr, curPath, expPath)
}

func TestRelPathCurPathTooShort(t *testing.T) {
	expr := "../../../foo/bar"
	curPath := xutils.NewPathType("/some/value")
	expPath := xutils.NewPathType("(unknown)/foo/bar")

	checkPath(t, expr, curPath, expPath)
}

func TestDeepRelPath(t *testing.T) {
	expr := "../../../foo/bar"
	curPath := xutils.NewPathType("/some/other/very/deep/leaf/value")
	expPath := xutils.NewPathType("/some/other/foo/bar")

	checkPath(t, expr, curPath, expPath)
}

func TestPrefixedPath(t *testing.T) {
	expr := "../../../pfx:foo/otherPfx:bar"
	curPath := xutils.NewPathType("/some/other/very/deep/leaf/value")
	expPath := xutils.NewPathType("/some/other/foo/bar")

	checkPath(t, expr, curPath, expPath)
}
