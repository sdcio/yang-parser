// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// These tests verify that the core functions required to implement
// XPATH, including the YANG-specific current() function, are working
// correctly.
//
// Note that the way the functions are coded to allow for different
// argument types and return types, we need to verify that the symbol
// table view of these matches the actual function definition.  To do
// this we have special validation code that is only used when we run
// a machine with the 'validate' flag, and we only do this at test time.
//
// Finally the last test here checks that all built-in functions have been
// run so we know that the type validation has been done for all of them.

package expr

import (
	"math"
	"testing"

	"github.com/iptecharch/yang-parser/xpath"
	"github.com/iptecharch/yang-parser/xpath/xpathtest"
	"github.com/iptecharch/yang-parser/xpath/xutils"
)

// Functions

// Check type conversions when calling functions ... ideally we need functions
// that return a type *different* to the types passed in to make sure we
// aren't accidentally converting to the return type (as was accidentally
// done first time around (-: )
func TestTypeConversionsToLiteral(t *testing.T) {
	// Number
	checkBoolResult(t, "contains(1234, 23)", true)
	checkBoolResult(t, "contains(1234, 32)", false)

	// Bool
	checkBoolResult(t, "contains(2+2=4,1+1=3)", false)
	checkBoolResult(t, "contains(2+2=5, 'als')", true)
}

func TestTypeConversionsToNumber(t *testing.T) {
	// Literal
	checkLiteralResult(t, "substring(12345, '1', 4)", "1234")

	// Bool
	checkLiteralResult(t, "substring((2 + 2 = 5), 2+2=4, 5)", "false")
	checkLiteralResult(t, "substring((2 + 2 = 5), 2+2=5, 3)", "fa")

	// Nodeset conversions handled in TestNodesetTypeConversion()
}

func TestTypeConversionsToBool(t *testing.T) {
	checkBoolResult(t, "not('hello')", false)
	checkBoolResult(t, "not('')", true)

	checkBoolResult(t, "not(1)", false)
	checkBoolResult(t, "not(0)", true)

	// Nodeset conversions handled in TestNodesetTypeConversion()
}

func TestNodesetTypeConversion(t *testing.T) {
	configTree := xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
		})

	// The string-value of a nodeset is the concatenation of the string-values
	// of all text node descendants of the first node in document order.
	checkLiteralResultWithContext(t, "string(../dataplane/name)", "dp0s1",
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))
	checkLiteralResultWithContext(t, "string(../dataplane/nom)", "",
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))
	// Yes, you read that right (-:
	checkLiteralResultWithContext(t, "string(..)",
		"dp0s11111dp0s221112222s15555lo2",
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))

	checkNumResultWithContext(t, "number(../serial/address)",
		5555,
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))
	checkNumResultWithContext(t, "round(../*/address)",
		1111,
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))
	checkNumResultWithContext(t, "number(../*/postcode)", math.NaN(),
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))

	// True if non-empty
	checkBoolResultWithContext(t, "boolean(../*/address)", true,
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))
	checkBoolResultWithContext(t, "boolean(../*/postcode)", false,
		configTree, xutils.PathType([]string{"/", "interface", "serial"}))
}

func TestWrongArgNums(t *testing.T) {
	checkParseError(t, "boolean()", []string{
		"Failed to compile 'boolean()'",
		"Parse Error: boolean() takes 1 args, not 0."})
	checkParseError(t, "true(1)", []string{
		"Failed to compile 'true(1)'",
		"Parse Error: true() takes 0 args, not 1."})
	checkParseError(t, "true(1, 2)", []string{
		"Failed to compile 'true(1, 2)'",
		"Parse Error: true() takes 0 args, not 2."})
	checkParseError(t, "true(1, '2', 3)", []string{
		"Failed to compile 'true(1, '2', 3)'",
		"Parse Error: true() takes 0 args, not 3."})
	checkParseError(t, "true(1, 2, 3, 4)", []string{
		"Failed to compile 'true(1, 2, 3, 4)'",
		"Parse Error: syntax error"})
}

// For the most part we have conversion functions, which are tested
// with the number, boolean and string tests and do not need to be
// repeated for each actual function here as the conversion is done
// when we pop data off the stack, generically, using string / boolean /
// number.

