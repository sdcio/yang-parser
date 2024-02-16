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

// Copyright (c) 2019-2020, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// Portions Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: MPL-2.0 and BSD-3-Clause

package parse

import (
	"fmt"
	"runtime"
	"strings"
)

type Scope struct {
	tenv *TEnv
	genv *GEnv
}

// RFC 6020; Sec 6.1.3, specifies a tab is 8 space characters
const tabSpaces int = 8
const wsSpaces int = 1

var BuiltinTenv *TEnv

func init() {
	BuiltinTenv = NewTEnv(nil)
	BuiltinTenv.Put("binary", nil)
	BuiltinTenv.Put("bits", nil)
	BuiltinTenv.Put("boolean", nil)
	BuiltinTenv.Put("decimal64", nil)
	BuiltinTenv.Put("empty", nil)
	BuiltinTenv.Put("enumeration", nil)
	BuiltinTenv.Put("identityref", nil)
	BuiltinTenv.Put("instance-identifier", nil)
	BuiltinTenv.Put("int8", nil)
	BuiltinTenv.Put("int16", nil)
	BuiltinTenv.Put("int32", nil)
	BuiltinTenv.Put("int64", nil)
	BuiltinTenv.Put("leafref", nil)
	BuiltinTenv.Put("string", nil)
	BuiltinTenv.Put("uint8", nil)
	BuiltinTenv.Put("uint16", nil)
	BuiltinTenv.Put("uint32", nil)
	BuiltinTenv.Put("uint64", nil)
	BuiltinTenv.Put("union", nil)
}

func OpenScope(p *Scope) *Scope {
	if p == nil {
		return &Scope{
			tenv: NewTEnv(BuiltinTenv),
			genv: NewGEnv(nil),
		}
	}
	return &Scope{
		tenv: NewTEnv(p.tenv),
		genv: NewGEnv(p.genv),
	}
}

// Tree is the representation of a single parsed template.
type Tree struct {
	Root      Node // top-level root of the tree.
	ParseName string
	extCard   NodeCardinality // Function to provide cardinality of extensions
	text      string          // text parsed to create the template (or its parent)
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int

	argInterner    *ArgInterner
	stringInterner *StringInterner
}

func Parse(name, text string, extCard NodeCardinality) (*Tree, error) {
	return ParseWithInterners(name, text, extCard, NewStringInterner(), NewArgInterner())
}

func ParseWithInterners(
	name, text string,
	extCard NodeCardinality,
	stringInterner *StringInterner,
	argInterner *ArgInterner,
) (*Tree, error) {
	t := NewWithInterners(name, extCard, stringInterner, argInterner)
	t.text = text
	defer t.done()
	_, err := t.Parse(text)
	return t, err
}

func (t *Tree) done() {
	var empty [3]item
	copy(t.token[:], empty[:])

	t.extCard = nil
	t.argInterner = nil
	t.stringInterner = nil
	t.lex = nil
}

func (t *Tree) String() string {
	return t.text
}

// next returns the next token.
func (t *Tree) next() item {
	if t.peekCount > 0 {
		t.peekCount--
	} else {
		t.token[0] = t.lex.nextItem()
	}
	return t.token[t.peekCount]
}

// backup backs the input stream up one token.
func (t *Tree) backup() {
	t.peekCount++
}

// backup2 backs the input stream up two tokens.
// The zeroth token is already there.
func (t *Tree) backup2(t1 item) {
	t.token[1] = t1
	t.peekCount = 2
}

// backup3 backs the input stream up three tokens
// The zeroth token is already there.
func (t *Tree) backup3(t2, t1 item) { // Reverse order: we're pushing back.
	t.token[1] = t1
	t.token[2] = t2
	t.peekCount = 3
}

// peek returns but does not consume the next token.
func (t *Tree) peek() item {
	if t.peekCount > 0 {
		return t.token[t.peekCount-1]
	}
	t.peekCount = 1
	t.token[0] = t.lex.nextItem()
	return t.token[0]
}

// nextNonSpace returns the next non-space token.
func (t *Tree) nextNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSep {
			break
		}
	}
	return token
}

