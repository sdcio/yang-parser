// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse_test

import (
	"testing"
)

func TestStatusDeprecated(t *testing.T) {
	input := `container testContainer {
            leaf testLeaf {
                type string;
                status deprecated;
            }
        }`

	expected := ContainerNodeChecker{
		Name: "testContainer",
		Body: []LeafNodeChecker{
			LeafNodeChecker{Name: "testLeaf", Status: "deprecated"},
		},
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "testContainer")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}

func TestStatusObsolete(t *testing.T) {
	input := `container testContainer {
            leaf testLeaf {
                type string;
                status obsolete;
            }
        }`

	expected := ContainerNodeChecker{
		Name: "testContainer",
		Body: []LeafNodeChecker{
			LeafNodeChecker{Name: "testLeaf", Status: "obsolete"},
		},
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "testContainer")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}
