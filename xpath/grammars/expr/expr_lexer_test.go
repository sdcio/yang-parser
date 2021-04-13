// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// TODO

// - Will need a way to pass multiple expected tokens in against single
//   expression as we need to be able to consume as we go.

package expr

import (
	"bytes"
	"encoding/xml"
	"testing"

	. "github.com/danos/yang/xpath/grammars/lexertest"
	"github.com/danos/yang/xpath/xutils"
)

// The UTF-8 encoding for 0xD800, which is invalid in a UTF-8 stream,
// but is allowed in a CESU-8 stream.
const badUTF8 = "\xED\xA0\x80"

// Test NUM parsing
func TestLexInt(t *testing.T) {
	lexLine := NewExprLex("123", nil, nil)

	CheckNumToken(t, lexLine, 123)
}

func TestLexLeadingZero(t *testing.T) {
	lexLine := NewExprLex("0101", nil, nil)

	CheckNumToken(t, lexLine, 101)
}

func TestLeadingDotLex(t *testing.T) {
	lexLine := NewExprLex(".987", nil, nil)

	CheckNumToken(t, lexLine, 0.987)
}

func TestLexIntermediateDot(t *testing.T) {
	lexLine := NewExprLex("123.987", nil, nil)

	CheckNumToken(t, lexLine, 123.987)
}

func TestLexTrailingDot(t *testing.T) {
	lexLine := NewExprLex("654.", nil, nil)

	CheckNumToken(t, lexLine, 654)
}

// E / e
func TestLexE(t *testing.T) {
	lexLine := NewExprLex("654E2", nil, nil)

	CheckNumToken(t, lexLine, 654e2)
}

func TestLexe(t *testing.T) {
	lexLine := NewExprLex("654e2", nil, nil)

	CheckNumToken(t, lexLine, 654e2)
}

// Test whitespace
func TestLexWhitespace(t *testing.T) {
	lexLine := NewExprLex(" \t\n\r", nil, nil)

	CheckToken(t, lexLine, xutils.EOF)
}

// Test punctuation
func TestLexPunctuation(t *testing.T) {
	lexLine := NewExprLex("()@,[]|", nil, nil)

	CheckToken(t, lexLine, int('('))
	CheckToken(t, lexLine, int(')'))
	CheckToken(t, lexLine, int('@'))
	CheckToken(t, lexLine, int(','))
	CheckToken(t, lexLine, int('['))
	CheckToken(t, lexLine, int(']'))
	CheckToken(t, lexLine, int('|'))

	CheckToken(t, lexLine, xutils.EOF)
}

// Test operators
func TestLexMathematicalOperators(t *testing.T) {
	lexLine := NewExprLex("+-", nil, nil)

	CheckToken(t, lexLine, int('+'))
	CheckToken(t, lexLine, int('-'))

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexBooleanOperators(t *testing.T) {
	lexLine := NewExprLex("= != > >= < <=", nil, nil)

	CheckToken(t, lexLine, EQ)
	CheckToken(t, lexLine, NE)
	CheckToken(t, lexLine, GT)
	CheckToken(t, lexLine, GE)
	CheckToken(t, lexLine, LT)
	CheckToken(t, lexLine, LE)

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexInvalidRelationalOperators(t *testing.T) {
	lexLine := NewExprLex("! ", nil, nil)
	CheckUnlexableToken(t, lexLine, "'!' only valid when followed by '='")
}

// Cannot have operatorName as first token, nor 2 adjacent operator names
func TestLexOperatorNames(t *testing.T) {
	lexLine := NewExprLex("0 and 1 or 2 mod 3 div 4", nil, nil)

	CheckNumToken(t, lexLine, 0)
	CheckToken(t, lexLine, AND)
	CheckNumToken(t, lexLine, 1)
	CheckToken(t, lexLine, OR)
	CheckNumToken(t, lexLine, 2)
	CheckToken(t, lexLine, MOD)
	CheckNumToken(t, lexLine, 3)
	CheckToken(t, lexLine, DIV)
	CheckNumToken(t, lexLine, 4)

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexAsteriskAsMultiply(t *testing.T) {
	lexLine := NewExprLex("number(10 * 2) + intfName * 20", nil, nil)

	CheckFuncToken(t, lexLine, "number")
	CheckToken(t, lexLine, int('('))
	CheckNumToken(t, lexLine, 10)
	CheckToken(t, lexLine, int('*'))
	CheckNumToken(t, lexLine, 2)
	CheckToken(t, lexLine, int(')'))
	CheckToken(t, lexLine, int('+'))
	CheckNameTestToken(t, lexLine, xml.Name{Local: "intfName"})
	CheckToken(t, lexLine, int('*'))
	CheckNumToken(t, lexLine, 20)

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexAsteriskAsNameTest(t *testing.T) {
	lexLine := NewExprLex(
		"@* ::* (* [* ,* "+ /* Disambiguation rule */
			"and* or* mod* div* "+ /* Operator Names */
			"** /* //* |* =* !=* <* <=* >* >=*", /* Operators */
		nil, nil)

	// Disambiguation rule ...
	CheckToken(t, lexLine, int('@'))
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, DBLCOLON)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, int('('))
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, int('['))
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, int(','))
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())

	// Operator names
	CheckToken(t, lexLine, AND)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, OR)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, MOD)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, DIV)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())

	// Operators
	CheckToken(t, lexLine, int('*'))
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, int('/'))
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, DBLSLASH)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, int('|'))
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())

	CheckToken(t, lexLine, EQ)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, NE)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, LT)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, LE)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, GT)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())
	CheckToken(t, lexLine, GE)
	CheckNameTestToken(t, lexLine, xutils.AllChildren.Name())

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexDots(t *testing.T) {
	lexLine := NewExprLex(". .. ... .2 2.2 2.2.2", nil, nil)

	CheckToken(t, lexLine, int('.'))
	CheckToken(t, lexLine, DOTDOT)
	CheckToken(t, lexLine, DOTDOT)
	CheckToken(t, lexLine, int('.'))
	CheckNumToken(t, lexLine, 0.2)
	CheckNumToken(t, lexLine, 2.2)
	CheckUnlexableToken(t, lexLine, "bad number \"2.2.2\"")
}

