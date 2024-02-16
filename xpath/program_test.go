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

// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This suite of tests differs from the parser_test suite.  The latter
// checks expressions are parsed and evaluated correctly, and that parsing
// errors are caught.  This set of tests check that the internals of the
// machine construction and execution work correctly.  There is overlap,
// but the focus is different, and concentrates as much on error handling
// as normal operation.

package xpath

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sdcio/yang-parser/xpath/xutils"
)

// Helper functions
func newTestContext() *context {
	return newCtx(nil, nil, nil, 1, 1, 0, "expr", "(no location)").
		EnableValidation()
}

func verifyNoCompileErrors(t *testing.T, progBldr *ProgBuilder) {
	if progBldr.parseErr != nil {
		t.Fatalf("Machine has compile error: %s", progBldr.parseErr.Error())
	}
}

func verifyNoRuntimeErrors(t *testing.T, ctx *context) {
	if ctx.res.runErr != nil {
		t.Fatalf("Machine has runtime error: %s", ctx.res.runErr.Error())
	}
}

func verifyNumResult(t *testing.T, ctx *context, expVal float64) {
	if actVal, _ := ctx.res.GetNumResult(); actVal != expVal {
		t.Fatalf("Wrong value stored.  Exp %v, got %v", expVal, actVal)
		return
	}
}

func verifyRuntimeError(t *testing.T, actErr, expErrMsg string) {
	if actErr == "" {
		t.Fatalf("Didn't get expected runtime errors: %s", expErrMsg)
		return
	}
	if !strings.Contains(actErr, expErrMsg) {
		t.Fatalf("Wrong error. \nGot: %s\nExp: %s",
			actErr, expErrMsg)
		return
	}
}

func getSymbol(t *testing.T, name string) *Symbol {
	sym, ok := LookupXpathFunction(
		name,
		false, /* Don't allow custom functions */
		nil /* user-provided custom function checker */)
	if !ok {
		t.Fatalf("Cannot find symbol to test.")
		return nil
	}
	if sym.name != name {
		t.Fatalf("Wrong symbol returned.")
		return nil
	}

	return sym
}

func instructionsMatch(exp, act Inst) (bool, error) {
	if exp.fnName != act.fnName {
		return false, fmt.Errorf("Function name: exp '%s', got '%s'",
			exp.fnName, act.fnName)
	}
	// Don't check function pointers.  Name is enough, and now that we
	// generate closures for the likes of numpush, it's essentially
	// impossible to check the pointer.  Instead, the parameter encoded
	// in the closure should be checked via the function name which
	// should have the parameter(s) contained in it.

	return true, nil
}

func verifyProgramInstructions(
	t *testing.T,
	progBldr *ProgBuilder,
	instrs []Inst,
) {
	prog, err := progBldr.GetMainProg()
	if err != nil {
		t.Fatalf("Cannot verify program: %s", err.Error())
		return
	}
	if len(instrs) != len(prog) {
		t.Fatalf("Machine instrn length %d doesn't match expected length %d",
			len(prog), len(instrs))
		return
	}

	for index, instr := range instrs {
		if ok, err := instructionsMatch(instr, prog[index]); !ok {
			t.Fatalf("Machine doesn't match expected instruction set\n\n%s",
				err.Error())
			return
		}
	}
}

func dataMatch(exp, act Datum) (bool, error) {
	if !exp.isSameType(act) {
		return false, fmt.Errorf("Mismatched types:\nExp: %v\nGot: %v\n",
			exp, act)
	}

	switch {
	case isNum(exp):
		expNum := exp.Number("dataMatch exp number")
		actNum := act.Number("dataMatch act number")
		if expNum != actNum {
			return false, fmt.Errorf("Number: exp '%v', got '%v'",
				expNum, actNum)
		}
	case isLiteral(exp):
		expLit := exp.Literal("dataMatch exp literal")
		actLit := act.Literal("dataMatch act literal")
		if expLit != actLit {
			return false, fmt.Errorf("Literal: exp '%s', got '%s'",
				expLit, actLit)
		}
	default:
		return false, fmt.Errorf("Unknown data type")
	}

	return true, nil
}

