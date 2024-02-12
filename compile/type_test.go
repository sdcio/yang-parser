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
// Copyright (c)2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests on the built-in types and derived types
// available in Yang (RFC 6020).

package compile_test

import (
	"testing"

	"github.com/sdcio/yang-parser/schema"
	"github.com/sdcio/yang-parser/testutils"
)

// Subsets of error strings that we use to match on actual error reported.
// This is to catch test cases that fail for the *wrong* reason - amazingly
// easy otherwise to not realise you are not testing what you think you are
// testing (-:
const (
	ErrorSeverityStr             = "Error: "
	BetweenErrStr                = "between "
	derivedRangeRestrictive      = "derived range must be restrictive"
	derivedTypeLengthRestrictive = "derived type length must be restrictive"
	derivedTypeRangeRestrictive  = "derived type range must be restrictive"
	EqualErrStr                  = "equal to "
	LenBetweenErrStr             = "have length between "
	LenOfErrStr                  = "have length of "
	missingFracDigits            = "missing fraction-digits"
	OneEnumValueErrStr           = "Must have value "
	OneLenErrStr                 = "Must have length of "
	OneOfValueErrStr             = "Must have one of the following values: "
	OneOfLengthErrStr            = "Must be one of the following: "
	OneRangeErrStr               = "Must have value between "
	OneValueErrStr               = "Must have value equal to "
	OnlyOneDefaultAllowedStr     = "only one 'default' statement is allowed"
	parseMinusOneInvalidSyntax   = "ParseUint: parsing \"-1\": invalid syntax"
	rangesAscending              = "must be in ascending order"
	rangesDisjoint               = "must be disjoint"
	rangeEndGtThanStart          = "start must be greater than"
	unableToParse                = "Unable to parse"
	valueOutOfRange              = "value out of range"
)

var AllTypesTest = []testutils.TestCase{
	{
		Description: "AllTypes: validation of basic syntax",
		Template:    ListTemplate,
		Schema: `leaf test_bits {
		    type bits {
			    bit testbit;
		}
		}
		leaf test_bool {
			type boolean;
		}
		leaf test_dec64 {
			default "6.6";
			description "test_dec64 description";
			type decimal64 {
				fraction-digits 3;
			}
		}
		leaf test_empty {
			type empty;
		}
		leaf test_enum {
			type enumeration {
				enum "first_val";
				enum "second_val";
			}
		}
		leaf test_int8 {
			type int8;
		}
		leaf test_int16 {
			type int16;
		}
		leaf test_int32 {
			type int32;
		}
		leaf test_int64 {
			type int64;
		}
		leaf test_string {
			type string;
		}
		leaf test_uint8 {
			type uint8;
		}
		leaf test_uint16 {
			type uint16;
		}
		leaf test_uint32 {
			type uint32;
		}
		leaf test_uint64 {
			type uint64;
		}
		leaf test_union {
			type union {
				type int8;
				type int16;
			}
		}`,
		ExpResult: true,
		NodesToValidate: []schema.NodeSpec{
			{
				Path: []string{"testContainer", "testList", "test_dec64"},
				Statement: schema.NodeSubSpec{
					Type: "leaf",
					Properties: []schema.NodeProperty{
						{"default", "6.6"},
						{"description", "test_dec64 description"}}},
				Data: schema.NodeSubSpec{
					Type: "decimal64",
					Properties: []schema.NodeProperty{
						{"name", "{builtin decimal64}"}}},
			},
		},
	},
}

func TestAllTypes(t *testing.T) {
	runTestCases(t, AllTypesTest)
}

var BitsPassTests = []testutils.TestCase{
	{
		Description: "BitsPassTests: validation of basic syntax",
		Template:    ContainerTemplate,
		Schema: `leaf test_bits {
			type bits {
				bit no_pos;
			}
		}
		leaf test_bits2 {
			type bits {
				bit first_pos {
					position 1;
				}
			}
		}
		leaf test_bits3 {
			type bits {
				bit last_pos {
					position 4294967295;
				}
				bit other_pos {
					position 10;
				}
			}
		}`,
		ExpResult: true,
	},
}

// bits:
// - MUST have 'bit' (1 or more)
// - MAY have position
// - If max value used, subsequent entries MUST have position specified.
// - Position values MUST be unique.
func TestBitsPass(t *testing.T) {
	runTestCases(t, BitsPassTests)
}

