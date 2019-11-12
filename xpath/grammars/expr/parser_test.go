// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// These tests verify that the basic grammar constructs work as
// expected.  They neither delve into the internals of machines (see
// machine_test.go) nor do they extend to tests of more complex
// scenarios where we need a context/config (see xpath_test.go).
//
// Functions and nodesets are tested in separate files to keep like with
// like and reduce individual file size.  There is of course some overlap!

package expr

import (
	"github.com/danos/yang/xpath"
	"math"
	"testing"
)

// Test Cases
func TestParseBlank(t *testing.T) {
	checkParseError(t, "", []string{
		"Empty XPATH expression has no value."})
}

// Simple number parsing
func TestParseNumber(t *testing.T) {
	checkNumResult(t, "123.4E3", 123400)
}

func TestParseUnaryMinus(t *testing.T) {
	checkNumResult(t, "-66", -66)
}

// Arithmetic operations
func TestAddNumber(t *testing.T) {
	checkNumResult(t, "123 + 456", 579)
}

func TestParseSubtractNumber(t *testing.T) {
	checkNumResult(t, "123 - 456", -333)
}

func TestParseMultiplyNumber(t *testing.T) {
	checkNumResult(t, "123 * 456", 56088)
}

func TestParseDivideNumber(t *testing.T) {
	checkNumResult(t, "126 div 4", 31.5)
	checkNumResult(t, "4 div 0", math.Inf(1))
}

func TestParseModNumber(t *testing.T) {
	// Examples from XPATH spec (-:
	checkNumResult(t, "5 mod 2", 1)
	checkNumResult(t, "5 mod -2", 1)
	checkNumResult(t, "-5 mod 2", -1)
	checkNumResult(t, "-5 mod -2", -1)
}

func TestParseParentheses(t *testing.T) {
	checkNumResult(t, "(124 div 4)", 31)
}

// BODMAS
func TestParseBodmas1(t *testing.T) {
	checkNumResult(t, "1 + 2 div 4", 1.5)
}

func TestParseBodmas2(t *testing.T) {
	checkNumResult(t, "(1 + 2) div 4", 0.75)
}

func TestParseBodmas3(t *testing.T) {
	checkNumResult(t, "1 div 2 * 3 div (7+5)", 0.125)
}

func TestParseBodmas4(t *testing.T) {
	checkNumResult(t, "1 div (2 * 3) div (1 div (7+5))", 2)
}

// Relational operators
func TestParseAnd(t *testing.T) {
	checkBoolResult(t, "2 and 3", true)  // Both non-zero
	checkBoolResult(t, "2 and -3", true) // Negative number
	checkBoolResult(t, "4 and 0", false) // Last zero
	checkBoolResult(t, "0 and 1", false) // First zero
	checkBoolResult(t, "0 and 0", false) // Both zero
	checkBoolResult(t, "boolean(2) and 0", false)
	checkBoolResult(t, "boolean(-0) and 1", false)
}

func TestParseOr(t *testing.T) {
	checkBoolResult(t, "1 or 4", true)   // Both non-zero
	checkBoolResult(t, "-1 or 1", true)  // Negative number
	checkBoolResult(t, "-0 or 1", true)  // First zero
	checkBoolResult(t, "2 or 0", true)   // Second zero
	checkBoolResult(t, "0 or -0", false) // Both zero
	checkBoolResult(t, "0 or boolean(2)", true)
	checkBoolResult(t, "boolean(0) or 1", true)
}

func TestParseOrEquals(t *testing.T) {
	checkBoolResult(t, "(0 or 2) = (1 or 0)", true)
	checkBoolResult(t, "(0 or 2) != (1 or 0)", false)
}

func TestParseBoolAddition(t *testing.T) {
	checkNumResult(t, "(1 or 2) + (6 and 0)", 1)
}

func TestParseRelationals(t *testing.T) {
	checkBoolResult(t, "1 < 2", true)
	checkBoolResult(t, "3 < 2", false)
	checkBoolResult(t, "2 < 2", false)

	checkBoolResult(t, "1 <= 2", true)
	checkBoolResult(t, "3 <= 2", false)
	checkBoolResult(t, "2 <= 2", true)

	checkBoolResult(t, "1 > 2", false)
	checkBoolResult(t, "3 > 2", true)
	checkBoolResult(t, "2 > 2", false)

	checkBoolResult(t, "1 >= 2", false)
	checkBoolResult(t, "3 >= 2", true)
	checkBoolResult(t, "2 >= 2", true)
}

