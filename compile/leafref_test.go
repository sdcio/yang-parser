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

// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests relating to the when and must XPATH statements.
// It checks that for all schema node types that support when/must (ie
// container, list, leaflist, leaf and choice), the statements are correctly
// compiled into executable machines.

package compile_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/sdcio/yang-parser/testutils"
)

func TestValidLeafRef(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref statement";
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type leafref {
                     path "../refLeaf";
                 }
             }
         }`))

	expected := NewLeafChecker("testLeaf")

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"leafrefContainer", "testLeaf"})
	expected.check(t, actual)
}

func TestValidRefinedLeafRef(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref typefef statement";
             typedef myRef {
                 type leafref {
                     path "/container/refLeaf";
                 }
             }
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type myRef;
             }
         }`))

	expected := NewLeafChecker("testLeaf")

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"leafrefContainer", "testLeaf"})
	expected.check(t, actual)
}

func TestInvalidRefinedLeafRef(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref typefef statement";
             typedef myRef {
                 type leafref {
                     path "/container/refLeaf";
                 }
             }
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type myRef {
                     path "../refLeaf";
                 }
             }
         }`))

	expected :=
		"schema0:20:17: type myRef: cannot refine path"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestNoPath(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref statement";
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type leafref;
             }
         }`))

	expected :=
		"schema0:15:17: type leafref: missing path"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestTwoPaths(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref statement";
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type leafref {
                     path "../refLeaf";
                     path "../refLeaf";
                 }
             }
         }`))

	expected :=
		"schema0:15:17: type leafref: cardinality mismatch: " +
			"only one 'path' statement is allowed"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestInvalidSyntax(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref statement";
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type leafref {
                     path "../refLeaf[foo =";
                 }
             }
         }`))

	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err,
		"schema0:15:17: type leafref: Failed to compile '../refLeaf[foo ='\n",
		"Parse Error: syntax error\n",
		"Got to approx [X] in '../refLeaf[foo = [X] '")
}

func TestValidPrefix(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref statement";
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type leafref {
                     path "../test:refLeaf";
                 }
             }
         }`))

	expected := NewLeafChecker("testLeaf")

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"leafrefContainer", "testLeaf"})
	expected.check(t, actual)
}

func TestInvalidPrefix(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container leafrefContainer {
		     description "Container with leafref statement";
             leaf refLeaf {
                 type string;
             }
             leaf testLeaf {
                 type leafref {
                     path "../unknown:refLeaf";
                 }
             }
         }`))

	expected := []string{
		"schema0:15:17: type leafref: Failed to compile '../unknown:refLeaf'",
		"Lexer Error: unknown import unknown",
	}
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected...)
}
