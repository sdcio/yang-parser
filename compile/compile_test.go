// Copyright (c) 2017,2019 by AT&T Intellectual Property
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/danos/yang/schema"
	"github.com/danos/yang/schema/schematests"
	"github.com/danos/yang/testutils"
	"github.com/danos/yang/xpath/xutils"
)

const SchemaNamespace = "urn:vyatta.com:test:yang-compile"

// Schema Template with '%s' at end for insertion of schema for each test.
const SchemaTemplate = `
module test-yang-compile {
	namespace "urn:vyatta.com:test:yang-compile";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema";
	}
	%s
}
`

// Templates
//
// These allow the text in each test case to be greatly reduced by
// extracting the common patterns.

var BlankTemplate string = "%s"

var ContainerTemplate string = `
container testContainer {
	%s
}`

var ListTemplate string = `
container testContainer {
	list testList {
		key "testKey";
		%s
	}
}`

var ListTemplateNested string = `
	list testList2 {
        key "listContainer/uniqueLeaf";
        %s
		container listContainer {
			leaf uniqueLeaf {
				type string;
			}
		}
        leaf someLeaf {
            type empty;
        }
        leaf keyLeaf {
            type string;
        }
	}`

var ListRefineTemplate string = `
	grouping target {
    container test_container {
		list testList {
            key listKey;
            leaf listKey {
                type string;
            }
            leaf leaf1 {
                type string;
            }
            leaf leaf2 {
                type string;
            }
        }
    }
}
%s
`

var LeafTemplate string = `
	container testContainer {
		leaf testLeaf {
			%s
		}
	}
	`

var LeafRefineTemplate string = `
grouping target {
	    container test_container {
		leaf test_uint8_def {
			type uint8;
		    default "99";
		}
		leaf test_string_mandatory {
			type string;
			mandatory "true";
		}
	}
}

container test_target {
	uses target {
    %s
	}
}`

var DefaultTemplate string = `
	container testContainer {
		typedef uint8_base_default {
			type uint8;
			default "99";
		}
		typedef uint8_base_range {
			type uint8 {
				range "1..100 | 200 .. 255";
			}
		}
		typedef uint8_base_range_and_default {
			type uint8 {
				range "1..100 | 200 .. 255";
			}
			default "99";
		}
		typedef string_base_length_and_default {
			type string {
				length "0..10";
			}
			default "a string";
		}
	%s
}`

var BooleanDefaultTemplate string = `
	container testContainer {
		typedef boolean_default {
			type boolean;
			%s
		}
		leaf testLeaf {
			type boolean_default;
		}
}`

var EnumDefaultTemplate string = `
	container testContainer {
		typedef enum_test {
			type enumeration {
			    enum foo;
			    enum foo2;
			    enum bar;
            }
		}
		leaf testLeaf {
			type enum_test;
			%s
		}
}`

var EnumSingleValueTemplate string = `
	container testContainer {
		typedef enum_test_sgl_val {
			type enumeration {
			    enum fool;
            }
		}
		leaf testLeaf {
			type enum_test_sgl_val;
			%s
		}
}`

// Allows for both PASS and FAIL cases to be specified with minimal text.
var TypedefTemplate = `
container testContainer {
	typedef base_dec64_with_range {
		type decimal64 {
			fraction-digits 4;
			range "1 .. 10 | 11 .. 11 | 12 | 13 .. 20 | 31 .. 40 | 51 .. 60";
		}
	}
	typedef base_int_with_range {
		type int32 {
			range "1 .. 10 | 11 .. 11 | 12 | 13 .. 20 | 31 .. 40 | 51 .. 60";
		}
	}
	typedef base_string_with_range {
		type string {
			length "0 .. 10 | 11 .. 11 | 12 | 13 .. 20 | 31 .. 40 | 51 .. 60";
		}
	}
	typedef base_uint_with_range {
		type uint32 {
			range "1 .. 10 | 11 .. 11 | 12 | 13 .. 20 | 31 .. 40 | 51 .. 60";
		}
	}
	typedef derived_with_range {
		type %s
	}
	leaf testleaf {
		type derived_with_range;
	}
}`

var GroupingTemplate = `
grouping target {
	leaf test_uint8 {
		type uint8;
	}
	leaf test_int64 {
		type int64;
	}
	uses target;
}
%s`

//
//  Helper Functions
//
func buildSchema(t *testing.T, schema_snippet string) schema.ModelSet {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate, schema_snippet))
	st, err := testutils.GetConfigSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error when parsing RPC schema: %s", err)
	}

	return st
}