func TestLexSlashes(t *testing.T) {
	lexLine := NewExprLex("/ // ///", nil, nil)

	CheckToken(t, lexLine, int('/'))
	CheckToken(t, lexLine, DBLSLASH)
	CheckToken(t, lexLine, DBLSLASH)
	CheckToken(t, lexLine, int('/'))

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexColons(t *testing.T) {
	lexLine := NewExprLex(":: :.", nil, nil)

	CheckToken(t, lexLine, DBLCOLON)
	CheckUnlexableToken(t, lexLine, "':' only supported in QNames")
}

// Need following '(' to be identified as functions.
func TestLexFunctions(t *testing.T) {
	lexLine := NewExprLex("round( true (", nil, nil)

	CheckFuncToken(t, lexLine, "round")
	CheckToken(t, lexLine, '(')
	CheckFuncToken(t, lexLine, "true")
	CheckToken(t, lexLine, '(')

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexUnrecognisedFunction(t *testing.T) {
	lexLine := NewExprLex("foo1(2)", nil, nil)

	CheckUnlexableToken(t, lexLine, "Unknown function or node type: 'foo1'")
}

// For full XPATH compliance we would allow prefixed function names.  However,
// for YANG we only support core functions plus current(), none of which are
// prefixed, so we allow it to be parsed as a NameTest for now.  May need
// revisiting.
func TestLexUnrecognisedPrefixedFunction(t *testing.T) {
	lexLine := NewExprLex("foo:bar1(2)", nil, nil)
	t.Skipf("Probably not needed ... see comment in file.")
	CheckUnlexableToken(t, lexLine,
		"Unknown function or node type: 'foo1'")
}

func TestLexSimpleLiterals(t *testing.T) {
	lexLine := NewExprLex("'ab c' \": 3\" ", nil, nil)

	CheckLiteralToken(t, lexLine, "ab c")
	CheckLiteralToken(t, lexLine, ": 3")

	CheckToken(t, lexLine, xutils.EOF)
}

// Check we ignore single within double and vice versa
func TestLexLiteralsInLiterals(t *testing.T) {
	lexLine := NewExprLex(" 'de\"fg\"12' \"'4$ :;p'q\" ", nil, nil)

	CheckLiteralToken(t, lexLine, "de\"fg\"12")
	CheckLiteralToken(t, lexLine, "'4$ :;p'q")

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexAdjacentLiterals(t *testing.T) {
	lexLine := NewExprLex("'1''2' '3'\"4\" \"5\"\"6\" \"7\"'8'", nil, nil)

	CheckLiteralToken(t, lexLine, "1")
	CheckLiteralToken(t, lexLine, "2")
	CheckLiteralToken(t, lexLine, "3")
	CheckLiteralToken(t, lexLine, "4")
	CheckLiteralToken(t, lexLine, "5")
	CheckLiteralToken(t, lexLine, "6")
	CheckLiteralToken(t, lexLine, "7")
	CheckLiteralToken(t, lexLine, "8")

	CheckToken(t, lexLine, xutils.EOF)
}

// Empty strings
func TestLexEmptyLiteral(t *testing.T) {
	lexLine := NewExprLex("'' \"\" \"\" ''", nil, nil)

	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")

	CheckToken(t, lexLine, xutils.EOF)
}

// You would not believe how fiddly the literal code is to get right.
func TestLexEmptyAdjacentLiterals(t *testing.T) {
	lexLine := NewExprLex("''\"\" \"\"'' '''' \"\"\"\"", nil, nil)

	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")
	CheckLiteralToken(t, lexLine, "")

	CheckToken(t, lexLine, xutils.EOF)

}

func TestLexIncompleteLiteralSingleQuotes(t *testing.T) {
	lexLine := NewExprLex("'Always plan ahea", nil, nil)

	CheckUnlexableToken(t, lexLine,
		"End of Literal token not detected.")
}

func TestLexIncompleteLiteralDoubleQuotes(t *testing.T) {
	lexLine := NewExprLex("\"Always plan ahea", nil, nil)

	CheckUnlexableToken(t, lexLine,
		"End of Literal token not detected.")
}

func TestAllValidNameCharacters(t *testing.T) {

	// The range D800 - DFFF are not valid unicode characters.
	// This range is used for "surrogate pairs" which allow one to
	// represent values outwith the BMP when using UTF-16.
	// i.e a valid pair of such characters represents a rune
	// beyond 0xFFFF
	//
	// Moreover, the high and low surrogate code points are illegal
	// in UTF-8 (that form being CESU-8), and so utf8.Encode Rune(),
	// and hence bytes.WriteRune() will encode such runes as a 3 byte
	// UTF-8 representation of the RuneError value.
	//
	// That would then decode upon a round trip as the
	// "Unicode replacement character"; and not signal an error.
	createString := func(t *testing.T, start, end rune) string {
		// Encode the surrogate test case manually.
		// The lexer will be corrected later to error upon
		// these invalid inputs.
		if start == end && end == 0xD800 {
			return badUTF8
		}
		var b bytes.Buffer
		for c := start; c <= end; c = c + 1 {
			_, err := b.WriteRune(c)
			if err != nil {
				t.Logf("Oops - can't write run 0x%x", c)
			}
		}
		return b.String()
	}

	CheckValidNameRange := func(
		t *testing.T,
		start, end rune,
		preToken, postToken int,
	) {
		t.Logf("Checking 0x%x to 0x%x\n", start, end)
		// First, Check all characters in range are accepted by creating
		// one (possibly quite big!) string with all of them in.
		s := createString(t, start, end)
		lexLine := NewExprLex(s, nil, nil)
		CheckNameTestToken(t, lexLine, xml.Name{Local: s})
		CheckToken(t, lexLine, xutils.EOF)

		// Next Check characters immediately before and after given range
		// are not accepted.  As we know NameStartChar in the XPATH spec
		// (or specs referenced from there) contains sets of ranges that
		// all have at least one character gap between them, we can do this.
		//
		// NB: pre / post character may be a DIFFERENT token type, rather than
		//     being rejected out of hand.  Hence need for pre/postToken types.
		pre := createString(t, start-1, start-1)
		lexLine = NewExprLex(pre, nil, nil)
		if preToken == xutils.ERR {
			t.Logf(" - Checking start(1) 0x%x", start-1)
			CheckUnlexableToken(t, lexLine, "unrecognised character")
		} else {
			t.Logf(" - Checking start(2) 0x%x", start-1)
			CheckToken(t, lexLine, preToken)
		}

		post := createString(t, end+1, end+1)
		lexLine = NewExprLex(post, nil, nil)
		if postToken == xutils.ERR {
			t.Logf(" - Checking end(1) 0x%x", end+1)
			CheckUnlexableToken(t, lexLine, "unrecognised character")
		} else {
			t.Logf(" - Checking end(2) 0x%x", end+1)
			CheckToken(t, lexLine, postToken)
		}
	}

	CheckValidNameRange(t, 'A', 'Z', int('@'), int('['))
	CheckValidNameRange(t, '_', '_', xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 'a', 'z', xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0xC0, 0xD6, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0xD8, 0xF6, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0xF8, 0x2FF, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0x370, 0x37D, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0x37F, 0x1FFF, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0x200C, 0x200D, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0x2070, 0x218F, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0x2C00, 0x2FEF, xutils.ERR, xutils.ERR)

	// NB: See "surrogate pair" comment in createString()
	CheckValidNameRange(t, 0x3001, 0xD7FF, xutils.ERR, xutils.EOF)

	CheckValidNameRange(t, 0xF900, 0xFDCF, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0xFDF0, 0xFFFD, xutils.ERR, xutils.ERR)
	CheckValidNameRange(t, 0x10000, 0xEFFFF, xutils.ERR, xutils.ERR)
}

func TestLexNodeType(t *testing.T) {
	lexLine := NewExprLex("comment( text( processing-instruction( node (",
		nil, nil)

	CheckNodeTypeToken(t, lexLine, "comment")
	CheckToken(t, lexLine, int('('))
	CheckNodeTypeToken(t, lexLine, "text")
	CheckToken(t, lexLine, int('('))
	CheckNodeTypeToken(t, lexLine, "processing-instruction")
	CheckToken(t, lexLine, int('('))
	CheckNodeTypeToken(t, lexLine, "node")
	CheckToken(t, lexLine, int('('))

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexAxisName(t *testing.T) {
	lexLine := NewExprLex(
		"ancestor-or-self :: attribute:: child:: descendant :: "+
			"descendant-or-self:: "+
			"following:: following-sibling:: namespace:: parent:: "+
			"preceding:: preceding-sibling:: self::", nil, nil)

	CheckAxisNameToken(t, lexLine, "ancestor-or-self")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "attribute")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "child")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "descendant")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "descendant-or-self")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "following")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "following-sibling")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "namespace")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "parent")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "preceding")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "preceding-sibling")
	CheckToken(t, lexLine, DBLCOLON)
	CheckAxisNameToken(t, lexLine, "self")
	CheckToken(t, lexLine, DBLCOLON)

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexIllegalAxisName(t *testing.T) {
	lexLine := NewExprLex("unknown-axis::", nil, nil)

	CheckUnlexableToken(t, lexLine,
		"Unknown axis name: 'unknown-axis'")
}

