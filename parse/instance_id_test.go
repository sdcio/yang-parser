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

func TestInstanceIdAccepted(t *testing.T) {

	input := `
  leaf foo {
    type instance-identifier;
  }
`

	expected := LeafNodeChecker{
		Name: "foo",
		Typ:  "instance-identifier",
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "foo")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}

func TestInstanceIdRequiredAccepted(t *testing.T) {

	input := `
  leaf foo {
    type instance-identifier {
        require-instance true;

    }
  }
`

	expected := LeafNodeChecker{
		Name: "foo",
		Typ:  "instance-identifier",
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "foo")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}

func TestInstanceIdNotRequiredAccepted(t *testing.T) {

	input := `
  leaf foo {
    type instance-identifier {
        require-instance false;

    }
  }
`

	expected := LeafNodeChecker{
		Name: "foo",
		Typ:  "instance-identifier",
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "foo")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}
