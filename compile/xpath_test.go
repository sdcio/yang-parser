// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
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

	"github.com/danos/yang/schema"
	"github.com/danos/yang/testutils"
)

func validateMachine(t *testing.T, expected, actual string) {
	if expected != actual {
		t.Fatalf("--- Expected machine ---\n%s\n--- Actual machine ---\n%s\n",
			expected, actual)
	}
}

const NoPathToEvalMachine = "--- machine start ---\n" +
	"storePathEval\n" +
	"---- machine end ----\n"

func wrapPathEvalText(machineBody string) string {
	return ("--- machine start ---\n" +
		machineBody +
		"storePathEval\n" +
		"---- machine end ----\n")
}

// In theory we can only have one WHEN, but in the case of a 'when' directly
// under an 'augment' we put the WHEN statement onto each child, potentially
// creating a second there.  This second when should have the runAsParent flag
// set.
type whenExp struct {
	machineText  string
	errMsg       string
	pathEvalText string
	runAsParent  bool
}

// Separately, musts may have custom error-messages.
type mustExp struct {
	machineText  string
	errMsg       string
	pathEvalText string
}

func checkWhens(expected []whenExp) checkFn {
	return func(t *testing.T, actual schema.Node) {
		whens := actual.Whens()

		if len(expected) == 0 {
			t.Fatalf("Must specify at least one 'when' statement to check!")
			return
		}

		if len(whens) != len(expected) {
			t.Fatalf("Different number of 'whens': exp %d, got %d\n",
				len(expected), len(whens))
			return
		}

		for index, when := range whens {
			validateMachine(t, expected[index].machineText,
				when.Mach.PrintMachine())
			validateMachine(t, expected[index].pathEvalText,
				when.PathEvalMach.PrintMachine())
			if when.ErrMsg != expected[index].errMsg {
				t.Fatalf("When '%s' errMsg\nExp: '%s'\nGot: '%s'\n",
					when.Mach.GetExpr(),
					expected[index].errMsg, when.ErrMsg)
			}
			if when.RunAsParent != expected[index].runAsParent {
				t.Fatalf("When '%s': runAsParent is %t, exp %t\n",
					when.Mach.GetExpr(), when.RunAsParent,
					expected[index].runAsParent)
			}
		}
	}
}

func checkMusts(expected []mustExp) checkFn {
	return func(t *testing.T, actual schema.Node) {
		musts := actual.Musts()

		if len(expected) == 0 {
			t.Fatalf("Must specify at least one 'must' statement to check!")
		}

		if len(musts) != len(expected) {
			t.Fatalf("Different number of musts: exp %d, got %d\n",
				len(expected), len(musts))
			return
		}

		for index, must := range musts {
			validateMachine(t, expected[index].machineText,
				must.Mach.PrintMachine())
			validateMachine(t, expected[index].pathEvalText,
				must.PathEvalMach.PrintMachine())
			if must.ErrMsg != expected[index].errMsg {
				t.Fatalf("Must '%s' errMsg\nExp: '%s'\nGot: '%s'\n",
					must.Mach.GetExpr(),
					expected[index].errMsg, must.ErrMsg)
			}
		}
	}
}

// WHEN

func TestXpathWhenChoice(t *testing.T) {
	t.Skipf("TBD when choice is implemented.")
}

func TestXpathWhenContainer(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
		     description "Container with when statement";
             leaf whenLeaf {
                 type string;
             }
             when "count(../interfaces/*) < 10";
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile interfaces}\n" +
			"nameTestPush\t{ *}\n" +
			"evalLocPath\n" +
			"bltin\t\tcount()\n" +
			"numpush\t\t10\n" +
			"lt\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile interfaces}\n" +
			"nameTestPush\t{ *}\n" +
			"locPathExists\n")

	expected := NewContainerChecker(
		"whenContainer",
		[]NodeChecker{
			NewLeafChecker("whenLeaf"),
		},
		checkWhens([]whenExp{
			{
				expMachine,
				"'when' condition is false: " +
					"'count(../interfaces/*) < 10'",
				expPathEvalMachine,
				false}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"whenContainer"})
	expected.check(t, actual)
}

func TestXpathWhenLeaf(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
		     description "Container with when statement";
             leaf whenLeaf {
                 type string;
                 when "count(../../interfaces/*) < 20";
             }
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile interfaces}\n" +
			"nameTestPush\t{ *}\n" +
			"evalLocPath\n" +
			"bltin\t\tcount()\n" +
			"numpush\t\t20\n" +
			"lt\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"pathOperPush\t..\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile interfaces}\n" +
			"nameTestPush\t{ *}\n" +
			"locPathExists\n")

	expected := NewLeafChecker(
		"whenLeaf",
		checkWhens([]whenExp{{expMachine,
			"'when' condition is false: " +
				"'count(../../interfaces/*) < 20'",
			expPathEvalMachine, false}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"whenContainer", "whenLeaf"})
	expected.check(t, actual)
}