func buildSchemaRetWarns(
	t *testing.T, schema_snippet string,
) (schema.ModelSet, []xutils.Warning, error) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate, schema_snippet))
	return testutils.GetConfigSchemaWithWarns(schema_text.Bytes())
}

type checkFn func(t *testing.T, actual schema.Node)

type NodeChecker struct {
	Name   string
	checks []checkFn
}

func (l NodeChecker) GetName() string {
	return l.Name
}

func (expected NodeChecker) check(t *testing.T, actual schema.Node) {
	for _, checker := range expected.checks {
		checker(t, actual)
	}
}

func (n NodeChecker) String() string {
	return n.Name
}

func findChildByName(nl []schema.Node, name string) schema.Node {
	for _, v := range nl {
		if v.Name() == name {
			return v
		}
	}
	return nil
}

func checkAllNodes(t *testing.T, node_name string, expected []NodeChecker, actual []schema.Node) {
	if len(expected) != len(actual) {
		t.Errorf("Node %s child count does not match\n  expect=%d - %s\n  actual=%d - %s",
			node_name, len(expected), expected, len(actual), actual)
	}
	for _, exp := range expected {
		actualLeaf := findChildByName(actual, exp.GetName())
		if actualLeaf == nil {
			t.Errorf("Expected leaf not found: %s\n", exp.GetName())
			continue
		}
		exp.check(t, actualLeaf)
	}
}

func CheckChildren(node_name string, expected []NodeChecker) checkFn {
	return func(t *testing.T, actual schema.Node) {
		checkAllNodes(t, node_name, expected, actual.Children())
	}
}

func CheckName(expected_name string) checkFn {
	return func(t *testing.T, actual schema.Node) {
		if expected_name != actual.Name() {
			t.Errorf("Node name does not match\n  expect=%s\n  actual=%s",
				expected_name, actual.Name())
		}
	}
}

func CheckSchemaType(expected_type string) checkFn {
	return func(t *testing.T, actual schema.Node) {
		actual_type := reflect.TypeOf(actual).String()
		if expected_type != actual_type {
			t.Errorf("Node type does not match\n  expect=%s\n  actual=%s",
				expected_type, actual_type)
		}
	}
}

func CheckType(expected_type string) checkFn {
	return func(t *testing.T, actual schema.Node) {
		switch actual.(type) {
		case schema.Leaf, schema.LeafList:
			actual_type := actual.Type().Name().Local
			if expected_type != actual_type {
				t.Errorf("Node type does not match\n  expect=%s\n  actual=%s",
					expected_type, actual_type)
			}
		default:
			t.Errorf("Attempt to check type of non-leaf node")
		}
	}
}

func CheckAnd(a, b checkFn) checkFn {
	return func(t *testing.T, actual schema.Node) {
		a(t, actual)
		b(t, actual)
	}
}

func CheckConfig(expect bool) checkFn {
	return func(t *testing.T, actual schema.Node) {
		if actual.Config() != expect {
			t.Errorf("Config mismatch for %s\n    expect: %t\n    actual: %t\n",
				actual.Name(), expect, actual.Config())
		}
	}
}

func CheckMandatory(expect bool) checkFn {
	return func(t *testing.T, actual schema.Node) {
		var act_mand bool
		switch v := actual.(type) {
		case schema.Leaf:
			act_mand = v.Mandatory()
		}
		if act_mand != expect {
			t.Errorf("Mandatory mismatch for %s\n    expect: %t\n    actual: %t\n",
				actual.Name(), expect, act_mand)
		}
	}
}

func CheckDescription(expect string) checkFn {
	return func(t *testing.T, actual schema.Node) {
		act_desc := actual.Description()
		if act_desc != expect {
			t.Errorf(
				"Description mismatch for %s\n    expect: %s\n    actual: %s\n",
				actual.Name(), expect, act_desc)
		}
	}
}

func NewLeafChecker(name string, checks ...checkFn) NodeChecker {
	checkList := append([]checkFn{
		CheckSchemaType("*schema.leaf"),
		CheckName(name)},
		checks...)
	return NodeChecker{name, checkList}
}

func NewKeyChecker(name string, checks ...checkFn) NodeChecker {
	checkList := append([]checkFn{
		CheckSchemaType("*compile.key"),
		CheckName(name)},
		checks...)
	return NodeChecker{name, checkList}
}

func NewLeafListChecker(name string, checks ...checkFn) NodeChecker {
	checkList := append([]checkFn{
		CheckSchemaType("*schema.leafList"),
		CheckName(name)},
		checks...)
	return NodeChecker{name, checkList}
}

