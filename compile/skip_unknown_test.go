// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"testing"

	"github.com/sdcio/yang-parser/schema"
	"github.com/sdcio/yang-parser/testutils"
)

func TestSkipUnknownTypes(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test;

		import missing { prefix gone; }

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		container one {
			leaf myLeaf{
				type gone:fishing;
			}
		}
	}`)

	expected := NewLeafChecker("myLeaf", CheckType("string"))

	st, err := testutils.GetConfigSchemaSkipUnknown(module_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	if actual := findSchemaNodeInTree(t, st,
		[]string{"one", "myLeaf"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestSkipUnknownGrouping(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test;

		import missing { prefix gone; }

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		uses gone:home;
	}`)

	_, err := testutils.GetConfigSchemaSkipUnknown(module_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}
}

func assertMissing(t *testing.T, st schema.ModelSet, path []string) {

	nodeToFind := schema.NodeSpec{
		Path: path,
	}
	actual, _, _ := st.FindOrWalk(nodeToFind, nodeMatcher, t)
	if actual != nil {
		t.Fatalf("Unexpectedly found node: %s", path)
	}
}

func TestSkipUnknownFeature(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test;

		import missing { prefix gone; }

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		leaf myLeaf {
			type string;
			if-feature gone:for-lunch;
		}
	}`)

	st, err := testutils.GetConfigSchemaSkipUnknown(module_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	assertMissing(t, st, []string{"myLeaf"})
}

func TestSkipUnknownAugment(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test;

		import missing { prefix gone; }

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		augment /gone:one {
			leaf myLeaf{
				type string;
			}
		}
	}`)

	st, err := testutils.GetConfigSchemaSkipUnknown(module_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	assertMissing(t, st, []string{"myLeaf"})
}
