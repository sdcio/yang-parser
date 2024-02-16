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

func TestOpdOptionMustRejected(t *testing.T) {
	schemaSnippet := `opd:option nomustallowed {
		description "no must statements allowed";
		type string;
		must "../nomustallowed";
	}`

	expected := "opd:option nomustallowed: cardinality mismatch: " +
		"invalid substatement 'must'"

	verifyExpectedFail(t, schemaSnippet, expected)
}

func TestOpdArgumentMustRejected(t *testing.T) {
	schemaSnippet := `opd:command nomustallowed {
		description "no must statements allowed";
		must "../nomustallowed";
	}`

	expected := "opd:command nomustallowed: cardinality mismatch: " +
		"invalid substatement 'must'"

	verifyExpectedFail(t, schemaSnippet, expected)
}

func TestOpdCommandMustRejected(t *testing.T) {
	schemaSnippet := `opd:command nomustallowed {
		    opd:argument nomusts {
		        description "no must statements allowed";
		        type string;
		        must "../nomusts";
		    }
	}`

	expected := "opd:argument nomusts: cardinality mismatch: " +
		"invalid substatement 'must'"

	verifyExpectedFail(t, schemaSnippet, expected)
}

func TestOpdOptionWhenRejected(t *testing.T) {
	schemaSnippet := `opd:option nomustallowed {
		description "no must statements allowed";
		type string;
		when "../nomustallowed";
	}`

	expected := "opd:option nomustallowed: cardinality mismatch: " +
		"invalid substatement 'when'"

	verifyExpectedFail(t, schemaSnippet, expected)
}

func TestOpdArgumentWhenRejected(t *testing.T) {
	schemaSnippet := `opd:command nomustallowed {
		description "no must statements allowed";
		when "../nomustallowed";
	}`

	expected := "opd:command nomustallowed: cardinality mismatch: " +
		"invalid substatement 'when'"

	verifyExpectedFail(t, schemaSnippet, expected)
}

func TestOpdCommandWhenRejected(t *testing.T) {
	schemaSnippet := `opd:command nomustallowed {
		    opd:argument nomusts {
		        description "no must statements allowed";
		        type string;
		        when "../nomusts";
		    }
	}`

	expected := "opd:argument nomusts: cardinality mismatch: " +
		"invalid substatement 'when'"

	verifyExpectedFail(t, schemaSnippet, expected)
}
