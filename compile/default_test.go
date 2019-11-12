// Copyright (c) 2017,2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests relating to the 'default' statement
// available in Yang (RFC 6020).

package compile_test

import (
	"testing"

	"github.com/danos/mgmterror/errtest"
	"github.com/danos/yang/schema"
	"github.com/danos/yang/testutils"
)

// Single typedef statement validation (Passing testcases) for ranged values
//
// - default in range (4 types)
func TestNumberDefaultPass(t *testing.T) {
	var NumberDefaultPassTests = []testutils.TestCase{
		{
			Description: "Typedef: default substatement within range",
			Template:    TypedefTemplate,
			Schema: `base_uint_with_range {
                range "2 .. 4 | 6 .. 8"; } default "3";`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testleaf"},
					Statement: schema.NodeSubSpec{
						Type:       "leaf",
						Properties: []schema.NodeProperty{{"default", "3"}},
					},
					Data: schema.NodeSubSpec{
						Type:       "uinteger",
						Properties: []schema.NodeProperty{{"default", "3"}},
					},
				},
			},
		},
		{
			Description: "Typedef: default substatement within range",
			Template:    TypedefTemplate,
			Schema: `base_int_with_range {
                range "2 .. 4 | 6 .. 8"; } default "4";`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testleaf"},
					Statement: schema.NodeSubSpec{
						Type:       "leaf",
						Properties: []schema.NodeProperty{{"default", "4"}},
					},
					Data: schema.NodeSubSpec{
						Type:       "integer",
						Properties: []schema.NodeProperty{{"default", "4"}},
					},
				},
			},
		},
		{
			Description: "Typedef: default substatement within range",
			Template:    TypedefTemplate,
			Schema: `base_dec64_with_range {
                range "2 .. 4 | 6 .. 8"; } default "6.6";`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testleaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "6.6"}},
					},
					Data: schema.NodeSubSpec{
						Type: "decimal64",
						Properties: []schema.NodeProperty{
							{"default", "6.6"}},
					},
				},
			},
		},
	}
	runTestCases(t, NumberDefaultPassTests)
}

// Single typedef statement validation (failing testcases)
//
// - 2 default statements
// - default outwith range (4 types)
func TestNumberDefaultFail(t *testing.T) {
	var NumberDefaultFailTests = []testutils.TestCase{
		{
			Description: "Typedef: 2 uint default substatements (illegal)",
			Template:    TypedefTemplate,
			Schema:      `base_uint_with_range; default "6"; default "7";`,
			ExpResult:   false,
			ExpErrMsg:   OnlyOneDefaultAllowedStr,
		},
		{
			Description: "Typedef: 2 int default substatements (illegal)",
			Template:    TypedefTemplate,
			Schema:      `base_int_with_range; default "6"; default "7";`,
			ExpResult:   false,
			ExpErrMsg:   OnlyOneDefaultAllowedStr,
		},
		{
			Description: "Typedef: 2 dec64 default substatements (illegal)",
			Template:    TypedefTemplate,
			Schema:      `base_dec64_with_range; default "6"; default "7";`,
			ExpResult:   false,
			ExpErrMsg:   OnlyOneDefaultAllowedStr,
		},
		{
			Description: "Typedef: default substatement outwith range (uint)",
			Template:    TypedefTemplate,
			Schema: `base_uint_with_range {
                range "2 .. 4 | 6 .. 8"; } default "66";`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, "test-yang-compile", "base_uint_with_range", "66",
				[]errtest.NodeLimits{
					{Min: "2", Max: "4"},
					{Min: "6", Max: "8"}}),
		},
		{
			Description: "Typedef: default substatement outwith range (int)",
			Template:    TypedefTemplate,
			Schema: `base_int_with_range {
                range "2 .. 4 | 6 .. 8"; } default "66";`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, "test-yang-compile", "base_int_with_range", "66",
				[]errtest.NodeLimits{
					{Min: "2", Max: "4"},
					{Min: "6", Max: "8"}}),
		},
		{
			Description: "Typedef: default substatement outwith range (dec64)",
			Template:    TypedefTemplate,
			Schema: `base_dec64_with_range {
                range "2 .. 4 | 6 .. 8"; } default "66";`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, "test-yang-compile", "base_dec64_with_range", "66",
				[]errtest.NodeLimits{
					{Min: "2.000000", Max: "4.000000"},
					{Min: "6.000000", Max: "8.000000"}}),
		},
		{
			Description: "Typedef: default not a number (uint)",
			Template:    TypedefTemplate,
			Schema: `base_uint_with_range {
                range "2 .. 4 | 6 .. 8"; } default "foo";`,
			ExpResult: false,
			ExpErrMsg: "is not an uint32",
		},
		{
			Description: "Typedef: default not a number (int)",
			Template:    TypedefTemplate,
			Schema: `base_int_with_range {
                range "2 .. 4 | 6 .. 8"; } default "foo";`,
			ExpResult: false,
			ExpErrMsg: "is not an int32",
		},
		{
			Description: "Typedef: default not a number (dec64)",
			Template:    TypedefTemplate,
			Schema: `base_dec64_with_range {
                range "2 .. 4 | 6 .. 8"; } default "foo";`,
			ExpResult: false,
			ExpErrMsg: "is not a decimal64",
		},
		{
			Description: "Typedef: empty default (uint)",
			Template:    TypedefTemplate,
			Schema: `base_uint_with_range {
                range "2 .. 4 | 6 .. 8"; } default "";`,
			ExpResult: false,
			ExpErrMsg: "is not an uint32",
		},
		{
			Description: "Typedef: empty default (int)",
			Template:    TypedefTemplate,
			Schema: `base_int_with_range {
                range "2 .. 4 | 6 .. 8"; } default "";`,
			ExpResult: false,
			ExpErrMsg: "is not an int32",
		},
		{
			Description: "Typedef: empty default (dec64)",
			Template:    TypedefTemplate,
			Schema: `base_dec64_with_range {
                range "2 .. 4 | 6 .. 8"; } default "";`,
			ExpResult: false,
			ExpErrMsg: "is not a decimal64",
		},
	}
	runTestCases(t, NumberDefaultFailTests)
}

