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

// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package leafref

import (
	"encoding/xml"
	"testing"

	. "github.com/sdcio/yang-parser/xpath/grammars/lexertest"
)

// Test basic characters
func TestLeafrefLexSlash(t *testing.T) {
	lexLine := NewLeafrefLex("/", nil, nil)

	CheckToken(t, lexLine, '/')
	CheckToken(t, lexLine, EOF)
}

func TestLeafrefLexDotDot(t *testing.T) {
	lexLine := NewLeafrefLex("..", nil, nil)

	CheckToken(t, lexLine, DOTDOT)
	CheckToken(t, lexLine, EOF)
}

func TestLeafrefLexEquals(t *testing.T) {
	lexLine := NewLeafrefLex("=", nil, nil)

	CheckToken(t, lexLine, EQ)
	CheckToken(t, lexLine, EOF)
}

func TestLeafrefLexSquareBrackets(t *testing.T) {
	lexLine := NewLeafrefLex("[]", nil, nil)

	CheckToken(t, lexLine, '[')
	CheckToken(t, lexLine, ']')
	CheckToken(t, lexLine, EOF)
}

// Test functions (well, only the one ...)
func TestLeafrefLexCurrent(t *testing.T) {
	lexLine := NewLeafrefLex("current()", nil, nil)

	CheckFuncToken(t, lexLine, "current")
	CheckToken(t, lexLine, '(')
	CheckToken(t, lexLine, ')')
	CheckToken(t, lexLine, EOF)
}

// Check other function name is not allowed, even when in the symbol table
// that we share across the grammars.
func TestLeafrefLexIllegalFunction(t *testing.T) {
	lexLine := NewLeafrefLex("true()", nil, nil)

	CheckUnlexableToken(t, lexLine,
		"Function 'true' is not valid here.")
}

// For our testing here we just reuse the prefix as the namespace
func lexLeafrefTestMapFn(prefix string) (namespace string, err error) {
	return prefix, nil
}

// Check we ignore whitespace.  Technically we are only meant to ignore this
// within the predicate, at least according to section 12 of the YANG spec,
// but elsewhere it suggests whitespace can be generally ignored, and as this
// relaxes the strictness of the checking, it shouldn't cause any harm.
func TestLeafRefLexIgnoreWhitespace(t *testing.T) {
	lexLine := NewLeafrefLex("[ foo:bar  = current() / .. / aaa / bbb ] ",
		nil, lexLeafrefTestMapFn)

	CheckToken(t, lexLine, '[')
	CheckNameTestToken(t, lexLine, xml.Name{Space: "foo", Local: "bar"})
	CheckToken(t, lexLine, EQ)
	CheckFuncToken(t, lexLine, "current")
	CheckToken(t, lexLine, '(')
	CheckToken(t, lexLine, ')')
	CheckToken(t, lexLine, '/')
	CheckToken(t, lexLine, DOTDOT)
	CheckToken(t, lexLine, '/')
	CheckNameTestToken(t, lexLine, xml.Name{Local: "aaa"})
	CheckToken(t, lexLine, '/')
	CheckNameTestToken(t, lexLine, xml.Name{Local: "bbb"})
	CheckToken(t, lexLine, ']')

	CheckToken(t, lexLine, EOF)
}

// Test path parsing
func TestLeafrefLexUnqualifiedName(t *testing.T) {
	lexLine := NewLeafrefLex("foo", nil, nil)

	CheckNameTestToken(t, lexLine, xml.Name{Local: "foo"})
	CheckToken(t, lexLine, EOF)
}

func TestLeafrefLexQualifiedName(t *testing.T) {
	lexLine := NewLeafrefLex("foo:bar", nil, lexLeafrefTestMapFn)

	CheckNameTestToken(t, lexLine, xml.Name{Space: "foo", Local: "bar"})
	CheckToken(t, lexLine, EOF)
}

func TestLeafrefLexNameCharacters(t *testing.T) {
	lexLine := NewLeafrefLex("_foo XXX _-. a-foo a.foo abc-def Z1.bar _XML XM",
		nil, nil)

	CheckNameTestToken(t, lexLine, xml.Name{Local: "_foo"})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "XXX"})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "_-."})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "a-foo"})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "a.foo"})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "abc-def"})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "Z1.bar"})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "_XML"})
	CheckNameTestToken(t, lexLine, xml.Name{Local: "XM"})
	CheckToken(t, lexLine, EOF)
}

func TestLeafrefLexIllegalNames(t *testing.T) {
	lexLine := NewLeafrefLex("XML", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Neither part of name may begin with XML: ':XML'")

	lexLine = NewLeafrefLex("xML", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Neither part of name may begin with XML: ':xML'")

	lexLine = NewLeafrefLex("xml", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Neither part of name may begin with XML: ':xml'")

	lexLine = NewLeafrefLex("foo:XML", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Neither part of name may begin with XML: 'foo:XML'")

	lexLine = NewLeafrefLex("bar:xML", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Neither part of name may begin with XML: 'bar:xML'")

	lexLine = NewLeafrefLex("bar:xml", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Neither part of name may begin with XML: 'bar:xml'")

	lexLine = NewLeafrefLex("foo:", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Name requires local part")

	lexLine = NewLeafrefLex("foo:1", nil, lexLeafrefTestMapFn)
	CheckUnlexableToken(t, lexLine,
		"Illegal local part start character: '1'")
}

// We (ab)use LexCommon and need to override handling of a few tokens.
func TestLeafrefLexTokensRejectedByParser(t *testing.T) {
	lexLine := NewLeafrefLex(".foo", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"'.' is not a valid token.")

	lexLine = NewLeafrefLex("1foo", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"Numbers are not valid tokens.")

	lexLine = NewLeafrefLex("-foo", nil, nil)
	CheckUnlexableToken(t, lexLine,
		"'-' is not a valid token.")

}
