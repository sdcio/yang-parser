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
