// Copyright (c) 2021, AT&T Intellectual Property. All rights reserved
//
// SPDX-License-Identifier: MPL-2.0

package yangutils

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

const maxDecimal64 = math.MaxInt64
const minDecimal64 = math.MinInt64

const ErrorStringNullInput = "Decimal64 must contain at least one decimal digit"
const ErrorStringSign = "Invalid input: Decimal64 values must begin with +/- or a decimal digit"
const ErrorStringExcessDecimalPoint = "Invalid input: More than one `.` rune was found"
const ErrorStringFractionDigitMismatch = "Number of fractional digits in input string is greater than int parameter"
const ErrorStringValueAboveMaximum = "Value is greater than maximum decimal64"
const ErrorStringValueBelowMinimum = "Value is less than minimum decimal64"

type ValidateDecimal64Error struct {
	err string
}

func (e *ValidateDecimal64Error) Error() string {
	return fmt.Sprintf("Invalid decimal64: %s", e.err)
}

func NewValidateDecimal64Error(s string) error {
	return &ValidateDecimal64Error{err: s}
}

// Helper method, finds 10^exponent. Does not sanitise input.
func pow10Int64(exponent int) int64 {
	product := int64(1)
	for i := 0; i < exponent; i++ {
		product *= 10
	}
	return product
}

// Implement validation for decimal64 values according to RFC6020: 9.3
func ValidateDecimal64String(s string, fractionDigitsExpected int) error {
	if len(s) == 0 {
		return NewValidateDecimal64Error(ErrorStringNullInput)
	}

	if s[0] != '+' && s[0] != '-' && !(s[0] >= '0' && s[0] <= '9') {
		return NewValidateDecimal64Error(ErrorStringSign)
	}

	sSplit := strings.Split(s, ".")
	if len(sSplit) == 1 {
		_, err := strconv.ParseInt(sSplit[0], 10, 64)
		if err != nil {
			return NewValidateDecimal64Error(fmt.Sprintf("Error parsing digits: %s", err))
		}
		return nil
	}
	if len(sSplit) > 2 {
		return NewValidateDecimal64Error(ErrorStringExcessDecimalPoint)
	}

	fractionDigitsActual := len(sSplit[1])
	// fractionDigitsExpected = 0 disables this check
	if fractionDigitsExpected != 0 && fractionDigitsActual > fractionDigitsExpected {
		return NewValidateDecimal64Error(ErrorStringFractionDigitMismatch)
	}

	upperBits, err := strconv.ParseInt(sSplit[0], 10, 64)
	if err != nil {
		return NewValidateDecimal64Error(fmt.Sprintf("Error parsing upper digits: %s", err))
	}
	if upperBits > maxDecimal64/pow10Int64(fractionDigitsActual) {
		return NewValidateDecimal64Error(ErrorStringValueAboveMaximum)
	}
	if upperBits < minDecimal64/pow10Int64(fractionDigitsActual) {
		return NewValidateDecimal64Error("Value is less than minimum decimal64")
	}

	lowerBits, err := strconv.ParseInt(sSplit[1], 10, 64)
	if err != nil {
		return NewValidateDecimal64Error(fmt.Sprintf("Error parsing lower digits: %s", err))
	}
	if upperBits == maxDecimal64/pow10Int64(fractionDigitsActual) {
		if lowerBits > maxDecimal64%pow10Int64(fractionDigitsActual) {
			return NewValidateDecimal64Error(ErrorStringValueAboveMaximum)
		}
	}
	if upperBits == minDecimal64/pow10Int64(fractionDigitsActual) {
		if lowerBits > maxDecimal64%pow10Int64(fractionDigitsActual)+1 {
			return NewValidateDecimal64Error(ErrorStringValueBelowMinimum)
		}
	}

	return nil
}