// peekNonSpace returns but does not consume the next non-space token.
func (t *Tree) peekNonSpace() (token item) {
	for {
		token = t.next()
		if token.typ != itemSep {
			break
		}
	}
	t.backup()
	return token
}

// Parsing.
// New allocates a new parse tree with the given name.
func New(name string, card NodeCardinality) *Tree {
	return NewWithInterners(name, card, NewStringInterner(), NewArgInterner())
}

func NewWithInterners(
	name string,
	card NodeCardinality,
	stringInterner *StringInterner,
	argInterner *ArgInterner,
) *Tree {

	if card == nil {
		card = func(n NodeType) map[NodeType]Cardinality { return nil }
	}
	return &Tree{
		ParseName:      name,
		extCard:        card,
		argInterner:    argInterner,
		stringInterner: stringInterner,
	}
}

func (t *Tree) ErrorContextPosition(pos int, ctx string) (location, context string) {
	text := t.text[:pos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	lineNum := 1 + strings.Count(text, "\n")
	context = ctx
	if len(context) > 20 {
		context = fmt.Sprintf("%.20s...", context)
	}
	if ctx == "" {
		return fmt.Sprintf("%s:%d:%d", t.ParseName, lineNum, byteNum), context
	}
	return fmt.Sprintf("%s:%d:%d: %s", t.ParseName, lineNum, byteNum, ctx), context
}

// errorf formats the error and terminates processing.
func (t *Tree) errorf(format string, args ...interface{}) {
	t.Root = nil
	pos := int(t.lex.lastPos)
	text := t.lex.input[:t.lex.lastPos]
	byteNum := strings.LastIndex(text, "\n")
	if byteNum == -1 {
		byteNum = pos // On first line.
	} else {
		byteNum++ // After the newline.
		byteNum = pos - byteNum
	}
	format = fmt.Sprintf("yang: %s:%d:%d: %s", t.ParseName, t.lex.lineNumber(), byteNum, format)
	panic(fmt.Errorf(format, args...))
}

// error terminates processing.
func (t *Tree) error(err error) {
	t.errorf("%s", err)
}

// expect consumes the next token and guarantees it has the required type.
func (t *Tree) expect(expected itemType, context string) item {
	token := t.nextNonSpace()
	if token.typ != expected {
		t.unexpected(token, context)
	}
	return token
}

// expectOneOf consumes the next token and guarantees it has one of the required types.
func (t *Tree) expectOneOf(expected1, expected2 itemType, context string) item {
	token := t.nextNonSpace()
	if token.typ != expected1 && token.typ != expected2 {
		t.unexpected(token, context)
	}
	return token
}

// unexpected complains about the token and terminates processing.
func (t *Tree) unexpected(token item, context string) {
	t.errorf("unexpected %s in %s", token, context)
}

// recover is the handler that turns panics into returns from the top level of Parse.
func (t *Tree) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		if t != nil {
			t.stopParse()
		}
		*errp = e.(error)
	}
	return
}

// startParse initializes the parser, using the lexer.
func (t *Tree) startParse(lex *lexer) {
	t.Root = nil
	t.lex = lex
}

// stopParse terminates parsing.
func (t *Tree) stopParse() {
	t.lex = nil
}

func (t *Tree) Parse(text string) (tree *Tree, err error) {
	defer t.recover(&err)
	t.startParse(lexWithInterner(t.ParseName, text, t.stringInterner))
	t.text = text
	t.parse()
	t.stopParse()
	return t, nil
}

func (tree *Tree) NewNode(id item, arg string, children []Node, s *Scope) Node {
	ntype := NodeTypeFromName(id.val, arg)
	return newNodeByType(ntype, tree, id, arg, children, s, tree.argInterner)
}

//file:
//	stmt stmt*
func (t *Tree) parse() {
	s := OpenScope(nil)
	t.Root = t.stmt("file", s)
	t.expect(itemEOF, "file")

	// (agj) TODO: Should we check it's module or submodule?

	//Fill out symbol tables top down, this enables us to check for shadowing
	//which is disallowed by yang.
	err, pos := t.Root.buildSymbols()
	if err != nil {
		s, _ := t.ErrorContextPosition(int(pos), "")
		panic(fmt.Errorf("%s: %s", s, err))
	}
	return
}

