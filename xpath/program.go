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

// This file contains the ProgBuilder object which is used by the parser to
// build up the set of Inst (instruction) objects representing the machine
// that can be executed to implement an XPATH statement.

package xpath

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"math"
	"strconv"
	"strings"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	log "github.com/sirupsen/logrus"

	"github.com/sdcio/yang-parser/xpath/xutils"
)

type Prog []Inst
type ProgStack []Prog

func (ps ProgStack) Count() int { return len(ps) }
func (ps ProgStack) Peek() Prog { return ps[len(ps)-1] }
func (ps *ProgStack) Push(p Prog) {
	(*ps) = append((*ps), p)
}

func (ps *ProgStack) Update(i Inst) {
	prog := ps.Pop()
	ps.Push(append(prog, i))
}
func (ps *ProgStack) Pop() Prog {
	if len(*ps) < 1 {
		panic(fmt.Errorf("Encoding PredicateEnd before PredicateStart!"))
	}
	p := (*ps)[len(*ps)-1]
	(*ps) = (*ps)[:len(*ps)-1]
	return p
}

// PROGBUILDER
type ProgBuilder struct {
	// When dealing with (nested) predicates, we can have multiple programs
	// in the process of construction.  When we move into a predicate, we
	// pause construction of the current program and start constructing a
	// new child program.  When we have finished processing the predicate,
	// we return to constructing the parent, and embed the child within
	// the parent as a sub-machine.  We can therefore only have a maximum
	// program stack of 2 unless we have nested predicates.  Also, there
	// should only be one program remaining when we have finished processing
	// an XPATH statement.
	progs   ProgStack
	refExpr string
	// Number of '[' seen, used for debug only.  NOT a count of nesting level.
	preds     int
	parseErr  error
	lineAtErr string
	// For path evaluation, we want to ignore anything inside a predicate.
	// It's an integer not a bool as nesting does count here.
	ignoreInsidePred int
}

func NewProgBuilder(refExpr string) *ProgBuilder {
	progBldr := &ProgBuilder{refExpr: refExpr}
	progBldr.progs.Push(Prog{})

	return progBldr
}

func (progBldr *ProgBuilder) CurrentFix() {
	strNamePrevFunc := progBldr.progs[len(progBldr.progs)-1][len(progBldr.progs[len(progBldr.progs)-1])-1].String()
	if strings.Contains(strNamePrevFunc, "current()") {
		progBldr.CodeFn(progBldr.EvalLocPath, "evalLocPath(afterCurrent)")
	}
}

// Extract relevant predicate from expression - 'preds' indicates which '['
// is the starting point.
func GetSubExpr(expr string, preds int) (retStr string) {
	start := strings.Index(expr, "[")
	if start == -1 {
		return expr
	}
	expr = expr[start:]
	var b bytes.Buffer
	var count int

	for index := 0; index < len(expr); index++ {
		b.WriteByte(expr[index])
		if expr[index] == '[' {
			count++
		}
		if expr[index] == ']' {
			count--
		}
		if count == 0 {
			// Is this the predicate we're looking for?
			if preds == 1 {
				return b.String()
			}
			preds--
			b.Reset()
			index = index + strings.Index(expr[index:], "[") - 1
		}
	}

	return b.String()
}

func GetProgListing(prog Prog, level int) string {
	var b bytes.Buffer

	var prefix string
	for i := 0; i < level; i++ {
		prefix += "\t"
	}

	b.WriteString(prefix + "--- machine start ---\n")
	for _, line := range prog {
		b.WriteString(prefix + line.fnName + "\n")
		if line.subMachine != "" {
			b.WriteString(line.subMachine)
		}
	}
	b.WriteString(prefix + "---- machine end ----\n")

	return b.String()
}

func (progBldr *ProgBuilder) GetMainProg() (prog Prog, err error) {
	if progBldr.progs.Count() > 1 {
		return nil, fmt.Errorf("Program still being compiled - %d!",
			progBldr.progs.Count())
	}

	return progBldr.progs.Peek(), nil
}