func NewListChecker(
	name string,
	children []NodeChecker,
	checks ...checkFn,
) NodeChecker {
	checkList := append([]checkFn{
		CheckName(name),
		CheckSchemaType("*schema.list"),
		CheckChildren(name, children)},
		checks...)
	return NodeChecker{name, checkList}
}

func NewContainerChecker(
	name string,
	children []NodeChecker,
	checks ...checkFn,
) NodeChecker {
	checkList := append([]checkFn{
		CheckName(name),
		CheckSchemaType("*schema.container"),
		CheckChildren(name, children)},
		checks...)
	return NodeChecker{name, checkList}
}

func NewTreeChecker(name string, children []NodeChecker, checks ...checkFn) NodeChecker {
	checkList := append([]checkFn{
		CheckSchemaType("*schema.tree"),
		CheckChildren(name, children)},
		checks...)
	return NodeChecker{name, checkList}
}

//
// Utility Functions
//

var testSchemaWalkEnabled bool

func enableTestSchemaWalk() { testSchemaWalkEnabled = true }

func disableTestSchemaWalk() { testSchemaWalkEnabled = false }

func testSchemaWalk() bool {
	return testSchemaWalkEnabled
}

// Check <err> contains ALL expected constituent strings in <expected>
func assertErrorContains(t *testing.T, err error, expected ...string) {
	if err == nil {
		t.Errorf(
			"Unexpected success when parsing schema and expecting:\n  %v",
			expected)
		return
	}
	for _, expStr := range expected {
		if !strings.Contains(err.Error(), expStr) {
			t.Errorf("Unexpected error output:\n    expect: %s\n    actual=%s",
				expected, err.Error())
		}
	}
}

func assertSuccess(t *testing.T, text string, err error) {
	if err != nil {
		t.Errorf(
			"Unexpected failure when parsing schema\n  %s\n\n  %s",
			text, err.Error())
		return
	}
}

//
//  Helper Functions
//
func assertLeafMatches(
	t *testing.T, st schema.ModelSet, node_name, node_type string, checks ...checkFn) {

	checks = append(checks, CheckType(node_type))
	expected := NewLeafChecker(node_name, checks...)
	actual := st.Child(node_name)

	expected.check(t, actual)
}

// Apply single schema using template provided in testcase and verify it
// compiles.
func applyAndVerifySchema(
	t *testing.T,
	testCase *testutils.TestCase,
	fullSchema bool,
) bool {
	// Schema is built up in 2 stages.  First we use the specific template
	// inside the test case (this allows us to avoid duplication between
	// tests using similar constructs).  Then we put the result into the
	// standard template that provides generic module, prefix etc.
	testSchema := fmt.Sprintf(testCase.Template, testCase.Schema)
	testSchema = fmt.Sprintf(SchemaTemplate, testSchema)
	sch := bytes.NewBufferString(testSchema).Bytes()

	return applyAndVerify(t, testCase, [][]byte{sch}, fullSchema)
}

// Apply multiple schemas from testcase and verify they compile.
func applyAndVerifySchemas(
	t *testing.T,
	testCase *testutils.TestCase,
	fullSchema bool,
) bool {
	var schemas = make([][]byte, len(testCase.Schemas))
	for index, schemaDef := range testCase.Schemas {
		schemas[index] = []byte(testutils.ConstructSchema(schemaDef))
	}

	return applyAndVerify(t, testCase, schemas, fullSchema)
}

