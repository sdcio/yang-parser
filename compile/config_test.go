// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/iptecharch/yang-parser/schema"
	"github.com/iptecharch/yang-parser/schema/schematests"
	. "github.com/iptecharch/yang-parser/testutils"
)

func verifyConfigStatement(
	t *testing.T,
	testschema string,
	nspec schema.NodeSpec, config bool) {
	// Compile the schema
	sch := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate, testschema))
	fullSchemaTree, err := GetFullSchema(sch.Bytes())

	if err != nil {
		t.Fatalf("Unable to compile schema; %s", err)
	}

	// Find a node in the FULL schema, we are looking for operational
	// state nodes
	sn, _, _ := fullSchemaTree.FindOrWalk(nspec, schematests.NodeFinder, t)

	if sn == nil {
		t.Fatalf("Unable to find node")
	} else {

		if sn.Config() != config {
			t.Errorf("Config statement mismatch")
			t.Logf("\nGot: %v\nExp: %v\n", sn.Config(), config)
			LogStack(t)
		}
	}

}

// Verify that a config statement of true is compiled as true,
// and false is false
func TestConfigTrueIsTrueFalseIsFalse(t *testing.T) {
	configTest := ` container truetestcontainer {
			config true;
			leaf teststatus {
				type string;
			}
		}
		container falsetestcontainer {
			config false;
			leaf teststatus {
				type string;
			}
		}
		container testcontainer {
			leaf teststatus {
				type string;
			}
		}
		`
	verifyConfigStatement(t, configTest,
		schema.NodeSpec{Path: []string{"truetestcontainer"}},
		true)
	verifyConfigStatement(t, configTest,
		schema.NodeSpec{Path: []string{"falsetestcontainer"}},
		false)
	verifyConfigStatement(t, configTest,
		schema.NodeSpec{Path: []string{"truetestcontainer", "teststatus"}},
		true)
	verifyConfigStatement(t, configTest,
		schema.NodeSpec{Path: []string{"falsetestcontainer", "teststatus"}},
		false)
	verifyConfigStatement(t, configTest,
		schema.NodeSpec{Path: []string{"testcontainer"}},
		true)
	verifyConfigStatement(t, configTest,
		schema.NodeSpec{Path: []string{"testcontainer", "teststatus"}},
		true)
}

func TestConfigInContainer(t *testing.T) {
	var configInContainerTest = []TestCase{
		{
			Description: "Config statement in a leaf is allowed",
			Template:    BlankTemplate,
			Schema: ` container test {
				config false;
				leaf teststatus {
					type string;
				}
			}
			`,
			ExpResult: true,
		},
	}
	runTestCasesFullSchema(t, configInContainerTest)
}

func TestConfigInLeaf(t *testing.T) {
	var configInLeafTest = []TestCase{
		{
			Description: "Config statement in a leaf is allowed",
			Template:    BlankTemplate,
			Schema: `container test {
				leaf teststatus {
					type string;
					config false;
				}
			}
			`,
			ExpResult: true,
		},
	}
	runTestCasesFullSchema(t, configInLeafTest)
}

func TestConfigInLeafList(t *testing.T) {
	var configInLeafListTest = []TestCase{
		{
			Description: "Config statement in a leaf-list is allowed",
			Template:    BlankTemplate,
			Schema: ` container test {
				leaf-list teststatus {
					type string;
					config false;
				}
			}
			`,
			ExpResult: true,
		},
	}
	runTestCasesFullSchema(t, configInLeafListTest)
}

func TestConfigInList(t *testing.T) {
	var configInListTest = []TestCase{
		{
			Description: "Config statement in a list is allowed",
			Template:    BlankTemplate,
			Schema: ` container test {
				list teststatus {
					key name;
					leaf name {
						type string;
					}
					leaf foo {
						type string;
					}
					config false;
				}
			}
			`,
			ExpResult: true,
		},
	}
	runTestCasesFullSchema(t, configInListTest)
}

func TestConfigInLeafInAList(t *testing.T) {
	var configInLeafInAListTest = []TestCase{
		{
			Description: "Config statement in a leaf which is in a " +
				"list is allowed",
			Template: BlankTemplate,
			Schema: ` container test {
				list testlist {
					key name;
					leaf name {
						type string;
					}
					leaf foo {
						type string;
					}
					leaf status {
						type string;
						config false;
					}
				}
			}
		`,
			ExpResult: true,
		},
	}
	runTestCasesFullSchema(t, configInLeafInAListTest)
}
func TestConfigFalseAllowedWhenParentFalse(t *testing.T) {
	var configFalseWithFalseParentTest = []TestCase{
		{
			Description: "Config: a config false is allowed as " +
				"a child of config false",
			Template: BlankTemplate,
			Schema: ` container test {
				config false;
				list testlist {
					config false;
					key name;
					leaf name {
						type string;
						config false;
					}
					leaf foo {
						type string;
					}
					leaf status {
						type string;
						config false;
					}
					container subcontainer {
						config false;
						leaf subleaf {
							type string;
							config false;
						}
					}
				}
			}
			`,
			ExpResult: true,
		},
	}
	runTestCasesFullSchema(t, configFalseWithFalseParentTest)
}
func TestConfigFalseAllowedWhenParentTrue(t *testing.T) {
	var configFalseWithTrueParentTest = []TestCase{
		{
			Description: "Config: a config false is allowed as " +
				"a child of config true",
			Template: BlankTemplate,
			Schema: ` container test {
				config true;
				list testlist {
					config false;
					key name;
					leaf name {
						type string;
						config false;
					}
					leaf foo {
						type string;
					}
					leaf status {
						type string;
						config false;
					}
					container subcontainer {
						config false;
						leaf subleaf {
							type string;
							config false;
						}
					}
				}
			}
			`,
			ExpResult: true,
		},
	}
	runTestCasesFullSchema(t, configFalseWithTrueParentTest)
}

func TestConfigGarbageValueReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true leaf node with config false parent.",
			Template:    BlankTemplate,
			Schema: ` container test {
				config foo;  \\ Not true or false
				leaf testleaf {
					type string;
					config true;
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "parsing \"foo\": invalid syntax",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}

func TestConfigTrueLeafInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true leaf node with config false parent.",
			Template:    BlankTemplate,
			Schema: ` container test {
				config false;
				leaf testleaf {
					type string;
					config true;
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}

func TestConfigTrueLeafListFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true leaf-list node with config false parent.",
			Template:    BlankTemplate,
			Schema: ` container test {
				config false;
				leaf-list testlist {
					type string;
					config true;
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueListInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true list node with config false parent.",
			Template:    BlankTemplate,
			Schema: ` container test {
				config false;
				list testlist {
					key "name";
					leaf name {
						type string;
					}
					leaf testdata {
						type string;
					}
					config true;
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueContainerInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true container node with config false parent.",
			Template:    BlankTemplate,
			Schema: ` container test {
				config false;
				leaf test {
					type string;
				}
				container subcontainer {
					config true;
					leaf subleaf {
						type string;
					}
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueLeafInContainerWithFalseParentContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true leaf node in a child container with " +
				"config false parent.",
			Template: BlankTemplate,
			Schema: ` container test {
				config false;
				leaf test {
					type string;
				}
				container subcontainer {
					leaf subleaf {
						type string;
						config true;
					}
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueGroupingUsedInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true node within a grouping used in a " +
				"config false container.",
			Template: BlankTemplate,
			Schema: `grouping target {
				leaf testleaf1 {
					type string;
				}
				leaf testleaf2 {
					type string;
					config true;
				}
			}
			container test {
				config false;
				leaf test {
					type string;
				}
				uses target;
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueInAugmentOfFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true node within an augment of a " +
				"config false container.",
			Template: BlankTemplate,
			Schema: `container test {
				config false;
				leaf test {
					type string;
				}
			}
			augment /test {
				leaf testleaf {
					type string;
					config true;
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueRefineLeafInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true refine of a leaf in a " +
				"config false container.",
			Template: BlankTemplate,
			Schema: `grouping target {
				leaf testleaf1 {
					type string;
				}
				leaf testleaf2 {
					type string;
				}
			}
			container test {
				config false;
				leaf test {
					type string;
				}
				uses target {
					refine testleaf2 {
						config true;
					}
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueRefineLeafListInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true refine of a leaf-list in a " +
				"config false container.",
			Template: BlankTemplate,
			Schema: `grouping target {
				leaf testleaf1 {
					type string;
				}
				leaf-list testleaf2 {
					type string;
				}
			}
			container test {
				config false;
				leaf test {
					type string;
				}
				uses target {
					refine testleaf2 {
						config true;
					}
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueRefineOfListInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true refine of a list in a " +
				"config false container.",
			Template: BlankTemplate,
			Schema: `grouping target {
				leaf testleaf1 {
					type string;
				}
				list testleaf2 {
					key "name";
					leaf name {
						type string;
					}
					leaf data {
						type string;
					}
				}
			}
			container test {
				config false;
				leaf test {
					type string;
				}
				uses target {
					refine testleaf2 {
						config true;
					}
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
func TestConfigTrueRefineInFalseListReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true refine of a leaf in a list in a " +
				"config false container.",
			Template: BlankTemplate,
			Schema: `grouping target {
				leaf testleaf1 {
					type string;
				}
				list testleaf2 {
					key "name";
					leaf name {
						type string;
					}
					leaf data {
						type string;
					}
				}
			}
			container test {
				config false;
				leaf test {
					type string;
				}
				uses target {
					refine testleaf2/data {
						config true;
					}
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}

func TestConfigTrueRefineInFalseContainerReject(t *testing.T) {
	var configFailTest = []TestCase{
		{
			Description: "Config true refine of a container in a " +
				"config false container.",
			Template: BlankTemplate,
			Schema: `grouping target {
				container testcontainer {
					leaf testleaf1 {
						type string;
					}
				}
			}
			container test {
				config false;
				leaf test {
					type string;
				}
				uses target {
					refine testcontainer {
						config true;
					}
				}
			}
			`,
			ExpResult: false,
			ExpErrMsg: "can't have a config false parent",
		},
	}
	runTestCasesFullSchema(t, configFailTest)
}