func (progBldr *ProgBuilder) UnsupportedName(tokenType int, token string) {
	progBldr.parseErr = fmt.Errorf("%s unsupported: %s",
		xutils.GetTokenName(tokenType), token)
}

// The 'codeXX' functions are called by the parser to build up the machine,
// inserting operations and operands for the operations into a slice.

func (progBldr *ProgBuilder) CodeFn(fn instFunc, fnName string) {
	newInstr := newInst(fn, fnName)
	progBldr.progs.Update(newInstr)
}

func (progBldr *ProgBuilder) CodeSubMachine(
	fn instFunc,
	fnName, subMachine string,
) {
	newInstr := newInstWithSubMachine(fn, fnName, subMachine)
	progBldr.progs.Update(newInstr)
}

func (progBldr *ProgBuilder) CodeNum(num float64) {
	numpush := func(ctx *context) {
		ctx.pushDatum(NewNumDatum(num))
	}
	progBldr.CodeFn(numpush, fmt.Sprintf("numpush\t\t%v", num))
}

func (progBldr *ProgBuilder) PushBool(b bool) {
	numpush := func(ctx *context) {
		ctx.pushDatum(NewBoolDatum(b))
	}
	progBldr.CodeFn(numpush, fmt.Sprintf("boolpush\t\t%v", b))
}

func (progBldr *ProgBuilder) PushNotFound() {
	nsetPush := func(ctx *context) {
		// use BTnkTEI1y8iFq01rk837 as the value for a non found element
		// ctx.pushDatum(NewLiteralDatum("BTnkTEI1y8iFq01rk837"))
		ctx.pushDatum(NewNodesetDatum([]xutils.XpathNode{}))
	}
	progBldr.CodeFn(nsetPush, fmt.Sprintf("nodesetpush\t\t[]"))
}

func (progBldr *ProgBuilder) CodeLiteral(lit string) {
	litpush := func(ctx *context) {
		ctx.pushDatum(NewLiteralDatum(lit))
	}
	progBldr.CodeFn(litpush, fmt.Sprintf("litpush\t\t'%s'", lit))
}

func (progBldr *ProgBuilder) CodePathSetCurrent() {
	pathSetCurrent := func(ctx *context) {
		_ = ctx.actualPathStack.PopPath()
		ctx.actualPathStack.NewPathFromCurrent()
	}
	progBldr.CodeFn(pathSetCurrent,
		fmt.Sprintf("pathsetcurrent"),
	)
}

func (progBldr *ProgBuilder) CodePathOper(elem int) {
	if progBldr.ignoreInsidePred > 0 {
		return
	}

	var pathOperPush func(ctx *context)

	switch elem {
	case '.':
		// noop
	case xutils.DOTDOT:
		pathOperPush = func(ctx *context) {
			ctx.actualPathStack.PushElem("..")
		}
	case xutils.DBLSLASH:
		// not implemented
	case '/':
		pathOperPush = func(ctx *context) {
			ctx.actualPathStack.PushElem("/")
		}
	default:
		// unknown
	}

	if pathOperPush != nil {
		progBldr.CodeFn(pathOperPush,
			fmt.Sprintf("MypathOperPush\t%s", xutils.GetTokenName(elem)))
		return
	}
	log.Debugf("skipped %s token", xutils.GetTokenName(elem))
}

func (progBldr *ProgBuilder) CodeNameTest(name xml.Name) {

	nameTestPush := func(ctx *context) {
		if ctx.predicateCount > 0 && ctx.predicateEvalPath%2 == 0 {
			ctx.pushDatum(NewLiteralDatum(name.Local))
		} else {
			//fmt.Println(utils.ToXPath(ctx.GetActualPath(),false))
			ctx.actualPathStack.PushElem(name.Local)
			//fmt.Println(utils.ToXPath(ctx.GetActualPath(),false))
		}
	}
	progBldr.CodeFn(nameTestPush,
		fmt.Sprintf("MyNameTestPush\t%s", name))
}