func TestBinaryDefault(t *testing.T) {
	t.Skipf("Default Test not yet implemented for Binary type.")
}

func TestBitsDefault(t *testing.T) {
	var BitsDefaultTests = []testutils.TestCase{
		{
			Description: "Bits: valid default",
			Template:    ContainerTemplate,
			Schema: `leaf bitLeaf {
			type bits {
			    bit testbit1;
			    bit testbit2;
			}
			default "testbit2";
		}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "bitLeaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "testbit2"}},
					},
					Data: schema.NodeSubSpec{
						Type: "Bits",
						Properties: []schema.NodeProperty{
							{"default", "testbit2"}},
					},
				},
			},
		},
		{
			Description: "Bits: invalid default",
			Template:    ContainerTemplate,
			Schema: `leaf bitLeaf {
			type bits {
			    bit testbit1;
			    bit testbit2;
			}
			default "testbit3";
		}`,
			ExpResult: false,
			ExpErrMsg: "",
		},
	}
	t.Skipf("Validate() not yet implemented for Bits type.")
	runTestCases(t, BitsDefaultTests)
}

func TestBooleanDefault(t *testing.T) {
	var BooleanDefaultTests = []testutils.TestCase{
		{
			Description: "Boolean: 'true' default",
			Template:    BooleanDefaultTemplate,
			Schema:      `default "true";`,
			ExpResult:   true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testLeaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "true"}}},
					Data: schema.NodeSubSpec{
						Type: "boolean",
						Properties: []schema.NodeProperty{
							{"default", "true"}}},
				},
			},
		},
		{
			Description: "boolean: 'false' default",
			Template:    BooleanDefaultTemplate,
			Schema:      `default "false";`,
			ExpResult:   true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testLeaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "false"}}},
					Data: schema.NodeSubSpec{
						Type: "boolean",
						Properties: []schema.NodeProperty{
							{"default", "false"}}},
				},
			},
		},
		{
			Description: "boolean: invalid default",
			Template:    BooleanDefaultTemplate,
			Schema:      `default "invalid";`,
			ExpResult:   false,
			ExpErrs: errtest.YangInvalidDefaultEnumOrBoolErrorStrings(
				t, errtest.Bltin, errtest.BoolType, "invalid",
				[]string{"true", "false"}),
		},
	}
	runTestCases(t, BooleanDefaultTests)
}

// In need of some actual implementation!
// Need to test mandatory vs default
func TestChoiceDefault(t *testing.T) {
	var ChoiceDefaultTests = []testutils.TestCase{
		{
			Description: "Choice Default: valid default",
			Template:    ContainerTemplate,
			Schema: `choice choiceTest {
			default "15";
			case firstCase {
				leaf first {
					type uint16;
					default 30;
					description "First Leaf Description";
				}
			}
		    case secondCase {
				leaf second {
					type uint8;
					default 10;
				}
			}
		}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "choiceTest", "first"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "30"},
							{"description", "First Leaf Description"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "111"}}},
				},
				{
					Path: []string{"testContainer", "choiceTest", "firstCase"},
					Statement: schema.NodeSubSpec{
						Type: "List",
						Properties: []schema.NodeProperty{
							{"default", "222"},
							{"description", "bar"}}},
					Data: schema.NodeSubSpec{
						Type: "decimal64",
						Properties: []schema.NodeProperty{
							{"default", "111"}}},
				},
			},
		},
	}
	// Remove choice default and use type one instead.
	t.Skipf("Skipping ChoiceDefaultTests")
	runTestCases(t, ChoiceDefaultTests)
}