func TestLexNameTestWildcard(t *testing.T) {
	lexLine := NewExprLex("*", nil, nil)

	CheckNameTestToken(t, lexLine, xml.Name{Local: "*"})

	CheckToken(t, lexLine, xutils.EOF)
}

// For our testing here we just reuse the prefix as the namespace
func lexerTestMapFn(prefix string) (namespace string, err error) {
	return prefix, nil
}

func TestLexNameTestPrefixedWildcard(t *testing.T) {
	lexLine := NewExprLex("NCName:*", nil, lexerTestMapFn)
	CheckNameTestToken(t, lexLine, xml.Name{Space: "NCName", Local: "*"})

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexNameTestPrefixed(t *testing.T) {
	lexLine := NewExprLex("Pfx:Local", nil, lexerTestMapFn)
	CheckNameTestToken(t, lexLine, xml.Name{Space: "Pfx", Local: "Local"})

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexNameTestUnprefixed(t *testing.T) {
	lexLine := NewExprLex("UnpfxName", nil, nil)
	CheckNameTestToken(t, lexLine, xml.Name{Local: "UnpfxName"})

	CheckToken(t, lexLine, xutils.EOF)
}

func TestLexNameTestUnterminated(t *testing.T) {
	lexLine := NewExprLex("Pfx: ", nil, nil)

	CheckUnlexableToken(t, lexLine, "Name requires local part")
}

// Rules for disambiguating:
//
// (a) If there is a preceding token, and said token is none of '@', '::',
//     '(', '[', ',' or an Operator, then '*' is the MultiplyOperator and
//     NCName must be recognised as an OperatorName
//
// (b) If the character following an NCName (possibly after intervening
//     whitespace) is '(', then the token must be recognized as a NodeType
//     or FunctionName
//
// (c) If an NCName is followed by '::' (possibly with intervening whitespace)
//     then the NCName must be recognised as an AxisName
//
// (d) In all other cases, the token must NOT be recognised as a Multiply
//     Operator, OperatorName, NodeType, FunctionName, or AxisName
//
func TestLexDisambiguation(t *testing.T) {
	t.Skipf("Implemented elsewhere, at least in part.")
	// Look at the must / must not and verify each one if not done elsewhere
	// (ref other tests if so)

	// Specifically, the likes of 'div' and other operator names followed by
	// valid Name characters are interesting, as we are meant to parse greedily
	// yet we perhaps could be more intelligent?  Can we get a valid statement
	// where 'divxxx' in the middle can be parsed as 'div xxx' and never as
	// 'divxxx' such that we can do this?
}

// Test parsing multiple tokens in one string
func TestLexMultipleTokens(t *testing.T) {
	lexLine := NewExprLex("12 +66", nil, nil)

	CheckNumToken(t, lexLine, 12)
	CheckToken(t, lexLine, int('+'))
	CheckNumToken(t, lexLine, 66)

	CheckToken(t, lexLine, xutils.EOF)
}

// Test illegal tokens
func TestLexIllegalNumber(t *testing.T) {
	lexLine := NewExprLex("1eE6", nil, nil)

	CheckUnlexableToken(t, lexLine, "bad number \"1eE6\"")
}

func TestInvalidOperator(t *testing.T) {
	lexLine := NewExprLex("%", nil, nil)

	CheckUnlexableToken(t, lexLine, "unrecognised character '%'")
}

func TestInvalidOperator2(t *testing.T) {
	lexLine := NewExprLex("2 % 4", nil, nil)

	CheckNumToken(t, lexLine, 2)
	CheckUnlexableToken(t, lexLine, "unrecognised character '%'")
}