func (progBldr *ProgBuilder) CodeBltin(sym *Symbol, numArgs int) {
	bltinOrCustom := func(ctx *context) {
		if (sym.custom && sym.customFunc == nil) ||
			(!sym.custom && sym.bltinFunc == nil) {
			ctx.execError("Cannot run null bltin/custom fn ptr", sym.name)
		}
		// Need to extract and convert operands, in reverse order
		numArgs := len(sym.argTypeCheckers)
		args := make([]Datum, numArgs)
		for index := numArgs - 1; index >= 0; index = index - 1 {
			d := ctx.popDatum()
			d = progBldr.convertArgType(ctx, d, index, sym)
			args[index] = d
		}

		var val Datum
		if sym.custom {
			val = sym.customFunc(args)
		} else {
			val = sym.bltinFunc(ctx, args)
		}

		ctx.verifyReturnType(sym, val)
		ctx.pushDatum(val)
	}

	if numArgs != len(sym.argTypeCheckers) {
		progBldr.parseErr = fmt.Errorf("%s() takes %d args, not %d.",
			sym.name, len(sym.argTypeCheckers), numArgs)
	}

	var fnType string
	if sym.custom {
		fnType = "custom"
	} else {
		fnType = "bltin"
	}
	progBldr.CodeFn(bltinOrCustom, fmt.Sprintf("%s\t\t%s()", fnType, sym.name))
}

func (progBldr *ProgBuilder) CodeEvalLocPathExists() {
	if progBldr.ignoreInsidePred > 0 {
		return
	}
	progBldr.CodeFn(progBldr.EvalLocPathExists, "locPathExists")
}

// Code:
//   - encode EvalLocPath
//   - start new (child) program
func (progBldr *ProgBuilder) CodePredStart() {
	// progBldr.CodeFn(progBldr.EvalLocPath, "evalLocPath(PredStart)")
	// progBldr.progs.Push(Prog{})
	// progBldr.preds++

	// if progBldr.progs.Count() > 2 {
	// 	progBldr.parseErr = fmt.Errorf("Nested predicates not yet supported.")
	// }
	instFn := func(ctx *context) {
		progBldr.NewPathStackFromActual()(ctx)
		ctx.predicateCount += 1
	}

	progBldr.CodeFn(instFn, "PREDSTART")

	//progBldr.CodeFn(progBldr.NewPathStackFromActual(), "PREDSTART - NewPathStackFromActual")
	// progBldr.CodeFn(progBldr.Store, "PREDSTART")
}

func (progBldr *ProgBuilder) NewPathStackFromActual() instFunc {
	return func(ctx *context) {
		ctx.actualPathStack.NewPathFromActual()
	}
}

func (progBldr *ProgBuilder) CodePredStartIgnore() {
	progBldr.ignoreInsidePred++
}

func (progBldr *ProgBuilder) CodePredEndIgnore() {
	progBldr.ignoreInsidePred--
}

// First parameter is 0-indexed in Go, whereas XPath position is
// 1-indexed. Here xpos is the XPath position, and pos is the Go
// index corresponding to it.
func predicateIsTrue(
	res *Result,
	ctx *context,
	pos int,
) bool {
	if isNum(res.value) {
		xpos, err := res.GetNumResult()
		if err != nil {
			ctx.execError(err.Error(), "")
			return false
		}
		if pos == int(xpos-1) {
			return true
		}
	} else {
		add, err := res.GetBoolResult()
		if err != nil {
			ctx.execError(err.Error(), "")
			return false
		}
		if add {
			return true
		}
	}
	return false
}

