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
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file provides a common superset of tokens so that common_lexer.go
// can use these definitions and each individual grammar can map the tokens
// it uses onto these via mapping functions.

package xutils

import (
	"fmt"
)

// The parser expects the lexer to return 0 on EOF.  Give it a name
// for clarity.  Also define 'special' tokens so they won't clash with
// other token values (ASCII value in most cases).  Just to ensure mapping
// is working correctly, YACC starts off at 0xE000 or thereabouts, so this
// enum is deliberately starting at a different place to catch bugs where
// we've forgotten to map and some tokens are ok, some not.
const (
	EOF = 0
	ERR = 0xF000 + iota
	NUM
	FUNC
	DOTDOT
	DBLSLASH
	DBLCOLON
	GT
	GE
	LT
	LE
	EQ
	NE
	NODETYPE
	AXISNAME
	NAMETEST
	LITERAL
	OR
	AND
	MOD
	DIV
	TEXTFUNC
)

// For testing and error pretty-printing, useful to be able to dump a string
// for token values.

// Need to map within each grammar so you lookup xpath:FUNC and get
//<grammar>:FUNC back to return ... or they get mapped back at 'base' in
// <grammar>.  Which is best?  Fn call here or map elsewhere?

var tokenNameMap = map[int]string{
	EOF:      "EOF",
	ERR:      "LexError",
	NUM:      "NUM",
	FUNC:     "FUNC",
	DOTDOT:   "..",
	DBLSLASH: "//",
	DBLCOLON: "::",
	GT:       ">",
	GE:       ">=",
	LT:       "<",
	LE:       "<=",
	EQ:       "=",
	NE:       "!=",
	NODETYPE: "NodeType",
	AXISNAME: "AxisName",
	NAMETEST: "NameTest",
	LITERAL:  "Literal",
	OR:       "or",
	AND:      "and",
	MOD:      "mod",
	DIV:      "div",
	TEXTFUNC: "text",
}

func GetTokenName(token int) string {
	if name, ok := tokenNameMap[token]; ok {
		return name
	}

	if token >= 32 && token <= 255 {
		return fmt.Sprintf("%c (%x)", token, token)
	} else {
		return fmt.Sprintf("0x%x", token)
	}
}