func TestBitsSkip(t *testing.T) {
	t.Skipf("Bits not fully implemented")

	// Test enforcement of specifying position if we've had max value used.

	// Test requirement that position values are unique.

	// Test MUST have at least 1 'bit' statement.

	// Check valid / invalid position numbers

	// Don't think we can test that correct bit value is assigned here?
}

var BitsFailTests = []testutils.TestCase{
	{
		Description: "BitsFailTests: illegal position (low)",
		Template:    LeafTemplate,
		Schema: `type bits {
			bit first_pos {
				position -1;
			}
		}`,
		ExpResult: false,
		ExpErrMsg: parseMinusOneInvalidSyntax,
	},
	{
		Description: "BitsFailTests: illegal position (high)",
		Template:    LeafTemplate,
		Schema: `type bits {
			bit first_pos {
				position 4294967296;
			}
		}`,
		ExpResult: false,
		ExpErrMsg: "ParseUint: parsing \"4294967296\": value out of range",
	},
}

func TestBitsFail(t *testing.T) {
	runTestCases(t, BitsFailTests)
}

var Dec64Passes = []testutils.TestCase{
	{
		Description: "Dec64: Passing Testcases",
		Template:    ListTemplate,
		Schema: `leaf test_dec1 {
				type decimal64 {
					fraction-digits 1;
				}
			}
			leaf test_dec2 {
				type decimal64 {
					fraction-digits 2;
				}
			}
			leaf test_dec3 {
				type decimal64 {
					fraction-digits 18;
				}
			}
			leaf test_dec4 {
				type decimal64 {
					fraction-digits 5;
					range "1 .. 4 | 5 | 6 .. 6 | 10 .. 20 | 26 .. 29";
				}
			}
			leaf test_dec5 {
				type decimal64 {
					fraction-digits 4;
					range "min .. 4 | 5 | 6 .. 6 | 10 .. 20 | 26 .. max";
				}
			}`,
		ExpResult: true,
	},
}

// decimal64:
// - MUST have fraction-digits (1-18)
// - MAY have range
// - MAY refine another decimal64
func TestDec64Pass(t *testing.T) {
	runTestCases(t, Dec64Passes)
}

var Dec64Failures = []testutils.TestCase{
	{
		Description: "Dec64: no fraction-digits",
		Template:    LeafTemplate,
		Schema:      `type decimal64;`,
		ExpResult:   false,
		ExpErrMsg:   missingFracDigits,
	},
	{
		Description: "Dec64: illegal fraction-digits (low)",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
		     fraction-digits 0; }`,
		ExpResult: false,
		ExpErrMsg: "fraction-digits 0: invalid argument: 0",
	},
	{
		Description: "Dec64: illegal fraction-digits (high)",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
		     fraction-digits 19; }`,
		ExpResult: false,
		ExpErrMsg: "fraction-digits 19: invalid argument: 19",
	},
}

func TestDec64Fail(t *testing.T) {
	runTestCases(t, Dec64Failures)
}

// For each int type, verify:
//
// - multiple ranges are accepted
// - min and max values for range are ok
var IntPasses = []testutils.TestCase{
	{
		Description: "Int: Pass cases",
		Template:    ListTemplate,
		Schema: `leaf test_int8_1 {
			type int8 {
				range "0 ..4 | 5 | 6 .. 6 |  10 .. 20 | 26 .. 29";
			}
		}
		leaf test_int8_2 {
			type int8 {
				range "-128 .. 127";
			}
		}
		leaf test_int16_1 {
			type int16 {
				range "0 ..4 | 5 | 6 .. 6 |  10 .. 20 | 26 .. 29";
			}
		}
		leaf test_int16_verify_limits {
			type int16 {
				range "-32768 .. 32767";
			}
		}
		leaf test_int16_verify_min_and_max {
			type int16 {
				range "min .. -32768 | 32767 .. max";
			}
		}
		leaf test_int32_1 {
			type int32 {
				range "0 ..4 | 5 | 6 .. 6 |  10 .. 20 | 26 .. 29";
			}
		}
		leaf test_int32_2 {
			type int32 {
				range "-2147483648 .. 2147483647";
			}
		}
		leaf test_int64_1 {
			type int64 {
				range "0 ..4 | 5 | 6 .. 6 |  10 .. 20 | 26 .. 29";
			}
		}
		leaf test_int64_2 {
			type int64 {
				range "-9223372036854775808 .. 9223372036854775807";
			}
		}`,
		ExpResult: true,
	},
}

