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

// Copyright (c) 2017-2019, AT&T Intellectual Property.  All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// These tests verify the path_eval machine operation.

package compile_test

import (
	"testing"

	"fmt"

	"github.com/sdcio/yang-parser/testutils"
	"github.com/sdcio/yang-parser/xpath/xpathtest"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

const (
	noTestPath = ""
	noDebug    = ""
)

func checkWarnings(
	t *testing.T,
	warns []xutils.Warning,
	expWarnings ...xutils.Warning,
) {
	if len(warns) != len(expWarnings) {
		t.Logf("Got %d warnings, expected %d\n", len(warns), len(expWarnings))
		t.Logf("Expected:\n%v\n", expWarnings)
		t.Fatalf("Actual:\n%v\n", warns)
		return
	}

	// Seemingly there's a map lurking in the compiler causing 'warns' to
	// be generated in random order.  So, can't just pair warns and expWarnings
	// in order.
	warnsMatched := make([]bool, len(warns))
	expWarnsFound := make([]bool, len(expWarnings))

	for expIx, expWarn := range expWarnings {
		for index, warn := range warns {
			if err := warn.Match(expWarn); err == nil {
				expWarnsFound[expIx] = true
				warnsMatched[index] = true
				break
			}
		}
	}

	// Just flag up first warning not found ... we'll note the rest soon
	// enough once first is fixed!
	for index := 0; index < len(expWarnings); index++ {
		if !expWarnsFound[index] {
			t.Logf("!!!At least one expected warning not found!!!\n")
			t.Logf("Expected:\n%v\n", expWarnings[index])
			t.Logf("Actual:\n")
			for index2, warn := range warns {
				if !warnsMatched[index2] {
					t.Logf("\n%s\n", warn)
				}
			}
			t.Fatalf("Warning(s) not matched.")
			return
		}
	}
}

func verifyPathEvalSchema(t *testing.T, pathEvalSchema string) {
	_, warns, err := buildSchemaRetWarns(t, pathEvalSchema)
	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	if len(warns) != 0 {
		t.Fatalf("Unexpected warnings: %v\n", warns)
	}
}

func verifyMultiplePathEvalSchemas(
	t *testing.T,
	schemaDefs []testutils.TestSchema,
) {
	warns, err := compileMultiplePathEvalSchemas(t, schemaDefs)
	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	if len(warns) != 0 {
		t.Fatalf("Unexpected warnings: %v\n", warns)
	}
}

func compileMultiplePathEvalSchemas(
	t *testing.T,
	schemaDefs []testutils.TestSchema,
) ([]xutils.Warning, error) {

	// Create full schemas from schemaDefs in-place (replace schema with
	// fully-processed one).
	var schemas = make([][]byte, len(schemaDefs))
	for index, schemaDef := range schemaDefs {
		schemas[index] = []byte(testutils.ConstructSchema(schemaDef))
	}

	_, warns, err := testutils.GetConfigSchemaWithWarns(schemas...)
	return warns, err
}

// Basic test schema, single namespace.  Two top level containers, representing
// the 'reference' YANG we are referring to, and the 'test' YANG we are
// starting from.
const baseSchema = `
	container refCont { // Line 10 (for current SchemaTemplate)
	presence "Avoid 'reference to NP container' warnings";
	leaf refLeaf {
		type string;
	}
	list refList {
		key refListName;
		leaf refListName {
			type string;
		}
		leaf refListLeaf {
			type string;
		}
	}
	leaf-list refLeafList {
		type string;
	}
	container emptyCont {
		presence "Container allowed without content";
	}
} // Line 30`

const (
	baseSchemaLastLine = 30
)

func schemaBasePlusNStr(n int) string {
	return fmt.Sprintf("schema0:%d", baseSchemaLastLine+n)
}

// This set of tests checks we can correctly find a target that exists.  First
// test makes sure a non-existent node generates a fail so we can be more
// confident of the pass results as we get little feedback from them.

func TestAbsoluteToNonExistentNode(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must '/nonExistentCont';
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)
	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont", "/nonExistentCont", "",
			"/"+SchemaNamespace+":nonExistentCont", ""))
}