// Code:
//   - encapsulate current program in 'parent' as EvalSubMachine() function
//   - remove this program from stack so next instruction goes on parent
//     machine.
//
// Run:
//   - EvalSubMachine()
func (progBldr *ProgBuilder) CodePredEnd() {

	// Must explicitly append request to store result
	//progBldr.CodeFn(progBldr.Store, "PREDEND")

	cFn := func(ctx *context) {
		//progBldr.Store(ctx)
		ctx.predicateCount -= 1
		ctx.predicateEvalPath = 0
		ctx.actualPathStack.PopPath()
	}

	progBldr.CodeFn(cFn, "PREDEND")
	// prog := progBldr.progs.Pop()
	// preds := progBldr.preds

	// evalSubMachine := func(ctx *context) {
	// 	inputNodeset := ctx.popNodeSet("evalSubMachine")
	// 	var outputNodeset []xutils.XpathNode

	// 	ctx.addDebug(ctx.pfx + "\t----\n")
	// 	size := len(inputNodeset)

	// 	for pos, node := range inputNodeset {
	// 		expr := GetSubExpr(progBldr.refExpr, preds)
	// 		res := newCtx(
	// 			prog, node, ctx.node,
	// 			pos+1, size, progBldr.progs.Count(),
	// 			expr, ctx.xpathStmtLoc).
	// 			SetDebug(ctx.debug).
	// 			SetAccessibleTree(ctx.filter).
	// 			Run()
	// 		ctx.addDebug(res.output)
	// 		if err := res.GetError(); err != nil {
	// 			ctx.execError(err.Error(), "")
	// 			return
	// 		}
	// 		if predicateIsTrue(res, ctx, pos) {
	// 			outputNodeset = append(outputNodeset, node)
	// 			ctx.addDebug("\tPred:\tMATCH\n")
	// 			ctx.addDebug("\t----\n")
	// 		} else {
	// 			ctx.addDebug("\tPred:\tNo Match\n")
	// 			ctx.addDebug("\t----\n")
	// 		}
	// 	}
	// 	ctx.pushDatum(NewNodesetDatum(outputNodeset))
	// 	ctx.stackedNodesets++
	// }

	// progBldr.CodeSubMachine(evalSubMachine, "evalSubMachine",
	// 	GetProgListing(prog, progBldr.progs.Count()))
}

func (progBldr *ProgBuilder) Store(ctx *context) {
	d := ctx.popDatum() // Current value to work on

	// Couple of sanity checks to make sure there don't appear to be any
	// loose ends after processing the XPATH statement that would suggest
	// a logic error somewhere ...
	if len(ctx.stack) > 0 {
		ctx.execError("Storing result when stack is not empty.", "")
		return
	}
	if ctx.stackedNodesets > 0 {
		ctx.execError("Storing result with error in stackedNodeset handling.",
			"")
		return
	}

	ctx.res.save(d)
}

// Unless we have one or more invalid paths (false on stack) then all is ok.
func (progBldr *ProgBuilder) StorePathEval(ctx *context) {
	for len(ctx.stack) > 0 {
		if !ctx.popBool("Path validation result") {
			ctx.res.save(NewBoolDatum(false))
			return
		}
	}
	ctx.res.save(NewBoolDatum(true))
}

// EvalLocPath is put on the stack whenever a path is finally defined, such that the value is to be resolved
// and put onto the stack for other functions to evaluate.
func (progBldr *ProgBuilder) EvalLocPath(ctx *context) {
	// if EvalLocPath is encountered within a predicat, we need to distinguish.
	// a predicat is the part in the square brackets "interface[name=current()/../something]/mtu"
	// within the curly brackets EvalLocPath will be called twice. first after "name" and then after "current()/../something"
	// the call for "name" is to be interrupted, since we need it as a key identifier in the path and not the resolved value.
	// hence, if we're actually within a predicate
	if ctx.predicateCount > 0 {
		// we add 1 to predicateEvalPath
		ctx.predicateEvalPath += 1
		// and the value of predicateEvalPath is uneven (hence the left side of the assignment [=], since we've already added 1 to predicateEvalPath early)
		// then we skip the resolution for the value
		//if ctx.predicateEvalPath%2 == 1 {
		//	return
		//}
	}

	path := ctx.actualPathStack.PopPath()

	valEntry, err := ctx.current.Navigate(path)
	if err != nil {
		ctx.res.runErr = err
		return
	}

	ctx.lastEvalPath = valEntry

	val, err := valEntry.GetValue()
	if err != nil {
		ctx.res.runErr = err
		return
	}
	ctx.pushDatum(val)

	// reset the path to the actual reference
	ctx.actualPathStack.NewPathFromActual()
}

