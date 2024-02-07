// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Utility functions required by the parse unit tests.

package parsetest

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	. "github.com/sdcio/yang-parser/parse"
	"github.com/sdcio/yang-parser/testutils"
)

// Parse a yang schema file
func ParseSchemaFile(fname string) (*Tree, error) {
	text, err := ioutil.ReadFile(fname)
	if err != nil && err != io.EOF {
		return nil, err
	}
	t, err := Parse(fname, string(text), nil)
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Parse a yang file, which is expected to contain a parsing error
// Check that the reported error is the expected error.
func VerifyParseErrorIsSeen(t *testing.T, fname, expectedError string) {
	_, err := ParseSchemaFile(fname)

	if err == nil {
		t.Errorf("Unexpected Parse Success: %s", fname)
		testutils.LogStack(t)
	} else if strings.Index(err.Error(), expectedError) == -1 {
		t.Errorf("Expected error was not seen:")
		t.Logf("Observed error: %s", err.Error())
		t.Logf("Should contain: %s", expectedError)
		testutils.LogStack(t)
	}
}

// Currently only supports arguments of top level nodes.
func VerifyExpectedArgument(t *testing.T, fname string, ntype NodeType, exp string) {
	tree, err := ParseSchemaFile(fname)
	if err != nil {
		t.Errorf("Unable to parse file: %s", fname)
		t.Log(err)
		testutils.LogStack(t)
		return
	}

	st := tree.Root.ChildByType(ntype).Argument().String()
	if st != exp {
		t.Errorf("Argument %s does not match expected.", ntype)
		t.Logf("Received:\n%q", st)
		t.Logf("Expected:\n%q", exp)
		testutils.LogStack(t)
	}
}