func TestParseEqualityOperators(t *testing.T) {
	// Numbers
	checkBoolResult(t, "10 = 10", true)
	checkBoolResult(t, "11 = 10", false)
	checkBoolResult(t, "10 != 10", false)
	checkBoolResult(t, "11 != 10", true)

	// Boolean
	checkBoolResult(t, "boolean(2) = boolean(3)", true)
	checkBoolResult(t, "boolean(2) != boolean(3)", false)
	checkBoolResult(t, "boolean(0) != boolean(3)", true)
	checkBoolResult(t, "boolean(0) = boolean(3)", false)

	// Strings
	checkBoolResult(t, "'some text' = \"some text\"", true)
	checkBoolResult(t, "'some text' != \"some text\"", false)
	checkBoolResult(t, "'some text' != 'other text'", true)
	checkBoolResult(t, "'some text' = 'other text'", false)
}

// When performing '=' or '!=' on mixed types, and neither type is a nodeset,
// then we convert to bool if either is bool, else as numbers if at least one
// is a number, otherwise compare as strings.
func TestParseEqualityMixedTypes(t *testing.T) {
	// Literal vs bool
	checkBoolResult(t, "'10' = true()", true)
	checkBoolResult(t, "'' != true()", true)

	// Number vs bool
	checkBoolResult(t, "-5 = false()", false)
	checkBoolResult(t, "true() = 0", false)

	// Literal vs number
	checkBoolResult(t, "'foo' = 10", false)
	checkBoolResult(t, "666 = '666'", true)
}

func TestParseEqualitySpecialNumbers(t *testing.T) {
	// These special numeric types can be a bit funny, so just sanity check
	// we give correct answers for equality operations.  We use the number()
	// function to generate the float64 from the given string as XPATH does
	// not directly parse 'NaN' or 'Infinity'.
	checkBoolResult(t, "0 = -0", true)
	checkBoolResult(t, "-0 != 0", false)

	checkBoolResult(t, "number('NaN') = number('NaN')", false)
	checkBoolResult(t, "number('NaN') != number('NaN')", true)

	checkBoolResult(t, "number('Infinity') = number('Infinity')", true)
	checkBoolResult(t, "number('Infinity') != number('Infinity')", false)

	checkBoolResult(t, "number('-Infinity') = number('-Infinity')", true)
	checkBoolResult(t, "number('-Infinity') != number('-Infinity')", false)

	checkBoolResult(t, "number('-Infinity') = number('Infinity')", false)
	checkBoolResult(t, "number('Infinity') != number('-Infinity')", true)
}

func TestParseLiteral(t *testing.T) {
	checkLiteralResult(t, "'some text'", "some text")
	checkLiteralResult(t, "\"more text\"", "more text")
	checkLiteralResult(t, "'text containing \"double quotes\" inside it'",
		"text containing \"double quotes\" inside it")
	checkLiteralResult(t, "\"text containing 'single quotes' inside it\"",
		"text containing 'single quotes' inside it")
}

// Check we get same result if we run an expression twice.  Helps ensure
// we are initialising correctly / cleaning up afterwards.

