// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved
//
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"testing"

	"github.com/danos/yang/schema"
)

func runValidateDecimal64StringAndCheckFails(t *testing.T, s string, fractionalDigits int) {
	err := schema.ValidateDecimal64String(s, fractionalDigits)
	if err == nil {
		t.Errorf("Input: %q failed to produce error where expected", s)
	}
}

func runValidateDecimal64StringAndCheckPasses(t *testing.T, s string, fractionalDigits int) {
	err := schema.ValidateDecimal64String(s, fractionalDigits)
	if err != nil {
		t.Errorf("Input: %q produced unexpected error: %s", s, err)
	}
}

type testNameAndInput struct {
	name                         string
	inputString                  string
	inputAllowedFractionalDigits int
}

func TestValidateDecimal64String(t *testing.T) {
	testsGoodInput := []testNameAndInput{
		{name: "No fractional digits in string",
			inputString: "+1000", inputAllowedFractionalDigits: 2},
		{name: "No leading sign (+/-)",
			inputString: "111333.12", inputAllowedFractionalDigits: 2},
		{name: "Minimum allowed value",
			inputString: "-9223372036854775.808", inputAllowedFractionalDigits: 3},
		{name: "Maximum allowed value",
			inputString: "+9223372036854775.807", inputAllowedFractionalDigits: 3},
		{name: "Positive value with 3 fractional digits",
			inputString: "+100.808", inputAllowedFractionalDigits: 3},
		{name: "Positive value with 6 fractional digits",
			inputString: "+100.808003", inputAllowedFractionalDigits: 6},
		{name: "Negative value with implied fractional digits",
			inputString: "-9223372036854775.8", inputAllowedFractionalDigits: 3},
		{name: "Positive value with 1 fractional digit",
			inputString: "-9223372036854375.0", inputAllowedFractionalDigits: 1},
		{name: "Negative value with 1 fractional digit",
			inputString: "+9223372036854775.0", inputAllowedFractionalDigits: 1},
		{name: "Positive value with 18 fractional digits",
			inputString: "+9.2233720368547750", inputAllowedFractionalDigits: 18},
		{name: "Negative value with 18 fractional digits",
			inputString: "-9.2233720362544750", inputAllowedFractionalDigits: 18},
		{name: "Zero, with 18 Fd",
			inputString: "+0.0000000000000000", inputAllowedFractionalDigits: 18},
		{name: "Zero, with 1 Fd",
			inputString: "+0.0", inputAllowedFractionalDigits: 1},
		{name: "Minus zero, with 18 Fd",
			inputString: "-0.0000000000000000", inputAllowedFractionalDigits: 18},
		{name: "Minus zero, with 1 Fd",
			inputString: "-0.0", inputAllowedFractionalDigits: 1},
		// The following cases are not technically correct,
		// however we allow them to avoid a regression. They would work for strconv.Float()
		{name: "Leading 0s",
			inputString: "0000009.2233720368547750", inputAllowedFractionalDigits: 18},
	}
	for _, test := range testsGoodInput {
		t.Run(test.name, func(t *testing.T) {
			runValidateDecimal64StringAndCheckPasses(t,
				test.inputString, test.inputAllowedFractionalDigits)
		})
	}

	testsBadInput := []testNameAndInput{
		{name: "Not Numeric", inputString: "foobar", inputAllowedFractionalDigits: 2},
		{name: "Hex number", inputString: "+FFFF.FF", inputAllowedFractionalDigits: 2},
		{name: "Two decimal points",
			inputString: "+1113.33.123", inputAllowedFractionalDigits: 2},
		{name: "Below minimum: (minimum - 0.002)",
			inputString: "-9223372036854775.810", inputAllowedFractionalDigits: 3},
		{name: "Above maximum: (maximum + 2)",
			inputString: "+9223372036854777.807", inputAllowedFractionalDigits: 3},
		{name: "Extra fractional digits",
			inputString: "+9223372036854766.7007", inputAllowedFractionalDigits: 3},
		{name: "Positive value above maximum with 1 fractional digit",
			inputString: "+922337203685477580.8", inputAllowedFractionalDigits: 1},
		{name: "Negative value below minimum with 1 fractional digit",
			inputString: "-922337203685477580.9", inputAllowedFractionalDigits: 1},
		{name: "Positive value above maximum with 18 fractional digits",
			inputString: "+9.223372036854775808", inputAllowedFractionalDigits: 18},
		{name: "Negative value below minimum with 18 fractional digits",
			inputString: "-9.223372036854775809", inputAllowedFractionalDigits: 18},
		{name: "Non-numeric fractional digits",
			inputString: "+922337203685477.BA08", inputAllowedFractionalDigits: 4},
		{name: "Excess integer digits",
			inputString: "+100922337203685477.5807", inputAllowedFractionalDigits: 4},
		{name: "Two signs",
			inputString: "-+9.223372036854775808", inputAllowedFractionalDigits: 18},
		{name: "Sign after decimal point",
			inputString: "+9.+223372036854775808", inputAllowedFractionalDigits: 18},
		{name: "More than 18 fractional digits",
			inputString: "+9.2233720368547758081", inputAllowedFractionalDigits: 19},
		{name: "No fractional digits",
			inputString: "+9.2233720368547758081", inputAllowedFractionalDigits: 0},
		{name: "Positive value above maximum with 18 fractional digits",
			inputString: "+9.9", inputAllowedFractionalDigits: 18},
		{name: "Above maximum: (maximum + 0.003)",
			inputString: "+9223372036854775.81", inputAllowedFractionalDigits: 3},
		{name: "Above maximum: (maximum + 0.093)",
			inputString: "+9223372036854775.9", inputAllowedFractionalDigits: 3},
		{name: "Above maximum: (maximum + 0.003)",
			inputString: "+9223372036854775.810", inputAllowedFractionalDigits: 3},
		{name: "Above maximum: (maximum + 0.093)",
			inputString: "+9223372036854775.900", inputAllowedFractionalDigits: 3},
		{name: "Excess fractional digits: Trailing 0s",
			inputString: "+9223372036854775.8070000", inputAllowedFractionalDigits: 3},
		{name: "No explicit fractional digits",
			inputString: "+1.", inputAllowedFractionalDigits: 1},
		{name: "No explicit integer digits",
			inputString: ".1", inputAllowedFractionalDigits: 1},
	}
	for _, test := range testsBadInput {
		t.Run(test.name, func(t *testing.T) {
			runValidateDecimal64StringAndCheckFails(t,
				test.inputString, test.inputAllowedFractionalDigits)
		})
	}

}