func TestEnumDefault(t *testing.T) {
	var EnumDefaultTests = []testutils.TestCase{
		{
			Description: "Enum: valid default",
			Template:    EnumDefaultTemplate,
			Schema:      `default "foo";`,
			ExpResult:   true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testLeaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "foo"}}},
					Data: schema.NodeSubSpec{
						Type: "enumeration",
						Properties: []schema.NodeProperty{
							{"default", "foo"}}},
				},
			},
		},
		{
			Description: "Enum: invalid default",
			Template:    EnumDefaultTemplate,
			Schema:      `default "foo3";`,
			ExpResult:   false,
			ExpErrs: errtest.YangInvalidDefaultEnumOrBoolErrorStrings(
				t, "test-yang-compile", "enum_test", "foo3",
				[]string{"foo", "foo2", "bar"}),
		},
		{
			Description: "Enum: invalid default, single line error",
			Template:    EnumSingleValueTemplate,
			Schema:      `default "foo3";`,
			ExpResult:   false,
			ExpErrs: errtest.YangInvalidDefaultEnumOrBoolErrorStrings(
				t, "test-yang-compile", "enum_test_sgl_val", "foo3",
				[]string{"fool"}),
		},
	}

	runTestCases(t, EnumDefaultTests)
}

func TestEmptyDefault(t *testing.T) {
	var EmptyDefaultTests = []testutils.TestCase{
		{
			Description: "Empty: invalid default",
			Template:    ContainerTemplate,
			Schema: `leaf emptyLeaf {
			type empty;
			default "66";
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultEmptyErrorStrings(
				t, errtest.Bltin, errtest.EmptyType, "66"),
		},
	}
	runTestCases(t, EmptyDefaultTests)
}

// This set of tests works through setting default (or not) for Leaf nodes,
// along with looking at refining a leaf which can change default and
// mandatory values.
//
// NB: The recursive tests use Leaves so some test cases are either duplicated
//     or not done here.
func TestLeafDefaultTest(t *testing.T) {
	var LeafDefaultTests = []testutils.TestCase{
		{
			Description: "Leaf Default: local default in range",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
			type uint8;
			default "66";
			description "topLeaf description";
		}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "topLeaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "66"},
							{"description", "topLeaf description"}}},
					Data: schema.NodeSubSpec{
						Type:       "uinteger",
						Properties: []schema.NodeProperty{}},
				},
			},
		},
		{
			Description: "Leaf Default: local default out of range",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
				type int8;
			    default "666";
			}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Int8Type, "666",
				[]errtest.NodeLimits{
					{Min: "-128", Max: "127"}}),
		},
		{
			Description: "Leaf Default: cannot have default and mandatory",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
				type int8;
			    default "66";
			    mandatory "true";
			}`,
			ExpResult: false,
			ExpErrMsg: "Leaf cannot have default and be mandatory",
		},
		{
			Description: "Leaf Default: can inherit default and be mandatory",
			Template:    DefaultTemplate,
			Schema: `leaf topLeaf {
				type uint8_base_default;
				description "topLeaf description";
			    mandatory "true";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "topLeaf"},
					DefaultPropNotPresent: true,
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", ""},
							{"mandatory", "true"},
						},
					},
					Data: schema.NodeSubSpec{
						Type:       "uinteger",
						Properties: []schema.NodeProperty{{"default", "99"}},
					},
				},
			},
		},
		{
			Description: "Leaf Default: invalid default (between ranges)",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
			type decimal64 {
				fraction-digits 6;
				range "1 .. 50 | 101 .. 150";
			}
			default "66";
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Dec64Type, "66",
				[]errtest.NodeLimits{
					{Min: "1.000000", Max: "50.000000"},
					{Min: "101.000000", Max: "150.000000"}}),
		},
		{
			Description: "Leaf Default: local default string not integer",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
			type uint8;
			default "not a number";
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultTypeErrorStrings(
				t, errtest.Bltin, errtest.Uint8Type, "not a number"),
		},
		{
			Description: "Leaf Default: refine leaf with default - " +
				"add mandatory",
			Template: LeafRefineTemplate,
			Schema: `refine test_container/test_uint8_def {
				mandatory "true";
			}`,
			ExpResult: false,
			ExpErrMsg: "Leaf cannot have default and be mandatory",
		},
		{
			Description: "Leaf Default: refine leaf with default - " +
				"make default invalid",
			Template: LeafRefineTemplate,
			Schema: `refine test_container/test_uint8_def {
				default "999";
			}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Uint8Type, "999",
				[]errtest.NodeLimits{
					{Min: "0", Max: "255"}}),
		},
		{
			Description: "Leaf Default: refine leaf with default - " +
				"change default",
			Template: LeafRefineTemplate,
			Schema: `refine test_container/test_uint8_def {
				default "22";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"test_target", "test_container",
						"test_uint8_def"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "22"},
							{"mandatory", "false"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "22"}}},
				},
			},
		},
		{
			Description: "Leaf Default: refine leaf with mandatory - " +
				"add default",
			Template: LeafRefineTemplate,
			Schema: `refine test_container/test_string_mandatory {
			    default "11";
			}`,
			ExpResult: false,
			ExpErrMsg: "Leaf cannot have default and be mandatory",
		},
		{
			Description: "Leaf Default: refine leaf with mandatory - " +
				"change mandatory",
			Template: LeafRefineTemplate,
			Schema: `refine test_container/test_string_mandatory {
				mandatory "false";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"test_target", "test_container",
						"test_string_mandatory"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"mandatory", "false"}}},
					Data: schema.NodeSubSpec{
						Type: "ystring",
						Properties: []schema.NodeProperty{
							{"name", "{builtin string}"}}},
				},
			},
		},
		{
			Description: "Leaf Default: refine leaf with mandatory - " +
				"rem man, add def",
			Template: LeafRefineTemplate,
			Schema: `refine test_container/test_string_mandatory {
				default "33";
				mandatory "false";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"test_target", "test_container",
						"test_string_mandatory"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "33"},
							{"mandatory", "false"}}},
					Data: schema.NodeSubSpec{
						Type: "ystring",
						Properties: []schema.NodeProperty{
							{"default", "33"}}},
				},
			},
		},
	}
	runTestCases(t, LeafDefaultTests)
}

// We process node 'attributes' in a fixed order, no matter what order they
// are specified in the YANG file ... at the point this test was written.
// As the validation code makes this assumption, we'd better just test that
// the order doesn't matter by testing here with default specified BEFORE
// range (elsewhere it's always after in this file).
//
// This test doubles as a test for the error messages as well.  Passing in
// an invalid default triggers the relevant error message without needing
// to run a config session which gets messier to test.
func TestRangeVsDefaultOrderTest(t *testing.T) {
	var RangeVsDefaultOrderTests = []testutils.TestCase{
		{
			Description: "RangeVsDefault: uint, default first, multiple ranges",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
			default "66";
			type uint8 {
				range "1 .. 50 | 101 .. 150";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Uint8Type, "66",
				[]errtest.NodeLimits{
					{Min: "1", Max: "50"},
					{Min: "101", Max: "150"}}),
		},
		{
			Description: "RangeVsDefault: uint, default first, single value",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
			default "66";
			type uint8 {
				range "65";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Uint8Type, "66",
				[]errtest.NodeLimits{
					{Min: "65", Max: "65"}}),
		},
		{
			Description: "RangeVsDefault: int, default first, multiple ranges",
			Template:    ContainerTemplate,
			Schema: `leaf topLeaf {
			default "77";
			type int8 {
				range "1 .. 50 | 101 .. 120";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Int8Type, "77",
				[]errtest.NodeLimits{
					{Min: "1", Max: "50"},
					{Min: "101", Max: "120"}}),
		},
		{
			Description: "RangeVsDefault: int, default first, " +
				"single valid value",
			Template: ContainerTemplate,
			Schema: `leaf topLeaf {
			default "77";
			type int8 {
				range "78";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Int8Type, "77",
				[]errtest.NodeLimits{
					{Min: "78", Max: "78"}}),
		},
		{
			Description: "RangeVsDefault: dec64, default first, " +
				"multiple ranges",
			Template: ContainerTemplate,
			Schema: `leaf topLeaf {
			default "66.6";
			type decimal64 {
				fraction-digits 4;
				range "1 .. 50 | 101 .. 150";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Dec64Type, "66.6",
				[]errtest.NodeLimits{
					{Min: "1.000000", Max: "50.000000"},
					{Min: "101.000000", Max: "150.000000"}}),
		},
		{
			Description: "RangeVsDefault: dec64, default first, " +
				"single value range",
			Template: ContainerTemplate,
			Schema: `leaf topLeaf {
			default "66.6";
			type decimal64 {
				fraction-digits 4;
				range "50";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Dec64Type, "66.6",
				[]errtest.NodeLimits{
					{Min: "50.000000", Max: "50.000000"}}),
		},
		{
			Description: "RangeVsDefault: string, default first, " +
				"multiple ranges",
			Template: ContainerTemplate,
			Schema: `leaf topLeaf {
			default "333333";
			type string {
				length "1 .. 5 | 7";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultLengthErrorStrings(
				t, errtest.Bltin, errtest.StringType, "333333",
				[]errtest.NodeLimits{
					{Min: "1", Max: "5"}, {Min: "7", Max: "7"}}),
		},
		{
			Description: "RangeVsDefault: string, default first, " +
				"single length valid",
			Template: ContainerTemplate,
			Schema: `leaf topLeaf {
			default "333333";
			type string {
				length "3";
			}
		}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultLengthErrorStrings(
				t, errtest.Bltin, errtest.StringType, "333333",
				[]errtest.NodeLimits{
					{Min: "3", Max: "3"}}),
		},
	}
	runTestCases(t, RangeVsDefaultOrderTests)
}

// Recursive default tests look at multiple level typedefs to ensure that
// defaults are correctly inherited and validated.
//
// Base typedefs are range, default, range_and_default (and none!)
func TestRecursiveDefault(t *testing.T) {
	var RecursiveDefaultTests = []testutils.TestCase{
		{
			Description: "RecursiveDefaultTest: invalid inherited default",
			Template:    DefaultTemplate,
			Schema: `typedef uint8_invalid_inherited_dflt {
				type uint8_base_range_and_default {
					range "1..90 | 200 .. 250";
				}
			}
			leaf invalid_inherited_dflt_leaf {
				type uint8_invalid_inherited_dflt;
			}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, "test-yang-compile", "uint8_base_range_and_default", "99",
				[]errtest.NodeLimits{
					{Min: "1", Max: "90"},
					{Min: "200", Max: "250"}}),
		},
		{
			Description: "RecursiveDefaultTest: valid inherited default",
			Template:    DefaultTemplate,
			Schema: `typedef uint8_valid_inherited_dflt {
				type uint8_base_range_and_default {
					range "1..100 | 220 .. 250";
				}
			}
			leaf valid_inherited_dflt_leaf {
				type uint8_valid_inherited_dflt;
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer",
						"valid_inherited_dflt_leaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "99"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "99"}}},
				},
			},
		},
		{
			Description: "RecursiveDefaultTest: override invalid inherited " +
				"default",
			Template: DefaultTemplate,
			Schema: `typedef uint8_override_invalid_dflt {
			    default "80";
				type uint8_base_range_and_default {
					range "1..90 | 200 .. 250";
				}
			}
			leaf override_invalid_inherited_dflt_leaf {
				type uint8_override_invalid_dflt;
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer",
						"override_invalid_inherited_dflt_leaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "80"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "80"}}},
				},
			},
		},
		{
			Description: "RecursiveDefaultTest: override valid " +
				"inherited default",
			Template: DefaultTemplate,
			Schema: `typedef uint8_override_valid_default {
				type uint8_base_range_and_default;
			    default "60";
			}
			leaf override_valid_inherited_dflt_leaf {
				type uint8_override_valid_default;
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer",
						"override_valid_inherited_dflt_leaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "60"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "60"}}},
				},
			},
		},
		{
			Description: "RecursiveDefaultTest: leaf override " +
				"inherited default",
			Template: DefaultTemplate,
			Schema: `typedef uint8_override_valid_default {
				type uint8_base_range_and_default;
			    default "60";
			}
			leaf leaf_override_inherited_dflt_leaf {
				type uint8_override_valid_default;
			    default "30";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer",
						"leaf_override_inherited_dflt_leaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "30"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "30"}}},
				},
			},
		},
		{
			Description: "RecursiveDefaultTest: inherited default, " +
				"leaf def is empty string",
			Template: DefaultTemplate,
			Schema: `leaf leaf_override_inherited_dflt_leaf_empty_str {
				type string_base_length_and_default;
			    default "";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer",
						"leaf_override_inherited_dflt_leaf_empty_str"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", ""}}},
					Data: schema.NodeSubSpec{
						Type:       "ystring",
						Properties: []schema.NodeProperty{{"default", ""}}},
				},
			},
		},
		{
			Description: "RecursiveDefaultTest: no inherited default",
			Template:    DefaultTemplate,
			Schema: `typedef uint8_no_default_inherited {
				type uint8_base_range {
					range "1..100 | 200 .. 255";
				}
			}
			leaf no_inherited_default {
				type uint8_no_default_inherited;
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{
						"testContainer", "no_inherited_default"},
					DefaultPropNotPresent: true,
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", ""}}},
					DataPropNotPresent: true,
					Data: schema.NodeSubSpec{
						Type:       "uinteger",
						Properties: []schema.NodeProperty{{"default", ""}}},
				},
			},
		},
		{
			Description: "RecursiveDefaultTest: no inherited default, " +
				"set locally",
			Template: DefaultTemplate,
			Schema: `typedef uint8_no_default_inherited {
				type uint8_base_range {
					range "1..80 | 220 .. 240";
				}
			}
			leaf no_inherited_dflt_set_on_leaf {
				type uint8_no_default_inherited;
			    default "33";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer",
						"no_inherited_dflt_set_on_leaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "33"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "33"}}},
				},
			},
		},
		{
			Description: "RecursiveDefaultTest: no inherited default " +
				"or range, set locally",
			Template: DefaultTemplate,
			Schema: `typedef uint8_no_default_inherited {
				type uint8_base_range;
			}
			leaf no_inherited_dflt_set_on_leaf {
				type uint8_no_default_inherited;
			default "33";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer",
						"no_inherited_dflt_set_on_leaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "33"}}},
					Data: schema.NodeSubSpec{
						Type: "uinteger",
						Properties: []schema.NodeProperty{
							{"default", "33"}}},
				},
			},
		},
	}
	runTestCases(t, RecursiveDefaultTests)
}

