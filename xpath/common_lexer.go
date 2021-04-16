// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Credit for the 'next' function, and initial 'lex' function go to whoever
// wrote the 'expr' YACC example in the Go source code.

// This file implements XPATH lexing / tokenisation for YANG.  Specifically,
// it diverges from a complete XPATH implementation as follows:
//
// (a) As YANG uses only the core function set, we do not accept fully-
//     qualified function names (eg prefix:fn_name())
//
// Different YANG statements use different subsets of XPATH.  This file
// contains the common lexing code, with customisations separated out into
// specific _lexer.go files that live with their associated <prefix>.y YACC
// grammar files.

package xpath

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"unicode/utf8"

	"github.com/danos/yang/xpath/xutils"
)

// Allow for different grammars to be compiled and run using Machine,
// sharing common lexing code where possible.
type XpathLexer interface {
	GetError() error
	SetError(err error)
	GetProgBldr() *ProgBuilder
	GetLine() []byte
	GetLexParams() *CommonSymType
	Parse()

	Next() rune
	SaveTokenType(tokenType int) int
	IsNameChar(c rune) bool
	IsNameStartChar(c rune) bool
	MapTokenValToCommon(tokenType int) int

	LexLiteral(yylval *CommonSymType, quote rune) int
	LexDot(c rune, yylval *CommonSymType) int
	LexNum(c rune, yylval *CommonSymType) int
	LexSlash() int
	LexColon() int
	LexAsterisk(yylval *CommonSymType) int
	LexRelationalOperator(c rune) int
	LexName(c rune, yylval *CommonSymType) int
	LexPunctuation(c rune) int
}

// COMMONSYMTYPE
type CommonSymType struct {
	sym     *Symbol  /* Symbol table entry */
	val     float64  /* Numeric value */
	name    string   /* NodeType or AxisName */
	xmlname xml.Name /* For NameTest */
}

func (cst *CommonSymType) GetSym() *Symbol {
	return cst.sym
}

func (cst *CommonSymType) SetSym(sym *Symbol) {
	cst.sym = sym
}

func (cst *CommonSymType) GetVal() float64 {
	return cst.val
}

func (cst *CommonSymType) GetName() string {
	return cst.name
}

func (cst *CommonSymType) SetXmlName(name xml.Name) {
	cst.xmlname = name
}

func (cst *CommonSymType) GetXmlName() xml.Name {
	return cst.xmlname
}

// COMMONLEX
type CommonLex struct {
	// Exported via accessors.
	line      []byte
	err       error
	mapFn     PfxMapFn
	progBldr  *ProgBuilder // Used to build the program to be run later.
	lexParams CommonSymType

	// Internal use only
	peek           rune
	precToken      int  // Preceding token type, if any (otherwise EOF)
	allowCustomFns bool // Expr may use custom XPATH functions
	userFnChecker  UserCustomFunctionCheckerFn
}

func NewCommonLex(
	line []byte,
	progBldr *ProgBuilder,
	mapFn PfxMapFn,
) CommonLex {
	return CommonLex{line: line, progBldr: progBldr, mapFn: mapFn}
}

func (lexer *CommonLex) AllowCustomFns() *CommonLex {
	lexer.allowCustomFns = true
	return lexer
}

func (lexer *CommonLex) SetUserFnChecker(
	userFnChecker UserCustomFunctionCheckerFn,
) {
	lexer.userFnChecker = userFnChecker
}

func (lexer *CommonLex) Parse() {
	panic("CommonLex doesn't implement Parse()")
}

func (lexer *CommonLex) CreateProgram(expr string) (prog []Inst, err error) {
	if lexer.progBldr.parseErr == nil && lexer.GetError() == nil {
		return lexer.progBldr.GetMainProg()
	}

	errors := fmt.Sprintf("Failed to compile '%s'\n", expr)
	currentPosInLine :=
		len(string(expr)) - len(string(lexer.progBldr.lineAtErr))
	parsedLine := string(expr)[:currentPosInLine]
	unParsedLine := string(expr)[currentPosInLine:]

	if lexer.GetError() != nil {
		errors += fmt.Sprintf("Lexer Error: %s\n",
			lexer.GetError().Error())
	}

	if lexer.progBldr.parseErr != nil {
		errors += fmt.Sprintf("Parse Error: %s\n",
			lexer.progBldr.parseErr.Error())
	}

	return nil, fmt.Errorf("%s\nGot to approx [X] in '%s [X] %s'\n", errors,
		parsedLine, unParsedLine)
}