func (progBldr *ProgBuilder) EvalLocPathExists(ctx *context) {
	if (ctx.pathOperPushes == 0) && (ctx.stackedNodesets == 0) {
		ctx.execError("Cannot evaluate zero length path.", "")
		return
	}

	var locPathElems = make([]pathElem, ctx.pathOperPushes)
	for ; ctx.pathOperPushes > 0; ctx.pathOperPushes-- {
		locPathElems[ctx.pathOperPushes-1] = ctx.popPathElem()
	}

	ctx.pushDatum(NewBoolDatum(ctx.validatePath(
		locPathElems, progBldr.refExpr)))
}

// When we reach the end of a filter expression, we need to check that
// the result pushed to the stack is a nodeset (it's an error according
// to the XPATH RFC if not).  Once checked, we push it back on the stack
// and increment stackedNodesets so subsequent path construction operations
// take this nodeset into account.
func (progBldr *ProgBuilder) FilterExprEnd(ctx *context) {

	// // // NOOP

	// d := ctx.popDatum()

	// ns, ok := d.(nodesetDatum)
	// if !ok {
	// 	ctx.execError("Filter Expression must evaluate to a nodeset.", "")
	// 	return
	// }

	// ctx.pushDatum(ns)
	// ctx.stackedNodesets++
}

func (progBldr *ProgBuilder) LRefEquals(ctx *context) {
	if ctx.pathOperPushes != 1 {
		ctx.execError("Unexpected number of key name elements.", "")
		return
	}
	ctx.pathOperPushes = 0
}

func (progBldr *ProgBuilder) LRefPredEnd(ctx *context) {
	// Stack should contain:
	//
	// - NSET: Nodeset (up to pred start '[')
	// - KEY:  Key name
	// - PTH:  Set of path operations giving a nodeset with single LEAF

	// We call EvalLocPath to convert the latter into a nodeset that should
	// contain a single leaf node (LEAFVAL).
	progBldr.EvalLocPath(ctx)

	// Check we have a single node
	leafNodeSet := ctx.popNodeSet("Leaf")
	if len(leafNodeSet) != 1 {
		ctx.execError(fmt.Sprintf(
			"Leafref statement key value not single leaf.  %d values",
			len(leafNodeSet)), "")
	}
	// ... that is a leaf (has no children).
	leafNode := leafNodeSet[0]
	if !leafNode.XIsLeaf() {
		ctx.execError(fmt.Sprintf(
			"Leafref pathKeyExpr is not a leaf."), "")
	}
	leafVal := leafNode.XValue()

	// We know from grammar definition that key must be a nameTestElem.
	// We validate that 'key' is indeed a key in FilterNodeset as it is
	// possible that it is a key for some elements and not others.
	key := ctx.popPathElem().(nameTestElem).value()

	// Now we need to take NSET and filter to leave only elements which have
	// KEY = LEAFVAL.  As we can have multiple keys, and it is the combination
	// that must be unique, we can get multiple nodes here.
	nset := ctx.popNodeSet("List entries")
	if ctx.debug {
		ctx.addDebug("----\n")
		ctx.addDebug(fmt.Sprintf("FilterNodeSet:\t\t[%s = %s]\n", key,
			leafVal))
		ns := NewNodesetDatum(nset)
		ctx.addDebug(ns.(nodesetDatum).Print(ctx.pfx))
	}
	filteredNodes, debugLog := xutils.FilterNodeset(
		nset, key, leafVal)
	ctx.addDebug(debugLog)

	// Put result on stack and don't forget to record it as the first
	// element of the next EvalLocPath operation.
	ctx.pushDatum(NewNodesetDatum(filteredNodes))
	ctx.stackedNodesets++
}

