// Copyright (c) 2019-2020, AT&T Intellectual Property
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests on grouping, uses, refines and augment options
// available in Yang (RFC 6020).

package compile_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/steiler/yang-parser/testutils"
)

var GroupingPass = []testutils.TestCase{
	{
		Description: "Grouping: Passing cases",
		Template:    BlankTemplate,
		Schema: `grouping target {
			container test_uint8_container {
				leaf test_uint8 {
					type uint8;
				}
			}
			container test_uint8_2_container {
				leaf test_uint8_2 {
					type uint8;
				}
			}
			container test_uint8_3_container {
				leaf test_uint8_3 {
					type uint8;
				}
			}
		}

		container testuses {
			uses target;
		}

		container testuses2 {
			uses target;
		}

		container testuses3 {
			uses target;
			list test_list {
				key "name";
				leaf name {
					type string;
				}
				uses target;
			}
		}

		container testuses4 {
			uses target {
				augment test_uint8_container {
					uses target {
						augment test_uint8_2_container {
							uses target;
						}
					}
				}
				augment test_uint8_3_container {
					uses target;
				}
			}
		}
		`,
		ExpResult: true,
	},
}

var GroupingFail = []testutils.TestCase{
	{
		Description: "Grouping: self-referential group",
		Template:    GroupingTemplate,
		Schema: `container testuses_referential_group {
					uses target;
				}`,
		ExpResult: false,
		ExpErrMsg: "Grouping cycle detected",
	},
	{
		Description: "Grouping: invalid group",
		Template:    BlankTemplate,
		Schema: `container testuses_invalid_group {
					uses TargeT;
				}`,
		ExpResult: false,
		ExpErrMsg: "Unknown grouping (grouping TargeT) referenced from " +
			"testuses_invalid_group",
	},
	{
		Description: "Grouping: indirect self-referential target",
		Template:    BlankTemplate,
		Schema: `
		grouping target3 {
			leaf test_ascend {
				type uint8;
			}
			uses target4;
		}

		grouping target4 {
			leaf test_ascend {
				type uint8;
			}
			uses target3;
		}

		container testuses_indirectly_self_referential_group {
			uses target3;
		}
		`,
		ExpResult: false,
		ExpErrMsg: "Grouping cycle detected",
	},
	{
		Description: "Grouping: duplicate leaf",
		Template:    GroupingTemplate,
		Schema: `container testuses_duplicate_leaf {
			leaf test_int64 {
				type int64;
			}
			uses target;
			leaf test_uint8 {
				type uint8;
			}
		}
		`,
		ExpResult: false,
		ExpErrMsg: "Grouping cycle detected",
	},
}

// NB: unclear on return on investment for this!
func TestGroupingSubstatements(t *testing.T) {
	t.Skipf("Substatements and cardinality for grouping.")
}

// Grouping
// - grouping, uses, augment
func TestGroupingPass(t *testing.T) {
	runTestCases(t, GroupingPass)
}

// Need to check we can refer to an external group, and ensure that if we
// have internal and imported groupings with same name that we can 'use'
// both without them being seen as same group now we check for infinite loops.
func TestGroupingSkip(t *testing.T) {
	t.Skipf("External grouping import, substatements not done.")
}

// Grouping must not reference self.
// Grouping name must be unique at level defined (module or submodule)
func TestGroupingFail(t *testing.T) {
	runTestCases(t, GroupingFail)
}

var RefinePass = []testutils.TestCase{
	{
		Description: "Refine: expected PASS cases",
		Template:    BlankTemplate,
		Schema: `grouping target {
			container test_container {
				leaf test_uint8 {
					type uint8 {
						range "1..30";
					}
					mandatory "true";
				}
				leaf-list test_leaf_list {
					type string {
						length "2..20";
					}
				}
				list server {
					key "name";
					unique "ip port";
					leaf name {
						type string;
					}
					leaf ip {
						type uint32;
					}
					leaf port {
						type uint16;
					}
				}
			}
		}

		container testuses {
			uses target {
				refine test_container/test_uint8 {
					description "New description";
				default 20;
					mandatory "false";
				}
			}
		}`,
		ExpResult: true,
	},
}

// Uses
//
// Substatements
//   - augment
//   - description
//   - if-feature
//   - refine
//   - reference
//   - status
//   - when
func TestUsesSubstatements(t *testing.T) {
	t.Skipf("Substatements and cardinality for 'uses'.")
}

