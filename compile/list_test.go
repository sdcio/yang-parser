// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"testing"

	"github.com/steiler/yang-parser/schema"
	"github.com/steiler/yang-parser/testutils"
)

func TestKeyShouldNotBeInUniqueStmt(t *testing.T) {
	t.Skipf("Treating this as warning not error for now ...")
	var KeyNotInUniqueStmtTest = []testutils.TestCase{
		{
			Description: "ListKey: Key should not be in unique statement",
			Template:    ListTemplate,
			Schema: `leaf testKey {
			type string;
		    }
		    unique testKey;`,
			ExpResult: false,
			ExpErrMsg: "List Key must not form part of unique statement",
		},
	}

	runTestCases(t, KeyNotInUniqueStmtTest)
}

// Test nested key works.
func TestNestedKeyAccepted(t *testing.T) {
	var NestedKeyAccepted = []testutils.TestCase{
		{
			Description: "ListKey: Nested key used in unique statement.",
			Template:    ListTemplateNested,
			Schema:      "",
			ExpResult:   true,
		},
	}

	runTestCases(t, NestedKeyAccepted)
}

// Test nested key within container used as unique key fails.
func TestNestedKeyShouldNotBeInUniqueStmt(t *testing.T) {
	t.Skipf("Treating this as warning not error for now ...")
	var NestedKeyNotInUniqueStmt = []testutils.TestCase{
		{
			Description: "ListKey: Nested key used in unique statement.",
			Template:    ListTemplateNested,
			Schema:      `unique "listContainer/uniqueLeaf";`,
			ExpResult:   false,
			ExpErrMsg:   "List Key must not form part of unique statement",
		},
	}

	runTestCases(t, NestedKeyNotInUniqueStmt)
}

func TestCannotRefineUniqueForList(t *testing.T) {
	var RefineUniqueForList = []testutils.TestCase{
		{
			Description: "ListKey: Cannot refine unique statement for list.",
			Template:    ListRefineTemplate,
			Schema: `container realContainer {
				uses target {
                    refine test_container/testList {
                        unique "leaf1 leaf2 listKey";
                    }
                }
			}`,
			ExpResult: false,
			ExpErrMsg: "refine test_container/testList: " +
				"invalid refinement unique for statement list",
		},
	}

	runTestCases(t, RefineUniqueForList)
}

// Check that a schema node has the expected default and mandatory values
func checkNodeMandatoryAndDefault(t *testing.T, st schema.ModelSet,
	path []string, mand, def bool,
	defValue string) {
	node := findSchemaNodeInTree(t, st, path)
	if node == nil {
		t.Fatalf("Unable to find node")
	}
	var sn schema.Node
	sn = node
	switch lf := sn.(type) {
	case schema.Leaf:
		str, bl := lf.Default()
		if def {
			if !bl || defValue != str {
				t.Fatalf("Default value not as expected")
			}
		} else if bl {
			t.Fatalf("Unexpected default encountered")
		}
		if lf.Mandatory() != mand {
			t.Fatalf("Unexpected mandatory value encountered\n"+
				"Expected: %t\nGot: %t", mand, lf.Mandatory())
		}
	default:
		t.Fatalf("Unexpected node type encountered")
	}

}

// Test that mandatory and default statements on a leaf which is a list key
// are ignored and do not appear on the compiled schema node.
func TestListKeyDefaultIgnored(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile {
                namespace "urn:vyatta.com:test:yang-compile";
                prefix test;

                organization "Brocade Communications Systems, Inc.";
                revision 2015-09-28 {
                        description "Test schema";
                }

		container test {
			list firstlist {
				key "testkey";
				leaf testkey {
					type string;
					default "IgnoredDefault";
				}
				leaf testleaf {
					type string;
					default "RetainDefault";
				}
			}

			list secondlist {
				key "testkey";
				leaf testkey {
					type string;
					mandatory true;
				}
				leaf testleaf {
					type string;
					mandatory true;
				}
			}
		}
        }`)

	st, err := testutils.GetConfigSchema(module_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	// Check that firstlist testkey has no default, even though
	// specified in the Yang
	checkNodeMandatoryAndDefault(t, st,
		[]string{"test", "firstlist", "testkey"},
		false, false, "")
	checkNodeMandatoryAndDefault(t, st,
		[]string{"test", "firstlist", "testleaf"},
		false, true, "RetainDefault")
	// Check secondlist testkey is mandatory false, even though the
	// yang stated mandatory true
	checkNodeMandatoryAndDefault(t, st,
		[]string{"test", "secondlist", "testkey"},
		false, false, "")
	checkNodeMandatoryAndDefault(t, st,
		[]string{"test", "secondlist", "testleaf"},
		true, false, "")
}

func TestListUniquePathToLeafCompiles(t *testing.T) {
	ListPassUniqueToLeaf := []testutils.TestCase{
		{
			Description: "List: Unique path to non-empty leaf",
			Template:    ListTemplate,
			Schema: `leaf testKey {
				type string;
			}
			unique server/port;
			container server {
				leaf port {
					type uint32;
				}
			}`,
			ExpResult: true,
		},
	}
	runTestCases(t, ListPassUniqueToLeaf)
}

func TestListOrderedByUserKeywordCompiles(t *testing.T) {
	ListPassOrderedByUser := []testutils.TestCase{
		{
			Description: "List: Support 'ordered-by user'",
			Template:    ListTemplate,
			Schema: `ordered-by user;
			leaf testKey {
				type string;
			}`,
			ExpResult: true,
		},
	}
	runTestCases(t, ListPassOrderedByUser)
}

var ListFail = []testutils.TestCase{
	{
		Description: "List: Unique path to non-leaf",
		Template:    ListTemplate,
		Schema: `leaf testKey {
			type string;
		}
		unique server/flags;
		container server {
			leaf-list flags {
				type string;
			}
		}`,
		ExpResult: false,
		ExpErrMsg: "non leaf descendant",
	},
	{
		Description: "List: Unique path to empty leaf",
		Template:    ListTemplate,
		Schema: `leaf testKey {
			type string;
		}
		unique server/flag;
		container server {
			leaf flag {
				type empty;
			}
		}`,
		ExpResult: false,
		ExpErrMsg: "empty leaf descendant",
	},
	{
		Description: "List: Unique path through nested list",
		Template:    ListTemplate,
		Schema: `leaf testKey {
			type string;
		}
		unique testcontainer/server/port;
		container testcontainer {
			list server {
				key name;
				leaf name {
					type string;
				}
				leaf port {
					type uint32;
				}
			}
		}`,
		ExpResult: false,
		ExpErrMsg: "list descendant",
	},
}

func TestListFail(t *testing.T) {
	runTestCases(t, ListFail)
}