func verifycontextStack(t *testing.T, ctx *context, data []Datum) {
	if len(data) != len(ctx.stack) {
		t.Errorf("Machine stack length %d doesn't match expected length %d",
			len(ctx.stack), len(data))
		return
	}

	for index, item := range data {
		if ok, err := dataMatch(item, ctx.stack[index].(Datum)); !ok {
			t.Errorf("Machine doesn't match expected data set\n\n%s",
				err.Error())
			return
		}
	}
}

// Test PathType strings are generated correctly.
func TestPathTypeString(t *testing.T) {
	absPath := xutils.PathType([]string{"/", "interface", "dataplane"})
	relPath := xutils.PathType([]string{"..", "interface", "dataplane"})

	absStr := "/interface/dataplane"
	relStr := "../interface/dataplane"

	if absStr != absPath.String() {
		t.Fatalf("PathString: exp '%s', got '%s'", absStr, absPath.String())
	}

	if relStr != relPath.String() {
		t.Fatalf("PathString: exp '%s', got '%s'", relStr, relPath.String())
	}
}

// Test program construction functions. (code Fn/Sym/Num/Bltin)
func TestCodeFn(t *testing.T) {
	testProgBldr := NewProgBuilder("dummy expression")

	testProgBldr.CodeFn(testProgBldr.Negate, "negate")

	expectedInstructions := []Inst{
		newInst(testProgBldr.Negate, "negate"),
	}
	verifyProgramInstructions(t, testProgBldr, expectedInstructions)
	verifyNoCompileErrors(t, testProgBldr)
}

func TestCodeNum(t *testing.T) {
	testProgBldr := NewProgBuilder("dummy expression")

	testProgBldr.CodeNum(66.6)

	expectedInstructions := []Inst{
		newInst(nil, fmt.Sprintf("numpush\t\t66.6")),
	}
	verifyProgramInstructions(t, testProgBldr, expectedInstructions)
	verifyNoCompileErrors(t, testProgBldr)
}

func TestCodeBltin(t *testing.T) {
	testProgBldr := NewProgBuilder("dummy expression")

	sym := getSymbol(t, "number")
	testProgBldr.CodeBltin(sym, 1)

	expectedInstructions := []Inst{
		newInst(nil, fmt.Sprintf("bltin\t\tnumber()")),
	}
	verifyProgramInstructions(t, testProgBldr, expectedInstructions)
	verifyNoCompileErrors(t, testProgBldr)
}

// Test operation of stack operations: push / pop

// Check popped value matched pushed one, and that stack length is correctly
// modified.
func popValueAndVerify(t *testing.T, ctx *context, pushed Datum) {
	initStackLen := len(ctx.stack)
	popped := ctx.popDatum()
	if len(ctx.stack) != initStackLen-1 {
		t.Fatalf("Incorrect stack length.")
		return
	}
	if err := popped.equalTo(pushed); err != nil {
		t.Fatalf("Popped value (%v) doesn't match pushed one (%v).\n%s",
			popped, pushed, err.Error())
		return
	}
}

func TestPushThenPop(t *testing.T) {
	testCtx := newTestContext()

	d := NewNumDatum(666)
	testCtx.pushDatum(d)

	expectedStack := []Datum{
		NewNumDatum(666),
	}
	verifycontextStack(t, testCtx, expectedStack)

	popValueAndVerify(t, testCtx, d)
}

func Test2PushesAndPops(t *testing.T) {
	testCtx := newTestContext()

	d1 := NewNumDatum(666)
	testCtx.pushDatum(d1)
	d2 := NewNumDatum(333)
	testCtx.pushDatum(d2)

	expectedStack := []Datum{
		NewNumDatum(666),
		NewNumDatum(333),
	}
	verifycontextStack(t, testCtx, expectedStack)
	popValueAndVerify(t, testCtx, d2)
	popValueAndVerify(t, testCtx, d1)
	verifyNoRuntimeErrors(t, testCtx)
}

