// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015, 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This package implements test support for the schema package, but keeps
// test-only code out of the shipping image.

package schematests

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/danos/yang/schema"
	. "github.com/danos/yang/testutils"
)

// Only enable logging once the self-test has been done.
var loggingEnabled = false

func EnableSchemaTestLogging() {
	loggingEnabled = true
}
func DisableSchemaTestLogging() {
	loggingEnabled = false
}

func testLog(t *testing.T, format string, params ...interface{}) {
	if loggingEnabled {
		t.Logf(format, params...)
	}
}

// In developing these tests, I got caught out when my nodeSpec had multiple
// properties and I returned true after first one matched, even when second
// actually should have failed.  So, the first time we call ValidateNodes(),
// we run a quick self-test.
const selfTestSchemaTemplate = `
module test-yang-compile {
	namespace "urn:vyatta.com:test:yang-schema-test";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for schematests";
	}
	%s
}
`
const selfTestSchema = `
container testContainer {
	list testList {
		key "testKey";
		leaf test_dec64 {
			default "6.6";
			description "test_dec64 description";
			type decimal64 {
				fraction-digits 3;
			}
		}
		leaf test_int {
			default "66";
			description "test_int description";
			type int8;
		}
		leaf test_string {
			description "test_string description";
			type string;
			mandatory "true";
		}
	}
}`

const propNotPresentTestSchema = `
container testContainer {
	list testList {
		key "testKey";
		leaf test_dec64 {
			description "test_dec64 description";
			type decimal64 {
				fraction-digits 3;
			}
		}
		leaf test_int {
			description "test_int description";
			type int8;
			default "66";
		}
	}
}`

// Need to test both passing and failing cases to ensure tests work correctly
// and must also verify errors are the expected ones!
var schemaSelfTests = []TestCase{
	{
		Description: "Wrong Node Property (Description)",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "wrong leaf property for description",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec65 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
	{
		Description: "Wrong Node Type",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "wrong type (leaf, got List)",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "List",
					Properties: []schema.NodeProperty{ // Should be leaf
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
	{
		Description: "Wrong Data Property (Name)",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "wrong type data for 'name'",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin integer}"}}},
			},
		},
	},
	{
		Description: "Unsupported Node Statement Type",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "invalid leaf property type 'unknown'",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"unknown", "runcible spoon"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
	{
		Description: "Wrong Data Type",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "wrong type (decimal64, got uinteger)",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "uinteger",
					Properties: []schema.NodeProperty{ // Should be Dec64
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
	{
		Description: "Unsupported Data Type",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "unsupported type property 'not_valid'",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"not_valid", "stuff and nonsense"}}},
			},
		},
	},
	{
		Description: "Non-existent node",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "Cannot find [Container testList test_dec64] node",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"Container", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
	{
		Description: "All nodes match",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   true,
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
			{
				Path: []string{"testContainer", "testList", "test_int"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "66"},
						{"description", "test_int description"},
						{"mandatory", "false"}}},
				Data: schema.NodeSubSpec{
					Type: "integer",
					Properties: []schema.NodeProperty{
						{"name", "{builtin int8}"}}},
			},
			{
				Path: []string{"testContainer", "testList", "test_string"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"description", "test_string description"},
						{"mandatory", "true"}}},
				Data: schema.NodeSubSpec{
					Type: "ystring",
					Properties: []schema.NodeProperty{
						{"name", "{builtin string}"}}},
			},
		},
	},
	{
		Description: "Only first node matches",
		Template:    selfTestSchemaTemplate,
		Schema:      selfTestSchema,
		ExpResult:   false,
		ExpErrMsg: "data for 'name' (exp '{builtin decimal64}', " +
			"got '{builtin int8}'",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
			{
				Path: []string{"testContainer", "testList", "test_int"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "66"},
						{"description", "test_int description"}}},
				Data: schema.NodeSubSpec{
					Type: "integer",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}}, // Wrong type
			},
		},
	},
}