// Refine
//
// - container, list, leaf, list-leaf, choice
//
// Substatements
//   - leaf/choice may get default (or replacement one)
//   - any node may get specialised description
//   - any node may get specialised reference
//   - any node may get different config statement
//   - leaf/anyxml/choice node may get different mandatory statement
//   - container node may get presence statement
//   - leaf/leaf-list/list/container/anyxml may get additional must expressions
//   - leaf-list/list may get different min-/max-elements statement
func TestRefineSubstatements(t *testing.T) {
	t.Skipf("Substatements and cardinality for 'refine'.")
}

func TestRefinePass(t *testing.T) {
	runTestCases(t, RefinePass)
}

func TestMissingGrouping(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping one {
			leaf one {
				type string;
			}
			uses two;
		}`))

	expected := "Unknown grouping (grouping two) " +
		"referenced from grouping one"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestSimpleGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping one {
			leaf one {
				type string;
			}
			uses one;
		}`))

	expected := "Grouping cycle detected in: grouping one"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestInnerGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping one {
			grouping inner {
				leaf one {
					type string;
				}
				uses inner;
			}
			uses inner;
		}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestComplexGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping one {
			leaf one {
				type string;
			}
			uses three;
		}
		grouping two {
			leaf two {
				type string;
			}
			uses one;
		}
		grouping three {
			leaf three {
				type string;
			}
			uses one;
		}`))

	expected := "Grouping cycle detected in: grouping "
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestSubmoduleGroupingCycle(t *testing.T) {

	module_text := bytes.NewBufferString(
		`module test-yang-compile {
		namespace "urn:vyatta.com:test:yang-compile";
		prefix test;
		include subone;

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}
	}`)

	submodule_text := bytes.NewBufferString(
		`submodule subone {
			belongs-to test-yang-compile { prefix prefix; }
			grouping one {
				leaf one {
					type string;
				}
				uses one;
			}
		}`)

	expected := "Grouping cycle detected in: grouping one"
	_, err := testutils.GetConfigSchema(module_text.Bytes(), submodule_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestContainerGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container bucket {
			grouping inner {
				leaf one {
					type string;
				}
				uses inner;
			}
			uses inner;
		}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestListGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`list bucket {
			key name;
			leaf name {
					type string;
			}
			grouping inner {
				leaf one {
					type string;
				}
				uses inner;
			}
			uses inner;
		}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestRpcGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc bucket {
			grouping inner {
				leaf one {
					type string;
				}
				uses inner;
			}
		}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestRpcInputGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc bucket {
            input {
				grouping inner {
					leaf one {
						type string;
					}
					uses inner;
				}
			}
		}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestRpcOutputGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`rpc bucket {
            output {
				grouping inner {
					leaf one {
						type string;
					}
					uses inner;
				}
			}
		}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestAugmentGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container test;
		augment /test {
			container inside {
				grouping inner {
					leaf one {
						type string;
					}
					uses inner;
				}
			}
		}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestNestedGroupingCycle(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container outer {
		list bucket {
			key name;
			leaf name {
					type string;
			}
			grouping inner {
				leaf one {
					type string;
				}
				uses inner;
			}
			uses inner;
		}}`))

	expected := "Grouping cycle detected in: grouping inner"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestSimpleGroupExpansion(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping g1 {
			leaf two {
				type string;
			}
		}
		container c1 {
			leaf one {
				type string;
			}
			uses g1;
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one"),
			NewLeafChecker("two"),
		})

	actual := getSchemaNodeFromPath(t, schema_text, []string{"c1"})

	expected.check(t, actual)
}

func TestSimpleRefine(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping g1 {
			leaf one {
				type string;
			}
		}
		container c1 {
			uses g1 {
				refine one {
					mandatory true;
				}
			}
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one", CheckMandatory(true)),
		})

	actual := getSchemaNodeFromPath(t, schema_text, []string{"c1"})

	expected.check(t, actual)
}