func TestPopCompareAndPushNum(t *testing.T) {
	testProgBldr := NewProgBuilder("dummy expression")

	testCtx := newTestContext()
	testCtx.pushDatum(NewNumDatum(111))
	testCtx.pushDatum(NewNumDatum(222))

	testProgBldr.Eq(testCtx)
	verifyNoRuntimeErrors(t, testCtx)
	popValueAndVerify(t, testCtx, NewBoolDatum(false))
}

func TestPopCompareAndPushBool(t *testing.T) {
	testProgBldr := NewProgBuilder("dummy expression")

	testCtx := newTestContext()
	testCtx.pushDatum(NewBoolDatum(true))
	testCtx.pushDatum(NewBoolDatum(true))

	testProgBldr.Eq(testCtx)
	verifyNoRuntimeErrors(t, testCtx)
	popValueAndVerify(t, testCtx, NewBoolDatum(true))
}

func TestPopCompareAndPushInvalidBool(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"eq(bool1): Unable to convert datum to a boolean.")
		}
	}()

	testProgBldr := NewProgBuilder("dummy expression")

	testCtx := newTestContext()
	testCtx.pushDatum(NewInvalidDatum())
	testCtx.pushDatum(NewBoolDatum(true))

	testProgBldr.Eq(testCtx)
	t.Fatalf("Failed to panic!")
}

func TestPopCompareAndPushInvalidNum(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"eq(num2): Unable to convert datum to a number.")
		}
	}()

	testProgBldr := NewProgBuilder("dummy expression")

	testCtx := newTestContext()
	testCtx.pushDatum(NewNumDatum(333))
	testCtx.pushDatum(NewInvalidDatum())

	testProgBldr.Eq(testCtx)
	t.Fatalf("Failed to panic!")
}

func TestPopCompareAndPushInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"'=' operator doesn't support 'INVALID = INVALID'")
		}
	}()

	testProgBldr := NewProgBuilder("dummy expression")

	testCtx := newTestContext()
	testCtx.pushDatum(NewInvalidDatum())
	testCtx.pushDatum(NewInvalidDatum())

	testProgBldr.Eq(testCtx)
	t.Fatalf("Failed to panic!")
}

func TestPopWithEmptyStack(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"Stack underflow")
		}
	}()

	testCtx := newTestContext()
	if len(testCtx.stack) != 0 {
		t.Errorf("Wrong starting length for stack.")
		return
	}
	_ = testCtx.popDatum()
	t.Fatalf("Failed to panic!")
}

// Test operation of built-in functions including error handling that run
// the Machine.
func TestStore(t *testing.T) {
	testProgBldr := NewProgBuilder("Dummy expression")
	testCtx := newTestContext()

	d := NewNumDatum(555)
	testCtx.pushDatum(d)

	testProgBldr.Store(testCtx)
	verifyNumResult(t, testCtx, 555)

	verifyNoRuntimeErrors(t, testCtx)
}

func TestStoreNotEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"Storing result when stack is not empty")
		}
	}()

	testProgBldr := NewProgBuilder("Dummy expression")
	testCtx := newTestContext()

	d := NewNumDatum(555)
	testCtx.pushDatum(d)
	testCtx.pushDatum(d)

	testProgBldr.Store(testCtx)
	t.Fatalf("Failed to panic!")
}

// numpush
func TestExecuteNumpush(t *testing.T) {
	testProgBldr := NewProgBuilder("Dummy expression")
	testCtx := newTestContext()

	testProgBldr.CodeNum(123)

	prog, err := testProgBldr.GetMainProg()
	if err != nil {
		t.Fatalf("Cannot get program: %s", err.Error())
		return
	}
	prog[0].fn(testCtx)

	expectedStack := []Datum{
		NewNumDatum(123),
	}
	verifycontextStack(t, testCtx, expectedStack)
	verifyNoCompileErrors(t, testProgBldr)
	verifyNoRuntimeErrors(t, testCtx)
}

// litpush
func TestExecuteLitpush(t *testing.T) {
	testProgBldr := NewProgBuilder("Dummy expression")
	testCtx := newTestContext()

	testProgBldr.CodeLiteral("a string")

	prog, err := testProgBldr.GetMainProg()
	if err != nil {
		t.Fatalf("Cannot get program: %s", err.Error())
		return
	}
	prog[0].fn(testCtx)

	expectedStack := []Datum{
		NewLiteralDatum("a string"),
	}
	verifycontextStack(t, testCtx, expectedStack)
	verifyNoCompileErrors(t, testProgBldr)
	verifyNoRuntimeErrors(t, testCtx)
}