func (lexer *CommonLex) GetError() error { return lexer.err }

func (lexer *CommonLex) SetError(err error) { lexer.err = err }

func (lexer *CommonLex) GetProgBldr() *ProgBuilder {
	return lexer.progBldr
}

func (lexer *CommonLex) GetLexParams() *CommonSymType {
	return &lexer.lexParams
}

func (lexer *CommonLex) GetLine() []byte { return lexer.line }

func (lexer *CommonLex) GetMapFn() PfxMapFn { return lexer.mapFn }

// The parser calls this method on a parse error.  It stores the error in the
// machine for later retrieval.
func (x *CommonLex) Error(s string) {
	if x.progBldr.parseErr != nil {
		// Use first error found, if more than one detected.
		return
	}
	x.progBldr.parseErr = fmt.Errorf("%s", s)
	if x.peek != xutils.EOF {
		x.progBldr.lineAtErr = string(x.peek) + string(x.line)
	} else {
		x.progBldr.lineAtErr = string(x.line)
	}
}

// Some parsing will produce different tokens depending on what came before
// so we need to keep track of this.
func (x *CommonLex) SaveTokenType(tokenType int) int {
	x.precToken = tokenType
	return tokenType
}

// The parser calls this method to get each new token.
//
// We store the token value so it is available as the preceding token
// value when parsing the next token.
func LexCommon(x XpathLexer, yylval *CommonSymType) int {
	for {
		c := x.Next()
		switch c {
		case xutils.EOF:
			return xutils.EOF

		case xutils.ERR:
			x.SetError(fmt.Errorf("Invalid UTF-8 input"))
			return xutils.ERR

		case '"', '\'':
			return x.SaveTokenType(x.LexLiteral(yylval, c))

		case '.':
			return x.SaveTokenType(x.LexDot(c, yylval))

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
			return x.SaveTokenType(x.LexNum(c, yylval))

		case '/':
			return x.SaveTokenType(x.LexSlash())

		case ':':
			return x.SaveTokenType(x.LexColon())

		case '*':
			return x.SaveTokenType(x.LexAsterisk(yylval))

		case '+', '-', '(', ')', '@', ',', '[', ']', '|':
			return x.SaveTokenType(x.LexPunctuation(c))

		case '=', '>', '<', '!':
			return x.SaveTokenType(x.LexRelationalOperator(c))

		case ' ', '\t', '\n', '\r':
			// Deal with whitespace by ignoring it
			continue
		}

		// Names of some form or another ... NameTest, NodeType,
		// OperatorName, FunctionName, or AxisName
		if x.IsNameStartChar(c) {
			return x.SaveTokenType(x.LexName(c, yylval))
		}

		x.SetError(fmt.Errorf("unrecognised character %q", c))
		return xutils.ERR
	}
}

// Separated out to allow us to override it.
func (x *CommonLex) LexPunctuation(c rune) int {
	return int(c)
}

func (x *CommonLex) LexDot(c rune, yylval *CommonSymType) int {
	// Could be '.', '..', or number
	next := x.Next()
	switch next {
	case '.':
		return xutils.DOTDOT
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		x.peek = next
		return x.LexNum(c, yylval)
	default:
		x.peek = next
		return '.'
	}
}

func (x *CommonLex) LexSlash() int {
	// Could be '/' or '//'.  NB - this is not 'divide', ever.
	next := x.Next()
	if next == '/' {
		return xutils.DBLSLASH
	}
	x.peek = next
	return '/'
}

func (x *CommonLex) LexColon() int {
	// Should be '::' as single colons are only allowed within QNames and
	// are not detected in main lexer loop.
	next := x.Next()
	if next == ':' {
		return xutils.DBLCOLON
	}
	// Part of a name, should have been detected elsewhere.
	x.peek = next
	x.SetError(fmt.Errorf("':' only supported in QNames"))
	return xutils.ERR
}

