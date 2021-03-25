// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// lexertest - test functions for lexing

package lexertest

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

const NoErrMsg = ""

type tokenCheckFnType func(*testing.T, xpath.CommonSymType)

func checkTokenInternal(
	t *testing.T,
	lexer xpath.XpathLexer,
	tokenType int,
	expectPass bool,
	expErrMsg string,
	tokenCheckFn tokenCheckFnType,
) {
	var lexVal xpath.CommonSymType

	line := lexer.GetLine()
	lexType := xpath.LexCommon(lexer, &lexVal)
	tokenType = lexer.MapTokenValToCommon(tokenType)

	// Pass or fail, we expect a token type (EOF if fail / end).
	if tokenType != lexType {
		t.Logf("Parsing '%s'.\n", line)
		t.Fatalf("Wrong token type.  Exp '%s', got '%s'",
			xutils.GetTokenName(tokenType), xutils.GetTokenName(lexType))
		return
	}

	// Some token types have specific checks, so call those if present...
	if tokenCheckFn != nil {
		tokenCheckFn(t, lexVal)
	}

	// Check error status and expected error message all match up.
	err := lexer.GetError()
	if err != nil {
		if expectPass {
			t.Logf("Parsing '%s'.\n", line)
			t.Fatalf("Unexpected failure lexing: %s", err.Error())
			return
		}

		if expErrMsg == "" {
			t.Logf("Parsing '%s'.\n", line)
			t.Fatalf("Expected error message must be non-null!")
			return
		}

		if !strings.Contains(err.Error(), expErrMsg) {
			t.Logf("Parsing '%s'.\n", line)
			t.Fatalf("Wrong result for : exp '%s', got '%s'",
				expErrMsg, err.Error())
			return
		}
	}

	if err == nil && !expectPass {
		t.Logf("Parsing '%s'.\n", line)
		t.Fatalf("Unexpected success lexing.  Should have got: '%s'",
			expErrMsg)
		return
	}
}

// Check token types where there is no associated value - eg single
// character symbols.
func CheckToken(t *testing.T, lexer xpath.XpathLexer, tokenType int) {
	checkTokenInternal(t, lexer, tokenType, true, NoErrMsg, nil)
}

func CheckNumToken(t *testing.T, lexer xpath.XpathLexer, tokenVal float64) {
	checkNum := func(t *testing.T, lexVal xpath.CommonSymType) {
		if tokenVal != lexVal.GetVal() {
			t.Fatalf("Wrong token value.  Exp %v, got %v", tokenVal,
				lexVal.GetVal())
		}
	}
	checkTokenInternal(t, lexer, xutils.NUM, true, NoErrMsg, checkNum)
}

func CheckFuncToken(t *testing.T, lexer xpath.XpathLexer, funcName string) {
	checkFunc := func(t *testing.T, lexVal xpath.CommonSymType) {
		if funcName != lexVal.GetSym().GetName() {
			t.Fatalf("Wrong function name.  Exp '%s', got '%s'", funcName,
				lexVal.GetSym().GetName())
		}
	}
	checkTokenInternal(t, lexer, xutils.FUNC, true, NoErrMsg, checkFunc)
}

func CheckStringToken(
	t *testing.T,
	lexer xpath.XpathLexer,
	tokenType int,
	name string,
) {
	checkString := func(t *testing.T, lexVal xpath.CommonSymType) {
		if name != lexVal.GetName() {
			t.Fatalf("Wrong %s name.  Exp '%s', got '%s'",
				xutils.GetTokenName(tokenType), name, lexVal.GetName())
		}
	}
	checkTokenInternal(t, lexer, tokenType, true, NoErrMsg, checkString)
}

func CheckLiteralToken(t *testing.T, lexer xpath.XpathLexer, literal string) {
	CheckStringToken(t, lexer, xutils.LITERAL, literal)
}

func CheckNodeTypeToken(t *testing.T, lexer xpath.XpathLexer, nodeType string) {
	CheckStringToken(t, lexer, xutils.NODETYPE, nodeType)
}

func CheckAxisNameToken(t *testing.T, lexer xpath.XpathLexer, axisName string) {
	CheckStringToken(t, lexer, xutils.AXISNAME, axisName)
}

func CheckNameTestToken(
	t *testing.T,
	lexer xpath.XpathLexer,
	xmlname xml.Name,
) {
	checkNameTest := func(t *testing.T, lexVal xpath.CommonSymType) {

		if xmlname.Space != lexVal.GetXmlName().Space {
			t.Fatalf("Wrong NameTest namespace.  Exp '%s', got '%s'",
				xmlname.Space, lexVal.GetXmlName().Space)
		}
		if xmlname.Local != lexVal.GetXmlName().Local {
			t.Fatalf("Wrong NameTest Local.  Exp '%s', got '%s'",
				xmlname.Local, lexVal.GetXmlName().Local)
		}
	}
	checkTokenInternal(t, lexer, xutils.NAMETEST, true,
		NoErrMsg, checkNameTest)
}

func CheckUnlexableToken(
	t *testing.T,
	lexer xpath.XpathLexer,
	expErrMsg string) {
	checkTokenInternal(t, lexer, xutils.EOF, false, expErrMsg, nil)
}
