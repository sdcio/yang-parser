// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// SPDX-License-Identifier: MPL-2.0 and BSD-3-Clause
package parse

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// item represents a token or text string returned from the scanner.
type item struct {
	typ itemType // The type of this item.
	pos Pos      // The starting position, in bytes, of this item in the input string.
	val string   // The value of this item.
}

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case i.typ > itemKeyword:
		return fmt.Sprintf("[%s] <%s>", i.typ, i.val)
	case len(i.val) > 80:
		return fmt.Sprintf("[%s] %.80q...", i.typ, i.val)
	}
	return fmt.Sprintf("[%s] %q", i.typ, i.val)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError      itemType = iota // error occurred; value is text of error
	itemEOF                        // EOF
	itemLeftBrace                  // left brace
	itemRightBrace                 // right brace
	itemSep                        // run of separating characters
	itemString                     // quoted string (includes quotes)
	itemSemiColon
	itemKeyword // used only to delimit the keywords
	itemPlus
	itemQuote
)

var types = [...]string{
	itemError:      "Error",
	itemEOF:        "EOF",
	itemLeftBrace:  "LBrace",
	itemRightBrace: "RBrace",
	itemSep:        "Separator",
	itemString:     "String",
	itemSemiColon:  "SemiColon",
	itemKeyword:    "Keyword",
	itemPlus:       "Plus",
	itemQuote:      "Quote",
}

func (i itemType) String() string {
	return types[i]
}

const eof = -1

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name         string    // the name of the input; used only for error reports
	input        string    // the string being scanned
	state        stateFn   // the next lexing function to enter
	pos          Pos       // current position in the input
	start        Pos       // start position of this item
	width        Pos       // width of last rune read from input
	lastPos      Pos       // position of most recent item returned by nextItem
	items        chan item // channel of scanned items
	bracketDepth int       // nesting depth of ( ) exprs
}

// next returns the next rune in the input.
func (l *lexer) next() rune {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.next()
	l.backup()
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.pos -= l.width
}

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

// ignore skips over the pending input before this point.
func (l *lexer) ignore() {
	l.start = l.pos
}

// accept consumes the next rune if it's from the valid set.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

// acceptRun consumes a run of runes from the valid set.
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

// lineNumber reports which line we're on, based on the position of
// the previous item returned by nextItem. Doing it this way
// means we don't have to worry about peek double counting.
func (l *lexer) lineNumber() int {
	return 1 + strings.Count(l.input[:l.lastPos], "\n")
}

// errorf returns an error token and terminates the scan by passing
// back a nil pointer that will be the next state, terminating l.nextItem.
func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexStmt; l.state != nil; {
		l.state = l.state(l)
	}
}

// state functions

const (
	leftComment  = "/*"
	rightComment = "*/"
	lineComment  = "//"
)

// lexComment scans a comment. The left comment marker is known to be present.
func lexComment(l *lexer) stateFn {
	l.pos += Pos(len(leftComment))
	i := strings.Index(l.input[l.pos:], rightComment)
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i + len(rightComment))
	l.ignore()
	return lexStmt
}

// lexCommentLine scans a comment. The left comment marker is known to be present.
func lexCommentLine(l *lexer) stateFn {
	l.pos += Pos(len(leftComment))
	i := strings.Index(l.input[l.pos:], "\n")
	if i < 0 {
		return l.errorf("unclosed comment")
	}
	l.pos += Pos(i + 1)
	l.ignore()
	return lexStmt
}

// lexStmt scans a statement
func lexStmt(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], leftComment) {
			return lexComment
		}
		if strings.HasPrefix(l.input[l.pos:], lineComment) {
			return lexCommentLine
		}

		switch r := l.next(); {
		case r == eof:
			l.backup()
			break
		case isSep(r):
			return lexSep
		case r == '"' || r == '\'':
			return lexQuote
		case r == '{':
			l.emit(itemLeftBrace)
			l.bracketDepth++
			return lexStmt
		case r == '}':
			l.emit(itemRightBrace)
			l.bracketDepth--
			if l.bracketDepth < 0 {
				return l.errorf("unexpected right bracket %#U", r)
			}
			return lexStmt
		case r == ';':
			l.emit(itemSemiColon)
			return lexStmt
		case r == '+':
			l.emit(itemPlus)
			return lexStmt
		default:
			l.backup()
			return lexString
		}

		if l.next() == eof {
			break
		}

		l.backup()
	}
	if l.bracketDepth > 0 {
		return l.errorf("unterminated statement block")
	}
	l.emit(itemEOF)
	return nil
}

// lexString scans a run of non-separator characters
func lexString(l *lexer) stateFn {
	for !isTerminator(l.peek()) {
		l.next()
	}
	l.emit(itemString)
	return lexStmt
}

// lexSpace scans a run of space characters.
// One space has already been seen.
func lexSep(l *lexer) stateFn {
	for isSep(l.peek()) {
		l.next()
	}
	l.emit(itemSep)
	return lexStmt
}

/*lexQuote scans a quoted string.
 *
 * Emits a Quote - String - Quote sequence
 *
 * Quote is one of (') or ("), the second quote emitted will match first.
 * String is everything between the opening and closing quotes.
 * A double quoted string supports escape sequences, where any character
 * following a backslash (\) is skipped, allowing a double quote (") to
 * appear in a double quoted string
 */
func lexQuote(l *lexer) stateFn {
	var qt rune = rune(l.input[l.start])
	l.emit(itemQuote)
Loop:
	for {
		switch l.next() {
		case '\\':
			if qt == '\'' {
				// no special characters for single
				// quoted strings
				break
			} else if r := l.next(); r != eof {
				break
			}
			fallthrough
		case eof:
			return l.errorf("unterminated quoted string")
		case qt:
			l.backup()
			break Loop
		}
	}
	l.emit(itemString)
	l.next()
	l.emit(itemQuote)
	return lexStmt
}

func isTerminator(r rune) bool {
	return isSep(r) || r == ';' || r == '{' || r == '"' || r == '}'
}

func isSep(r rune) bool {
	return isSpace(r) || isEndOfLine(r)
}

// isSpace reports whether r is a space character.
func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

// isEndOfLine reports whether r is an end-of-line character.
func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

// isAlphaNumeric reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