func TestXpathWhenLeafList(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
		     description "Container with when statement";
             leaf-list whenLeafList {
                 type string;
                 when "../anotherLeaf";
             }
             leaf anotherLeaf {
                 type string;
             }
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile anotherLeaf}\n" +
			"evalLocPath\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile anotherLeaf}\n" +
			"locPathExists\n")

	expected := NewLeafListChecker(
		"whenLeafList",
		checkWhens([]whenExp{{expMachine,
			"'when' condition is false: " +
				"'../anotherLeaf'",
			expPathEvalMachine, false}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"whenContainer", "whenLeafList"})
	expected.check(t, actual)
}

func TestXpathWhenList(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
			description "Container with when statement";
			list whenList {
				key "keyLeaf";
				leaf keyLeaf {
					type string;
				}
				leaf anotherLeaf {
					type string;
				}
				when "../yetAnotherLeaf = 1234";
			}
			leaf yetAnotherLeaf {
				type uint32;
			}
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile yetAnotherLeaf}\n" +
			"evalLocPath\n" +
			"numpush\t\t1234\n" +
			"eq\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile yetAnotherLeaf}\n" +
			"locPathExists\n")

	expected := NewListChecker(
		"whenList",
		[]NodeChecker{
			NewKeyChecker("keyLeaf"),
			NewLeafChecker("anotherLeaf"),
		},
		checkWhens([]whenExp{{expMachine,
			"'when' condition is false: " +
				"'../yetAnotherLeaf = 1234'",
			expPathEvalMachine, false}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"whenContainer", "whenList"})
	expected.check(t, actual)
}

// Check 'runAsParent' set only on augment when, not on leaf
func TestXpathWhenAugmentAndLeaf(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
			description "Container with when statement";
			leaf whenLeaf {
				type string;
			}
		}
		augment /whenContainer {
			description "Augment with top-level when";
			when "false()"; // Content irrelevant here too.
			leaf augmentLeaf {
				type string;
				when "true()"; // Content irrelevant here.
			}
		}`))

	expWhens := []whenExp{
		{
			machineText: "--- machine start ---\n" +
				"bltin\t\ttrue()\n" +
				"store\n" +
				"---- machine end ----\n",
			errMsg:       "'when' condition is false: 'true()'",
			pathEvalText: NoPathToEvalMachine,
			runAsParent:  false,
		},
		{
			machineText: "--- machine start ---\n" +
				"bltin\t\tfalse()\n" +
				"store\n" +
				"---- machine end ----\n",
			errMsg:       "'when' condition is false: 'false()'",
			pathEvalText: NoPathToEvalMachine,
			runAsParent:  true,
		},
	}

	expected := NewLeafChecker(
		"augmentLeaf",
		checkWhens(expWhens))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"whenContainer", "augmentLeaf"})
	expected.check(t, actual)
}

// Check 'runAsParent' set only on augment when
func TestXpathWhenAugmentAndGrouping(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
			description "Container with when statement";
			leaf whenLeaf {
				type string;
			}
		}

		grouping whenGroup {
			leaf groupLeaf {
				type string;
				when "true()"; // Content irrelevant here.
			}
		}

		augment /whenContainer {
			description "Augment with top-level when";
			when "false()"; // Content irrelevant here too.
			uses whenGroup;
		}`))

	expWhens := []whenExp{
		{
			machineText: "--- machine start ---\n" +
				"bltin\t\ttrue()\n" +
				"store\n" +
				"---- machine end ----\n",
			errMsg:       "'when' condition is false: 'true()'",
			pathEvalText: NoPathToEvalMachine,
			runAsParent:  false,
		},
		{
			machineText: "--- machine start ---\n" +
				"bltin\t\tfalse()\n" +
				"store\n" +
				"---- machine end ----\n",
			errMsg:       "'when' condition is false: 'false()'",
			pathEvalText: NoPathToEvalMachine,
			runAsParent:  true,
		},
	}

	expected := NewLeafChecker(
		"groupLeaf",
		checkWhens(expWhens))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"whenContainer", "groupLeaf"})
	expected.check(t, actual)
}

// MUST

func TestXpathMustChoice(t *testing.T) {
	t.Skipf("TBD when choice is implemented.")
}

