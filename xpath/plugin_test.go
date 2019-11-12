// Copyright 2019, AT&T Intellectual Property. All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package xpath_test

import (
	"fmt"
	"plugin"
	"testing"

	"github.com/danos/yang/xpath"
)

// Expects bool arg.  Returns string 'PASS' if arg is true, 'FAIL' if not.
func workingFunction(
	args []xpath.Datum,
) xpath.Datum {
	bool0 := args[0].Boolean("workingPluginFn")
	if bool0 {
		return xpath.NewLiteralDatum("PASS")
	}
	return xpath.NewLiteralDatum("FAIL")
}

// Expects literal arg.  Panics regardless of input.
func panickingFunction(
	args []xpath.Datum,
) xpath.Datum {
	lit0 := args[0].Literal("panickingPluginFn")
	if lit0 == "keep calm and carry on" {
		return xpath.NewLiteralDatum("Stayed calm")
	}
	panic("Oops - someone didn't write their plugin very well ...")
}

var customFnInfoData = []xpath.CustomFunctionInfo{
	{
		Name:          "working-function",
		FnPtr:         workingFunction,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsBool},
		RetType:       xpath.TypeIsLiteral,
		DefaultRetVal: xpath.NewLiteralDatum(""),
	},
	{ // Invalid name
		Name:          "1st-char-num-function",
		FnPtr:         nil,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsLiteral},
		RetType:       xpath.TypeIsLiteral,
		DefaultRetVal: xpath.NewLiteralDatum("invalid name"),
	},
	{
		Name:          "panicking-function",
		FnPtr:         panickingFunction,
		Args:          []xpath.DatumTypeChecker{xpath.TypeIsLiteral},
		RetType:       xpath.TypeIsLiteral,
		DefaultRetVal: xpath.NewLiteralDatum("DON'T PANIC!"),
	},
}

type testXpathPlugin struct {
	customFnInfoTbl []xpath.CustomFunctionInfo
	name            string
}

func (txp *testXpathPlugin) Name() string { return txp.name }
func (txp *testXpathPlugin) Lookup(name string) (plugin.Symbol, error) {
	if name == "RegistrationData" {
		return &txp.customFnInfoTbl, nil
	}
	return nil, fmt.Errorf("Unhandled lookup: %s\n", name)
}

var testXPlugins = []xpath.XpathPlugin{
	&testXpathPlugin{
		customFnInfoTbl: customFnInfoData,
		name:            "TestPlugin"}}

// TESTS

func verifyXpathFunctionAbsent(t *testing.T, name string) {
	if _, ok := xpath.LookupXpathFunction(name, false, nil); ok {
		t.Fatalf("Found '%s' in standard symbol table", name)
	}
}

func verifyCustomFunctionPresent(t *testing.T, name string) {
	if _, ok := xpath.LookupXpathFunction(name, true, nil); !ok {
		t.Fatalf("Unable to find '%s' in custom symbol table", name)
	}
}

func verifyCustomFunctionAbsent(t *testing.T, name string) {
	if _, ok := xpath.LookupXpathFunction(name, true, nil); ok {
		t.Fatalf("Found '%s' in custom symbol table", name)
	}
}

func TestPluginRegistration(t *testing.T) {
	xpath.RegisterCustomFunctions(xpath.GetCustomFunctionInfo(testXPlugins))

	// Check symbol table content
	verifyCustomFunctionPresent(t, "working-function")
	verifyXpathFunctionAbsent(t, "working-function")
	verifyCustomFunctionAbsent(t, "non-existent-function")
	verifyCustomFunctionAbsent(t, "1st-char-num-function")
}

func TestCallWorkingPlugin(t *testing.T) {
	xpath.RegisterCustomFunctions(xpath.GetCustomFunctionInfo(testXPlugins))

	sym, ok := xpath.LookupXpathFunction("working-function", true, nil)
	if !ok {
		t.Fatalf("Unable to find workingPlugin function")
	}

	fn := sym.CustomFunc()
	res := fn([]xpath.Datum{xpath.NewBoolDatum(true)})
	if ok, _ := xpath.TypeIsLiteral(res); !ok {
		t.Fatalf("Wrong type returned by plugin fn")
	}
	if res.Literal("only-used-for-debug") != "PASS" {
		t.Fatalf("Plugin returned wrong result:\n\tExp 'PASS', got '%s'.",
			res.Literal(""))
	}

	res = fn([]xpath.Datum{xpath.NewBoolDatum(false)})
	if ok, _ = xpath.TypeIsLiteral(res); !ok {
		t.Fatalf("Wrong type returned by plugin fn")
	}
	if res.Literal("only-used-for-debug") != "FAIL" {
		t.Fatalf("Plugin returned wrong result:\n\tExp 'FAIL', got '%s'.",
			res.Literal(""))
	}

}

func TestCallPanicPlugin(t *testing.T) {
	xpath.RegisterCustomFunctions(xpath.GetCustomFunctionInfo(testXPlugins))

	sym, ok := xpath.LookupXpathFunction("panicking-function", true, nil)
	if !ok {
		t.Fatalf("Unable to find panickingPlugin function")
	}

	// First check we don't panic if we pass in the magic string ...
	fn := sym.CustomFunc()
	res := fn([]xpath.Datum{
		xpath.NewLiteralDatum("keep calm and carry on")})
	if ok, _ := xpath.TypeIsLiteral(res); !ok {
		t.Fatalf("Wrong type returned by plugin fn")
	}
	if res.Literal("only-used-for-debug") != "Stayed calm" {
		t.Fatalf(
			"Plugin returned wrong result:\n\tExp 'Stayed calm', got '%s'.",
			res.Literal(""))
	}

	// Now check we panic in other cases ... we should get empty string back.
	res = fn([]xpath.Datum{xpath.NewLiteralDatum("panic")})
	if ok, _ = xpath.TypeIsLiteral(res); !ok {
		t.Fatalf("Wrong type returned by plugin fn")
	}
	if res.Literal("only-used-for-debug") != "DON'T PANIC!" {
		t.Fatalf(
			"Plugin returned wrong result:\n\tExp 'DON'T PANIC!', got '%s'.",
			res.Literal(""))
	}
}