// Bltin - really external function calls we know about.
func TestExecuteBltin(t *testing.T) {
	testProgBldr := NewProgBuilder("Dummy expression")
	testCtx := newTestContext()

	testCtx.pushDatum(NewNumDatum(3))
	testProgBldr.CodeBltin(getSymbol(t, "number"), 1)
	prog, err := testProgBldr.GetMainProg()
	if err != nil {
		t.Fatalf("Cannot get program: %s", err.Error())
		return
	}
	prog[0].fn(testCtx)

	expectedStack := []Datum{
		NewNumDatum(3),
	}
	verifycontextStack(t, testCtx, expectedStack)
	verifyNoCompileErrors(t, testProgBldr)
	verifyNoRuntimeErrors(t, testCtx)
}

func TestBltinVerifyArgTypes(t *testing.T) {
	testCtx := newTestContext()

	testCtx.verifyArgNumAndTypes("foo",
		[]Datum{
			NewLiteralDatum("barbar"), NewNumDatum(123), NewBoolDatum(true)},
		[]DatumTypeChecker{TypeIsLiteral, TypeIsNumber, TypeIsBool})

	// Passes if it doesn't throw a panic.
}

func TestBltinWrongArgTypes(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"foo has mismatched arg type [1]: "+
					"using NUMBER, expect BOOL")
		}
	}()

	testCtx := newTestContext()

	testCtx.verifyArgNumAndTypes("foo",
		[]Datum{
			NewLiteralDatum("barbar"), NewNumDatum(123), NewBoolDatum(true)},
		[]DatumTypeChecker{TypeIsLiteral, TypeIsBool, TypeIsNumber})

	t.Fatalf("Failed to panic!")
}

func TestBltinWrongArgNums(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"foo has mismatched arg nums: "+
					"using 3, expect 2")
		}
	}()

	testCtx := newTestContext()

	testCtx.verifyArgNumAndTypes("foo",
		[]Datum{
			NewLiteralDatum("barbar"), NewNumDatum(123), NewBoolDatum(true)},
		[]DatumTypeChecker{TypeIsLiteral, TypeIsNumber})

	t.Fatalf("Failed to panic!")
}

func TestBltinWrongReturnType(t *testing.T) {
	t.Skipf("Write this.")
}

// Exercises error handling for popNumber()
func TestIllegalAdd(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"Failure to pop number (add (operand2)):: "+
					"Unable to convert datum to a number.")
		}
	}()

	testProgBldr := NewProgBuilder("dummy expression")

	testCtx := newTestContext()
	testCtx.pushDatum(NewInvalidDatum())
	testCtx.pushDatum(NewInvalidDatum())

	testProgBldr.Add(testCtx)
	t.Fatalf("Failed to panic!")
}

// Exercises error handling for popBool()
func TestIllegalOr(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"Failure to pop boolean (or (operand2)):: "+
					"Unable to convert datum to a boolean.")
		}
	}()

	testProgBldr := NewProgBuilder("dummy expression")

	testCtx := newTestContext()
	testCtx.pushDatum(NewInvalidDatum())
	testCtx.pushDatum(NewInvalidDatum())

	testProgBldr.Or(testCtx)
	t.Fatalf("Failed to panic!")
}

// Can only have CodePredEnd in sub-machine, not main program
func TestCodePredEndInMainProgram(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			verifyRuntimeError(t, fmt.Sprintf("%v", r),
				"Encoding PredicateEnd before PredicateStart!")
		}
	}()

	testProgBldr := NewProgBuilder("dummy expression")
	testProgBldr.CodePredEnd()
	t.Fatalf("Failed to panic!")

}

func TestTypeConversionsToDo(t *testing.T) {
	t.Skipf("boolean() and number() conversions from nodesets")
	t.Skipf("popCompareAndPush() conversions.")
}