func TestXpathMustContainer(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container mustContainer {
		     description "Container with must statement";
             leaf mustLeaf {
                 type string;
             }
             must "mustLeaf and contains(mustLeaf, 'hello world')";
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
			"evalLocPath\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
			"evalLocPath\n" +
			"litpush\t\t'hello world'\n" +
			"bltin\t\tcontains()\n" +
			"and\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
			"locPathExists\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
			"locPathExists\n")

	expected := NewContainerChecker(
		"mustContainer",
		[]NodeChecker{
			NewLeafChecker("mustLeaf"),
		},
		checkMusts([]mustExp{
			{
				expMachine,
				"'must' condition is false: 'mustLeaf and " +
					"contains(mustLeaf, 'hello world')'",
				expPathEvalMachine}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"mustContainer"})
	expected.check(t, actual)
}

func TestXpathMustLeaf(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container mustContainer {
		     description "Container with must statement";
             leaf mustLeaf {
                 type string;
                 must "../anotherLeaf";
             }
             leaf anotherLeaf {
                 type string;
             }
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile anotherLeaf}\n" +
			"evalLocPath\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile anotherLeaf}\n" +
			"locPathExists\n")

	expected := NewLeafChecker(
		"mustLeaf",
		checkMusts([]mustExp{
			{
				expMachine, "'must' condition is false: '../anotherLeaf'",
				expPathEvalMachine,
			}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"mustContainer", "mustLeaf"})
	expected.check(t, actual)
}

func TestXpathMustLeafList(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container mustContainer {
		     description "Container with must statement";
             leaf-list mustLeafList {
                 type string;
                 must "../anotherLeaf";
             }
             leaf anotherLeaf {
                 type string;
             }
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile anotherLeaf}\n" +
			"evalLocPath\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile anotherLeaf}\n" +
			"locPathExists\n")

	expected := NewLeafListChecker(
		"mustLeafList",
		checkMusts([]mustExp{
			{
				expMachine, "'must' condition is false: '../anotherLeaf'",
				expPathEvalMachine,
			}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"mustContainer", "mustLeafList"})
	expected.check(t, actual)
}

func TestXpathMustList(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container mustContainer {
			description "Container with must statement";
			list mustList {
				key "keyLeaf";
				leaf keyLeaf {
					type string;
				}
				leaf anotherLeaf {
					type string;
				}
				must "../yetAnotherLeaf";
			}
			leaf yetAnotherLeaf {
				type uint32;
			}
         }`))

	expMachine :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile yetAnotherLeaf}\n" +
			"evalLocPath\n" +
			"store\n" +
			"---- machine end ----\n"

	expPathEvalMachine := wrapPathEvalText(
		"pathOperPush\t..\n" +
			"nameTestPush\t{urn:vyatta.com:test:yang-compile yetAnotherLeaf}\n" +
			"locPathExists\n")

	expected := NewListChecker(
		"mustList",
		[]NodeChecker{
			NewKeyChecker("keyLeaf"),
			NewLeafChecker("anotherLeaf"),
		},
		checkMusts([]mustExp{
			{
				expMachine,
				"'must' condition is false: '../yetAnotherLeaf'",
				expPathEvalMachine,
			}}))

	actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"mustContainer", "mustList"})
	expected.check(t, actual)
}

// Error handling - make sure we get the expected error message(s).
func TestXpathWhenError(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
		     description "Container with when statement";
             leaf whenLeaf {
                 type string;
             }
             when "unknownFn() = count(../interfaces)";
         }`))

	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err,
		"Failed to compile 'unknownFn() = count(../interfaces)'\n",
		"Parse Error: syntax error\n",
		"Got to approx [X] in 'unknownFn [X] () = count(../interfaces)'\n",
		"Lexer Error: Unknown function or node type: 'unknownFn'")
}

func TestXpathMustError(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container whenContainer {
		     description "Container with when statement";
             leaf whenLeaf {
                 type string;
                 must "../auth/";
             }
         }`))

	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err,
		"Failed to compile '../auth/'\n",
		"Parse Error: syntax error\n",
		"Got to approx [X] in '../auth/ [X] '")

}

// Check 2 different error messages get parsed and assigned correctly.
func TestXpathMustCustomErrors(t *testing.T) {
	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate,
		`container mustContainer {
			description "Container with must statements";
			leaf mustLeaf {
				type uint8;
			}
			leaf mustLeaf2 {
				type uint8;
			}
			must "mustLeaf < 8" {
				error-message "mustLeaf must have value < 8";
			}
			must "mustLeaf2 > mustLeaf" {
				error-message "mustLeaf2 must have value > mustLeaf";
			}
		}`))

	expMusts := []mustExp{
		{
			machineText: "--- machine start ---\n" +
				"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
				"evalLocPath\n" +
				"numpush\t\t8\n" +
				"lt\n" +
				"store\n" +
				"---- machine end ----\n",
			pathEvalText: wrapPathEvalText(
				"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
					"locPathExists\n"),
			errMsg: "mustLeaf must have value < 8",
		},
		{
			machineText: "--- machine start ---\n" +
				"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf2}\n" +
				"evalLocPath\n" +
				"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
				"evalLocPath\n" +
				"gt\n" +
				"store\n" +
				"---- machine end ----\n",
			pathEvalText: wrapPathEvalText(
				"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf2}\n" +
					"locPathExists\n" +
					"nameTestPush\t{urn:vyatta.com:test:yang-compile mustLeaf}\n" +
					"locPathExists\n"),
			errMsg: "mustLeaf2 must have value > mustLeaf",
		},
	}

	expected := NewContainerChecker(
		"mustContainer",
		[]NodeChecker{
			NewLeafChecker("mustLeaf"),
			NewLeafChecker("mustLeaf2"),
		},
		checkMusts(expMusts))

	if actual := getSchemaNodeFromPath(t, schema_text,
		[]string{"mustContainer"}); actual != nil {
		expected.check(t, actual)
	}
}
