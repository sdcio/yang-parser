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

func TestIdentityAccepted(t *testing.T) {

	input := `
  identity schema-format {
    description
      "Base identity for data model schema languages.";
  }

  identity xsd {
    base schema-format;
  }

  identity yang {
    base schema-format;
  }

  identity yin {
    base schema-format;
  }

  leaf foo {
    type identityref {
        base schema-format;
    }
  }
`

	expected := LeafNodeChecker{
		Name: "foo",
		Typ:  "identityref",
	}

	tree := createParseTreeFromYang(t, input)
	actual := findNode(tree, "foo")
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}