func (progBldr *ProgBuilder) Add(ctx *context) {
	op2 := ctx.popNumber("add (operand2)")
	op1 := ctx.popNumber("add (operand1)")
	ctx.pushDatum(NewNumDatum(op1 + op2))
}

func (progBldr *ProgBuilder) Sub(ctx *context) {
	op2 := ctx.popNumber("subtract (operand2)")
	op1 := ctx.popNumber("subtract (operand1)")
	ctx.pushDatum(NewNumDatum(op1 - op2))
}

func (progBldr *ProgBuilder) Mul(ctx *context) {
	op2 := ctx.popNumber("multiply (operand2)")
	op1 := ctx.popNumber("multiply (operand1)")
	ctx.pushDatum(NewNumDatum(op1 * op2))
}

func (progBldr *ProgBuilder) Div(ctx *context) {
	denom := ctx.popNumber("div (denominator)")
	numer := ctx.popNumber("div (numerator)")
	if denom == 0.0 {
		ctx.pushDatum(NewNumDatum(math.Inf(1)))
		return
	}
	ctx.pushDatum(NewNumDatum(numer / denom))
}

func (progBldr *ProgBuilder) Mod(ctx *context) {
	denom := ctx.popNumber("mod (denominator)")
	numer := ctx.popNumber("mod (numerator)")
	ctx.pushDatum(NewNumDatum(math.Mod(numer, denom)))
}

func (progBldr *ProgBuilder) Negate(ctx *context) {
	op := ctx.popNumber("negate")
	ctx.pushDatum(NewNumDatum(-op))
}

func (progBldr *ProgBuilder) And(ctx *context) {
	op2 := ctx.popBool("and (operand2)")
	op1 := ctx.popBool("and (operand1)")
	ctx.pushDatum(NewBoolDatum(op1 && op2))
}

func (progBldr *ProgBuilder) Or(ctx *context) {
	op2 := ctx.popBool("or (operand2)")
	op1 := ctx.popBool("or (operand1)")
	ctx.pushDatum(NewBoolDatum(op1 || op2))
}

func (progBldr *ProgBuilder) Eq(ctx *context) {
	// test if d2 (leftmost) is nodeset then handle leaflist evaluation

	// if not, continue
	switch {
	// being out of predicate, this is an equality check
	case ctx.predicateCount == 0:
		boolFn := func(d1, d2 Datum) bool {
			return d1.Boolean("eq(bool1)") == d2.Boolean("eq(bool2)")
		}
		litFn := func(d1, d2 Datum) bool {
			return d1.Literal("eq(lit1)") == d2.Literal("eq(lit2)")
		}
		numFn := func(d1, d2 Datum) bool {
			// Some values of NaN are more equal than others, but if either
			// n1 or n2 is NaN, then n1 != n2.
			n1 := d1.Number("eq(num1)")
			n2 := d2.Number("eq(num2)")
			return (n1 == n2) && !math.IsNaN(n1) && !math.IsNaN(n2)
		}
		ctx.popCompareEqualityAndPush(boolFn, litFn, numFn, "=")
	case ctx.predicateCount > 0:
		// being within a predicate this is an assignment
		d1 := ctx.popDatum()
		d2 := ctx.popDatum()
		_ = d2
		ctx.actualPathStack.PopPath()
		ctx.actualPathStack.PushElem(d1.Literal(""))
		ctx.actualPathStack.NewPathFromActual()
	}
}

func (progBldr *ProgBuilder) Ne(ctx *context) {
	boolFn := func(d1, d2 Datum) bool {
		return d1.Boolean("ne(bool1)") != d2.Boolean("ne(bool2)")
	}
	litFn := func(d1, d2 Datum) bool {
		return d1.Literal("ne(lit1)") != d2.Literal("ne(lit2)")
	}
	numFn := func(d1, d2 Datum) bool {
		// If either n1 or n2 is NaN, then n1 != n2.
		n1 := d1.Number("ne(num1)")
		n2 := d2.Number("ne(num2)")
		return (n1 != n2) || math.IsNaN(n1) || math.IsNaN(n2)
	}
	ctx.popCompareEqualityAndPush(boolFn, litFn, numFn, "!=")
}

