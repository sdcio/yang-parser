// Copyright (c) 2018-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file holds the go generate command to run yacc on the grammar in
// xpath.y. !!!  DO NOT REMOVE THE NEXT TWO LINES !!!

//go:generate goyacc -o xpath.go -p "expr" xpath.y

// "expr" in the above line is the 'prefix' that must match the 'expr' prefix
// on the exprLex type below.

package expr

import (
	"fmt"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

// The parser uses the type <prefix>Lex as a lexer.  It must provide
// the methods Lex(*<prefix>SymType) int and Error(string).  The former
// is implemented in this file as it is grammar-specific, whereas the latter
// can be implemented generically on CommonLex instead.
type exprLex struct {
	xpath.CommonLex
}

func NewExprLex(
	expr string,
	progBldr *xpath.ProgBuilder,
	mapFn xpath.PfxMapFn,
) *exprLex {
	return &exprLex{
		xpath.NewCommonLex([]byte(expr), progBldr, mapFn)}
}

func (lexer *exprLex) Parse() {
	exprParse(lexer)
}

func getProgBldr(lexer exprLexer) *xpath.ProgBuilder {
	return lexer.(*exprLex).GetProgBldr()
}

// Wrapper around CommonLex to map between exprSymType and the common
// lexParams.
func (x *exprLex) Lex(yylval *exprSymType) int {
	lexParams := x.GetLexParams()

	retval := xpath.LexCommon(x, lexParams)
	yylval.sym = lexParams.GetSym()
	yylval.val = lexParams.GetVal()
	yylval.name = lexParams.GetName()
	yylval.xmlname = lexParams.GetXmlName()
	return mapCommonTokenValToExpr(retval)
}

const EOF = 0

var commonToExprTokenMap = map[int]int{
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

func mapCommonTokenValToExpr(val int) int {
	if retval, ok := commonToExprTokenMap[val]; ok {
		return retval
	}
	return val
}

var exprToCommonTokenMap = map[int]int{
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

func (expr *exprLex) MapTokenValToCommon(val int) int {
	if retval, ok := exprToCommonTokenMap[val]; ok {
		return retval
	}
	return val
}

// Create a machine that can run the full XPATH grammar for 'when' and
// 'must' statements.  'expr' matches the name given to this full
// grammar (as in Expr / ExprToken in the XPATH RFC)
func NewExprMachine(
	expr string,
	mapFn xpath.PfxMapFn,
) (*xpath.Machine, error) {
	return newExprMachineInternal(expr, mapFn, false)
}

func NewExprMachineWithCustomFunctions(
	expr string,
	mapFn xpath.PfxMapFn,
) (*xpath.Machine, error) {
	return newExprMachineInternal(expr, mapFn, true)
}

func newExprMachineInternal(
	expr string,
	mapFn xpath.PfxMapFn,
	allowCustomFns bool,
) (*xpath.Machine, error) {

	if len(expr) == 0 {
		return nil, fmt.Errorf("Empty XPATH expression has no value.")
	}
	progBldr := xpath.NewProgBuilder(expr)
	lexer := NewExprLex(expr, progBldr, mapFn)
	if allowCustomFns {
		lexer.AllowCustomFns()
	}
	lexer.Parse()
	prog, err := lexer.CreateProgram(expr)
	if err != nil {
		return nil, err
	}
	return xpath.NewMachine(expr, prog, "exprMachine"), nil
}
