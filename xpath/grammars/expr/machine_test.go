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

// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This suite of tests differs from the parser_test suite.  The latter
// checks expressions are parsed and evaluated correctly, and that parsing
// errors are caught.  This set of tests check that the internals of the
// machine construction and execution work correctly.  There is overlap,
// but the focus is different, and concentrates as much on error handling
// as normal operation.

package expr

import (
	"testing"
)

// Check all valid options in a machine are printed correctly.
// Ensures this function keeps working in case it is needed for debug!
func TestMachinePrint(t *testing.T) {
	testMachine, _ := NewExprMachine("10 + number(substring('1234', 1, 2))",
		nil)

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"numpush\t\t10\n" +
			"litpush\t\t'1234'\n" +
			"numpush\t\t1\n" +
			"numpush\t\t2\n" +
			"bltin\t\tsubstring()\n" +
			"bltin\t\tnumber()\n" +
			"add\n" +
			"store\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}