func (x *CommonLex) LexAsterisk(yylval *CommonSymType) int {
	if x.tokenCanBeOperator() {
		return '*'
	}

	// This is the global wildcard representing all child nodes, regardless
	// of module.
	yylval.xmlname = xutils.AllChildren.Name()
	return xutils.NAMETEST
}

func (x *CommonLex) LexRelationalOperator(c rune) int {
	switch c {
	case '=':
		return xutils.EQ
	case '>':
		next := x.Next()
		if next == '=' {
			return xutils.GE
		}
		x.peek = next
		return xutils.GT

	case '<':
		next := x.Next()
		if next == '=' {
			return xutils.LE
		}
		x.peek = next
		return xutils.LT

	case '!':
		next := x.Next()
		if next == '=' {
			return xutils.NE
		}
		x.peek = next
		x.SetError(fmt.Errorf("'!' only valid when followed by '='"))
		return xutils.ERR
	default:
		x.SetError(fmt.Errorf("Invalid relational operator"))
		return xutils.ERR
	}
}

// Lex a non-literal name (ie something textual that isn't quoted).
//
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
func (x *CommonLex) LexName(c rune, yylval *CommonSymType) int {
	nameMatcher := func(c rune) bool {
		if x.IsNameChar(c) {
			return true
		}
		return false
	}

	// Next get 'NCName'
	name := x.ConstructToken(c, nameMatcher, "NAME")

	// If there's a preceding token, and it's not '@', '::', '(', '[', ',' or
	// an Operator then NCName is an OperatorName
	if x.tokenCanBeOperator() {
		return x.getOperatorName(name.String())
	}

	// If next non-whitespace character is '(' then this must be a NodeType
	// or a FunctionName
	if x.NextNonWhitespaceStringIs("(") {
		if x.nameIsNodeType(name.String()) {
			yylval.name = name.String()
			return xutils.NODETYPE
		}
		fn, ok := LookupXpathFunction(
			name.String(),
			x.allowCustomFns,
			x.userFnChecker)
		if ok {
			yylval.sym = fn
			return xutils.FUNC
		}
		x.SetError(fmt.Errorf("Unknown function or node type: '%s'",
			name.String()))
		return xutils.ERR
	}

	// If next non-whitespace token is '::', NCName is an AxisName.
	if x.NextNonWhitespaceStringIs("::") {
		if x.nameIsAxisName(name.String()) {
			yylval.name = name.String()
			return xutils.AXISNAME
		}
		x.SetError(fmt.Errorf("Unknown axis name: '%s'", name.String()))
		return xutils.ERR
	}

	// If none of the above applies, it's a NameTest token.  Question is
	// whether it's a Prefixed or Unprefixed ... so let's see if we have a
	// colon following.  As we already checked for '::', we can safely check
	// for single ':'
	var namespace, localPart, prefix string
	if x.NextNonWhitespaceStringIs(":") {
		// Prefixed, so it's either Prefix:LocalPart or Prefix:*
		// Next token had better be a ':' when formally extracted ...
		if c := x.NextNonWhitespace(); c != ':' {
			x.SetError(fmt.Errorf(
				"Badly formatted QName (exp ':', got '%c'", c))
			return xutils.ERR
		}

		// Now we need the local part - or wildcard (*).  Note that in the
		// latter case this must be 'NCName:*' - the global wildcard '*' is
		// handled by LexAsterisk().
		if x.NextNonWhitespaceStringIs("*") {
			// Next token had better be a '*' when formally extracted ...
			if c := x.NextNonWhitespace(); c != '*' {
				x.SetError(fmt.Errorf("Badly formatted QName (*)."))
				return xutils.ERR
			}
			prefix = name.String()
			localPart = "*"
		} else {
			// We need to extract the second part of the name.
			c := x.NextNonWhitespace()
			if c == xutils.EOF {
				x.err = fmt.Errorf("Name requires local part.")
				return xutils.ERR
			}
			localPartBuf := x.ConstructToken(c, nameMatcher, "NAME")
			localPart = localPartBuf.String()
			prefix = name.String()
		}

	} else {
		localPart = name.String()
	}

	// If we have a mapping function, map the locally-scoped (within namespace)
	// prefix name (if present) to a globally scoped namespace.
	if x.mapFn != nil {
		var err error
		namespace, err = x.mapFn(prefix)
		if err != nil {
			x.SetError(err)
			return xutils.ERR
		}
	}

	yylval.xmlname = xml.Name{Space: namespace, Local: localPart}

	return xutils.NAMETEST
}