// boolean()
func TestParseBooleanFn(t *testing.T) {
	checkBoolResult(t, "boolean(111.1)", true)
	checkBoolResult(t, "boolean(1 - 1)", false)
	checkBoolResult(t, "boolean(1 and 10)", true)
	checkBoolResult(t, "boolean(0 <= 1)", true)
	checkBoolResult(t, "boolean(\"a string\")", true)
	checkBoolResult(t, "boolean('')", false)

	// Nodesets covered in TestNodesetTypeConversion
}

// ceiling() (round up)
func TestParseCeilingFn(t *testing.T) {
	checkNumResult(t, "ceiling(0.5)", 1)
	checkNumResult(t, "ceiling(9.49)", 10)
	checkNumResult(t, "ceiling(0.99)", 1)
	checkNumResult(t, "ceiling(1)", 1)
	checkNumResult(t, "ceiling(-1)", -1)
	checkNumResult(t, "ceiling(-0.5)", 0)
	checkNumResult(t, "ceiling(-3.3)", -3)
	checkNumResult(t, "ceiling(-1.99)", -1)
}

// concat()
func TestParseConcatFn(t *testing.T) {
	checkLiteralResult(t, "concat('', '')", "")
	checkLiteralResult(t, "concat('', 'abc')", "abc")
	checkLiteralResult(t, "concat('def', '')", "def")
	checkLiteralResult(t, "concat('321', '456')", "321456")
}

// contains()
func TestParseContainsFn(t *testing.T) {
	checkBoolResult(t, "contains('foo', 'bar')", false)
	checkBoolResult(t, "contains('fool', 'fool')", true)
	checkBoolResult(t, "contains('fool', 'foo')", true)
	checkBoolResult(t, "contains('fool', 'oo')", true)
	checkBoolResult(t, "contains('fool', 'ool')", true)
	checkBoolResult(t, "contains('fool', 'f')", true)
	checkBoolResult(t, "contains('fool', 'l')", true)
	checkBoolResult(t, "contains('fool', '')", true)
	checkBoolResult(t, "contains('fool', 'foolish')", false)
}

// count()
func TestParseCount(t *testing.T) {
	configTree := xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "dataplane/name+dp0s2", "address@4321"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
			{"protocols", "mpls", "max-label+1000000"},
		})

	// dp0s1 and dp0s2 list elements
	checkNumResultWithContext(t, "count(../interface/dataplane)", 2,
		configTree, xutils.PathType([]string{"/", "protocols"}))

	// dp0s1, dp0s2, s1 and lo2 list elements
	checkNumResultWithContext(t, "count(../interface/*)", 4,
		configTree, xutils.PathType([]string{"/", "protocols"}))

	// name and address
	checkNumResultWithContext(t, "count(../interface/*/*)", 6,
		configTree, xutils.PathType([]string{"/", "protocols"}))

	checkNumResultWithContext(t, "count(../interface)", 1,
		configTree, xutils.PathType([]string{"/", "protocols"}))

	checkNumResultWithContext(t, "count(../interface/nonexistent)", 0,
		configTree, xutils.PathType([]string{"/", "protocols"}))
}

func TestParseCurrent(t *testing.T) {
	configTree := xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@5678"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
			{"protocols", "mpls", "min-label+15"},
		})

	checkNodeSetResult(t, "current()", nil,
		configTree, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeaf(
				nil, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}),
				"", "min-label", "15")}))
}

func TestParseCurrentOnEphemeral(t *testing.T) {
	configTree := xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"top", "ephemeral$"},
		})

	checkNodeSetResult(t, "current()", nil,
		configTree, xutils.PathType([]string{"/", "top", "ephemeral"}),
		xpathtest.TNodeSet{})
}

// floor() (round down)
func TestParseFloorFn(t *testing.T) {
	checkNumResult(t, "floor(0.5)", 0)
	checkNumResult(t, "floor(9.49)", 9)
	checkNumResult(t, "floor(0.99)", 0)
	checkNumResult(t, "floor(1)", 1)
	checkNumResult(t, "floor(-1)", -1)
	checkNumResult(t, "floor(-0.5)", -1)
	checkNumResult(t, "floor(-1.4)", -2)
	checkNumResult(t, "floor(-1.99)", -2)
}

// last() - properly tested with predicates; token call here to ensure that
// TestAllFunctionsTested() passes.
func TestParseLast(t *testing.T) {
	checkNumResult(t, "last()", 1)
}

