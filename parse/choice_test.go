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
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse_test

import (
	"testing"
)

func TestChoiceAccepted(t *testing.T) {

	input := `choice options {
        case option1 {
            leaf choice1 {
                type boolean;
                description "Option 1";
            }
        }
        case option2 {
            leaf choice2 {
                type boolean;
                description "Option 2";
            }
        }
	}`
	expected := ChoiceNodeChecker{
		Name: "options",
		Cases: []CaseNodeChecker{
			CaseNodeChecker{
				Name: "option1",
				Body: []LeafNodeChecker{LeafNodeChecker{Name: "choice1"}}},
			CaseNodeChecker{
				Name: "option2",
				Body: []LeafNodeChecker{LeafNodeChecker{Name: "choice2"}}},
		},
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "choice options")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}

func TestChoiceAcceptedImplicitCase(t *testing.T) {

	t.Skip("This is known to fail right now")

	/*
	 * If the "case" statement is missing, an implicit
	 * one is added with the leaf name
	 */
	input := `choice options {
        case option1 {
            leaf choice1 {
                type boolean;
                description "Option 1";
            }
        }
        leaf choice2 {
            type boolean;
            description "Option 2";
        }
	}`

	expected := ChoiceNodeChecker{
		Name: "options",
		Cases: []CaseNodeChecker{
			CaseNodeChecker{
				Name: "option1",
				Body: []LeafNodeChecker{LeafNodeChecker{Name: "choice1"}}},
			CaseNodeChecker{
				Name: "choice2",
				Body: []LeafNodeChecker{LeafNodeChecker{Name: "choice2"}}},
		},
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "choice options")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}
