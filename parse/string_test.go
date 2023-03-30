// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse_test

import (
	"testing"

	"github.com/steiler/yang-parser/parse"
	. "github.com/steiler/yang-parser/parse/parsetest"
)

// This is what the contents of the "contact" field in string_pass_concat.yang
// after it has been concatenated and trimmed of whitespace
const expectedConcat = "AT&T\n" +
	"Postal: 208 S. Akard Street\n" +
	"        Dallas, TX 75202\n" +
	"\t Web: www.att.com\r\n" +
	"   Some additional test data SingleQuote with space and newline to be preserved\n" +
	"       Second SingleQuote DoubleQuote With WS with NewLine" +
	" Trim Next 3 WS\nTrim before here"

// Parse a schema which contains an example of a string which requires
// multiple concatenations and WS trimming.
// Validate the parsed string is as expected
func TestStringConcatenation(t *testing.T) {
	VerifyExpectedArgument(t, "testschemas/string_pass_concat.yang",
		parse.NodeContact, expectedConcat)
}

// This is what the contents of the "contact" field in string_pass_esc_seq.yang
// after escape sequence substitution
const expectedEscSeq = "AT&T\n" +
	"Postal: 208 S. Akard Street\n" +
	"        Dallas, TX 75202\n" +
	"\t Web: www.att.com\r\n" +
	"   Some additional test data" +
	"\\\\ \\n \\t \\r \\n \\. \\*\\* \\\\\\\\\\\\\\ Left untouched!!\\ \\" +
	"\t\t    Two LineBreak Variants\n\r\nTwo LinesLine 1:\nLine 2:\r\n" +
	" Let's quote something \"2 Bee or not 2 Bee\"  "

// Parse a schema which contains escape sequences and verify that escaping is
// performed correctly
func TestStringEscSeq(t *testing.T) {
	VerifyExpectedArgument(t, "testschemas/string_pass_esc_seq.yang",
		parse.NodeContact, expectedEscSeq)
}

func TestXmlIdentifierRejected(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_xml.yang",
		"invalid identifier, not allowed to start with xml: XmLtestcontainer")
}

func TestDoublePlusRejected(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_double_plus.yang",
		"unexpected [Plus] <+> in argument of contact")
}

func TestDanglingPlusRejected(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_dangling_plus.yang",
		"unexpected [SemiColon] \";\" in argument of contact")
}

func TestMissingPlusRejected(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_miss_plus.yang",
		"unexpected [Quote] <\"> in argument of contact")
}

func TestNonQuoteConcatRejected(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_quotes.yang",
		"unexpected [String] \"nonQuoteString\" in argument of contact")
}

func TestMissingSemiColon(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_miss_semicolon.yang",
		"unexpected [String] \"revision\" in argument of contact")
}

func TestRejectSquoteInSquote(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_squote_in_squote.yang",
		"unexpected [String] \"No_single_quote_escaping\" in argument of contact")
}

func TestRejectSpaceInUnquotedString(t *testing.T) {
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_space.yang",
		"unexpected [String] \"lean\" in type boo")

}

func TestBadEscSeqRejected(t *testing.T) {
	t.Skipf("Escape Sequences substitution does not yet enforce illegal sequences")
	VerifyParseErrorIsSeen(t,
		"testschemas/string_fail_bad_esc_seq.yang",
		"Illegal escape sequence found in string <\\*>")
}