// Tests for property not found (either statement or data).  Note we are
// testing for absence of ANY value for the property.
//
// Test cases (repeated for DATA and STATEMENT)
//
// - PASS when property should not be present and is not present
// - FAIL when property should not be present, non-zero value given
//        (to avoid any confusion - if value is specified, then caller
//         might not be clear what they are testing!)
// - FAIL when property should not be present but is present
var propNotPresentTests = []TestCase{
	{
		Description: "Node found, specified data prop NOT present, " +
			"otherwise all matches",
		Template:  selfTestSchemaTemplate,
		Schema:    propNotPresentTestSchema,
		ExpResult: true,
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{
					"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"description", "test_dec64 description"}}},
				DataPropNotPresent: true,
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"default", ""}}},
			},
		},
	},
	{
		Description: "Node found, requested 'absent' property has " +
			"value specified",
		Template:  selfTestSchemaTemplate,
		Schema:    propNotPresentTestSchema,
		ExpResult: false,
		ExpErrMsg: "Node Data value should be empty when testing absence",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{
					"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"description", "test_dec64_description"}}},
				DataPropNotPresent: true,
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"default", "66.6"}}},
			},
		},
	},
	{
		Description: "Node found, specified data prop present when shouldn't be",
		Template:    selfTestSchemaTemplate,
		Schema:      propNotPresentTestSchema,
		ExpResult:   false,
		ExpErrMsg:   "Data property should not exist",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{
					"testContainer", "testList", "test_int"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"description", "test_int description"}}},
				DataPropNotPresent: true,
				Data: schema.NodeSubSpec{
					Type: "integer",
					Properties: []schema.NodeProperty{
						{"default", ""}}},
			},
		},
	},
	{
		Description: "Node found, specified stmt prop NOT present, " +
			"otherwise all matches",
		Template:  selfTestSchemaTemplate,
		Schema:    propNotPresentTestSchema,
		ExpResult: true,
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{
					"testContainer", "testList", "test_dec64"},
				DefaultPropNotPresent: true,
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", ""}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
	{
		Description: "Node found, requested 'absent' stmt property has " +
			"value specified",
		Template:  selfTestSchemaTemplate,
		Schema:    propNotPresentTestSchema,
		ExpResult: false,
		ExpErrMsg: "Stmt value should be empty when testing absence",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{
					"testContainer", "testList", "test_dec64"},
				DefaultPropNotPresent: true,
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
	{
		Description: "Node found, specified stmt prop present when " +
			"shouldn't be",
		Template:  selfTestSchemaTemplate,
		Schema:    propNotPresentTestSchema,
		ExpResult: false,
		ExpErrMsg: "Statement property should not exist",
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{
					"testContainer", "testList", "test_int"},
				DefaultPropNotPresent: true,
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", ""}}},
				Data: schema.NodeSubSpec{
					Type: "integer",
					Properties: []schema.NodeProperty{
						{"name", "{builtin int8}"}}},
			},
		},
	},
}

func runTests(
	t *testing.T,
	tests []TestCase) {

	for _, tc := range tests {
		sch := bytes.NewBufferString(fmt.Sprintf(
			tc.Template, tc.Schema))
		compiledSchemaTree, err := GetConfigSchema(sch.Bytes())

		if err != nil {
			t.Errorf("Self-test has failed (schema compilation)!!!\n")
			return
		}

		ok, failReason := validateNodesInternal(
			t, tc.NodesToValidate, compiledSchemaTree)
		if (tc.ExpResult == true) && (ok == false) {
			t.Errorf("%s: should have passed.\nReason: %s",
				tc.Description, failReason)
		} else if (tc.ExpResult == false) && (ok == true) {
			t.Errorf("%s: should have failed.\n", tc.Description)
			return
		} else if !ok {
			if (len(tc.ExpErrMsg) == 0) ||
				!strings.Contains(failReason[0].(string), tc.ExpErrMsg) {
				t.Errorf("%s failed:\n  Exp: %s\n  Got: %s\n",
					tc.Description, tc.ExpErrMsg, failReason)
			}
		}
	}
}

func selfTest(t *testing.T) {
	runTests(t, schemaSelfTests)
	runTests(t, propNotPresentTests)
}

// Utility function to log failures (if logging enabled)
func logPropertyMatchFail(t *testing.T, nodePath []string,
	property schema.NodeProperty, nodeTypeStr string, actVal interface{}) {
	testLog(t, "Fail: '%s' on '%s' (%s):\nExp: '%v'\nGot: '%v'\n",
		property.NodeProp, nodePath,
		nodeTypeStr, property.NodeValue, actVal)
}

