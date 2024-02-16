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
