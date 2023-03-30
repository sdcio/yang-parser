// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"testing"

	"github.com/steiler/yang-parser/schema"
)

// This returns a standard checker function that can be used from NodeChecker
func checkReqInst(expected_val bool) checkFn {
	return func(t *testing.T, actual schema.Node) {
		actual_val := actual.Type().(schema.InstanceId).Require()
		if expected_val != actual_val {
			t.Errorf("Node require-instance value does not match\n"+
				"  expect = %t\n"+
				"  actual = %t",
				expected_val, actual_val)
		}
	}
}

// Test Cases
func TestInstanceIdAccepted(t *testing.T) {

	schema_snippet := `
  leaf foo {
    type instance-identifier;
  }
`

	st := buildSchema(t, schema_snippet)
	assertLeafMatches(t, st, "foo", "instance-identifier", checkReqInst(true))
}

func TestInstanceIdRequiredAccepted(t *testing.T) {

	schema_snippet := `
  leaf foo {
    type instance-identifier {
        require-instance true;

    }
  }
`

	st := buildSchema(t, schema_snippet)
	assertLeafMatches(t, st, "foo", "instance-identifier", checkReqInst(true))
}

func TestInstanceIdNotRequiredAccepted(t *testing.T) {

	schema_snippet := `
  leaf foo {
    type instance-identifier {
        require-instance false;

    }
  }
`

	st := buildSchema(t, schema_snippet)
	assertLeafMatches(t, st, "foo", "instance-identifier", checkReqInst(false))
}

func TestInstanceIdInheritedRequiredAccepted(t *testing.T) {

	schema_snippet := `
  typedef test-type {
    type instance-identifier {
        require-instance false;

    }
  }

  leaf foo {
    type test-type;
  }

  leaf bar {
    type test-type {
      require-instance true;
    }
  }
`

	st := buildSchema(t, schema_snippet)
	assertLeafMatches(t, st, "foo", "test-type", checkReqInst(false))
	assertLeafMatches(t, st, "bar", "test-type", checkReqInst(true))
}