// Int:
//   - MAY have range
//   - NB: We mostly test int32 as most handling is common.  The only parts
//     we test all ints for are the extreme bounds of ranges, and out of
//     bounds values.
func TestIntPass(t *testing.T) {
	runTestCases(t, IntPasses)
}

// Check range is correctly validated for each type, bottom and top.
var IntFails = []testutils.TestCase{
	{
		Description: "Int8: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type int8 {
			range "-129 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Int8: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type int8 {
			range "10 .. 128";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Int16: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type int16 {
			range "-32769 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Int16: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type int16 {
			range "10 .. 32768";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Int32: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "-2147483649 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Int32: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "10 .. 2147483648";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Int64: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type int64 {
			range "-9223372036854775809 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: valueOutOfRange,
	},
	{
		Description: "Int64: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type int64 {
			range "10 .. 9223372036854775808";
		}`,
		ExpResult: false,
		ExpErrMsg: valueOutOfRange,
	},
}

func TestIntFail(t *testing.T) {
	runTestCases(t, IntFails)
}

var StringPasses = []testutils.TestCase{
	{
		Description: "String: Pass cases",
		Template:    ListTemplate,
		Schema: `leaf test_string_basic {
				type string;
			}
			leaf test_string_length {
				type string {
					length "1 ..4 | 5 | 6 .. 6 | 10 .. 20 | 26 .. 29";
				}
			}
			leaf test_string_pattern {
				type string {
					pattern "[0-9a-fA-F]*";
				}
			}`,
		ExpResult: true,
	},
}

func TestString(t *testing.T) {
	runTestCases(t, StringPasses)
}

// For each Uint type, verify:
//
// - multiple ranges are accepted
// - min and max values for range are ok
var UintPasses = []testutils.TestCase{
	{
		Description: "Uint: Pass cases",
		Template:    ListTemplate,
		Schema: `leaf test_uint8_1 {
			type uint8 {
				range "0 ..4 | 5 | 6 .. 6 | 10 .. 20 | 26 .. 29";
			}
		}
		leaf test_uint8_2 {
			type uint8 {
				range "0 .. 255";
			}
		}
		leaf test_uint8_3 {
			type uint8 {
				range "min .. max";
			}
		}
		leaf test_uint16_1 {
			type uint16 {
				range "0 ..4 | 5 | 6 .. 6 |  10 .. 20 | 26 .. 29";
			}
		}
		leaf test_uint16_2 {
			type uint16 {
				range "0 .. 65535";
			}
		}
		leaf test_uint32_1 {
			type uint32 {
				range "0 ..4 | 5 | 6 .. 6 |  10 .. 20 | 26 .. 29";
			}
		}
		leaf test_uint32_2 {
			type uint32 {
				range "0 .. 4294967295";
			}
		}
		leaf test_uint64_1 {
			type uint64 {
				range "0 ..4 | 5 | 6 .. 6 |  10 .. 20 | 26 .. 29";
			}
		}
		leaf test_uint64_2 {
			type uint64 {
				range "0 .. 18446744073709551615";
			}
		}`,
		ExpResult: true,
	},
}

func TestUintPass(t *testing.T) {
	runTestCases(t, UintPasses)
}

// Check range is correctly validated for each type, bottom and top.
var UintFails = []testutils.TestCase{
	{
		Description: "Uint8: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type uint8 {
			range "-1 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: parseMinusOneInvalidSyntax,
	},
	{
		Description: "Uint8: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type uint8 {
			range "10 .. 256";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Uint16: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type uint16 {
			range "-1 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: parseMinusOneInvalidSyntax,
	},
	{
		Description: "Uint16: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type uint16 {
			range "10 .. 65536";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Uint32: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type uint32 {
			range "-1 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: parseMinusOneInvalidSyntax,
	},
	{
		Description: "Uint32: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type uint32 {
			range "10 .. 4294967296";
		}`,
		ExpResult: false,
		ExpErrMsg: derivedTypeRangeRestrictive,
	},
	{
		Description: "Uint64: invalid range (low)",
		Template:    LeafTemplate,
		Schema: `type uint64 {
			range "-1 .. 2";
		}`,
		ExpResult: false,
		ExpErrMsg: parseMinusOneInvalidSyntax,
	},
	{
		Description: "Uint64: invalid range (high)",
		Template:    LeafTemplate,
		Schema: `type uint64 {
			range "10 .. 18446744073709551616";
		}`,
		ExpResult: false,
		ExpErrMsg: valueOutOfRange,
	},
}

