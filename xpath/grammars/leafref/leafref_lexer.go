// Copyright (c) 2018-2021,  AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file holds the go generate command to run yacc on the grammar in
// xpath.y. !!!  DO NOT REMOVE THE NEXT TWO LINES !!!

//go:generate goyacc -o leafref.go -p "leafref" leafref.y

// "leafref" in the above line is the 'prefix' that must match the 'leafref'
// prefix on the leafrefLex type below.

package leafref

import (
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

// The parser uses the type <prefix>Lex as a lexer.  It must provide
// the methods Lex(*<prefix>SymType) int and Error(string).  The former
// is implemented in this file as it is grammar-specific, whereas the latter
// can be implemented generically on commonLex instead.
type leafrefLex struct {
	xpath.CommonLex
}

func NewLeafrefLex(
	leafref string,
	progBldr *xpath.ProgBuilder,
	mapFn xpath.PfxMapFn,
) *leafrefLex {
	return &leafrefLex{xpath.NewCommonLex([]byte(leafref), progBldr, mapFn)}
}

func (lexer *leafrefLex) Parse() {
	leafrefParse(lexer)
	if lexer.GetLexParams().GetLastToken() != xutils.EOF {
		lexer.Error("syntax error")
	}
}

func getProgBldr(lexer leafrefLexer) *xpath.ProgBuilder {
	return lexer.(*leafrefLex).GetProgBldr()
}

// Wrapper around CommonLex to map between leafrefSymType and the common
// lexParams.
func (x *leafrefLex) Lex(yylval *leafrefSymType) int {
	lexParams := x.GetLexParams()

	retval := xpath.LexCommon(x, lexParams)
	yylval.sym = lexParams.GetSym()
	yylval.val = lexParams.GetVal()
	yylval.xmlname = lexParams.GetXmlName()
	lexParams.SetLastToken(retval)
	return mapCommonTokenValToLeafref(retval)
}

const EOF = 0

var commonToLeafrefTokenMap = map[int]int{
	xutils.EOF:      EOF,
	xutils.EQ:       EQ,
	xutils.FUNC:     FUNC,
	xutils.DOTDOT:   DOTDOT,
	xutils.NAMETEST: NAMETEST,
}

func mapCommonTokenValToLeafref(val int) int {
	if retval, ok := commonToLeafrefTokenMap[val]; ok {
		return retval
	}
	return val
}

var leafrefToCommonTokenMap = map[int]int{
	EOF:      xutils.EOF,
	EQ:       xutils.EQ,
	FUNC:     xutils.FUNC,
	DOTDOT:   xutils.DOTDOT,
	NAMETEST: xutils.NAMETEST,
}

func (leafref *leafrefLex) MapTokenValToCommon(val int) int {
	if retval, ok := leafrefToCommonTokenMap[val]; ok {
		return retval
	}
	return val
}

func (x *leafrefLex) IsNameStartChar(c rune) bool {
	switch {
	case (c >= 'A') && (c <= 'Z'):
		return true
	case c == '_':
		return true
	case (c >= 'a') && (c <= 'z'):
		return true
	default:
		return false
	}
}

func (x *leafrefLex) IsNameChar(c rune) bool {
	switch {
	case x.IsNameStartChar(c):
		return true
	case c == '-' || c == '.':
		return true
	case (c >= '0') && (c <= '9'):
		return true
	default:
		return false
	}

}

func startsWithXML(name string) bool {
	if len(name) < 3 {
		return false
	}

	if strings.ToLower(name)[0:3] == "xml" {
		return true
	}

	return false
}

func (x *leafrefLex) LexPunctuation(c rune) int {
	switch c {
	case '[', ']', '(', ')':
		return int(c)
	default:
		x.SetError(fmt.Errorf("'%c' is not a valid token.", c))
		return xutils.EOF
	}
}

// '..' is only valid token that starts with a dot.
func (x *leafrefLex) LexDot(c rune, yylval *xpath.CommonSymType) int {
	next := x.Next()
	switch next {
	case '.':
		return xutils.DOTDOT
	default:
		x.SetError(fmt.Errorf("'.' is not a valid token."))
		return EOF
	}
}

func (x *leafrefLex) LexNum(c rune, yylval *xpath.CommonSymType) int {
	x.SetError(fmt.Errorf("Numbers are not valid tokens."))
	return EOF
}

func (x *leafrefLex) LexName(c rune, yylval *xpath.CommonSymType) int {
	nameMatcher := func(c rune) bool {
		if x.IsNameChar(c) {
			return true
		}
		return false
	}

	// Next get 'NCName'
	name := x.ConstructToken(c, nameMatcher, "NAME")

	// If next non-whitespace character is '(' then this must be a function
	if x.NextNonWhitespaceStringIs("(") {
		if name.String() != "current" {
			x.SetError(fmt.Errorf("Function '%s' is not valid here.",
				name.String()))
			return xutils.EOF
		}
		fn, ok := xpath.LookupXpathFunction(name.String(),
			false, /* no custom functions allowed here */
			nil /* no user-provided checker fn */)
		if ok {
			yylval.SetSym(fn)
			return xutils.FUNC
		}
		x.SetError(fmt.Errorf("Unable to resolve 'current' function."))
		return xutils.EOF
	}

	// OK, it's a NameTest token.  Question is whether it's a Prefixed or
	// Unprefixed ... so let's see if we have a colon following.
	// for single ':'
	var namespace, localPart, prefix string
	if x.NextNonWhitespaceStringIs(":") {
		// Prefixed ...
		// Next token had better be a ':' when formally extracted ...
		if c := x.NextNonWhitespace(); c != ':' {
			x.SetError(fmt.Errorf(
				"Badly formatted QName (exp ':', got '%c'", c))
			return xutils.EOF
		}

		// Now we need the local part.  No wildcards here
		c := x.NextNonWhitespace()
		if c == xutils.EOF {
			x.SetError(fmt.Errorf("Name requires local part."))
			return xutils.EOF
		}
		if !x.IsNameStartChar(c) {
			x.SetError(fmt.Errorf(
				"Illegal local part start character: '%c'", c))
			return xutils.EOF
		}
		localPartBuf := x.ConstructToken(c, nameMatcher, "NAME")
		localPart = localPartBuf.String()
		prefix = name.String()
	} else {
		localPart = name.String()
	}

	// Need to check neither prefix nor localPart begin with XML (case-
	// insensitive)
	if startsWithXML(prefix) || startsWithXML(localPart) {
		x.SetError(fmt.Errorf(
			"Neither part of name may begin with XML: '%s:%s'",
			prefix, localPart))
		return xutils.EOF
	}

	// If we have a mapping function, map the locally-scoped (within namespace)
	// prefix name (if present) to a globally scoped namespace.  Otherwise
	// we ignore the prefix.
	if x.GetMapFn() != nil {
		var err error
		namespace, err = x.GetMapFn()(prefix)
		if err != nil {
			x.SetError(err)
			return xutils.EOF
		}
	}

	yylval.SetXmlName(xml.Name{Space: namespace, Local: localPart})

	return xutils.NAMETEST
}

// Create a machine that can run the full XLEAFREF grammar for 'when' and
// 'must' statements.  'leafref' matches the name given to this full
// grammar (as in Leafref / LeafrefToken in the XLEAFREF RFC)
func NewLeafrefMachine(
	leafref string,
	mapFn xpath.PfxMapFn,
) (*xpath.Machine, error) {

	if len(leafref) == 0 {
		return nil, fmt.Errorf("Empty XPATH expression has no value.")
	}
	progBldr := xpath.NewProgBuilder(leafref)
	lexer := NewLeafrefLex(leafref, progBldr, mapFn)
	lexer.Parse()
	prog, err := lexer.CreateProgram(leafref)
	if err != nil {
		return nil, err
	}
	return xpath.NewMachine(leafref, prog, "leafrefMachine"), nil
}
