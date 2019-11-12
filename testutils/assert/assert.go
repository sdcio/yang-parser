// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Useful test functions for validating (mostly) string outputs match
// what is expected.

package assert

import (
	"bytes"
	"strings"
	"testing"
)

type expectedError struct {
	expected string
}

func NewExpectedError(expect string) *expectedError {
	return &expectedError{expected: expect}
}

func (e *expectedError) Matches(t *testing.T, actual error) {
	if actual == nil {
		t.Fatalf("Unexpected success")
	}

	CheckStringDivergence(t, e.expected, actual.Error())
}

type ExpectedMessages struct {
	expected []string
}

func NewExpectedMessages(expect ...string) *ExpectedMessages {
	return &ExpectedMessages{expected: expect}
}

func (e *ExpectedMessages) ContainedIn(t *testing.T, actual string) {
	if len(actual) == 0 {
		t.Fatalf("No output in which to search for expected message(s).")
		return
	}

	for _, exp := range e.expected {
		if !strings.Contains(actual, exp) {
			t.Fatalf("Actual output doesn't contain expected output:\n"+
				"Exp:\n%s\nAct:\n%v\n", exp, actual)
		}
	}
}

func (e *ExpectedMessages) NotContainedIn(t *testing.T, actual string) {
	if len(actual) == 0 {
		t.Fatalf("No output in which to search for expected message(s).")
		return
	}

	for _, exp := range e.expected {
		if strings.Contains(actual, exp) {
			t.Fatalf("Actual output contain unexpected output:\n"+
				"NotExp:\n%s\nAct:\n%v\n", exp, actual)
		}
	}
}

// Check each expected message appears in at least one of the actual strings.
func (e *ExpectedMessages) ContainedInAny(t *testing.T, actual []string) {
	if len(actual) == 0 {
		t.Fatalf("No output in which to search for expected message(s).")
		return
	}

outerLoop:
	for _, exp := range e.expected {
		for _, act := range actual {
			if strings.Contains(act, exp) {
				continue outerLoop
			}
		}

		t.Fatalf("Actual output doesn't contain expected output:\n"+
			"Exp:\n%s\nAct:\n%v\n", exp, actual)
	}
}

// Very useful when debugging outputs that don't match up.
func CheckStringDivergence(t *testing.T, expOut, actOut string) {
	if expOut == actOut {
		return
	}

	var expOutCopy = expOut
	var act bytes.Buffer
	var charsToDump = 10
	var expCharsToDump = 10
	var actCharsLeft, expCharsLeft int
	for index, char := range actOut {
		if len(expOutCopy) > 0 {
			if char == rune(expOutCopy[0]) {
				act.WriteByte(byte(char))
			} else {
				act.WriteString("###") // Mark point of divergence.
				expCharsLeft = len(expOutCopy)
				actCharsLeft = len(actOut) - index
				if expCharsLeft < charsToDump {
					expCharsToDump = expCharsLeft
				}
				if actCharsLeft < charsToDump {
					charsToDump = actCharsLeft
				}
				act.WriteString(actOut[index : index+charsToDump])
				break
			}
		} else {
			t.Logf("Expected output terminates early.\n")
			t.Fatalf("Exp:\n%s\nGot extra:\n%s\n",
				expOut[:index], act.String()[index:])
		}
		expOutCopy = expOutCopy[1:]
	}

	// Useful to print whole output first for reference (useful when debugging
	// when you don't want to have to construct the expected output up front).
	t.Logf("Actual output:\n%s\n--- ENDS ---\n", actOut)

	// After that we then print up to the point of divergence so it's easy to
	// work out what went wrong ...
	t.Fatalf("Unexpected output.\nGot:\n%s\nExp at ###:\n'%s ...'\n",
		act.String(), expOutCopy[:expCharsToDump])
}
