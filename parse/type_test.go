// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Testing on the type statement ... note that the list of supported
// substatements in the original RFC (6020) is wrong, and has been
// updated in the errata.

package parse_test

import (
	"testing"
)

func TestTypeDescriptionRejected(t *testing.T) {

	schemaSnippet := `type string {
		description "not allowed";
	}`

	expected := "cardinality mismatch: invalid substatement 'description'"

	verifyExpectedFail(t, schemaSnippet, expected)
}