func checkPropertyValueAndLog(
	t *testing.T,
	path []string,
	property schema.NodeProperty,
	name string,
	actualValue interface{},
) (success bool, failReason string) {
	var match bool

	switch actualValue.(type) {
	case bool:
		if ((property.NodeValue == "false") && (actualValue.(bool) == false)) ||
			((property.NodeValue == "true") && (actualValue.(bool) == true)) {
			match = true
		}
	case string:
		if property.NodeValue == actualValue.(string) {
			match = true
		}
	default:
		errMsg := fmt.Sprintf(
			"%s - unsupported %s property type %s (%s) being checked.",
			path, name, reflect.TypeOf(actualValue), property.NodeProp)
		return false, errMsg
	}

	if !match {
		logPropertyMatchFail(t, path, property, name, actualValue)
		errMsg := fmt.Sprintf(
			"%s - wrong %s property for %s (%s, got %s)",
			path, name, property.NodeProp,
			property.NodeValue, actualValue)
		return false, errMsg
	}

	return true, ""
}

// Statement validation functions.  General aim is to:
//   - implement new properties and node types on demand
//   - return false if property doesn't match or property isn't handled.
func validateLeafProperties(
	t *testing.T,
	targetNode schema.Node,
	nodeToFind schema.NodeSpec,
) (success bool, failReason string) {

	var errMsg string

	for _, property := range nodeToFind.Statement.Properties {
		switch property.NodeProp {
		case "default":
			if nodeToFind.DefaultPropNotPresent && (property.NodeValue != "") {
				return false, "Stmt value should be empty when testing absence"
			}
			def, ok := targetNode.(schema.Leaf).Default()
			if !ok {
				if nodeToFind.DefaultPropNotPresent {
					return true, "Stmt property absent as expected"
				}
				logPropertyMatchFail(t, nodeToFind.Path, property,
					"leaf", def)
				errMsg = fmt.Sprintf(
					"%s - no leaf property for %s (expected '%s')",
					nodeToFind.Path, property.NodeProp, property.NodeValue)
				return false, errMsg
			}
			if ok && nodeToFind.DefaultPropNotPresent {
				return false, "Statement property should not exist"
			}
			ok, errMsg := checkPropertyValueAndLog(
				t, nodeToFind.Path, property, "leaf", def)
			if !ok {
				return false, errMsg
			}
		case "description":
			ok, errMsg := checkPropertyValueAndLog(
				t, nodeToFind.Path, property, "leaf", targetNode.(schema.Leaf).Description())
			if !ok {
				return false, errMsg
			}
		case "mandatory":
			ok, errMsg := checkPropertyValueAndLog(
				t, nodeToFind.Path, property, "leaf",
				targetNode.(schema.Leaf).Mandatory())
			if !ok {
				return false, errMsg
			}
		default:
			testLog(t, "Unsupported leaf property type %s\n", property.NodeProp)
			errMsg = fmt.Sprintf("%s - invalid leaf property type '%s'",
				nodeToFind.Path, property.NodeProp)
			return false, errMsg
		}
	}
	return true, "Leaf properties validated."
}

func validateListProperties(
	t *testing.T,
	targetNode schema.Node,
	nodeToFind schema.NodeSpec,
) (success bool, failReason string) {

	var errMsg string

	for _, property := range nodeToFind.Statement.Properties {
		switch property.NodeProp {
		default:
			testLog(t, "Unsupported List property type %s\n", property.NodeProp)
			errMsg = fmt.Sprintf(
				"%s - unsupported list property '%s'",
				nodeToFind.Path, property.NodeProp)
		}
	}
	return false, errMsg
}

