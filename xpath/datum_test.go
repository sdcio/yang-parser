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
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This set of tests covers low level datum functions that aren't
// necessarily used much, if at all, but need to be implemented for the
// specific datum type to implement the datum interface.  Worth ensuring
// they are correct in case they do eventually get used.

package xpath

import (
	"testing"

	"github.com/sdcio/yang-parser/xpath/xpathtest"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

// This covers isSameType() as well as equalTo()
func verifyEqual(t *testing.T, d1, d2 Datum, expResult bool) {
	err := d1.equalTo(d2)
	if (err != nil) && (expResult == true) {
		t.Fatalf("Comparing %s with %s - should be equal but aren't: %s",
			d1.name(), d2.name(), err.Error())
	}
	if (err == nil) && (expResult == false) {
		t.Fatalf("Comparing %s with %s - shouldn't be equal but are.",
			d1.name(), d2.name())
	}
}

func TestDatumBoolEqualTo(t *testing.T) {
	b1 := NewBoolDatum(true)
	b2 := NewBoolDatum(true)
	b3 := NewBoolDatum(false)
	n4 := NewNumDatum(4)

	verifyEqual(t, b1, b2, true)
	verifyEqual(t, b1, b3, false)
	verifyEqual(t, b1, n4, false)
}

func TestDatumLiteralEqualTo(t *testing.T) {
	l1 := NewLiteralDatum("foo")
	l2 := NewLiteralDatum("foo")
	l3 := NewLiteralDatum("bar")
	n4 := NewNumDatum(4)

	verifyEqual(t, l1, l2, true)
	verifyEqual(t, l1, l3, false)
	verifyEqual(t, l1, n4, false)
}

func TestDatumNodesetEqualTo(t *testing.T) {
	node := xpathtest.NewTContainer(nil, xutils.PathType([]string{"foo"}),
		"testModule", "test node")
	ns1 := NewNodesetDatum([]xutils.XpathNode{})
	ns2 := NewNodesetDatum([]xutils.XpathNode{})
	ns3 := NewNodesetDatum([]xutils.XpathNode{node})
	n4 := NewNumDatum(4)

	verifyEqual(t, ns1, ns2, true)
	verifyEqual(t, ns1, ns3, false)
	verifyEqual(t, ns1, n4, false)
}

func TestDatumNumberEqualTo(t *testing.T) {
	n1 := NewNumDatum(1.5)
	n2 := NewNumDatum(1.5)
	n3 := NewNumDatum(3)
	b4 := NewBoolDatum(true)

	verifyEqual(t, n1, n2, true)
	verifyEqual(t, n1, n3, false)
	verifyEqual(t, n1, b4, false)
}