// local-name()
func TestParseLocalName(t *testing.T) {
	configTree := xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@5678"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
		})

	// Leaf
	checkLiteralResultWithContext(t, "local-name(.)", "min-label",
		configTree, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}))

	// List
	checkLiteralResultWithContext(t, "local-name(.)", "dataplane",
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))

	// LeafList
	checkLiteralResultWithContext(t, "local-name(.)", "address",
		configTree, xutils.PathType([]string{"/", "interface", "dataplane", "address"}))

	// Container
	checkLiteralResultWithContext(t, "local-name(.)", "interface",
		configTree, xutils.PathType([]string{"/", "interface"}))

	// Up a level ...
	checkLiteralResultWithContext(t, "local-name(..)", "mpls",
		configTree, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}))

	// ... or down
	checkLiteralResultWithContext(t, "local-name(dataplane)", "dataplane",
		configTree, xutils.PathType([]string{"/", "interface"}))
}

func TestParseNormalizeSpaceFn(t *testing.T) {
	// No-op
	checkLiteralResult(t, "normalize-space('aaa')", "aaa")

	// Leading characters
	checkLiteralResult(t, "normalize-space('\t\n\r aaa')", "aaa")

	// Trailing characters
	checkLiteralResult(t, "normalize-space('aaa\r \n\t')", "aaa")

	// Intermediate characters (replaced by single space)
	checkLiteralResult(t, "normalize-space('a\t\ta\ra\nb  b b')", "a a a b b b")

	// Whitespace everywhere
	checkLiteralResult(
		t, "normalize-space(' \t \r\n123\t\r\r\n \n 456\n\t\r ')", "123 456")
}

// not()
func TestParseNotFn(t *testing.T) {
	checkBoolResult(t, "not(1 + 1 = 2)", false)
	checkBoolResult(t, "not(2 + 2 = 5)", true)
}

// number()
func TestParseNumberFn(t *testing.T) {
	checkNumResult(t, "number(333)", 333)
	checkNumResult(t, "number(1 + 1)", 2)
	checkNumResult(t, "number(1 and 10)", 1)
	checkNumResult(t, "number(0 <= 1)", 1)
	checkNumResult(t, "number(\"a string\")", math.NaN())
	checkNumResult(t, "number('123')", 123)
	checkNumResult(t, "number(' 123.4 ')", 123.4)    // WS before and after
	checkNumResult(t, "number(' -123')", -123)       // -ve, WS before
	checkNumResult(t, "number('-6b2 ')", math.NaN()) // -ve, WS after
	checkNumResult(t, "number('NaN')", math.NaN())

	// Nodesets covered in TestNodesetTypeConversion
}

// position() - properly tested with predicates; token call here to ensure that
// TestAllFunctionsTested() passes.

func TestParsePosition(t *testing.T) {
	checkNumResult(t, "position()", 1)
}

// round()
func TestParseRoundFn(t *testing.T) {
	checkNumResult(t, "round(0.5)", 1)
	checkNumResult(t, "round(9.49)", 9)
	checkNumResult(t, "round(0.99)", 1)
	checkNumResult(t, "round(0)", 0)
	checkNumResult(t, "round(-0)", -0)
	checkNumResult(t, "floor(-0 + 0.5)", 0) // Note different to previous.
	checkNumResult(t, "round(-0.25)", -0)
	checkNumResult(t, "floor(-0.5 + 0.5)", 0) // Note different to previous.
	checkNumResult(t, "round(-0.55)", -1)
	checkNumResult(t, "round(-666.6)", -667)
	t.Skipf("More cases from spec to implement ...")
	checkNumResult(t, "round(-0.5)", -0) // -ve <x>.5 not right yet
	checkNumResult(t, "round(number('NaN'))", math.NaN())
}

