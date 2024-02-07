// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017-2018 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

//go:generate goyacc -o path_eval.go -p "pathEval" path_eval.y

// "pathEval" in the above line is the 'prefix' that must match the
// 'pathEval' prefix on the pathEvalLex type below.

package path_eval

import (
	"encoding/xml"
	"fmt"

	"github.com/sdcio/yang-parser/xpath"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

// The parser uses the type <prefix>Lex as a lexer.  It must provide
// the methods Lex(*<prefix>SymType) int and Error(string).  The former
// is implemented in this file as it is grammar-specific, whereas the latter
// can be implemented generically on CommonLex instead.
type pathEvalLex struct {
	xpath.CommonLex
}

func NewPathEvalLex(
	expr string,
	progBldr *xpath.ProgBuilder,
	mapFn xpath.PfxMapFn,
) *pathEvalLex {
	return &pathEvalLex{
		xpath.NewCommonLex([]byte(expr), progBldr, mapFn)}
}

func (lexer *pathEvalLex) Parse() {
	pathEvalParse(lexer)
}

func getProgBldr(lexer pathEvalLexer) *xpath.ProgBuilder {
	return lexer.(*pathEvalLex).GetProgBldr()
}

// Wrapper around CommonLex to map to pathEvalSymType fields
func (x *pathEvalLex) Lex(yylval *pathEvalSymType) int {
	tok, val := xpath.LexCommon(x)

	switch v := val.(type) {
	case nil:
		/* No value */
	case float64:
		yylval.val = v
	case string:
		yylval.name = v
	case *xpath.Symbol:
		yylval.sym = v
	case xml.Name:
		yylval.xmlname = v
	default:
		tok = xutils.ERR
	}

	return mapCommonTokenValToPathEval(tok)
}

const EOF = 0

var commonToPathEvalTokenMap = map[int]int{
	xutils.EOF:      EOF,
	xutils.ERR:      ERR,
	xutils.NUM:      NUM,
	xutils.FUNC:     FUNC,
	xutils.DOTDOT:   DOTDOT,
	xutils.DBLSLASH: DBLSLASH,
	xutils.DBLCOLON: DBLCOLON,
	xutils.GT:       GT,
	xutils.GE:       GE,
	xutils.LT:       LT,
	xutils.LE:       LE,
	xutils.EQ:       EQ,
	xutils.NE:       NE,
	xutils.NODETYPE: NODETYPE,
	xutils.AXISNAME: AXISNAME,
	xutils.NAMETEST: NAMETEST,
	xutils.LITERAL:  LITERAL,
	xutils.OR:       OR,
	xutils.AND:      AND,
	xutils.MOD:      MOD,
	xutils.DIV:      DIV,
}

func mapCommonTokenValToPathEval(val int) int {
	if retval, ok := commonToPathEvalTokenMap[val]; ok {
		return retval
	}
	return val
}

var pathEvalToCommonTokenMap = map[int]int{
	EOF:      xutils.EOF,
	ERR:      xutils.ERR,
	NUM:      xutils.NUM,
	FUNC:     xutils.FUNC,
	DOTDOT:   xutils.DOTDOT,
	DBLSLASH: xutils.DBLSLASH,
	DBLCOLON: xutils.DBLCOLON,
	GT:       xutils.GT,
	GE:       xutils.GE,
	LT:       xutils.LT,
	LE:       xutils.LE,
	EQ:       xutils.EQ,
	NE:       xutils.NE,
	NODETYPE: xutils.NODETYPE,
	AXISNAME: xutils.AXISNAME,
	NAMETEST: xutils.NAMETEST,
	LITERAL:  xutils.LITERAL,
	OR:       xutils.OR,
	AND:      xutils.AND,
	MOD:      xutils.MOD,
	DIV:      xutils.DIV,
}

func (expr *pathEvalLex) MapTokenValToCommon(val int) int {
	if retval, ok := pathEvalToCommonTokenMap[val]; ok {
		return retval
	}
	return val
}

// Create a machine that can run the full XPATH grammar for 'when' and
// 'must' statements.  'pathEval' matches the name given to this full
// grammar (as in Expr / ExprToken in the XPATH RFC)
func NewPathEvalMachine(
	expr string,
	mapFn xpath.PfxMapFn,
	location string,
) (*xpath.Machine, error) {
	return newPathEvalMachineInternal(expr, mapFn, location, false, nil)
}

func NewPathEvalMachineWithCustomFns(
	expr string,
	mapFn xpath.PfxMapFn,
	location string,
	userFnChecker xpath.UserCustomFunctionCheckerFn,
) (*xpath.Machine, error) {
	return newPathEvalMachineInternal(
		expr, mapFn, location, true, userFnChecker)
}

func newPathEvalMachineInternal(
	expr string,
	mapFn xpath.PfxMapFn,
	location string,
	allowCustomFns bool,
	userFnChecker xpath.UserCustomFunctionCheckerFn,
) (*xpath.Machine, error) {

	if len(expr) == 0 {
		return nil, fmt.Errorf("Empty XPATH expression has no value.")
	}
	progBldr := xpath.NewProgBuilder(expr)
	lexer := NewPathEvalLex(expr, progBldr, mapFn)
	if allowCustomFns {
		lexer.AllowCustomFns()
	}
	lexer.SetUserFnChecker(userFnChecker)
	lexer.Parse()
	prog, err := lexer.CreateProgram(expr)
	if err != nil {
		return nil, err
	}
	return xpath.NewMachineWithLocation(
		expr, location, prog, "pathEvalMachine"), nil
}
