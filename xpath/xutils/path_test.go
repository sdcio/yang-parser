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

	"github.com/sdcio/yang-parser/xpath/xutils"
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