func TestAbsoluteToContainer(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteToLeaf(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont/refLeaf";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteToList(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont/refList";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteToListKey(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont/refList/refListName";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteToListLeaf(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont/refList/refListLeaf";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteToLeafList(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont/refLeafList";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteWithWildcard(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/*/refLeafList";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteWithWildcardInEmptyCont(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont/emptyCont/*";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestAbsoluteToEmptyContChild(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/refCont/emptyCont/nonExistent";
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)
	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont", "/refCont/emptyCont/nonExistent", "",
			"/"+SchemaNamespace+":refCont"+
				"/"+SchemaNamespace+":emptyCont"+
				"/"+SchemaNamespace+":nonExistent", ""))
}

// These tests check we are correctly interpreting the starting node for
// relative paths.  Destination is irrelevant here so long as it exists.
// As above, initial check is for non-existent node to raise confidence
// in tests we expect to pass.

func TestRelativeToNonExistent(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must '../nonExistentCont';
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)
	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont", "../nonExistentCont", "",
			"../"+SchemaNamespace+":nonExistentCont", ""))
}

func TestRelativeFromContainer(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "../refCont/refLeaf";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestRelativeFromLeaf(t *testing.T) {
	testSchema := `
		container testCont {
		leaf testLeaf {
			type string;
			must "../../refCont/refLeaf";
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestRelativeFromList(t *testing.T) {
	testSchema := `
		container testCont {
		list testList {
			must "../../refCont/refLeaf";
			key testListName;
			leaf testListName {
				type string;
			}
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestRelativeFromListKey(t *testing.T) {
	testSchema := `
		container testCont {
		list testList {
			key testListName;
			leaf testListName {
				type string;
				must "../../../refCont/refLeaf";
			}
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestRelativeFromListLeaf(t *testing.T) {
	testSchema := `
		container testCont {
		list testList {
			key testListName;
			leaf testListName {
				type string;
			}
			leaf testListLeaf {
				type string;
				must "../../../refCont/refLeaf";
			}
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestRelativeFromLeafList(t *testing.T) {
	testSchema := `
		container testCont {
		leaf-list testLeafList {
			type string;
			must "../../refCont/refLeaf";
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestRelativeWithMultipleDotDots(t *testing.T) {
	testSchema := `
		container testCont {
		list testList {
			key testListName;
			leaf testListName {
				type string;
			}
			leaf testListLeaf {
				type string;
				must "../../testList/../../refCont/refList/refListName/" +
					"../refListLeaf";
			}
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

// The following tests involve more complex 'must' expressions, eg involving
// not just paths, and some with functions inside (or enclosing) paths.

// Slight cheat - current() will in fact be ignored in parsing, and rest of
// path works without current().
func TestPathWithFunctionAtStart(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "current()/../refCont";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

// Unclear if we can really have a function located like this - think it's ok
// but current parser doesn't allow it?
func TestPathWithFunctionEmbedded(t *testing.T) {
	_ = `
		container testCont {
		presence "Required for test";
		must "../func()/refCont";
	}`
	t.Skipf("TBD")
}

// Right now we are ignoring paths inside predicates, but need to evaluate
// them.  Will need logic similar to current way of handling predicates, and
// think we can use the stackedNodeset concept here.
func TestPathWithPredicate(t *testing.T) {
	_ = `
		container testCont {
		presence "Required for test";
		must "../refCont/refList[name = 'foo']";
	}`
	t.Skipf("TBD")
}

func TestFunctionWithPath(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "count(/refCont/refLeaf)";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestComplexMustStatement(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "count(/refCont/refLeaf) + number(/refCont/refLeafList) > 6";
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

// Tests with prefixes.  For these we need to specify import statements and
// have multiple schemas.
func TestSimplePathWithCorrectPrefix(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must "/baseFromTest:refCont/baseFromTest:refList";
	}`

	var testSchemas = []testutils.TestSchema{
		{
			Name: testutils.NameDef{
				Namespace: "prefix-base",
				Prefix:    "base"},
			SchemaSnippet: baseSchema,
		},
		{
			Name: testutils.NameDef{
				Namespace: "prefix-test",
				Prefix:    "test"},
			Imports: []testutils.NameDef{{
				Namespace: "prefix-base",
				Prefix:    "baseFromTest"}},
			SchemaSnippet: testSchema,
		},
	}
	verifyMultiplePathEvalSchemas(t, testSchemas)
}

func TestSimplePathWithMissingPrefix(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must '/baseFromTest:refCont/refList';
	}`

	var testSchemas = []testutils.TestSchema{
		{
			Name: testutils.NameDef{
				Namespace: "prefix-base",
				Prefix:    "base"},
			SchemaSnippet: baseSchema,
		},
		{
			Name: testutils.NameDef{
				Namespace: "prefix-test",
				Prefix:    "test"},
			Imports: []testutils.NameDef{{
				Namespace: "prefix-base",
				Prefix:    "baseFromTest"}},
			SchemaSnippet: testSchema,
		},
	}
	warns, err := compileMultiplePathEvalSchemas(t, testSchemas)
	if err != nil {
		t.Fatalf("Unable to validate path: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.MissingOrWrongPrefix,
			"/testCont", "/baseFromTest:refCont/refList", "",
			"/urn:vyatta.com:test:prefix-base:refCont/"+
				"urn:vyatta.com:test:prefix-test:refList",
			""))
}

// These tests check that the post-compiler run to evaluate all PathEval
// machines does indeed run all of them.  To get explicit evidence that
// the machines have run, we put invalid paths in the must statements.
// We test with the start node being each type of node - the target node
// here is not under test (that is done above) and is in any case invalid
// to trigger test failure.
//
// These tests are NOT meant to check complex must statements are handled
// correctly - they are just intended to check correct reporting.

func TestFullSchemaFailMultiplePathsOneBadSingleNode(t *testing.T) {
	testSchema := `
		container testCont {
		presence "Required for test";
		must '../refCont/refList and ../nonExistentCont and ../refCont';
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont",
			"../refCont/refList and ../nonExistentCont and ../refCont", "",
			"../"+SchemaNamespace+":nonExistentCont", ""))
}

func TestFullSchemaFailMultiplePathsAllBadSingleNode(t *testing.T) {
	testSchema := `
		container testCont {
		leaf-list testLeafList {
			type string;
			must '../../refList and ../nonExistentCont';
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testLeafList",
			"../../refList and ../nonExistentCont", "",
			"../../"+SchemaNamespace+":refList", ""),
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testLeafList",
			"../../refList and ../nonExistentCont", "",
			"../"+SchemaNamespace+":nonExistentCont", ""))
}

func TestFullSchemaFailMultiplePathsOneBadPerNode(t *testing.T) {
	testSchema := `
		container testCont {
		list testList {
			must "../nonExistent1";
			key testListLeaf;
			leaf testListLeaf {
				type string;
				must "../../nonExistent2";
			}
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testList",
			"../nonExistent1", "",
			"../"+SchemaNamespace+":nonExistent1", ""),
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testList/testListLeaf",
			"../../nonExistent2", "",
			"../../"+SchemaNamespace+":nonExistent2", ""))
}

func TestFullSchemaFailDebug(t *testing.T) {
	testSchema := `
		container testCont {
		leaf testLeaf {
			type string;
			must "../../nonExistent or ../notThereEither";
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testLeaf",
			"../../nonExistent or ../notThereEither", schemaBasePlusNStr(4),
			"../../"+SchemaNamespace+":nonExistent",
			xpathtest.Brk+
				xpathtest.ValPath+"'/testCont/testLeaf'\n"+
				xpathtest.T_ApPO+"..\n"+
				xpathtest.Tab3+"/testCont\n"+
				xpathtest.T_ApPO+"..\n"+
				xpathtest.Tab3+"(root)\n"+
				xpathtest.T_ApNT_B+SchemaNamespace+" nonExistent}\n"+
				xpathtest.Tab3+"(empty)\n"),
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testLeaf",
			"../../nonExistent or ../notThereEither", schemaBasePlusNStr(4),
			"../"+SchemaNamespace+":notThereEither",
			xpathtest.Brk+
				xpathtest.ValPath+"'/testCont/testLeaf'\n"+
				xpathtest.T_ApPO+"..\n"+
				xpathtest.Tab3+"/testCont\n"+
				xpathtest.T_ApNT_B+SchemaNamespace+" notThereEither}\n"+
				xpathtest.Tab3+"(empty)\n"))
}

func TestFullSchemaFailMissingPrefixDebug(t *testing.T) {
	refSchema := `
		container testCont {
		leaf testLeaf {
			type string;
			must "../wrong:testLeaf2";
		}
		leaf testLeaf2 {
			type string;
		}
	}`
	wrongSchema := `
		container wrongCont {
		leaf wrongLeaf {
			type string;
		}
	}`

	var testSchemas = []testutils.TestSchema{
		{
			Name: testutils.NameDef{
				Namespace: "prefix-ref",
				Prefix:    "ref"},
			Imports: []testutils.NameDef{{
				Namespace: "prefix-wrong",
				Prefix:    "wrong"}},
			SchemaSnippet: refSchema,
		},
		{
			Name: testutils.NameDef{
				Namespace: "prefix-wrong",
				Prefix:    "wrong"},
			SchemaSnippet: wrongSchema,
		},
	}

	warns, err := compileMultiplePathEvalSchemas(t, testSchemas)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	ns := "urn:vyatta.com:test:prefix-wrong"

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.MissingOrWrongPrefix,
			"/testCont/testLeaf",
			"../wrong:testLeaf2", "schema0:25",
			"../"+ns+":testLeaf2",
			xpathtest.Brk+
				xpathtest.ValPath+"'/testCont/testLeaf'\n"+
				xpathtest.T_ApPO+"..\n"+
				xpathtest.Tab3+"/testCont\n"+
				xpathtest.T_ApNT_B+ns+" testLeaf2}\n"+
				xpathtest.Tab3+"(empty)\n"+
				xpathtest.IgnPfxs+
				xpathtest.T_ApPO+"..\n"+
				xpathtest.Tab3+"/testCont\n"+
				xpathtest.T_ApNT_B+ns+" testLeaf2}\n"+
				xpathtest.Tab3+"/testCont/testLeaf2\n",
		))
}

// Grouping tests
func TestInvalidPathInGroupingGeneratesWarningForEachUse(t *testing.T) {
	testSchema := `
	grouping testGroup {
		leaf tgLeaf {
			must "/neverValid";
			type string;
		}
	}
	container testCont {
		uses testGroup;
		container testSubCont {
			uses testGroup; // '../sometimesValid' invalid from here.
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/tgLeaf",
			"/neverValid", schemaBasePlusNStr(3),
			"/"+SchemaNamespace+":neverValid", ""),
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testSubCont/tgLeaf",
			"/neverValid", schemaBasePlusNStr(3),
			"/"+SchemaNamespace+":neverValid", ""))
}

// Test detection of 'sometimes valid' paths - eg where we have a shared
// grouping used in different locations.  Verify we suppress the warning.
func TestWhereGroupingPathSometimesValidSeparateMusts(t *testing.T) {
	testSchema := `
	grouping testGroup {
		leaf tgLeaf {
			must "/alwaysValid";
			must "/neverValid";
			must "../../alwaysValid/sometimesValid";
			type string;
		}
	}
	container alwaysValid {
		presence "For testing";
		leaf sometimesValid {
			type string;
		}
	}
	container testCont {
		uses testGroup; // '../../sometimesValid' valid from here
		container testSubCont {
			uses testGroup; // '../../sometimesValid' invalid from here.
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/tgLeaf",
			"/neverValid", schemaBasePlusNStr(4),
			"/"+SchemaNamespace+":neverValid", ""),
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testSubCont/tgLeaf",
			"/neverValid", schemaBasePlusNStr(4),
			"/"+SchemaNamespace+":neverValid", ""))
}

// Here the check is to ensure we still get the warnings for 'neverValid'
// and don't filter those out as well.
func TestWhereGroupingPathSometimesValidCompoundMust(t *testing.T) {
	testSchema := `
	grouping testGroup {
		leaf tgLeaf {
			must "/alwaysValid and /neverValid " +
				"and ../../alwaysValid/sometimesValid";
			type string;
		}
	}
	container alwaysValid {
		presence "For testing";
		leaf sometimesValid {
			type string;
		}
	}
	container testCont {
		uses testGroup; // '../../sometimesValid' valid from here
		container testSubCont {
			uses testGroup; // '../../sometimesValid' invalid from here.
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/tgLeaf",
			"/alwaysValid and /neverValid and ../../alwaysValid/sometimesValid",
			schemaBasePlusNStr(3),
			"/"+SchemaNamespace+":neverValid", ""),
		xutils.NewWarning(xutils.DoesntExist,
			"/testCont/testSubCont/tgLeaf",
			"/alwaysValid and /neverValid and ../../alwaysValid/sometimesValid",
			schemaBasePlusNStr(3),
			"/"+SchemaNamespace+":neverValid", ""))

	warns = xutils.RemoveNPContainerWarnings(warns)
	if len(warns) != 2 {
		t.Fatalf("Filtering NP warnings removed valid warnings.\n")
	}

}

// This checks that where a grouping is shared across multiple modules, any
// unprefixed paths that are validly used at least on some occasions don't
// get marked as false positives.
func TestSometimesValidWithDifferentPrefixesPass(t *testing.T) {
	mainSchema := `
		grouping mainGroup {
		leaf mainLeaf {
			type string;
			must "../../mainCont";
		}
	}
	container mainCont {
		presence "Stop non-presence warning for must statement";
		uses mainGroup;
	}`
	addSchema := `
		container addCont {
		container addSubCont {
			uses mainFromAdd:mainGroup;
		}
	}`

	var testSchemas = []testutils.TestSchema{
		{
			Name: testutils.NameDef{
				Namespace: "prefix-main",
				Prefix:    "main"},
			SchemaSnippet: mainSchema,
		},
		{
			Name: testutils.NameDef{
				Namespace: "prefix-add",
				Prefix:    "add"},
			Imports: []testutils.NameDef{{
				Namespace: "prefix-main",
				Prefix:    "mainFromAdd"}},
			SchemaSnippet: addSchema,
		},
	}
	verifyMultiplePathEvalSchemas(t, testSchemas)
	// Check what filters currently weed out, then look at what else needs
	// to be filtered out.
}

func TestSometimesValidWithDifferentPrefixesFail(t *testing.T) {
	// Check what filters currently weed out, then look at what else needs
	// to be filtered out.
}

func TestDetectionMustReferencingNPContainer(t *testing.T) {
	testSchema := `
	container presenceCont {
		presence "For testing";
		must "../nonpresenceCont";
		leaf testLeaf {
			type string;
		}
	}
	container nonpresenceCont {
		leaf npLeaf {
			type string;
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.RefNPContainer,
			"/presenceCont",
			"../nonpresenceCont", schemaBasePlusNStr(3),
			noTestPath, noDebug))

	warns = xutils.RemoveNPContainerWarnings(warns)
	if len(warns) != 0 {
		t.Fatalf("Filtering ref to NP container warning failed.\n")
	}
}

func TestNonDetectionOfMustReferencingPContainer(t *testing.T) {
	testSchema := `
	container presenceCont {
		presence "For testing";
		must "../nonpresenceCont";
		leaf testLeaf {
			type string;
		}
	}
	container nonpresenceCont {
		presence "For testing";
		leaf npLeaf {
			type string;
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

// Check we get 2 errors
func TestDetectionOfMustStatementsOnNPContainer(t *testing.T) {
	testSchema := `
		container topNPCont {
		container subNPCont {
			must "true()";
			must "npLeaf";
			leaf npLeaf {
				type string;
			}
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.MustOnNPContainer,
			"/topNPCont/subNPCont",
			"true()",
			schemaBasePlusNStr(3),
			"(n/a)", noDebug),
		xutils.NewWarning(xutils.MustOnNPContainer,
			"/topNPCont/subNPCont",
			"npLeaf",
			schemaBasePlusNStr(4),
			"(n/a)", noDebug))
}

func TestDetectionOfMustStatementOnNPContainerWithMaskedDefaults(t *testing.T) {
	testSchema := `
	container topNPCont {
		container subNPCont {
			must "true()";
			leaf npLeaf {
				type string;
			}
			container subContWithDefault {
				presence "Hides default";
				leaf subContLeaf1 {
					type string;
					default "a value";
				}
				leaf subContLeaf2 {
					type string;
					mandatory "true";
				}
			}
			list aList {
				key "name";
				leaf name {
					type string;
				}
				leaf defaultLeaf {
					default "abc";
					type string;
				}
				leaf mandatoryLeaf {
					mandatory "true";
					type string;
				}
			}
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.MustOnNPContainer,
			"/topNPCont/subNPCont",
			"true()",
			schemaBasePlusNStr(3),
			"(n/a)", noDebug))
}

// default node causes NP container to exist always, so we don't warn
// about must statement as you would hope(!) any testing would have
// revealed that must statement is always evaluated.  Flagging such nodes
// would result in a lot of noise.
func TestIgnoreMustStatementOnNPContainerWithDefault(t *testing.T) {
	testSchema := `
	container topNPCont {
		container subNPCont {
			must "true()";
			leaf npLeaf {
				type string;
				default "something";
			}
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

// Check we also detect a default on a lower level container and thus do not
// flag an error.
func TestIgnoreMustStatementOnNPContainerWithLowerLevelDefault(t *testing.T) {
	testSchema := `
	container topNPCont {
		container subNPCont {
			must "true()";
			container subSubNPCont {
				leaf npLeaf {
					type string;
					default "something";
				}
			}
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

// mandatory node causes NP container to exist always, so we don't warn about
// must statement.
func TestIgnoreMustStatementOnNPContainerWithMandatory(t *testing.T) {
	testSchema := `
	container topNPCont {
		container subNPCont {
			must "true()";
			leaf npLeaf {
				type string;
				mandatory "true";
			}
		}
	}`

	verifyPathEvalSchema(t, baseSchema+testSchema)
}

func TestDetectionOfMustStatementsOnNPContainerWithChild(t *testing.T) {
	testSchema := `
	container topNPCont {
		container subNPCont {
			must "npLeaf";
			leaf npLeaf {
				type string;
			}
			container subSubNPCont {
				must "true()";
				// Should generate NPCont warning on this node, and childNPCont
				// warning on parent node
			}
		}
	}`

	_, warns, err := buildSchemaRetWarns(t, baseSchema+testSchema)

	if err != nil {
		t.Fatalf("Failed to compile schema: %s\n", err.Error())
		return
	}

	checkWarnings(t, warns,
		xutils.NewWarning(xutils.MustOnNPContainer,
			"/topNPCont/subNPCont/subSubNPCont",
			"true()",
			schemaBasePlusNStr(8),
			"(n/a)", noDebug),
		xutils.NewWarning(xutils.MustOnNPContWithNPChild,
			"/topNPCont/subNPCont",
			"npLeaf",
			schemaBasePlusNStr(3),
			"(n/a)", noDebug))

	warns = xutils.RemoveNPContainerWarnings(warns)
	if len(warns) != 0 {
		t.Fatalf("Filtering NP container warnings failed.\n")
	}
}
