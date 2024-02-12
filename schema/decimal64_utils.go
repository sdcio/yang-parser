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

// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const (
	maxDecimal64 = math.MaxInt64
	minDecimal64 = math.MinInt64
)
const (
	maxFractionalDigits = 18
	minFractionalDigits = 1
)
const (
	errorStringNullInput =
		 "Decimal64 must contain at least one decimal digit"
	errorStringSign =
		"Invalid input: Decimal64 values must begin with +/- or a decimal digit"
	errorStringExcessDecimalPoint =
		"Invalid input: More than one `.` rune was found"
	errorStringFractionDigitMismatch =
		"Number of fractional digits in input string is greater than int parameter"
	errorStringValueAboveMaximum =
		"Value is greater than maximum decimal64"
	errorStringValueBelowMinimum =
		"Value is less than minimum decimal64"
	errorStringInvalidFractionDigitsParam =
		"Parameter fractionDigitsAllowed must be 1 <= n <= 18"
	errorStringMissingDigits =
		"Decimal64 must have at least 1 integer digit and at least 1 decimal digit"
)

type validateDecimal64Error struct {
	err string
}

func (e *validateDecimal64Error) Error() string {
	return fmt.Sprintf("Invalid decimal64: %s", e.err)
}

func newValidateDecimal64Error(s string) error {
	return &validateDecimal64Error{err: s}
}

// Implement validation for decimal64 values according to RFC6020: 9.3
func validateDecimal64String(s string, fractionDigitsAllowed int) error {
	if fractionDigitsAllowed < minFractionalDigits ||
		fractionDigitsAllowed > maxFractionalDigits {

		return newValidateDecimal64Error(errorStringInvalidFractionDigitsParam)
	}

	if len(s) == 0 {
		return newValidateDecimal64Error(errorStringNullInput)
	}

	if s[0] != '+' && s[0] != '-' && !(s[0] >= '0' && s[0] <= '9') {
		return newValidateDecimal64Error(errorStringSign)
	}

	sSplit := strings.Split(s, ".")
	if len(sSplit) == 1 {
		_, err := strconv.ParseInt(sSplit[0], 10, 64)
		if err != nil {
			return newValidateDecimal64Error(
				fmt.Sprintf("Error parsing digits: %s", err))
		}
		return nil
	}
	if len(sSplit) > 2 {
		return newValidateDecimal64Error(errorStringExcessDecimalPoint)
	}

	if len(sSplit[0]) == 0 || len(sSplit[1]) == 0 {
		return newValidateDecimal64Error(errorStringMissingDigits)
	}

	fractionDigitsActual := len(sSplit[1])
	if fractionDigitsActual > fractionDigitsAllowed {
		return newValidateDecimal64Error(errorStringFractionDigitMismatch)
	}
	// To make sure .1 > 0.01, we must pad the value before parsing as an integer
	if fractionDigitsActual < fractionDigitsAllowed {
		sSplit[1] = sSplit[1] +
			strings.Repeat("0", fractionDigitsAllowed - fractionDigitsActual)
	}

	denominator := int64(math.Pow10(fractionDigitsAllowed))

	upperBits, err := strconv.ParseInt(sSplit[0], 10, 64)
	if err != nil {
		return newValidateDecimal64Error(
			fmt.Sprintf("Error parsing upper digits: %s", err))
	}
	if upperBits > maxDecimal64/denominator {
		return newValidateDecimal64Error(errorStringValueAboveMaximum)
	}
	if upperBits < minDecimal64/denominator {
		return newValidateDecimal64Error("Value is less than minimum decimal64")
	}

	lowerBits, err := strconv.ParseInt(sSplit[1], 10, 64)
	if err != nil {
		return newValidateDecimal64Error(
			fmt.Sprintf("Error parsing lower digits: %s", err))
	}
	if upperBits == maxDecimal64/denominator {
		if lowerBits > maxDecimal64%denominator {
			return newValidateDecimal64Error(errorStringValueAboveMaximum)
		}
	}
	if upperBits == minDecimal64/denominator {
		if lowerBits > maxDecimal64%denominator+1 {
			return newValidateDecimal64Error(errorStringValueBelowMinimum)
		}
	}

	return nil
}