func (progBldr *ProgBuilder) Lt(ctx *context) {
	// All relational comparisons are done as numbers
	numFn := func(d1, d2 Datum) bool {
		return d1.Number("lt(op1)") < d2.Number("lt(op2)")
	}
	boolFn := func(d1, d2 Datum) bool {
		return numFn(NewBoolDatum(d1.Boolean("lt(bool1)")),
			NewBoolDatum(d2.Boolean("lt(bool2")))
	}
	litFn := func(d1, d2 Datum) bool {
		return numFn(NewLiteralDatum(d1.Literal("lt(bool1)")),
			NewLiteralDatum(d2.Literal("lt(bool2")))
	}

	ctx.popCompareRelationalAndPush(boolFn, litFn, numFn, "<")
}

func (progBldr *ProgBuilder) Gt(ctx *context) {
	// All relational comparisons are done as numbers
	numFn := func(d1, d2 Datum) bool {
		return d1.Number("gt(op1)") > d2.Number("gt(op2)")
	}
	boolFn := func(d1, d2 Datum) bool {
		return numFn(NewBoolDatum(d1.Boolean("gt(bool1)")),
			NewBoolDatum(d2.Boolean("gt(bool2")))
	}
	litFn := func(d1, d2 Datum) bool {
		return numFn(NewLiteralDatum(d1.Literal("gt(bool1)")),
			NewLiteralDatum(d2.Literal("gt(bool2")))
	}

	ctx.popCompareRelationalAndPush(boolFn, litFn, numFn, ">")
}

func (progBldr *ProgBuilder) Le(ctx *context) {
	// All relational comparisons are done as numbers
	numFn := func(d1, d2 Datum) bool {
		return d1.Number("le(op1)") <= d2.Number("le(op2)")
	}
	boolFn := func(d1, d2 Datum) bool {
		return numFn(NewBoolDatum(d1.Boolean("le(bool1)")),
			NewBoolDatum(d2.Boolean("le(bool2")))
	}
	litFn := func(d1, d2 Datum) bool {
		return numFn(NewLiteralDatum(d1.Literal("le(bool1)")),
			NewLiteralDatum(d2.Literal("le(bool2")))
	}

	ctx.popCompareRelationalAndPush(boolFn, litFn, numFn, "<=")
}

func (progBldr *ProgBuilder) Ge(ctx *context) {
	// All relational comparisons are done as numbers
	numFn := func(d1, d2 Datum) bool {
		return d1.Number("ge(op1)") >= d2.Number("ge(op2)")
	}
	boolFn := func(d1, d2 Datum) bool {
		return numFn(NewBoolDatum(d1.Boolean("lt(bool1)")),
			NewBoolDatum(d2.Boolean("lt(bool2")))
	}
	litFn := func(d1, d2 Datum) bool {
		return numFn(NewLiteralDatum(d1.Literal("lt(bool1)")),
			NewLiteralDatum(d2.Literal("lt(bool2")))
	}

	ctx.popCompareRelationalAndPush(boolFn, litFn, numFn, ">=")
}

func (progBldr *ProgBuilder) Union(ctx *context) {
	op2 := ctx.popNodeSet("union (operand2)")
	op1 := ctx.popNodeSet("union (operand1)")

	ctx.pushDatum(NewNodesetDatum(append(op1, op2...)))
}