// starts-with()
func TestParseStartsWithFn(t *testing.T) {
	// Simple case
	checkBoolResult(t, "starts-with('abcdef', 'abc')", true)

	// Matches exactly
	checkBoolResult(t, "starts-with('abcdef', 'abcdef')", true)

	// Includes but doesn't start-with
	checkBoolResult(t, "starts-with('abcdef', 'bcd')", false)

	// Matches up to a point
	checkBoolResult(t, "starts-with('abcdef', 'abf')", false)

	// Trying to find superset
	checkBoolResult(t, "starts-with('abcdef', 'abcdefg')", false)

	// Doesn't even include
	checkBoolResult(t, "starts-with('abcdef', 'dcb')", false)

	// Empty string we're checking the start of.
	checkBoolResult(t, "starts-with('', 'a')", false)

	// Start with empty string always true!
	checkBoolResult(t, "starts-with('abcdef', '')", true)

	// ... especially when we're comparing against nothing.
	checkBoolResult(t, "starts-with('', '')", true)
}

// string()
func TestParseStringFn(t *testing.T) {
	// Strings including empty string
	checkLiteralResult(t, "string(\"a string\")", "a string")
	checkLiteralResult(t, "string(\"\")", "")

	// Numbers, including NaN, +/-0, and Infinity
	checkLiteralResult(t, "string(123)", "123")
	checkLiteralResult(t, "string(-456)", "-456")
	checkLiteralResult(t, "string(-0.987654321)", "-0.987654321")
	checkLiteralResult(t, "string(number('NaN'))", "NaN")
	checkLiteralResult(t, "string(0)", "0")
	checkLiteralResult(t, "string(-0)", "0")
	checkLiteralResult(t, "string(number('Infinity'))", "Infinity")
	checkLiteralResult(t, "string(number('-Infinity'))", "-Infinity")

	// Booleans
	checkLiteralResult(t, "string(0 = 0)", "true")
	checkLiteralResult(t, "string(2 + 2 = 5)", "false")

	// Nodesets covered in TestNodesetTypeConversion
}

func TestStringLength(t *testing.T) {
	checkNumResult(t, "string-length('')", 0)
	checkNumResult(t, "string-length('string with length')", 18)
}

// substring()
//
// Examples for test taken from XPATH spec and illustrate 'various unusual
// cases' (-:
func TestParseSubstring(t *testing.T) {
	checkLiteralResult(t, "substring('12345', 1, 0)", "")
	checkLiteralResult(t, "substring('12345', 1, 1)", "1")
	checkLiteralResult(t, "substring('12345', 1, 4)", "1234")
	checkLiteralResult(t, "substring('12345', 1, 5)", "12345")
	checkLiteralResult(t, "substring('12345', 0, 0)", "")
	checkLiteralResult(t, "substring('12345', 0, 1)", "")
	checkLiteralResult(t, "substring('12345', 0, 2)", "1")
	checkLiteralResult(t, "substring('12345', 2, 2)", "23")
	checkLiteralResult(t, "substring('12345', 1, 0)", "")
	checkLiteralResult(t, "substring('12345', 2, 3)", "234")

	checkLiteralResult(t, "substring('', 10, 20)", "")
	checkLiteralResult(t, "substring('12345', 10, 5)", "")

	// Unusual cases from XPATH spec.
	checkLiteralResult(t, "substring('12345', 1.5, 2.6)", "234")
	checkLiteralResult(t, "substring('12345', 0, 3)", "12")
	checkLiteralResult(t, "substring('12345', 0 div 0, 3)", "")
	checkLiteralResult(t, "substring('12345', 1, 0 div 0)", "")
	checkLiteralResult(t, "substring('12345', -42, 1 div 0)", "12345")
	checkLiteralResult(t, "substring('12345', -1 div 0, 1 div 0)", "")

	t.Skipf("Not yet supported.")
	checkLiteralResult(t, "substring('12345', 2)", "2345") // Omit optional arg
}

// substring-after()
func TestParseSubstringAfterFn(t *testing.T) {
	// Single character
	checkLiteralResult(t, "substring-after('10.11.12.13', '.')", "11.12.13")

	// String
	checkLiteralResult(t, "substring-after('10.11.12.13', '.12')", ".13")

	// Last character
	checkLiteralResult(t, "substring-after('10.11.12.13', '3')", "")

	// Empty string
	checkLiteralResult(t, "substring-after('', '1')", "")

	// Empty substring - return all of first string
	checkLiteralResult(t, "substring-after('10.11.12.13', '')", "10.11.12.13")

	// Empty string and substring
	checkLiteralResult(t, "substring-after('', '')", "")

	// Non-existent string means we return empty string.
	checkLiteralResult(t, "substring-after('10.11.12.13', 'X')", "")

	// RFC examples
	checkLiteralResult(t, "substring-after('1999/04/01', '/')", "04/01")
	checkLiteralResult(t, "substring-after('1999/04/01', '19')", "99/04/01")
}