func TestRefineListWithKeyBeforeList(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping g1 {
			list aList {
				key name; // Must be before leaf 'name' for this test.
				leaf name {
					type string;
				}
			}
		}
		container c1 {
			uses g1 {
				refine aList/name {
					must "true()";
				}
			}
		}`))

	expMachine :=
		"--- machine start ---\n" +
			"bltin\t\ttrue()\n" +
			"store\n" +
			"---- machine end ----\n"

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewListChecker(
				"aList",
				[]NodeChecker{
					NewKeyChecker("name",
						checkMusts([]mustExp{
							{
								expMachine,
								"'must' condition is false: 'true()'",
								defaultMustAppTag,
								NoPathToEvalMachine}})),
				}),
		})

	actual := getSchemaNodeFromPath(t, schema_text, []string{"c1"})

	expected.check(t, actual)
}

func TestOverrideRefine(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping g1 {
			leaf one {
				type string;
					mandatory true;
			}
		}
		container c1 {
			uses g1 {
				refine one {
					mandatory false;
				}
			}
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one", CheckMandatory(false)),
		})

	actual := getSchemaNodeFromPath(t, schema_text, []string{"c1"})

	expected.check(t, actual)
}

func TestForbiddenRefine(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping g1 {
			list testlist {
				key name;
				unique "server port";
				leaf name {
					type string;
				}
				leaf server {
					type string;
				}
				leaf port {
					type int32;
				}
			}
		}
		uses g1 {
			refine testlist {
				unique post;
			}
		}`))

	expected := "refine testlist: invalid refinement unique for statement list"
	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestComplexGroupExpansion(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping g1 {
			leaf two {
				type string;
			}
			container box {
				leaf one {
					type string;
				}
			}
		}
		grouping g2 {
			uses g1 {
				refine two {
					config false;
				}
				augment "box" {
					leaf two {
						type string;
					}
				}
			}
			leaf three {
				type string;
			}
		}
		container c1 {
			leaf one {
				type string;
			}
			uses g2 {
				refine two {
					config true;
				}
				augment "box" {
					leaf three {
						type string;
					}
				}
			}
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one"),
			NewLeafChecker("two", CheckConfig(true)),
			NewContainerChecker("box",
				[]NodeChecker{
					NewLeafChecker("one"),
					NewLeafChecker("two"),
					NewLeafChecker("three")}),
			NewLeafChecker("three"),
		})

	actual := getSchemaNodeFromPath(t, schema_text, []string{"c1"})

	expected.check(t, actual)
}

func TestUsesAugmentUses(t *testing.T) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`grouping g1 {
			leaf two {
				type string;
			}
			container box {
				leaf one {
					type string;
				}
			}
		}
		grouping g2 {
			uses g1 {
				refine two {
					config false;
				}
				augment "box" {
					leaf two {
						type string;
					}
				}
			}
			leaf three {
				type string;
			}
		}
		grouping g4 {
			leaf four {
				type string;
			}
		}
		container c1 {
			leaf one {
				type string;
			}
			uses g2 {
				refine two {
					config true;
				}
				augment "box" {
					leaf three {
						type string;
					}
					uses g4 {
						refine four {
							mandatory true;
						}
					}
				}
			}
		}`))

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("one"),
			NewLeafChecker("two", CheckConfig(true)),
			NewContainerChecker("box",
				[]NodeChecker{
					NewLeafChecker("one"),
					NewLeafChecker("two"),
					NewLeafChecker("three"),
					NewLeafChecker("four", CheckMandatory(true))}),
			NewLeafChecker("three"),
		})

	actual := getSchemaNodeFromPath(t, schema_text, []string{"c1"})

	expected.check(t, actual)
}

func TestUsingExternalGrouping(t *testing.T) {
	module1_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test;

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		grouping one {
			leaf foo { type string; }
		}
	}`)

	module2_text := bytes.NewBufferString(
		`module test-yang-compile2 {
		namespace "urn:vyatta.com:test:yang-compile2";
		prefix test;

		import test-yang-compile1 { prefix compile1; }

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		uses compile1:one;
	}`)

	expected := NewLeafChecker("foo")

	st, err := testutils.GetConfigSchema(
		module1_text.Bytes(),
		module2_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	if actual := findSchemaNodeInTree(t, st,
		[]string{"foo"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestMissingImport(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test;

		import dont-exist {
			prefix error;
		}

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		grouping one {
			leaf foo { type string; }
		}
	}`)

	expected := "import dont-exist: module not found"
	_, err := testutils.GetConfigSchema(module_text.Bytes())

	assertErrorContains(t, err, expected)
}