// Convert arg type according to XPATH rules into required type for passing
// into next function...
func (progBldr *ProgBuilder) convertArgType(
	ctx *context,
	d Datum,
	argNum int,
	sym *Symbol,
) Datum {
	// Quick check to see if we don't need to convert.
	if ok, _ := sym.argTypeCheckers[argNum](d); ok {
		return d
	}

	// Conversion is required, so work through the possibilities.  Cannot
	// convert *to* a nodeset, so if 'd' is not already a nodeset then
	// we have a problem.
	if ok, _ := sym.argTypeCheckers[argNum](NewNumDatum(0)); ok {
		n := d.Number(
			fmt.Sprintf("%s(): cannot convert %s to number. ",
				sym.name, d.name()))
		return NewNumDatum(n)
	} else if ok, _ := sym.argTypeCheckers[argNum](NewLiteralDatum("")); ok {
		l := d.Literal(
			fmt.Sprintf("%s(): cannot convert %s to literal. ",
				sym.name, d.name()))
		return NewLiteralDatum(l)
	} else if ok, _ := sym.argTypeCheckers[argNum](NewBoolDatum(true)); ok {
		b := d.Boolean(
			fmt.Sprintf("%s(): cannot convert %s to boolean. ",
				sym.name, d.name()))
		return NewBoolDatum(b)
	}

	_, name := sym.argTypeCheckers[argNum](NewBoolDatum(true))
	ctx.execError(fmt.Sprintf(
		"Fn '%s' takes %s, not %s as arg %d.\n", sym.name,
		name, d.name(), argNum),
		"")

	return NewInvalidDatum()
}

// to be replaced with data-server/utils.TypedValueToString
func tvToString(tv *sdcpb.TypedValue) string {
	switch tv.Value.(type) {
	case *sdcpb.TypedValue_AnyVal:
		return string(tv.GetAnyVal().GetValue()) // questionable...
	case *sdcpb.TypedValue_AsciiVal:
		return tv.GetAsciiVal()
	case *sdcpb.TypedValue_BoolVal:
		return strconv.FormatBool(tv.GetBoolVal())
	case *sdcpb.TypedValue_BytesVal:
		return string(tv.GetBytesVal()) // questionable...
	case *sdcpb.TypedValue_DecimalVal:
		d := tv.GetDecimalVal()
		digitsStr := strconv.FormatInt(d.Digits, 10)
		negative := false
		if d.Digits < 0 {
			negative = true
			digitsStr = digitsStr[1:] // Remove the "-" sign for processing
		}
		// Add leading zeros if necessary
		for uint32(len(digitsStr)) <= d.Precision {
			digitsStr = "0" + digitsStr
		}
		// Insert the decimal point
		if d.Precision > 0 {
			decimalPointIndex := len(digitsStr) - int(d.Precision)
			digitsStr = digitsStr[:decimalPointIndex] + "." + digitsStr[decimalPointIndex:]
		}
		// Add back the negative sign if necessary
		if negative {
			digitsStr = "-" + digitsStr
		}
		return digitsStr
	case *sdcpb.TypedValue_DoubleVal:
		return strconv.FormatFloat(tv.GetDoubleVal(), byte('e'), -1, 64)
	case *sdcpb.TypedValue_FloatVal:
		return strconv.FormatFloat(float64(tv.GetFloatVal()), byte('e'), -1, 64)
	case *sdcpb.TypedValue_IntVal:
		return strconv.Itoa(int(tv.GetIntVal()))
	case *sdcpb.TypedValue_JsonIetfVal:
		return string(tv.GetJsonIetfVal())
	case *sdcpb.TypedValue_JsonVal:
		return string(tv.GetJsonVal())
	case *sdcpb.TypedValue_LeaflistVal:
		rs := make([]string, 0, len(tv.GetLeaflistVal().GetElement()))
		for _, lfv := range tv.GetLeaflistVal().GetElement() {
			rs = append(rs, tvToString(lfv))
		}
		return strings.Join(rs, ",")
	case *sdcpb.TypedValue_ProtoBytes:
		return string(tv.GetProtoBytes()) // questionable
	case *sdcpb.TypedValue_StringVal:
		return tv.GetStringVal()
	case *sdcpb.TypedValue_UintVal:
		return strconv.Itoa(int(tv.GetUintVal()))
	}
	return ""
}
