// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved
//
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"testing"

	"github.com/danos/yang/yangutils"
)

func runValidateDecimal64StringAndCheckFails(t *testing.T, s string, fractionalDigits int) {
	err := yangutils.ValidateDecimal64String(s, fractionalDigits)
	if err == nil {
		t.Errorf("Input: %q failed to produce error where expected", s)
	}
}

func runValidateDecimal64StringAndCheckPasses(t *testing.T, s string, fractionalDigits int) {
	err := yangutils.ValidateDecimal64String(s, fractionalDigits)
	if err != nil {
		t.Errorf("Input: %q produced unexpected error: %s", s, err)
	}
}

func TestValidateDecimal64String(t *testing.T) {
	const badInputNotNumeric = "potato"
	runValidateDecimal64StringAndCheckFails(t, badInputNotNumeric, 2)

	const badInputHexInput = "+FFFF.FF"
	runValidateDecimal64StringAndCheckFails(t, badInputHexInput, 2)

	const goodInputNoFractionalDigits = "+1000"
	runValidateDecimal64StringAndCheckPasses(t, goodInputNoFractionalDigits, 2)

	const goodInputNoSign = "111333.12"
	runValidateDecimal64StringAndCheckPasses(t, goodInputNoSign, 2)

	const badInput2DecimalPoint = "+1113.33.123"
	runValidateDecimal64StringAndCheckFails(t, badInput2DecimalPoint, 3)

	// Minimum int64 is -9223372036854775808
	const goodInputNegativeMinimum3Fd = "-9223372036854775.808"
	runValidateDecimal64StringAndCheckPasses(t, goodInputNegativeMinimum3Fd, 3)

	const badInputBelowMinimumMinusPt002 = "-9223372036854775.810"
	runValidateDecimal64StringAndCheckFails(t, badInputBelowMinimumMinusPt002, 3)

	// Maximum int64 is +9223372036854775807
	const goodInputPositiveMaximum3Fd = "+9223372036854775.807"
	runValidateDecimal64StringAndCheckPasses(t, goodInputPositiveMaximum3Fd, 3)

	const badInputAboveMaximumPlus2 = "+9223372036854777.807"
	runValidateDecimal64StringAndCheckFails(t, badInputAboveMaximumPlus2, 3)

	const badInputWrongFractionDigits = "+9223372036854777.8007"
	runValidateDecimal64StringAndCheckFails(t, badInputWrongFractionDigits, 3)

	const goodInputPositive3Fd = "+100.808"
	runValidateDecimal64StringAndCheckPasses(t, goodInputPositive3Fd, 3)

	const goodInputPositive6Fd = "+100.808003"
	runValidateDecimal64StringAndCheckPasses(t, goodInputPositive6Fd, 6)

	const goodInputNegative1Fd = "+9223372036854775.0"
	runValidateDecimal64StringAndCheckPasses(t, goodInputNegative1Fd, 1)

	const goodInputNegativeImplicitDigit3Fd = "-9223372036854775.8"
	runValidateDecimal64StringAndCheckPasses(t, goodInputNegativeMinimum3Fd, 3)
}