func applyAndVerify(
	t *testing.T,
	testCase *testutils.TestCase,
	schemas [][]byte,
	fullSchema bool,
) bool {
	var compiledSchemaTree schema.ModelSet
	var err error

	if fullSchema {
		compiledSchemaTree, err = testutils.GetFullSchema(schemas...)
	} else {
		compiledSchemaTree, err = testutils.GetConfigSchema(schemas...)
	}

	if testCase.ExpResult && len(testCase.ExpErrMsg) != 0 {
		t.Fatalf("Cannot set expected error for passing test!")
		return false
	}

	if (testCase.ExpResult && err != nil) ||
		(!testCase.ExpResult && err == nil) {
		// Unexpected result
		t.Logf("TEST: %s\n", testCase.Description)
		if testCase.ExpResult {
			t.Logf("Expected schema to work but it failed.\n")
			t.Log(err)
		} else {
			t.Logf("Expected schema to fail but it worked.\n")
		}
		// If you want the calling stack, enable the line below.
		//testutils.LogStack(t)
		t.Fail()
		return false
	} else if !testCase.ExpResult {
		// Expected failure.
		//
		// It's very easy to have test cases failing for the wrong reason,
		// so we provide the expected error message (if feasible), or at
		// least a subset, to match on.
		if len(testCase.ExpErrMsg) == 0 && len(testCase.ExpErrs) == 0 {
			t.Logf("Must specify expected error message for failure.\n")
			t.Logf("Got : %s\n", err.Error())
			// If you want the calling stack, enable the line below.
			// testutils.LogStack(t)
			t.Fail()
			return false
		}

		var expErrs []string
		if testCase.ExpErrMsg != "" {
			expErrs = append(expErrs, testCase.ExpErrMsg)
		} else {
			expErrs = append(expErrs, testCase.ExpErrs...)
		}

		for _, expErr := range expErrs {
			if !strings.Contains(err.Error(), expErr) {
				t.Logf("Exp : %s\n", expErr)
				t.Logf("Got : %s\n", err.Error())
				// If you want the calling stack, enable the line below.
				// testutils.LogStack(t)
				t.Fail()
				return false
			}
		}
	} else if testCase.NodesToValidate != nil {
		// Expected pass, with nodeSpec(s) to validate.
		ok, fail_reason := schematests.ValidateNodes(
			t, testCase.NodesToValidate, compiledSchemaTree)
		if !ok {
			t.Logf("TEST '%s' failed:\n", testCase.Description)
			t.Logf("REASON: %s\n", fail_reason)
			t.Fail()
			return false
		}

		if testSchemaWalk() {
			schematests.WalkNodes(t, compiledSchemaTree)
		}
	}

	return true
}

// Simple wrapper so that running test cases is a one-liner ...
func runTestCases(t *testing.T, testCases []testutils.TestCase) {
	for _, tc := range testCases {
		applyAndVerifySchema(t, &tc, false)
	}
}

// Simple wrapper so that running test cases is a one-liner ...
func runTestCasesFullSchema(t *testing.T, testCases []testutils.TestCase) {
	for _, tc := range testCases {
		applyAndVerifySchema(t, &tc, true)
	}
}

// Copied from schematests.go
func pathsEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}

	for (len(first) > 0) && (len(second) > 0) {
		if first[0] != second[0] {
			return false
		}
		first = first[1:]
		second = second[1:]
	}

	return true
}

// Return done == true if/when we find a matching node.
func nodeMatcher(
	targetNode schema.Node,
	parentNode *schema.XNode,
	nodeToFind schema.NodeSpec,
	path []string,
	param interface{},
) (bool, bool, []interface{}) {
	tmp_path := append(path, targetNode.Name())
	if !pathsEqual(tmp_path, nodeToFind.Path) {
		return false, true, nil
	}
	return true, true, nil
}

func findSchemaNodeInTree(
	t *testing.T,
	st schema.ModelSet,
	path []string,
) schema.Node {

	nodeToFind := schema.NodeSpec{
		Path: path,
	}
	actual, ok, _ := st.FindOrWalk(nodeToFind, nodeMatcher, t)
	if !ok || actual == nil {
		t.Fatalf("Unable to find node: %s", path)
		return nil
	}

	return actual
}

func getSchemaNodeFromPath(
	t *testing.T,
	schema_text *bytes.Buffer,
	path []string,
) schema.Node {

	st, err := testutils.GetFullSchema(schema_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error when parsing schema: %s", err)
	}

	return findSchemaNodeInTree(t, st, path)
}

// Tests
//
// Originally this file had the top level calls to run tests, with those
// tests being defined as long tables in separate files.  In theory that
// gave a good overview of what was being tested, but this doesn't match
// the structure elsewhere, and with good file names is not necessary.  It
// also stops us breaking down the tabulated tests into individual tests,
// which hinders precise location of any breakage.
//
// This file should now only contain generic test helpers and definitions,
// with test data and invocations in separate feature-specific files.  Tests
// remaining below are those yet to be implemented ...

func TestModules(t *testing.T) {
	t.Skipf("Modules, submodules, includes and imports")
}

func TestLeafList(t *testing.T) {
	t.Skipf("Need tests for leaf list")
}

func TestChoice(t *testing.T) {
	t.Skipf("Need tests for choice")
}

// Need to implement PASS and FAIL tests
func TestBinaryFail(t *testing.T) {
	t.Skipf("Binary not fully implemented")
}

// Need to implement PASS and FAIL tests
func TestIdentityRef(t *testing.T) {
	t.Skipf("IdentityRef not fully implemented")
}

// Need to implement PASS and FAIL tests
func TestInstanceId(t *testing.T) {
	t.Skipf("Instance-identifier not fully implemented")
}

// Need to implement PASS and FAIL tests
func TestLeafRef(t *testing.T) {
	t.Skipf("leafref not fully implemented")
}