func TestStringDefaultTest(t *testing.T) {
	var StringDefaultTests = []testutils.TestCase{
		{
			Description: "Typedef: string default substatement within range",
			Template:    TypedefTemplate,
			Schema: `base_string_with_range {
                length "2..4 | 6 .. 8"; } default "66";`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testleaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "66"}}},
					Data: schema.NodeSubSpec{
						Type: "ystring",
						Properties: []schema.NodeProperty{
							{"default", "66"}}},
				},
			},
		},
		{
			Description: "Typedef: empty string default within range",
			Template:    TypedefTemplate,
			Schema: `base_string_with_range {
                length "0 .. 4 | 6 .. 8"; } default "";`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "testleaf"},
					Statement: schema.NodeSubSpec{
						Type:       "leaf",
						Properties: []schema.NodeProperty{{"default", ""}}},
					Data: schema.NodeSubSpec{
						Type:       "ystring",
						Properties: []schema.NodeProperty{{"default", ""}}},
				},
			},
		},
		{
			Description: "Typedef: 2 string default substatements (illegal)",
			Template:    TypedefTemplate,
			Schema:      `base_string_with_range; default "6"; default "7";`,
			ExpResult:   false,
			ExpErrMsg:   OnlyOneDefaultAllowedStr,
		},
		{
			Description: "Typedef: default substatement outwith " +
				"valid lengths (string)",
			Template: TypedefTemplate,
			Schema: `base_string_with_range {
                length "1..2 | 6 .. 8"; } default "666";`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultLengthErrorStrings(
				t, "test-yang-compile", "base_string_with_range", "666",
				[]errtest.NodeLimits{
					{Min: "1", Max: "2"}, {Min: "6", Max: "8"}}),
		},
	}
	runTestCases(t, StringDefaultTests)
}