// Lex 'literal' string contained in single or double quotes
func (x *CommonLex) LexLiteral(yylval *CommonSymType, quote rune) int {
	literalMatcher := func(c rune) bool {
		if c != quote {
			return true
		}
		return false
	}

	// Skip initial quote - start from 'next'.  As constructToken always
	// adds first character, we also need to detect empty strings here.
	var b bytes.Buffer
	c := x.Next()
	if c != quote {
		b = x.ConstructToken(c, literalMatcher,
			xutils.GetTokenName(xutils.LITERAL))
		// Skip final quote character.
		x.Next()
	}

	yylval.name = b.String()
	if x.err != nil {
		return xutils.ERR
	}
	return xutils.LITERAL
}

// Lex a number.
func (x *CommonLex) LexNum(c rune, yylval *CommonSymType) int {
	numMatcher := func(c rune) bool {
		switch c {
		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.', 'e', 'E':
			return true
		}
		return false
	}
	b := x.ConstructToken(c, numMatcher, xutils.GetTokenName(xutils.NUM))
	val, err := strconv.ParseFloat(b.String(), 10)

	yylval.val = val
	if err != nil {
		x.SetError(fmt.Errorf("bad number %q", b.String()))
		return xutils.ERR
	}
	return xutils.NUM
}

// An operator cannot follow a specific set of other tokens, which include
// other operators (quite reasonably).  See XPATH section 3.7.
func (x *CommonLex) tokenCanBeOperator() bool {
	// Split into 3 cases to avoid wrapping and to match the order in the
	// XPATH spec (ambiguity rules for ExprToken)
	switch x.precToken {
	case xutils.EOF, '@', xutils.DBLCOLON, '(', '[', ',':
		return false

	case xutils.AND, xutils.OR, xutils.MOD, xutils.DIV:
		return false

	case '*', '/', xutils.DBLSLASH, '|', '+', '-',
		xutils.EQ, xutils.NE, xutils.LT, xutils.LE, xutils.GT, xutils.GE:
		return false
	}

	return true
}

// Useful for any multi-character token in conjunction with constructToken()
type tokenMatcherFn func(c rune) bool

// Given first character in token and function to identify further elements,
// return full token and set x.peek to the correct character.
func (x *CommonLex) ConstructToken(
	c rune,
	tokenMatcher tokenMatcherFn,
	tokenName string,
) bytes.Buffer {

	add := func(b *bytes.Buffer, c rune) {
		if _, err := b.WriteRune(c); err != nil {
			x.SetError(fmt.Errorf("WriteRune: %s", err))
		}
	}
	var b bytes.Buffer
	add(&b, c)

	for {
		c = x.Next()
		if tokenMatcher(c) {
			// As a sanity check against rogue tokenMatcher functions that fail
			// to spot EOF and claim a match, trap it here.  It's also rather
			// easier to spot here in the guts of the processing anyway.
			if c == xutils.EOF {
				x.SetError(fmt.Errorf("End of %s token not detected.",
					tokenName))
				break
			}
			add(&b, c)
		} else {
			break
		}
	}

	x.peek = c

	return b
}

func (x *CommonLex) IsNameStartChar(c rune) bool {
	switch {
	case (c >= 'A') && (c <= 'Z'):
		return true
	case c == '_':
		return true
	case (c >= 'a') && (c <= 'z'):
		return true
	case (c >= 0xC0) && (c <= 0xD6):
		return true
	case (c >= 0xD8) && (c <= 0xF6):
		return true
	case (c >= 0xF8) && (c <= 0x2FF):
		return true
	case (c >= 0x370) && (c <= 0x37D):
		return true
	case (c >= 0x37F) && (c <= 0x1FFF):
		return true
	case (c >= 0x200C) && (c <= 0x200D):
		return true
	case (c >= 0x2070) && (c <= 0x218F):
		return true
	case (c >= 0x2C00) && (c <= 0x2FEF):
		return true
	case (c >= 0x3001) && (c <= 0xD7FF):
		return true
	case (c >= 0xF900) && (c <= 0xFDCF):
		return true
	case (c >= 0xFDF0) && (c <= 0xFFFD):
		return true
	case (c >= 0x10000) && (c <= 0xEFFFF):
		return true
	default:
		return false
	}
}