func TestMultipleMachines(t *testing.T) {
	expr1 := "10+number(substring('3456', 1, 2))"
	mach1, err1 := NewExprMachine(expr1, nil)
	if err1 != nil {
		t.Fatalf("Unexpected error creating machine1: %s\n", err1.Error())
		return
	}

	res1 := xpath.NewCtxFromMach(mach1, nil).EnableValidation().Run()
	actResult1, err1 := res1.GetNumResult()
	if err1 != nil {
		t.Fatalf("Unexpected error getting result for %s: %s\n",
			expr1, err1.Error())
		return
	}
	if actResult1 != 44 {
		t.Fatalf("Wrong result for machine1. Exp 44, got %v\n", actResult1)
		return
	}

	expr2 := "15 + 2 * 3"
	mach2, err2 := NewExprMachine(expr2, nil)
	if err2 != nil {
		t.Fatalf("Unexpected error creating machine2: %s\n", err2.Error())
		return
	}

	res2 := xpath.NewCtxFromMach(mach2, nil).EnableValidation().Run()
	actResult2, err2 := res2.GetNumResult()
	if err2 != nil {
		t.Fatalf("Unexpected error getting result for %s: %s\n",
			expr2, err2.Error())
		return
	}

	if actResult2 != 21 {
		t.Fatalf("Wrong result for machine2.  Exp 21, got %v\n", actResult2)
		return
	}

	res1 = xpath.NewCtxFromMach(mach1, nil).EnableValidation().Run()
	actResult1_2, _ := res1.GetNumResult()
	if actResult1_2 != actResult1 {
		t.Fatalf("Machine 1 result not repeatable!")
		return
	}

	res2 = xpath.NewCtxFromMach(mach2, nil).EnableValidation().Run()
	actResult2_2, _ := res2.GetNumResult()
	if actResult2_2 != actResult2 {
		t.Fatalf("Machine 2 result not repeatable!")
		return
	}
}

// Valid but incomplete expression
func TestParseIncompleteExpression(t *testing.T) {
	errMsgs := []string{
		"Parse Error: syntax error",
		"Got to approx [X] in 'string(45div999 [X] ) +'"}
	checkParseError(t, "string(45div999) +", errMsgs)
}

// Some expressions that are not valid.
func TestParseIllegalExpression(t *testing.T) {
	parseErr := "Parse Error: syntax error"

	// Function name before closing parenthesis
	checkParseError(t, "string)",
		[]string{parseErr, "Got to approx [X] in 'string) [X] '"})
	// Path / nametest
	checkParseError(t, "strings)",
		[]string{parseErr, "Got to approx [X] in 'strings) [X] '"})
	// Number before closing square bracket
	checkParseError(t, "10]",
		[]string{parseErr, "Got to approx [X] in '10] [X] '"})
	// As above but further ignored expression after closing parenthesis
	checkParseError(t, "10 ) + 1234",
		[]string{parseErr, "Got to approx [X] in '10 ) [X]  + 1234'"})
	checkParseError(t, "10()",
		[]string{parseErr, "Got to approx [X] in '10( [X] )'"})
}

func TestParseMismatchedParenthesis2(t *testing.T) {
	errMsgs := []string{
		"Parse Error: syntax error",
		"Got to approx [X] in ') [X] '"}
	checkParseError(t, ")", errMsgs)
}

// Illegal character (as we don't support variables for YANG XPATH)
func TestParseUnrecognisedCharacter(t *testing.T) {
	errMsgs := []string{
		"Parse Error: syntax error",
		"Lexer Error: unrecognised character '$'",
	}
	checkParseError(t, "$", errMsgs)
}

// Illegal character mid-way through expression as with lookahead this can be
// handled differently.
func TestParseUnrecognisedCharacter2(t *testing.T) {
	errMsgs := []string{
		"Failed to compile '10 && 24'",
		"Lexer Error: unrecognised character '&'",
	}
	checkParseError(t, "10 && 24", errMsgs)
}

func TestParseErrorShowingPosition(t *testing.T) {
	errMsgs := []string{
		"Parse Error: syntax error",
		"Got to approx [X] in 'square [X] (10 + 2) * (2 + 3)'",
		"Lexer Error: Unknown function or node type: 'square'",
	}
	checkParseError(t, "square(10 + 2) * (2 + 3)", errMsgs)
}

func TestUnusedTokens(t *testing.T) {
	// To keep a handle on token types we've yet to implement, we have grammar
	// productions that explicitly catch these and generate an error.
	// This will help users understand what is wrong and let them know that
	// their XPath expression is correct but not yet supported, rather than
	// incorrect.
	errMsgs := []string{"AxisName unsupported: ancestor-or-self"}
	checkParseError(t, "ancestor-or-self::foo", errMsgs)

	errMsgs = []string{"NodeType unsupported: comment"}
	checkParseError(t, "comment()", errMsgs)

	// Double slash
	errMsgs = []string{"// unsupported: not yet implemented"}
	checkParseError(t, "../foo//bar", errMsgs)

	// @
	errMsgs = []string{"@ (40) unsupported: not yet implemented"}
	checkParseError(t, "../@foo", errMsgs)
}