// Verify that the properties of the statement as given in the NodeSpec are
// as required.
func validateStatementProperties(
	t *testing.T,
	targetNode schema.Node,
	nodeToFind schema.NodeSpec,
) (success bool, failReason string) {

	var errMsg string

	switch targetNode.(type) {
	case schema.Leaf:
		success, failReason = validateLeafProperties(
			t, targetNode, nodeToFind)
		if !success {
			return false, failReason
		}
	case schema.List:
		success, failReason = validateListProperties(
			t, targetNode, nodeToFind)
		if !success {
			return false, failReason
		}
	default:
		testLog(t, "Unsupported statement type %s - cannot validate.\n",
			reflect.TypeOf(targetNode))
		errMsg = fmt.Sprintf(
			"%s - unsupported statement type '%s'",
			nodeToFind.Path, reflect.TypeOf(targetNode))
		return false, errMsg
	}
	return true, "Statement properties validated."
}

// Data node validation functions - many schema node can have an underlying
// data type (one of the built-in types), and these functions help to validate
// this part of the schema tree.  Very similar logic to the statement
// verification.
func validateDataNodeProperties(
	t *testing.T,
	dataNode schema.Type,
	property schema.NodeProperty,
	nodeName []string,
	dataType string,
	shouldNotBePresent bool,
) (success bool, failReason string) {

	var errMsg string

	switch property.NodeProp {
	case "default":
		def, ok := dataNode.Default()
		if !ok {
			if shouldNotBePresent {
				return true, "Data property absent as expected"
			}
			return false, "Node has no default"
		}
		if ok && shouldNotBePresent {
			return false, "Data property should not exist"
		}
		if def == property.NodeValue {
			return true, "Default matched"
		}
		errMsg = fmt.Sprintf(
			"%s - wrong type data for 'default' (exp '%s', got '%s')",
			nodeName, property.NodeValue, def)
		logPropertyMatchFail(t, nodeName, property, dataType, def)
	case "name":
		if shouldNotBePresent {
			return false, "Name property ALWAYS exists!"
		}
		name := fmt.Sprintf("%s", dataNode.Name())
		if name == property.NodeValue {
			return true, "Name matched"
		}
		errMsg = fmt.Sprintf(
			"%s - wrong type data for 'name' (exp '%s', got '%s')",
			nodeName, property.NodeValue, name)
		logPropertyMatchFail(t, nodeName, property,
			dataType, name)
	default:
		testLog(t, "Unsupported: DataNode property '%s'\n", property.NodeProp)
		errMsg = fmt.Sprintf(
			"%s - unsupported type property '%s'",
			nodeName, property.NodeProp)
	}

	return false, errMsg
}

// Verify that the properties of the datatype as given in the NodeSpec are
// as required.
func validateDataProperties(
	t *testing.T,
	dataNode schema.Type,
	nodeToFind schema.NodeSpec,
) (success bool, failReason string) {

	for _, property := range nodeToFind.Data.Properties {
		if nodeToFind.DataPropNotPresent && (property.NodeValue != "") {
			return false, "Node Data value should be empty when testing absence"
		}
		success, failReason := validateDataNodeProperties(
			t, dataNode, property, nodeToFind.Path, nodeToFind.Data.Type,
			nodeToFind.DataPropNotPresent)
		if !success {
			return false, failReason
		}
	}

	return true, "Data properties validated ok"
}

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

