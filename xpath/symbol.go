// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package xpath

import (
	"bytes"
	"fmt"
	"math"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/steiler/yang-parser/xpath/xutils"
)

type bltinFn func(*context, []Datum) Datum
type CustomFn func([]Datum) Datum

type Symbol struct {
	name            string // Useful when symbol is referenced outside of map.
	argTypeCheckers []DatumTypeChecker
	retTypeChecker  DatumTypeChecker

	val        float64
	bltinFunc  bltinFn
	customFunc CustomFn
	custom     bool // This is a custom function, not core XPATH
}

func (sym *Symbol) GetName() string { return sym.name }

func NewFnSym(
	name string,
	fn bltinFn,
	argTypeCheckers []DatumTypeChecker,
	retTypeChecker DatumTypeChecker,
) *Symbol {
	return &Symbol{
		name:            name,
		bltinFunc:       fn,
		argTypeCheckers: argTypeCheckers,
		retTypeChecker:  retTypeChecker,
	}
}

func NewCustomFnSym(
	name string,
	fn CustomFn,
	argTypeCheckers []DatumTypeChecker,
	retTypeChecker DatumTypeChecker,
) *Symbol {
	return &Symbol{
		name:            name,
		customFunc:      fn,
		argTypeCheckers: argTypeCheckers,
		retTypeChecker:  retTypeChecker,
		custom:          true,
	}
}

func NewDummyFnSym(name string) *Symbol {
	return &Symbol{
		name:   name,
		custom: true,
	}
}

type symbolTable map[string]*Symbol

// These functions are the core XPATH functions, along with current() as
// defined in the YANG spec.  Where a function name has 'x' as prefix, this
// is to avoid namespace clashes with either Golang (eg string) or with
// internal functions called by these functions (eg round, boolean etc).
var xpathFunctionTable = symbolTable{
	"boolean": NewFnSym("boolean", xBoolean,
		[]DatumTypeChecker{TypeIsObject}, TypeIsBool),
	"ceiling": NewFnSym("ceiling", ceiling,
		[]DatumTypeChecker{TypeIsNumber}, TypeIsNumber),
	"concat": NewFnSym("concat", concat,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsLiteral}, TypeIsLiteral),
	"contains": NewFnSym("contains", contains,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsLiteral}, TypeIsBool),
	"re-match": NewFnSym("re-match", re_match,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsLiteral}, TypeIsBool),
	"count": NewFnSym("count", count,
		[]DatumTypeChecker{TypeIsNodeset}, TypeIsNumber),
	"current": NewFnSym("current", current,
		[]DatumTypeChecker{}, TypeIsNodeset),
	"false": NewFnSym("false", xFalse,
		[]DatumTypeChecker{}, TypeIsBool),
	"floor": NewFnSym("floor", floor,
		[]DatumTypeChecker{TypeIsNumber}, TypeIsNumber),
	"last": NewFnSym("last", last,
		[]DatumTypeChecker{}, TypeIsNumber),
	"local-name": NewFnSym("local-name", localName,
		[]DatumTypeChecker{TypeIsNodeset}, TypeIsLiteral),
	"normalize-space": NewFnSym("normalize-space", normalizeSpace,
		[]DatumTypeChecker{TypeIsLiteral}, TypeIsLiteral),
	"not": NewFnSym("not", not,
		[]DatumTypeChecker{TypeIsBool}, TypeIsBool),
	"number": NewFnSym("number", xNumber,
		[]DatumTypeChecker{TypeIsObject}, TypeIsNumber),
	"round": NewFnSym("round", round,
		[]DatumTypeChecker{TypeIsNumber}, TypeIsNumber),
	"position": NewFnSym("position", position,
		[]DatumTypeChecker{}, TypeIsNumber),
	"starts-with": NewFnSym("starts-with", startsWith,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsLiteral},
		TypeIsBool),
	"string": NewFnSym("string", xString,
		[]DatumTypeChecker{TypeIsObject}, TypeIsLiteral),
	"string-length": NewFnSym("string-length", stringLength,
		[]DatumTypeChecker{TypeIsLiteral}, TypeIsNumber),
	"substring": NewFnSym("substring", substring,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsNumber, TypeIsNumber},
		TypeIsLiteral),
	"substring-after": NewFnSym("substring-after", substringAfter,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsLiteral},
		TypeIsLiteral),
	"substring-before": NewFnSym("substring-before", substringBefore,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsLiteral},
		TypeIsLiteral),
	"sum": NewFnSym("sum", sum,
		[]DatumTypeChecker{TypeIsNodeset}, TypeIsNumber),
	"translate": NewFnSym("translate", translate,
		[]DatumTypeChecker{TypeIsLiteral, TypeIsLiteral, TypeIsLiteral},
		TypeIsLiteral),
	"true": NewFnSym("true", xTrue,
		[]DatumTypeChecker{}, TypeIsBool),
}

