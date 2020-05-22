// Copyright (c) 2020, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Testing of the opd:option statement.

package parse_test

import (
	"testing"
)

func TestOpdOptionMissingTypeRejected(t *testing.T) {
	schemaSnippet := `opd:option missingtype {
		description "opd:option is missing a type";
	}`

	expected := "opd:option missingtype: cardinality mismatch: " +
		"missing required 'type' statement"

	verifyExpectedFail(t, schemaSnippet, expected)

}