//stmt:
//	identifier argument stmtBody
//|	identifier '{' stmtStar '}' //special case for in and out
func (t *Tree) stmt(ctx string, s *Scope) Node {
	var arg string
	id := t.expect(itemString, ctx)
	i := t.peekNonSpace()
	switch i.typ {
	case itemLeftBrace:
		break
	default:
		arg = t.argument("argument of " + id.val)
	}
	//Link scopes as we walk the tree, we will fill
	//out the symbol tables in a separate pass top down
	ns := OpenScope(s)
	body := t.stmtBody(id.val+" "+arg, ns)
	n := t.NewNode(id, arg, body, s)

	//Validate cardinality, ordering, and arguemnt syntax
	e := n.check()
	if e != nil {
		s, _ := n.ErrorContext()
		panic(fmt.Errorf("%s: %s", s, e))
	}

	return n
}

// These are the only four valid escape sequences permitted in a
// double quoted string, refer to RFC 6020; sec 6.1.3 and
// RFC 6536, Erratta
var sMap = map[string]string{
	"n":  "\n",
	"r":  "\r",
	"t":  "\t",
	"\"": "\"",
	"\\": "\\",
}

func escapeSequenceSubstitution(s string, tree *Tree) string {
	var rs string
	var skip bool

	if s == "" {
		return s
	}

	/*
	 * Break up in to list of string, with backslash as seperator
	 * We now have a list of strings, with an implied backslash preceeding
	 * all but the first in the list.
	 *
	 * A nil string ("") implies a double backslash was seen,
	 * except first in list, which is a solitary backslash.
	 *
	 * Traverse list, looking up first character in each string, and
	 * replace with EscSeq substitute, if valid, while also accounting for
	 * double backslash sequences
	 */
	lines := strings.Split(s, `\`)
	for i, st := range lines {
		if st == "" {
			// Double Backslash was seen, inject backslash
			// unless first in list or double nil string
			if skip == false && i > 0 {
				rs += "\\"
				skip = true
			} else {
				skip = false
			}
			continue
		}

		// Only do EscSeq sub if not first in list and no double
		// backslash preceeds this string
		if i > 0 && skip == false {
			sub, found := sMap[st[:1]]
			if !found {
				// TODO: may Need to make illegal for Yang 1.1
				//
				// ignore any backslash sequences that
				// are not explicitly substituted by Yang 1.0,
				// restore the backslash character.
				rs += sub + "\\" + st
			} else {
				rs += sub + st[1:]
			}
		} else {
			// first in list or preceeded by double backslash
			// leave untouched
			rs += st
			skip = false
		}
	}

	return rs
}

// Get the count of WS that needs to be trimmed from lines to ensure they
// line up with the first character after the strings opening double quote
func openQuotePos(t *Tree, s string) (quotePos int) {

	// Get position of the quote
	posStart := strings.LastIndex(t.lex.input[:t.lex.lastPos], s)

	// Get position one after previous line-break
	lnBgn := strings.LastIndex(t.lex.input[:posStart], "\n") + 1

	// Get all the line up the the opening quote
	leadUp := t.lex.input[lnBgn:posStart]

	for _, c := range leadUp {
		if c == '\t' {
			// TODO: (pac) handle as tab-stop?
			quotePos += tabSpaces
		} else {
			// all other characters are 1 space
			quotePos += wsSpaces
		}
	}

	return quotePos
}

// Strip up to trimLen WS from begining of string
// If removing a tab would result in excess WS being trimmed,
// substitute with WS as required
func trimLeadWS(s string, trimLen int) string {
	var tabOfWS string = "        "
	var wsCount int

	for i, c := range s {
		switch c {
		case ' ':
			wsCount += wsSpaces
		case '\t':
			wsCount += tabSpaces
		default:
			// Reached a non-WS
			return s[i:]
		}

		if wsCount >= trimLen {
			//return string; if a tab would cause excess trimming
			// replace sufficient WS to re-align with open quote
			return tabOfWS[:wsCount-trimLen] + s[i+1:]
		}
	}

	// Must be all WS, return empty string
	return ""
}

// Trim all WS as required by RFC 6020; Sec 6.1.3
// - all WS at the start of each line up to the column of the
// strings opening double quote in the yang file
// - all WS at the end of each line, immediately before the a
// line-break, it one exists
func trimWhitespace(t *Tree, s string) string {
	var trimmed string
	var cr int

	// Two line-break variants, LF and CRLF
	lineBreaks := [2]string{"\n", "\r\n"}

	// Perform any special character substitution before trimming
	sub := escapeSequenceSubstitution(s, t)

	// only trim whitespace if a line-break is present
	if !strings.Contains(sub, "\n") {
		return sub
	}

	// Get quote position, using pre-substitution string
	quotePos := openQuotePos(t, s)

	// Break up into lines seperated LF
	lines := strings.Split(sub, "\n")
	for i, st := range lines {
		str := st

		if i > 0 {
			// No WS trimming on first line
			str = trimLeadWS(str, quotePos)
		}

		if len(str) == 0 {
			continue
		}

		// Handle a CRLF line-break
		if rune(str[len(str)-1]) == '\r' {
			cr = 1
		}

		if i != len(lines)-1 {
			// trim trailing WS for all but last string
			str = strings.TrimRight(str[:len(str)-cr], " \t") + lineBreaks[cr]
		}

		trimmed += str
		cr = 0
	}
	return trimmed
}

//argument:
// (string / (quotedString *([sep] '+' [sep] quotedString))) (';' / '{')
func (t *Tree) argument(ctx string) string {
	var i item
	var s string

	i = t.peekNonSpace()
	switch i.typ {
	case itemLeftBrace:
		fallthrough
	case itemSemiColon:
		return s
	case itemString:
		i = t.nextNonSpace()
		s = i.val
	case itemQuote:
		i = t.nextNonSpace()
		s = t.argumentQuoted(ctx)
	default:
		t.unexpected(i, ctx)
	}

	return s
}

//argumentQuoted:
// Quoted string; the leading quote has been removed
// strip whitespace of a double quoted string
// that contains a line break - RFC 6020; Sec 6.1.3
func (t *Tree) argumentQuoted(ctx string) string {
	var i item
	var s string

	i = t.peekNonSpace()
	switch i.typ {
	case itemString:
		i = t.nextNonSpace()
		// Quoted string must be terminated by a quote
		qt := t.expect(itemQuote, ctx)
		if qt.val == "\"" {
			s = trimWhitespace(t, i.val) + t.argumentConcatenate(ctx)
		} else {
			s = i.val + t.argumentConcatenate(ctx)
		}
	case itemQuote:
		i = t.nextNonSpace()
		s = t.argumentConcatenate(ctx)
	default:
		t.unexpected(i, ctx)
	}
	return s
}

//argumentConcatenate:
// Check if we need to concatenate another string, indicated by a '+'
func (t *Tree) argumentConcatenate(ctx string) string {
	var i item
	var s string

	i = t.peekNonSpace()
	switch i.typ {
	case itemLeftBrace:
		fallthrough
	case itemSemiColon:
		return s
	case itemPlus:
		i = t.nextNonSpace()
		// must be followed by [sep] quote
		t.expect(itemQuote, ctx)
		s = t.argumentQuoted(ctx)
	default:
		t.unexpected(i, ctx)
	}
	return s
}

//stmtBody:
//	';'
//| '{' stmtStar '}'
func (t *Tree) stmtBody(ctx string, s *Scope) []Node {
	var out []Node
	delim := t.expectOneOf(itemSemiColon, itemLeftBrace, ctx)
	switch delim.typ {
	case itemLeftBrace:
		out = t.stmtStar(ctx, s)
		t.expect(itemRightBrace, ctx)
	case itemSemiColon:
	default:
	}

	return out
}

//stmtStar
//	stmt*
func (t *Tree) stmtStar(ctx string, s *Scope) []Node {
	//0 stmts
	if i := t.peekNonSpace(); i.typ == itemRightBrace {
		return nil
	}

	//1 or more stmts
	out := make([]Node, 0)
	for n := t.stmt(ctx, s); n != nil; n = t.stmt(ctx, s) {
		out = append(out, n)
		if i := t.peekNonSpace(); i.typ == itemRightBrace {
			break
		}
	}
	if len(out) == 0 {
		return nil
	} else {
		children := make([]Node, len(out))
		copy(children, out)
		return children
	}
}