func TestUintFail(t *testing.T) {
	runTestCases(t, UintFails)
}

var Dec64RangeExceptions = []testutils.TestCase{
	{
		Description: "Dec64: single range, decreasing",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
		     fraction-digits 2;
		     range "5 .. 3"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "Dec64: 2 ranges, second decreasing",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
			fraction-digits 2;
			range "1 .. 5 | 15 .. 13"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "Dec64: 2 ranges, second overlaps first",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
			fraction-digits 2;
			range "1 .. 5 | 4 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Dec64: 2 ranges, second starts below first",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
			fraction-digits 2;
			range "3 .. 5 | 2 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesAscending,
	},
	{
		Description: "Dec64: 2 ranges, second starts with end value of first",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
			fraction-digits 2;
			range "3 .. 5 | 5 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Dec64: 3 ranges, third overlaps second",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
			fraction-digits 2;
			range "3 .. 5 | 6 .. 10 | 9 .. 12"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Dec64: min .. max repeated",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
			fraction-digits 2;
			range "min .. max | min .. max"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
}

func TestDec64RangeExceptions(t *testing.T) {
	runTestCases(t, Dec64RangeExceptions)
}

var Uint64RangeExceptions = []testutils.TestCase{
	{
		Description: "Uint64: single range, decreasing",
		Template:    LeafTemplate,
		Schema: `type uint64 {
		     range "5 .. 3"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "Uint64: 2 ranges, second decreasing",
		Template:    LeafTemplate,
		Schema: `type uint64 {
			range "1 .. 5 | 15 .. 13"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "Uint64: 2 ranges, second overlaps first",
		Template:    LeafTemplate,
		Schema: `type uint64 {
			range "1 .. 5 | 4 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Uint64: 2 ranges, second starts below first",
		Template:    LeafTemplate,
		Schema: `type uint64 {
			range "3 .. 5 | 2 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesAscending,
	},
	{
		Description: "Uint64: 2 ranges, second starts with end value of first",
		Template:    LeafTemplate,
		Schema: `type uint64 {
			range "3 .. 5 | 5 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Uint64: 3 ranges, third overlaps second",
		Template:    LeafTemplate,
		Schema: `type uint64 {
			range "3 .. 5 | 6 .. 10 | 9 .. 12"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Dec64: min .. max repeated",
		Template:    LeafTemplate,
		Schema: `type decimal64 {
			fraction-digits 2;
			range "min .. max | min .. max"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
}

func TestUint64RangeExceptions(t *testing.T) {
	runTestCases(t, Uint64RangeExceptions)
}

var Int32RangeExceptions = []testutils.TestCase{
	{
		Description: "Int32: single range, decreasing",
		Template:    LeafTemplate,
		Schema: `type int32 {
		     range "5 .. 3"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "Int32: 2 ranges, second decreasing",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "1 .. 5 | 15 .. 13"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "Int32: 2 ranges, second overlaps first",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "1 .. 5 | 4 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Int32: 2 ranges, second starts below first",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "3 .. 5 | 2 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesAscending,
	},
	{
		Description: "Int32: 2 ranges, second starts with end value of first",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "3 .. 5 | 5 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Int32: 3 ranges, third overlaps second",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "3 .. 5 | 6 .. 10 | 9 .. 12"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "Int32: min .. max repeated",
		Template:    LeafTemplate,
		Schema: `type int32 {
			range "min .. max | min .. max"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
}

func TestInt32RangeExceptions(t *testing.T) {
	runTestCases(t, Int32RangeExceptions)
}

var StringRangeExceptions = []testutils.TestCase{
	{
		Description: "String: single range, decreasing",
		Template:    LeafTemplate,
		Schema: `type string {
		     length "5 .. 3"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "String: 2 ranges, second decreasing",
		Template:    LeafTemplate,
		Schema: `type string {
			length "1 .. 5 | 15 .. 13"; }`,
		ExpResult: false,
		ExpErrMsg: rangeEndGtThanStart,
	},
	{
		Description: "String: 2 ranges, second overlaps first",
		Template:    LeafTemplate,
		Schema: `type string {
			length "1 .. 5 | 4 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "String: 2 ranges, second starts below first",
		Template:    LeafTemplate,
		Schema: `type string {
			length "3 .. 5 | 2 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesAscending,
	},
	{
		Description: "String: 2 ranges, second starts with end value of first",
		Template:    LeafTemplate,
		Schema: `type string {
			length "3 .. 5 | 5 .. 7"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "String: 3 ranges, third overlaps second",
		Template:    LeafTemplate,
		Schema: `type string {
			length "3 .. 5 | 6 .. 10 | 9 .. 12"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
	{
		Description: "String: min .. max repeated",
		Template:    LeafTemplate,
		Schema: `type string {
			length "min .. max | min .. max"; }`,
		ExpResult: false,
		ExpErrMsg: rangesDisjoint,
	},
}

func TestStringRangeExceptions(t *testing.T) {
	runTestCases(t, StringRangeExceptions)
}

// TypedefTemplate has:
//
//	"1 .. 10 | 11 .. 11 | 12 | 13 .. 20 | 31 .. 40 | 51 .. 60"
var TypedefPass = []testutils.TestCase{
	{
		Description: "Typedef: decimal64 restrictive range, no overlap",
		Template:    TypedefTemplate,
		Schema: `base_dec64_with_range {
            range "2 .. 9 | 13 .. 20 | 51 .. 59"; }`,
		ExpResult: true,
	},
	// dec64 contiguous test should FAIL so is in TypedefFail (-:
	{
		Description: "Typedef: decimal64 restrictive range, 2 in one range",
		Template:    TypedefTemplate,
		Schema:      `base_dec64_with_range { range "2 .. 4 | 6 .. 8"; }`,
		ExpResult:   true,
	},
	{
		Description: "Typedef: int restrictive range, no overlap",
		Template:    TypedefTemplate,
		Schema: `base_int_with_range {
            range "2 .. 9 | 11 .. 20 | 51 .. 59"; }`,
		ExpResult: true,
	},
	{
		Description: "Typedef: int restrictive range, " +
			"overlap (contiguous base ranges)",
		Template:  TypedefTemplate,
		Schema:    `base_int_with_range { range "2 .. 18"; }`,
		ExpResult: true,
	},
	{
		Description: "Typedef: int restrictive range, 2 in one range",
		Template:    TypedefTemplate,
		Schema:      `base_int_with_range { range "2 .. 4 | 6 .. 8"; }`,
		ExpResult:   true,
	},
	{
		Description: "Typedef: string restrictive range, no overlap",
		Template:    TypedefTemplate,
		Schema: `base_string_with_range {
            length "2 .. 9 | 11 .. 20 | 51 .. 59"; }`,
		ExpResult: true,
	},
	{
		Description: "Typedef: string restrictive range, " +
			"overlap (contiguous base ranges)",
		Template:  TypedefTemplate,
		Schema:    `base_string_with_range { length "2 .. 18"; }`,
		ExpResult: true,
	},
	{
		Description: "Typedef: string restrictive length, 2 in one range",
		Template:    TypedefTemplate,
		Schema:      `base_string_with_range { length "2 .. 4 | 6 .. 8"; }`,
		ExpResult:   true,
	},
	{
		Description: "Typedef: uint restrictive range, no overlap",
		Template:    TypedefTemplate,
		Schema: `base_uint_with_range {
            range "2 .. 9 | 11 .. 20 | 51 .. 59"; }`,
		ExpResult: true,
	},
	{
		Description: "Typedef: uint restrictive range, " +
			"overlap (contiguous base ranges)",
		Template:  TypedefTemplate,
		Schema:    `base_uint_with_range { range "2 .. 18"; }`,
		ExpResult: true,
	},
	{
		Description: "Typedef: uint restrictive range, 2 in one range",
		Template:    TypedefTemplate,
		Schema:      `base_uint_with_range { range "2 .. 4 | 6 .. 8"; }`,
		ExpResult:   true,
	},
}

func TestTypedefPass(t *testing.T) {
	runTestCases(t, TypedefPass)
}

var TypedefFail = []testutils.TestCase{
	{
		Description: "Typedef: dec64 less restrictive range",
		Template:    TypedefTemplate,
		Schema:      `base_dec64_with_range { range "51 .. 61"; }`,
		ExpResult:   false,
		ExpErrMsg:   derivedTypeRangeRestrictive,
	},
	{
		Description: "Typedef: dec64 restrictive range, " +
			"overlap (contiguous base ranges)",
		Template:  TypedefTemplate,
		Schema:    `base_dec64_with_range { range "2 .. 18"; }`,
		ExpResult: false,
		ExpErrMsg: derivedRangeRestrictive,
	},
	{
		Description: "Typedef: dec64 range covers 2 sub ranges " +
			"AND gap between them",
		Template:  TypedefTemplate,
		Schema:    `base_dec64_with_range { range "35 .. 55"; }`,
		ExpResult: false,
		ExpErrMsg: derivedRangeRestrictive,
	},
	{
		Description: "Typedef: int less restrictive range",
		Template:    TypedefTemplate,
		Schema:      `base_int_with_range { range "51 .. 61"; }`,
		ExpResult:   false,
		ExpErrMsg:   derivedTypeRangeRestrictive,
	},
	{
		Description: "Typedef: int range covers 2 sub ranges " +
			"AND gap between them",
		Template:  TypedefTemplate,
		Schema:    `base_int_with_range { range "35 .. 55"; }`,
		ExpResult: false,
		ExpErrMsg: derivedRangeRestrictive,
	},
	{
		Description: "Typedef: string less restrictive range",
		Template:    TypedefTemplate,
		Schema:      `base_string_with_range { length "51 .. 61"; }`,
		ExpResult:   false,
		ExpErrMsg:   derivedTypeLengthRestrictive,
	},
	{
		Description: "Typedef: string range covers 2 sub ranges " +
			"AND gap between them",
		Template:  TypedefTemplate,
		Schema:    `base_string_with_range { length "35 .. 55"; }`,
		ExpResult: false,
		ExpErrMsg: derivedRangeRestrictive,
	},
	{
		Description: "Typedef: uint less restrictive range",
		Template:    TypedefTemplate,
		Schema:      `base_uint_with_range { range "51 .. 61"; }`,
		ExpResult:   false,
		ExpErrMsg:   derivedTypeRangeRestrictive,
	},
	{
		Description: "Typedef: uint range covers 2 sub ranges " +
			"AND gap between them",
		Template:  TypedefTemplate,
		Schema:    `base_uint_with_range { range "35 .. 55"; }`,
		ExpResult: false,
		ExpErrMsg: derivedRangeRestrictive,
	},
	{
		Description: "Typedef: recursive ranges, middle one is invalid",
		Template:    ContainerTemplate,
		Schema: `typedef nearly_a_dozen {
			type uint32 {
				range "1..5 | 7..12";
			}
		}
		typedef dozen {
			type nearly_a_dozen {
				range "1 .. 3 | 4 .. 7 | 8 .. 12";
			}
		}
		typedef nearly_a_dozen2 {
			type dozen {
				range "1 .. 4 | 5 .. 5 | 8 .. 12";
			}
		}
		leaf typedef_test {
			type nearly_a_dozen2;
		}`,
		ExpResult: false,
		ExpErrMsg: derivedRangeRestrictive,
	},
	{
		Description: "Typedef: recursive ranges, last one is invalid",
		Template:    ContainerTemplate,
		Schema: `typedef nearly_a_dozen {
			type uint32 {
				range "1..5 | 7..12";
			}
		}
		typedef dozen {
			type nearly_a_dozen {
				range "1 .. 3 | 4 .. 5 | 8 .. 12";
			}
		}
		typedef nearly_a_dozen2 {
			type dozen {
				range "1 .. 4 | 5 .. 6 | 8 .. 12";
			}
		}
		leaf typedef_test {
			type nearly_a_dozen2;
		}`,
		ExpResult: false,
		ExpErrMsg: derivedRangeRestrictive,
	},
}

func TestTypedefFail(t *testing.T) {
	runTestCases(t, TypedefFail)
}

// Verify supported and unsupported substatements.  Would be nice to
// automate this in terms of generating test cases, rather than writing
// each one out by hand as this is a very large amount of work otherwise.
// Also unclear about return on investment in doing this versus other
// testing, so 'skip' for now.
// Also check cardinality.
func TestTypeSubstatements(t *testing.T) {
	t.Skipf("Verify [un]supported substatements for each type")
}