// This test checks that we correctly expand 'uses' inside groupings when
// the 'uses' is at the top level of the first grouping.
// when the 'uses' statement for the grouping comes before the grouping
// definition.  Unlike C / C++, order of definition vs reference in the YANG
// file does not matter.
func TestUsesExpansionWhenChildOfGrouping(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container before-cont {
			uses params-grp;
		}

		grouping params-grp {
			leaf direct-param-leaf {
				type empty;
			}
			uses used-params-grp;
		}

		grouping used-params-grp {
			leaf used-param-leaf {
				type empty;
			}
			uses doubly-used-params-grp;
		}

		grouping doubly-used-params-grp {
			leaf doubly-used-param-leaf {
				type empty;
			}
		}

		container after-cont {
			uses params-grp;
		}`))

	expectedBefore :=
		NewContainerChecker(
			"before-cont",
			[]NodeChecker{
				NewLeafChecker("direct-param-leaf"),
				NewLeafChecker("used-param-leaf"),
				NewLeafChecker("doubly-used-param-leaf")})

	actualBefore := getSchemaNodeFromPath(t, schema_text,
		[]string{"before-cont"})
	expectedBefore.check(t, actualBefore)

	expectedAfter :=
		NewContainerChecker(
			"after-cont",
			[]NodeChecker{
				NewLeafChecker("direct-param-leaf"),
				NewLeafChecker("used-param-leaf"),
				NewLeafChecker("doubly-used-param-leaf")})

	actualAfter := getSchemaNodeFromPath(t, schema_text,
		[]string{"after-cont"})
	expectedAfter.check(t, actualAfter)
}

// Forward reference to grouping that itself has a forward reference (not
// at top-level of grouping).
func TestUsesExpansionDoublyForwardRef(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container before-cont {
			uses params-grp;
		}

		grouping params-grp {
			container params-cont {
				leaf direct-param-leaf {
					type empty;
				}
				uses used-params-grp;
			}
		}

		grouping used-params-grp {
			container used-param-cont {
				leaf used-param-leaf {
					type empty;
				}
				uses doubly-used-params-grp;
			}
		}

		grouping doubly-used-params-grp {
			leaf doubly-used-param-leaf {
				type empty;
			}
		}

		container after-cont {
			uses params-grp;
		}`))

	expected :=
		NewContainerChecker(
			"params-cont",
			[]NodeChecker{
				NewLeafChecker("direct-param-leaf"),
				NewContainerChecker(
					"used-param-cont",
					[]NodeChecker{
						NewLeafChecker("used-param-leaf"),
						NewLeafChecker("doubly-used-param-leaf")})})

	actualBefore := getSchemaNodeFromPath(t, schema_text,
		[]string{"before-cont", "params-cont"})
	expected.check(t, actualBefore)

	actualAfter := getSchemaNodeFromPath(t, schema_text,
		[]string{"after-cont", "params-cont"})
	expected.check(t, actualAfter)
}

// As a further check on full expansion, make sure we can refine the leaves
// that have been expanded via forward reference.
func TestRefineExpansionOnDoubleForwardRef(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container before-cont {
			uses params-grp {
				refine used-param-leaf {
					description "refined leaf (direct uses)";
				}
				refine doubly-used-param-leaf {
					description "refined leaf (direct uses (dbl))";
				}
				refine params-cont/used-param-leaf {
					description "refined leaf (contained uses)";
				}
				refine params-cont/doubly-used-param-leaf {
					description "refined leaf (contained/direct uses)";
				}
				refine params-cont/used-params-cont/doubly-used-param-leaf {
					description "refined leaf (contained uses (dbl))";
				}
				refine used-params-cont/doubly-used-param-leaf {
					description "refined leaf (contained/direct uses (dbl))";
				}
			}
		}

		grouping params-grp {
			uses used-params-grp;
			container params-cont {
				uses used-params-grp;
			}
		}

		grouping used-params-grp {
			leaf used-param-leaf {
				type empty;
			}
			uses doubly-used-params-grp;
			container used-params-cont {
				uses doubly-used-params-grp;
			}
		}

		grouping doubly-used-params-grp {
			leaf doubly-used-param-leaf {
				type empty;
			}
		}`))

	expectedBefore :=
		NewContainerChecker(
			"before-cont",
			[]NodeChecker{
				NewLeafChecker("used-param-leaf",
					CheckDescription("refined leaf (direct uses)")),
				NewLeafChecker("doubly-used-param-leaf",
					CheckDescription("refined leaf (direct uses (dbl))")),
				NewContainerChecker(
					"params-cont",
					[]NodeChecker{
						NewLeafChecker("used-param-leaf",
							CheckDescription("refined leaf (contained uses)")),
						NewLeafChecker("doubly-used-param-leaf",
							CheckDescription(
								"refined leaf (contained/direct uses)")),
						NewContainerChecker(
							"used-params-cont",
							[]NodeChecker{
								NewLeafChecker("doubly-used-param-leaf",
									CheckDescription(
										"refined leaf (contained uses (dbl))")),
							}),
					}),
				NewContainerChecker(
					"used-params-cont",
					[]NodeChecker{
						NewLeafChecker("doubly-used-param-leaf",
							CheckDescription(
								"refined leaf (contained/direct uses (dbl))")),
					}),
			})

	actualBefore := getSchemaNodeFromPath(t, schema_text,
		[]string{"before-cont"})
	expectedBefore.check(t, actualBefore)
}