type UserCustomFunctionCheckerFn func(name string) (*Symbol, bool)

// LookupXpathFunction - return Symbol if 'name' is found in symbol table
//
// Core XPATH functions are in the function table with sym.custom = false.
// Custom XPATH functions available in the ISO via plugins (with corresponding
// INI file describing them) are also in the table, with sym.custom = true.
//
// Additionally there are circumstances (eg DRAM) where we wish to temporarily
// add function names to the table to prevent erroneous compiler errors.
// This is done via the optional userFnCheckFn which will return a dummy
// symbol if needed.
func LookupXpathFunction(
	name string,
	customFnsAllowed bool,
	userFnCheckFn UserCustomFunctionCheckerFn,
) (*Symbol, bool) {
	if !pluginsLoaded {
		RegisterCustomFunctions(openPlugins())
	}
	if sym, ok := xpathFunctionTable[name]; ok {
		if !sym.custom || customFnsAllowed {
			return sym, true
		}
	} else if userFnCheckFn != nil {
		return userFnCheckFn(name)
	}

	return nil, false
}

var testedFunctionTable = make(map[string]bool)

func markFunctionAsTested(name string) {
	testedFunctionTable[name] = true
}

func CheckAllFunctionsWereTested() error {
	for name, _ := range xpathFunctionTable {
		if _, ok := testedFunctionTable[name]; !ok {
			return fmt.Errorf("Function '%s' has not been tested!", name)
		}
	}

	return nil
}

// Functions: these are the core functions in XPATH Section 4, plus
// current() as described in YANG RFC 6020 Section 6.4.1
//
// Some names have the form 'xFn' - this is either to avoid namespace
// clashes, or where the underlying functionality is useful internally
// with a different prototype, in machine.go, and so 'xFn' is then a wrapper
// around 'fn'
//
// Functions may take specific types, or 'objects' which means any of the
// 4 types (literal, number, boolean and nodeset), and may return one of the
// 4 specific types.  Arguments of the 'wrong' type should be converted,
// generating an error if the conversion fails.  It is also an error to pass
// in the wrong number of arguments.
//
// NB: Just to spice this up a little, in some functions the last argument
//     is optional.
// NB: To further spice it up, concat() is variadic, taking 2 or more args.
//
// Argument count is dealt with at compile time.  codeBltin() is called with
// the number of arguments as matched by the YACC parser.  At this point we
// verify the number of arguments matches that encoded in FunctionTable and
// fail if there's a mismatch.
//
// This is also where we inject optional arguments, as type DATUM_OPT.  This
// means that there is no need to check argument numbers at runtime.  The
// only risk regarding argument numbers is if the FunctionTable and function
// definition are mismatched.  We check the number and type of arguments
// by the ctx.assert() call in each function that is only activated under
// test to avoid a runtime hit.  This localises the risk of mismatch as we now
// have the function definition and the check in the same place and can
// visually check they align.
//
// Arguments are pushed to the context object's stack, then popped by the
// bltin() function.  At this point they are converted to the expected type,
// and an error raised if this fails.  This means that when we call literal(),
// number() etc below, we do NOT need to deal with the error, as any error
// will have already been raised.  If type conversion worked, then when we
// call boolean / literal etc this time, we are getting the 'native' type
// back, unconverted, which will never generate an error.  The exception is
// when a function takes an OBJECT, in which case we do need to convert.
//
// Verification of the return type is again controlled by the 'testing' flag
// in the context and is done when we get the return value in bltin().
//

// Round UP to nearest integer.
func ceiling(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("ceiling",
		args, []DatumTypeChecker{TypeIsNumber})

	num0 := args[0].Number("ceiling()")
	return NewNumDatum(float64(math.Ceil(num0)))
}

