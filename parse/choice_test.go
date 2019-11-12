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