// Helper function that can be passed into the schema.FindorWalk() function
// that finds a Node matching a path
func NodeFinder(
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

// Helper function that can be passed into the schema.FindorWalk() function
// that returns true if the targetNode matches the NodeSpec.
//
// We use param to get the 'testing.T' object here from ValidateNodes without
// the 'schema' package needing to know anything about it (-:
//
// Return: { done(true)/continue(false), pass/fail, fail_reason_string }
func nodeMatcher(
	targetNode schema.Node,
	parentNode *schema.XNode,
	nodeToFind schema.NodeSpec,
	path []string,
	param interface{},
) (bool, bool, []interface{}) {
	// First check - fully-qualified name must match.  There can only be
	// one node with a given fully-qualified name.
	tmp_path := append(path, targetNode.Name())
	if !pathsEqual(tmp_path, nodeToFind.Path) {
		return false, true, nil
	}

	// Second check - does data type match, and if so, do properties?
	// NB: we check data before statement as this could be considered a
	//     'lower-level' check and thus for code development / debug a
	//     failure in statement may reflect an error seen also in data, but
	//     not vice versa.
	dataType := reflect.TypeOf(targetNode.Type()).String()
	dataType = dataType[strings.Index(dataType, ".")+1:]
	if dataType != nodeToFind.Data.Type {
		errMsg := fmt.Sprintf(
			"%s - wrong type (%s, got %s)",
			nodeToFind.Path, dataType, nodeToFind.Data.Type)
		return true, false, convertToInterface(errMsg)
	}

	var success bool
	var failReason string
	switch targetNode.(type) {
	case schema.Leaf:
		success, failReason = validateDataProperties(
			param.(*testing.T), targetNode.(schema.Leaf).Type(), nodeToFind)
	default:
		success = false
		failReason = "Unsupported node type."
	}

	if !success {
		return true, false, convertToInterface(failReason)
	}

	// Third check - does statement type match, along with its properties?
	statementType := reflect.TypeOf(targetNode).String()
	statementType = statementType[strings.Index(statementType, ".")+1:]
	if statementType != nodeToFind.Statement.Type {
		errMsg := fmt.Sprintf(
			"%s - wrong type (%s, got %s)",
			nodeToFind.Path, statementType, nodeToFind.Statement.Type)
		return true, false, convertToInterface(errMsg)
	}
	success, failReason = validateStatementProperties(
		param.(*testing.T), targetNode, nodeToFind)
	if !success {
		return true, false, convertToInterface(failReason)
	}

	// OK - we have the correct node and properties, if any, match.
	return true, true, convertToInterface("Node name and properties match")
}

func convertToInterface(msg string) []interface{} {
	var intf []interface{}
	return append(intf, msg)
}

var selfTestRun = false

// Using the provided NodeSpec(s), check all nodes exist with the required
// properties.  Return true if so (or if no nodes in the spec), otherwise
// false to indicate failure.
func ValidateNodes(
	t *testing.T,
	spec []schema.NodeSpec,
	st schema.ModelSet,
) (bool, []interface{}) {

	if !selfTestRun {
		selfTest(t)
		selfTestRun = true
	}

	if len(spec) == 0 {
		return true, convertToInterface("No nodes to match - guaranteed pass!")
	}
	if st == nil {
		return false, convertToInterface("No schema tree to validate against!")
	}

	return validateNodesInternal(t, spec, st)
}

// selfTest() needs to call this, and is called from ValidateNodes(), so
// to avoid an infinite loop we need to put this functionality into a common
// function called by both!
func validateNodesInternal(
	t *testing.T,
	spec []schema.NodeSpec,
	st schema.ModelSet,
) (bool, []interface{}) {
	// For each node to be validated first find the node that has the given
	// type and name.  Then check all properties required match.  If not,
	// keep looking.
	for _, valNode := range spec {
		node, success, failReason := st.FindOrWalk(valNode, nodeMatcher, t)
		if node == nil {
			testLog(t, " => Cannot find %s node (%s / %s).\n",
				valNode.Path, valNode.Statement.Type, valNode.Data.Type,
				failReason)
			errMsg := fmt.Sprintf(
				"Cannot find %s node", valNode.Path)
			return false, convertToInterface(errMsg)
		}
		if success == false {
			testLog(t, " => Invalid %s node data (%s / %s).  %s.\n",
				valNode.Path, valNode.Statement.Type, valNode.Data.Type,
				failReason[0].(string))
			return false, failReason // Must match ALL nodes to pass.
		}
	}

	// If we get here, we found a node with matching name and type, and
	// either all properties matched, or there are none to match.
	//
	// Alternatively, there were no nodes to check.
	return true, nil
}

// Helper function for WalkNodes() that does the actual printing of each
// node in the schema tree.
func nodePrinter(
	targetNode schema.Node,
	parentNode *schema.XNode,
	nodeToFind schema.NodeSpec,
	path []string,
	param interface{},
) (bool, bool, []interface{}) {

	fmt.Printf("%*s%v %s: %s / %s\n",
		len(path)*2, " ", len(path), targetNode.Name(),
		reflect.TypeOf(targetNode),
		reflect.TypeOf(targetNode.Type()))
	return false, true, nil
}

// Useful if you want an indented dump of nodes in the schema tree.
func WalkNodes(t *testing.T, st schema.ModelSet) {
	var dummyNode schema.NodeSpec

	if st == nil {
		return
	}

	st.FindOrWalk(dummyNode, nodePrinter, t)
}