func (x *CommonLex) IsNameChar(c rune) bool {
	switch {
	case x.IsNameStartChar(c):
		return true
	case c == '-' || c == '.':
		return true
	case (c >= '0') && (c <= '9'):
		return true
	case c == 0xB7:
		return true
	case (c >= 0x300) && (c <= 0x36F):
		return true
	case (c >= 0x203F) && (c <= 0x2040):
		return true
	default:
		return false
	}
}

func (x *CommonLex) getOperatorName(name string) int {
	switch name {
	case "and":
		return xutils.AND
	case "or":
		return xutils.OR
	case "mod":
		return xutils.MOD
	case "div":
		return xutils.DIV
	}

	x.SetError(fmt.Errorf("Unrecognised operator name: '%s'", name))
	return xutils.ERR
}

func (x *CommonLex) nameIsNodeType(name string) bool {
	switch name {
	case "comment", "text", "processing-instruction", "node":
		return true
	}

	return false
}

func (x *CommonLex) nameIsAxisName(name string) bool {
	switch name {
	case "ancestor-or-self", "attribute", "child", "descendant",
		"descendant-or-self", "following", "following-sibling",
		"namespace", "parent", "preceding", "preceding-sibling", "self":
		return true
	}

	return false
}

// Return the next rune for the lexer.  'peek' may have been set if we
// needed to look ahead but then didn't consume the character.  In other
// words, what remains to be parsed when we call Next() is:
//
//   x.peek (if not EOF) + x.line
//
func (x *CommonLex) Next() rune {
	if x.peek != xutils.EOF {
		r := x.peek
		x.peek = xutils.EOF
		return r
	}
	if len(x.line) == 0 {
		return xutils.EOF
	}
	c, size := utf8.DecodeRune(x.line)
	x.line = x.line[size:]
	if c == utf8.RuneError && size == 1 {
		return xutils.ERR
	}
	return c
}

func (x *CommonLex) isWhitespace(c rune) bool {
	switch c {
	case '\t', '\r', '\n', ' ':
		return true
	}

	return false
}

func (x *CommonLex) NextNonWhitespace() rune {
	c := x.Next()

	for c != xutils.EOF && x.isWhitespace(c) {
		c = x.Next()
	}

	return c
}

func next(line []byte) (rune, []byte) {
	if len(line) == 0 {
		return xutils.EOF, nil
	}
	c, size := utf8.DecodeRune(line)
	line = line[size:]
	if c == utf8.RuneError && size == 1 {
		return xutils.ERR, nil
	}
	return c, line
}

// Won't handle string containing whitespace.
// For now we only need this to match '(', '::', ':' and '*'.
// This assumes the passed in string consists of ASCII bytes
func (x *CommonLex) NextNonWhitespaceStringIs(expr string) bool {

	// First check peek (if in use) and if not whitespace, compare.
	// The ASCII assumption is here, it could be written using
	// utf8.RuneLen(), but that is not necessary.
	if (x.peek != xutils.EOF) && !x.isWhitespace(x.peek) {
		if len(expr) == 0 {
			return true
		}
		if x.peek != rune(expr[0]) {
			return false
		}
		if len(expr) == 1 {
			return true
		}
		expr = expr[1:]
	}

	// Next, skip any whitespace
	lc, line := next(x.line)
	for x.isWhitespace(lc) {
		lc, line = next(line)
	}

	// Now compare the rest of the string against the input
	for _, ec := range expr {
		if lc == xutils.EOF || lc == xutils.ERR {
			return false
		}
		if ec != lc {
			return false
		}
		lc, line = next(line)
	}

	return true
}