// Limited version taking 2 arguments rather than variadic ...
func concat(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("concat",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsLiteral})

	lit0 := args[0].Literal("concat()")
	lit1 := args[1].Literal("concat()")

	return NewLiteralDatum(lit0 + lit1)
}

func contains(ctx *context, args []Datum) (retBool Datum) {
	ctx.verifyArgNumAndTypes("contains",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsLiteral})

	lit0 := args[0].Literal("contains()")
	lit1 := args[1].Literal("contains()")

	return NewBoolDatum(strings.Contains(lit0, lit1))
}

func re_match(ctx *context, args []Datum) (retBool Datum) {
	ctx.verifyArgNumAndTypes("contains",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsLiteral})

	lit0 := args[0].Literal("re_match()")
	lit1 := args[1].Literal("re_match()")
	rx, err := regexp.Compile(lit1)
	if err != nil {
		log.Error(err)
		return NewBoolDatum(true)
	}

	return NewBoolDatum(rx.MatchString(lit0))
}

func count(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("count",
		args, []DatumTypeChecker{TypeIsNodeset})
	ns0 := args[0].Nodeset("count()")
	return NewNumDatum(float64(len(ns0)))
}

func current(ctx *context, args []Datum) (retNodeSet Datum) {
	// reset the path to the current path
	ctx.ActualPathReset()
	return NewNodesetDatum([]xutils.XpathNode{})
}

// Round DOWN to nearest integer
func floor(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("floor",
		args, []DatumTypeChecker{TypeIsNumber})

	num0 := args[0].Number("floor()")
	return NewNumDatum(float64(math.Floor(num0)))
}

func last(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("last",
		args, []DatumTypeChecker{})

	return NewNumDatum(float64(ctx.size))
}

func localName(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("local-name",
		args, []DatumTypeChecker{TypeIsNodeset})

	ns0 := args[0].Nodeset("local-name()")

	// TODO - optional argument

	if len(ns0) == 0 {
		return NewLiteralDatum("")
	}

	return NewLiteralDatum(ns0[0].XName())
}

func normalizeSpace(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("normalize-space",
		args, []DatumTypeChecker{TypeIsLiteral})

	lit0 := args[0].Literal("normalizeSpace()")

	fields := strings.Fields(lit0)
	var b bytes.Buffer
	for _, field := range fields {
		b.WriteString(field)
		b.WriteString(" ")
	}
	retStr := b.String()
	retStr = retStr[:len(retStr)-1] // Remove last space
	return NewLiteralDatum(retStr)
}

func not(ctx *context, args []Datum) (retBool Datum) {
	ctx.verifyArgNumAndTypes("not",
		args, []DatumTypeChecker{TypeIsBool})

	bool0 := args[0].Boolean("not()")
	return NewBoolDatum(!bool0)
}

func round(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("round",
		args, []DatumTypeChecker{TypeIsNumber})

	num0 := args[0].Number("round()")

	// Trunc() rounds towards zero.
	var rounded = 0.0
	if num0 >= 0 {
		rounded = float64(math.Trunc(0.5 + num0))
	} else {
		rounded = -float64(math.Trunc(0.5 - num0))
	}

	return NewNumDatum(rounded)
}

func position(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("position",
		args, []DatumTypeChecker{})

	return NewNumDatum(float64(ctx.pos))
}

func startsWith(ctx *context, args []Datum) (retBool Datum) {
	ctx.verifyArgNumAndTypes("starts-with",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsLiteral})

	lit0 := args[0].Literal("starts-with()")
	lit1 := args[1].Literal("starts-with()")

	if strings.HasPrefix(lit0, lit1) {
		return NewBoolDatum(true)
	}
	return NewBoolDatum(false)
}

func stringLength(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("string-length",
		args, []DatumTypeChecker{TypeIsLiteral})

	lit0 := args[0].Literal("string-length()")

	return NewNumDatum(float64(len(lit0)))
}

