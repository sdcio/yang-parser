// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"testing"
)

//
//  Test Cases
//
func TestIdentitySuccessSimple(t *testing.T) {
	schema_snippet := `
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
	st := buildSchema(t, schema_snippet)
	assertLeafMatches(t, st, "foo", "identityref")
}