// Union type does not inherit a default from constituent parts, but
// presumably we can have a default specified in a typedef that uses
// a type which is a union.
func TestUnionDefault(t *testing.T) {
	var UnionDefaultTests = []testutils.TestCase{
		{
			Description: "UnionDefaultTest: member type has invalid default",
			Template:    ContainerTemplate,
			Schema: `typedef string_with_default {
				type string;
			    default "a string";
			}
			typedef int8_with_invalid_default {
				type int8;
				default "129";
			}
			leaf union_leaf {
				type union {
					type string_with_default;
					type int8_with_invalid_default;
				}
			}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Int8Type, "129",
				[]errtest.NodeLimits{
					{Min: "-128", Max: "127"}}),
		},
		{
			Description: "UnionDefaultTest: invalid default, " +
				"multiple range error",
			Template: ContainerTemplate,
			Schema: `typedef string_with_default {
				type string;
			    default "a string";
			}
			typedef int8_with_invalid_default {
				type int8 {
					range "1 .. 10 | 100 | 120 .. 127";
				}
				default "129";
			}
			leaf union_leaf {
				type union {
					type string_with_default;
					type int8_with_invalid_default;
				}
			}`,
			ExpResult: false,
			ExpErrs: errtest.YangInvalidDefaultValueErrorStrings(
				t, errtest.Bltin, errtest.Int8Type, "129",
				[]errtest.NodeLimits{
					{Min: "1", Max: "10"},
					{Min: "100", Max: "100"},
					{Min: "120", Max: "127"}}),
		},
		{
			Description: "UnionDefaultTest: inherited default ignored",
			Template:    ContainerTemplate,
			Schema: `typedef string_with_default {
				type string;
			    default "a string";
			}
			typedef int8_with_valid_default {
				type int8;
				default "127";
			}
			leaf union_leaf {
				type union {
					type string_with_default;
					type int8_with_valid_default;
				}
				description "union_leaf description";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "union_leaf"},
					Statement: schema.NodeSubSpec{
						Type:       "leaf",
						Properties: []schema.NodeProperty{}},
					DataPropNotPresent: true,
					Data: schema.NodeSubSpec{
						Type: "union",

						Properties: []schema.NodeProperty{{"default", ""}}},
				},
			},
		},
		{
			Description: "UnionDefaultTest: local default in typedef allowed",
			Template:    ContainerTemplate,
			Schema: `typedef string_with_default {
				type string;
			    default "a string";
			}
			typedef int8_with_valid_default {
				type int8;
				default "127";
			}
			leaf union_leaf {
				type union {
					type string_with_default;
					type int8_with_valid_default;
				}
				default "125";
			}`,
			ExpResult: true,
			NodesToValidate: []schema.NodeSpec{
				{
					Path: []string{"testContainer", "union_leaf"},
					Statement: schema.NodeSubSpec{
						Type: "leaf",
						Properties: []schema.NodeProperty{
							{"default", "125"}}},
					Data: schema.NodeSubSpec{
						Type: "union",
						Properties: []schema.NodeProperty{
							{"default", "125"}}},
				},
			},
		},
	}

	runTestCases(t, UnionDefaultTests)
}

func TestIdentityRefDefault(t *testing.T) {
	t.Skipf("TBD: Default Test for identityref.")
}

func TestInstanceIdentifierDefault(t *testing.T) {
	t.Skipf("TBD: Default Test for instance-identifier.")
}

func TestLeafRefDefault(t *testing.T) {
	t.Skipf("TBD: Default Test for leafref.")
}