// substring-before()
func TestParseSubstringBeforeFn(t *testing.T) {
	// Single character
	checkLiteralResult(t, "substring-before('10.11.12.13', '.')", "10")

	// String
	checkLiteralResult(t, "substring-before('10.11.12.13', '.12')", "10.11")

	// First character
	checkLiteralResult(t, "substring-before('10.11.12.13', '1')", "")

	// Empty string
	checkLiteralResult(t, "substring-before('', '1')", "")

	// Empty substring
	checkLiteralResult(t, "substring-before('10.11.12.13', '')", "")

	// Empty string and substring
	checkLiteralResult(t, "substring-before('', '')", "")

	// Non-existent string means we return empty string.
	checkLiteralResult(t, "substring-before('10.11.12.13', 'X')", "")

	// RFC example
	checkLiteralResult(t, "substring-before('1999/04/01', '/')", "1999")
}

// sum()
func TestParseSum(t *testing.T) {
	configTree := xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@5678"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
			{"protocols", "mpls", "max-label+1000"},
		})

	// Simple case
	checkNumResultWithContext(t, "sum(/protocols/mpls/*)", (1000 + 16),
		configTree, xutils.PathType([]string{"/", "interface"}))

	// Union of 2 nodesets
	checkNumResultWithContext(
		t, "sum(/protocols/mpls/* | /interface/dataplane/address)",
		(16 + 1000 + 5678 + 1234),
		configTree, xutils.PathType([]string{"/", "interface"}))

	// String-value of 2 nodes merged
	checkNumResultWithContext(t, "sum(/protocols/mpls)", 161000,
		configTree, xutils.PathType([]string{"/", "interface"}))

	// String that can't be made into a number
	checkNumResultWithContext(t, "sum(/interfaces/dataplane/name)", 0,
		configTree, xutils.PathType([]string{"/", "interface"}))

	// Non-existent node
	checkNumResultWithContext(t, "sum(/interfaces/dataplane/unknown)", 0,
		configTree, xutils.PathType([]string{"/", "interface"}))

	// Invalid type (not nodeset)
	expErrMsgs := []string{
		"Fn 'sum' takes NODESET, not NUMBER as arg 0.",
	}
	checkExecuteError(t, "sum(1234)", expErrMsgs)

}

func TestParseTranslate(t *testing.T) {
	// Simple replacement of a, b, and c with d, e, and f respectively.
	checkLiteralResult(t, "translate('abcdef aaddbb', 'abc', 'def')",
		"defdef ddddee")
	// XPATH doc example
	checkLiteralResult(t, "translate('bar', 'abc', 'ABC')", "BAr")

	// Second string longer than third
	checkLiteralResult(t, "translate('abcdef aaddbb', 'abc', 'de')",
		"dedef ddddee")
	// XPATH example
	checkLiteralResult(t, "translate('--aaa--', 'abc-', 'ABC')", "AAA")

	// Third string longer than second
	checkLiteralResult(t, "translate('abcdef aaddbb', 'abc', 'defg')",
		"defdef ddddee")

	// First string empty
	checkLiteralResult(t, "translate('', 'abc', 'def')", "")

	// Second string empty
	checkLiteralResult(t, "translate('abcdef aaddbb', '', 'def')",
		"abcdef aaddbb")

	// Third string empty
	checkLiteralResult(t, "translate('abcdef aaddbb', 'abc', '')",
		"def dd")

	// Repeated character in second string
	checkLiteralResult(t, "translate('ababab', 'aa', 'AB')", "AbAbAb")

	// Sneakier - replace first character with itself to ensure second
	// replacement is not called!
	checkLiteralResult(t, "translate('ababab', 'aa', 'aB')", "ababab")
}

// true()
func TestParseTrue(t *testing.T) {
	checkBoolResult(t, "true()", true)
}

// This must be LAST, or it won't catch all function invocations...
func TestAllFunctionsTested(t *testing.T) {
	if err := xpath.CheckAllFunctionsWereTested(); err != nil {
		t.Fatalf("%s", err.Error())
	}
}