// Returns substring of arg[0] starting with the character at position arg[1],
// and of length arg[2].  If arg[2] isn't specified, return remainder of
// string.
func substring(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("substring",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsNumber, TypeIsNumber})

	lit0 := args[0].Literal("substring()")
	num1 := args[1].Number("substring()")
	num2 := args[2].Number("substring()")

	substrLen := len(lit0)
	if substrLen == 0 {
		return NewLiteralDatum("")
	}

	// NB: XPATH uses 1 as first index in string, not zero, so we have to
	//     subtract one here.  We also need to ensure both start and end Pos
	//     are >= 0.
	startPos := int(math.Trunc(num1+0.5)) - 1
	endPos := int(math.Trunc(num2+0.5)) + startPos
	if startPos < 0 {
		// Only do this AFTER calculating endPos as the spec says we calculate
		// length based on the rounded difference of the two params.
		startPos = 0
	}
	if startPos >= substrLen {
		return NewLiteralDatum("")
	}
	if endPos < 0 {
		endPos = 0
	}
	if endPos > substrLen {
		endPos = substrLen
	}
	substr := lit0[startPos:endPos]
	return NewLiteralDatum(substr)
}

func substringAfter(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("substring-after",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsLiteral})

	lit0 := args[0].Literal("substring-after()")
	lit1 := args[1].Literal("substring-after()")

	// If second string is empty, return first
	if len(lit1) == 0 {
		return NewLiteralDatum(lit0)
	}

	// idx == -1: not found
	if idx := strings.Index(lit0, lit1); idx >= 0 {
		return NewLiteralDatum(lit0[idx+len(lit1):])
	}
	return NewLiteralDatum("")
}

func substringBefore(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("substring-before",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsLiteral})

	lit0 := args[0].Literal("substring-before()")
	lit1 := args[1].Literal("substring-before()")

	// idx == -1: not found
	// idx ==  0: string at start, so empty before string
	if idx := strings.Index(lit0, lit1); idx > 0 {
		return NewLiteralDatum(lit0[:idx])
	}
	return NewLiteralDatum("")
}

func sum(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("sum",
		args, []DatumTypeChecker{TypeIsNodeset})

	ns0 := args[0].Nodeset("sum()")
	total := 0.0

	nsStrings := xutils.GetStringValues(ns0, false)

	for _, nodeStr := range nsStrings {
		total = total + numberFromString(nodeStr)
	}

	return NewNumDatum(total)
}

// Replace first string with characters in the second string replaced by
// the character in the equivalent position in the third string.  If the
// second string is longer than the third, then characters with no equivalent
// in the third string are removed.  If a character appears twice in the first
// string then the first occurrence determines the replacement to be used.
func translate(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("translate",
		args, []DatumTypeChecker{TypeIsLiteral, TypeIsLiteral, TypeIsLiteral})

	src := args[0].Literal("translate()")
	from := args[1].Literal("translate()")
	to := args[2].Literal("translate()")

	if len(src) == 0 || len(from) == 0 {
		return NewLiteralDatum(src)
	}

	var toChar string
	var alreadyTranslated = make(map[string]bool)
	for index, fromChar := range from {
		// Ensure we don't translate twice.
		if _, present := alreadyTranslated[string(fromChar)]; present {
			continue
		}
		alreadyTranslated[string(fromChar)] = true

		// Work out required replacement / removal
		if index < len(to) {
			toChar = to[index : index+1]
		} else {
			toChar = ""
		}

		src = strings.Replace(src, string(fromChar), toChar,
			-1 /* replace all */)
	}

	return NewLiteralDatum(src)
}

func xBoolean(ctx *context, args []Datum) Datum {
	ctx.verifyArgNumAndTypes("boolean",
		args, []DatumTypeChecker{TypeIsObject})
	return NewBoolDatum(args[0].Boolean("boolean()"))
}

func xFalse(ctx *context, args []Datum) (retBool Datum) {
	ctx.verifyArgNumAndTypes("false",
		args, []DatumTypeChecker{})

	return NewBoolDatum(false)
}

func xNumber(ctx *context, args []Datum) (retNum Datum) {
	ctx.verifyArgNumAndTypes("number",
		args, []DatumTypeChecker{TypeIsObject})

	return NewNumDatum(args[0].Number("number()"))
}

func xString(ctx *context, args []Datum) (retLit Datum) {
	ctx.verifyArgNumAndTypes("string",
		args, []DatumTypeChecker{TypeIsObject})

	return NewLiteralDatum(args[0].Literal("string()"))
}

func xTrue(ctx *context, args []Datum) (retBool Datum) {
	ctx.verifyArgNumAndTypes("true",
		args, []DatumTypeChecker{})

	return NewBoolDatum(true)
}
