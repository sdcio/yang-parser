// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This suite of tests verifies the basic function of the path_eval
// grammar which is designed to extract paths from XPATH statements
// and discard anything else.
//
// Initial version does not parse paths in predicates, or deal with
// functions inside paths - that's for the next revision ...

package path_eval

import (
	"testing"
)

// Add in a decent subset of elements that we are ignoring to check that they
// do get ignored, and don't cause parser errors.
func TestMachineNoPaths(t *testing.T) {
	testMachine, _ := NewPathEvalMachine(
		"(10 + number(substring('1234', 1, 2))) or (10 <= 6)",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

func TestMachineSimpleAbsolutePath(t *testing.T) {
	testMachine, _ := NewPathEvalMachine("/interfaces/dataplane/address",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush\t/ (2f)\n" +
			"nameTestPush\t{ interfaces}\n" +
			"nameTestPush\t{ dataplane}\n" +
			"nameTestPush\t{ address}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

// TBD: change processing thus:
// evalLocPathCanExist: can/does this generate immediate error?
//  - can't generate error at compile as we may not have all yang by now!
// stack true / false on stack
// store changes to check all stack values (if any) and store false if
// any stack value is false.
// func should store 'something' (well known value) if, and only if, inside
// a path ... so we need tracking for path start / end ...
// prob want to rename bltin as needs to run different functionality...
// (notes path may or may not be valid ...)
//
// Once all compiled, need to walk schema (not config!) tree and run all
// pathValidate machines.

func TestMachineSimpleRelativePath(t *testing.T) {
	testMachine, _ := NewPathEvalMachine("../dataplane/address",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{ dataplane}\n" +
			"nameTestPush\t{ address}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

func TestMachinePathWithPredicate(t *testing.T) {
	testMachine, _ := NewPathEvalMachine(
		"../dataplane[tagnode='dp0s2']/address",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush\t..\n" +
			"nameTestPush\t{ dataplane}\n" +
			"nameTestPush\t{ address}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

// For now, ALL functions, even in paths, are ignored.  The problem is
// working out how to spot they are at the start of a path rather than
// used in other scenarios.
func TestMachinePathWithFunction(t *testing.T) {
	testMachine, _ := NewPathEvalMachine(
		"current()/interfaces/address",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"nameTestPush\t{ interfaces}\n" +
			"nameTestPush\t{ address}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

func TestMachineFunctionWithPath(t *testing.T) {
	testMachine, _ := NewPathEvalMachine(
		"current() = 2 or /interfaces/dataplane",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush\t/ (2f)\n" +
			"nameTestPush\t{ interfaces}\n" +
			"nameTestPush\t{ dataplane}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

func TestMachinePathInsideFunction(t *testing.T) {
	testMachine, _ := NewPathEvalMachine(
		"count(/interfaces/dataplane/tagnode) = 0",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush\t/ (2f)\n" +
			"nameTestPush\t{ interfaces}\n" +
			"nameTestPush\t{ dataplane}\n" +
			"nameTestPush\t{ tagnode}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

func TestMachinePathFunctionWithPath(t *testing.T) {
	testMachine, _ := NewPathEvalMachine(
		// Path isn't meant to exist - just designed to test the specific
		// combination of elements: func(path)/path
		"count(/interfaces/dataplane)/service/ssh/port",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush\t/ (2f)\n" +
			"nameTestPush\t{ interfaces}\n" +
			"nameTestPush\t{ dataplane}\n" +
			"locPathExists\n" +
			"nameTestPush\t{ service}\n" +
			"nameTestPush\t{ ssh}\n" +
			"nameTestPush\t{ port}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

func TestFunctionWithPathEmbeddedInPath(t *testing.T) {
	t.Skipf("Can we have function in the middle of a path, not just at start")
}

func TestMachineMultiplePaths(t *testing.T) {
	testMachine, _ := NewPathEvalMachine(
		"/interfaces or count(/interfaces/dataplane/tagnode) = 0",
		nil, "(no location)")

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush\t/ (2f)\n" +
			"nameTestPush\t{ interfaces}\n" +
			"locPathExists\n" +
			"pathOperPush\t/ (2f)\n" +
			"nameTestPush\t{ interfaces}\n" +
			"nameTestPush\t{ dataplane}\n" +
			"nameTestPush\t{ tagnode}\n" +
			"locPathExists\n" +
			"storePathEval\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}
