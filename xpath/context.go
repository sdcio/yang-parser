// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains the 'context' object used to run a machine with a
// specific context.

package xpath

import (
	"bytes"
	"fmt"
	"github.com/danos/yang/xpath/xutils"
)

// CONTEXT
//
// Context on which to run machine, so one machine can be run multiple times
// concurrently.
//
// As we run, the stack content varies, but must end up empty as we take off
// the remaining element with the 'store' instruction.  For the likes of
// the EvalLocPath operation that takes a set of path operations and name-
// tests, we need to track what currently stacked objects are to be used by
// the operation.  These objects could be either a previously calculated
// nodeset (eg where we are about to do a predicate operation) or a set of
// raw path operations (DOTDOT, nametests etc) or a combination.  Note that
// there can only ever be one stacked nodeset per operation, and that the
// stacked nodeset will always come before any path operations.
//
// Expressed in logic, the set of objects is:
//
// [Nodeset IF (stackedNodesets > 0)] + pathOperPushes PathElements
//
// stackedNodesets could be >1 where we have nested predicates.
type context struct {
	// XPATH context node data
	node  xutils.XpathNode
	pos   int
	size  int
	level int
	// For predicates etc, we need to know the initial context node for
	// current() to work.  The initial node is the node that the XPATH
	// statement belongs to.
	initNode xutils.XpathNode
	filter   xutils.MatchType // Accessible tree filter

	pathOperPushes  int
	stackedNodesets int
	stack           []stackable
	prog            []Inst

	res *Result

	validate     bool // Used to enable type checking etc.
	debug        bool // For logging, dump instructions and stack as we go ...
	b            bytes.Buffer
	pfx          string // Prefix when printing
	refExpr      string // Expression being evaluated
	xpathStmtLoc string // Module:line of original xpath statement.
}

// As well as the initial context created when we start to evaluate an Xpath
// expression, we also need to cater for the context created when evaluating
// a predicate.  In this case, we need to track the initial context node as
// well as the local one, and also note the position and size of the context.
// For example, if we are evaluating nodeset[key = 'foo'] and <nodeset> has
// 4 elements, then size is 4 for each node, and pos ranges from 1 to 4.

// NewCtxFromMach - return new context in which to run instance of machine.
//
// Use for creating context for top-level machine - machines for nested
// predicates etc need fine-tuning.
//
func NewCtxFromMach(mach *Machine, ctxNode xutils.XpathNode) *context {
	return &context{
		res:          NewResult(),
		node:         ctxNode,
		initNode:     ctxNode,
		validate:     false,
		debug:        false,
		filter:       xutils.FullTree,
		pos:          1,
		size:         1,
		level:        0,
		refExpr:      mach.refExpr,
		prog:         mach.prog,
		xpathStmtLoc: mach.location,
	}
}

// newCtx - create customised context, necessary for predicates etc.
func newCtx(
	prog []Inst,
	ctxNode, initNode xutils.XpathNode,
	pos, size, level int,
	refExpr, location string,
) *context {
	ctx := &context{
		res:          NewResult(),
		node:         ctxNode,
		initNode:     initNode,
		validate:     false,
		debug:        false,
		filter:       xutils.FullTree,
		pos:          pos,
		size:         size,
		level:        level,
		refExpr:      refExpr,
		prog:         prog,
		xpathStmtLoc: location,
	}
	for i := 0; i < level; i++ {
		ctx.pfx += "\t"
	}
	return ctx
}

// These Enable / Set methods are designed to be chained together.
func (ctx *context) EnableDebug() *context {
	ctx.debug = true
	return ctx
}
func (ctx *context) SetDebug(debug bool) *context {
	ctx.debug = debug
	return ctx
}
func (ctx *context) EnableValidation() *context {
	ctx.validate = true
	return ctx
}
func (ctx *context) SetValidation(validate bool) *context {
	ctx.validate = validate
	return ctx
}
func (ctx *context) SetAccessibleTree(filter xutils.MatchType) *context {
	ctx.filter = filter
	return ctx
}
func (ctx *context) AccessibleTreeConfigOnly() *context {
	ctx.filter = xutils.ConfigOnly
	return ctx
}

// panic() seems reasonable as it is a run-time error that we shouldn't
// get.  Alternative requires a lot of careful unwinding and/or putting
// sensible values on the stack such that we continue to correctly run
// the remaining instructions.
//
// Alternatively, we could check error status each time we loop through
// the instructions in the machine, in the Execute() function.
func (ctx *context) execError(desc, param string) {
	panic(fmt.Sprintf("%s %s", desc, param))
}

func (ctx *context) addDebug(entry string) {
	if ctx.debug {
		ctx.b.WriteString(entry)
	}
}

func (ctx *context) addDebugNodeset(ns []xutils.XpathNode) {
	if ctx.debug {
		ctx.b.WriteString(
			NewNodesetDatum(ns).(nodesetDatum).Print(ctx.pfx))
	}
}

func (ctx *context) formatAndAddDebug(format string, params ...interface{}) {
	if ctx.debug {
		ctx.b.WriteString(fmt.Sprintf(format, params...))
	}
}

func (ctx *context) addDebugProgHeader(refExpr string) {
	if ctx.debug {
		ctx.b.WriteString(ctx.pfx + "Run\t'")
		ctx.b.WriteString(refExpr)
		ctx.b.WriteString("' on:\n\t" + ctx.pfx)
		ctx.b.WriteString(xutils.NodeString(ctx.node))
		ctx.b.WriteString("\n" + ctx.pfx + "----\n")
	}
}

// Add any current error to our debug and store on the context result for
// future reference.
func (ctx *context) saveDebug() {
	if ctx.debug {
		if ctx.res.runErr != nil {
			errMsg := fmt.Sprintf("%sError\t", ctx.pfx)
			ctx.addDebug(errMsg + ctx.res.runErr.Error() + "\n----\n")
		}
		ctx.res.output = ctx.b.String()
	}
}

// Dump saved debug to the log file (dump to stdout achieves this).
func (ctx *context) logDebug() {
	if ctx.debug {
		fmt.Println(ctx.b.String())
	}
}

func (ctx *context) addDebugInstrAndStack(instrName string) {
	if ctx.debug {
		ctx.b.WriteString(ctx.pfx + "Instr:\t")
		ctx.b.WriteString(instrName)
		ctx.b.WriteString("\n")
		ctx.b.WriteString(ctx.printStack(ctx.pfx))
	}
}

// In theory, we have populated the function table with correct return types
// and arg types (and number) so that we will never have any problems with
// mismatches at runtime as we can catch at compile time.
//
// However, if we do want to be paranoid, this function checks for us.  We use
// the 'testing' flag on the context so we only do this when running UT.
func (ctx *context) verifyArgNumAndTypes(
	fnName string,
	args []Datum,
	expArgTypeCheckers []DatumTypeChecker,
) {
	if ctx.validate == false {
		return
	}
	markFunctionAsTested(fnName)

	if len(args) != len(expArgTypeCheckers) {
		ctx.execError(fmt.Sprintf(
			"%s has mismatched arg nums: using %d, expect %d",
			fnName, len(args), len(expArgTypeCheckers)),
			"")
		return
	}

	for argNum, arg := range args {
		if ok, name := expArgTypeCheckers[argNum](arg); !ok {
			ctx.execError(fmt.Sprintf(
				"%s has mismatched arg type [%d]: using %s, expect %s",
				fnName, argNum, arg.name(), name), "")
			return
		}
	}
}

func (ctx *context) verifyReturnType(sym *Symbol, d Datum) Datum {
	if ctx.validate == false {
		return d
	}

	if ok, name := sym.retTypeChecker(d); ok {
		return d
	} else {
		ctx.execError(fmt.Sprintf(
			"%s has mismatched ret type: exp %s, got %s",
			sym.name, d.name(), name),
			"")
		return NewInvalidDatum()
	}
}

// Print top of stack (last entry) FIRST.
// <prefix> allows for indentation (eg string of tabs) so when we have
// nested machines the stack dump is correctly aligned.
func (ctx *context) printStack(prefix string) string {
	if len(ctx.stack) == 0 {
		return prefix + "Stack:\t(empty)\n"
	}

	var b bytes.Buffer
	for index, _ := range ctx.stack {
		b.WriteString(prefix)
		if index == 0 {
			b.WriteString("Stack:")
		}
		b.WriteString(fmt.Sprintf("\t%s\n", ctx.stack[len(ctx.stack)-index-1]))
	}

	return b.String()
}

type stackable interface {
}

func (ctx *context) pushInternal(s stackable) {
	ctx.stack = append(ctx.stack, s)
}

func (ctx *context) pushDatum(d Datum) {
	ctx.pushInternal(d)
}

func (ctx *context) pushPathElem(p pathElem) {
	ctx.pushInternal(p)
}

func (ctx *context) popInternal() stackable {
	if len(ctx.stack) == 0 {
		ctx.execError("Stack underflow", "")
		return nil
	}

	retval := ctx.stack[len(ctx.stack)-1]
	ctx.stack = ctx.stack[:len(ctx.stack)-1]
	return retval
}

func (ctx *context) popDatum() Datum {
	if d, ok := ctx.popInternal().(Datum); ok {
		return d
	}
	ctx.execError("Cannot unstack path operation element as datum.", "")
	return nil
}

func (ctx *context) popPathElem() pathElem {
	if p, ok := ctx.popInternal().(pathElem); ok {
		return p
	}
	ctx.execError("Cannot unstack datum element as path operation.", "")
	return nil
}

func (ctx *context) popNumber(desc string) float64 {
	return ctx.popDatum().Number(
		fmt.Sprintf("Failure to pop number (%s):", desc))
}

func (ctx *context) popBool(desc string) bool {
	return ctx.popDatum().Boolean(
		fmt.Sprintf("Failure to pop boolean (%s):", desc))
}

func (ctx *context) popNodeSet(desc string) []xutils.XpathNode {
	return ctx.popDatum().Nodeset(
		fmt.Sprintf("Failure to pop nodeset (%s):", desc))
}

type datumCompFn func(d1, d2 Datum) bool
type equalFn func(b bool) bool

// Common comparison logic for equality and relational operators when one
// or both operands is a nodeset.  Note that a nodeset may be empty, in
// which case the result of ANY comparison, even '!=', is FALSE.
func (ctx *context) compareNodesetsAndPush(
	boolCompare datumCompFn,
	litCompare datumCompFn,
	numCompare datumCompFn,
	operator string,
	op1, op2 Datum,
) {
	set1, set2 := []Datum{op1}, []Datum{op2}

	if isNodeset(op1) {
		if len(op1.(nodesetDatum).nodes) == 0 {
			ctx.pushDatum(NewBoolDatum(false))
			return
		}
		set1 = op1.(nodesetDatum).literalSlice()
	}

	if isNodeset(op2) {
		if len(op2.(nodesetDatum).nodes) == 0 {
			ctx.pushDatum(NewBoolDatum(false))
			return
		}
		set2 = op2.(nodesetDatum).literalSlice()
	}

	ctx.compareAndPushNodesets(
		set1, set2, boolCompare, litCompare, numCompare)
}

// For basic equality operators (ie '=' and '!='), we need to convert to
// a common type that can vary with the type of the operands.  (By
// comparison, relational operators (<, <=, >, >=) work only on numbers.)
//
// This function takes comparison functions and pops the 2 operands, works
// out what type comparison to do according to the rules in the XPATH spec
// (Section 3.4 Booleans) and pushes the result of that comparison.
//
// Nodeset case is called out to a separate function.  If neither operand
// is a nodeset, then both are converted to a single common type, based on
// the precedence of boolean wins over number wins over string.
//
func (ctx *context) popCompareEqualityAndPush(
	boolCompare datumCompFn,
	litCompare datumCompFn,
	numCompare datumCompFn,
	operator string,
) {
	op2 := ctx.popDatum()
	op1 := ctx.popDatum()

	op1IsNodeset := isNodeset(op1)
	op2IsNodeset := isNodeset(op2)

	switch {
	case op1IsNodeset || op2IsNodeset:
		ctx.compareNodesetsAndPush(boolCompare, litCompare, numCompare,
			operator, op1, op2)

	case isBool(op1) || isBool(op2):
		ctx.pushDatum(NewBoolDatum(boolCompare(op1, op2)))

	case isNum(op1) || isNum(op2):
		ctx.pushDatum(NewBoolDatum(numCompare(op1, op2)))

	case isLiteral(op1) || isLiteral(op2):
		// Given we only support 4 types, if we get here then both ought
		// to be Literals, but we keep the default just in case ...
		ctx.pushDatum(NewBoolDatum(litCompare(op1, op2)))

	default:
		ctx.execError(fmt.Sprintf("'%s' operator doesn't support '%s %s %s'",
			operator, op1.name(), operator, op2.name()), "")

	}
}

// Unlike for the equality operators where we need to handle each type
// separately, for relational operators we treat everything as a number.
// While nodesets appear to be handled differently, boolCompare and litCompare
// actually call numCompare under the covers after performing a type
// conversion.
func (ctx *context) popCompareRelationalAndPush(
	boolFn datumCompFn,
	litFn datumCompFn,
	numFn datumCompFn,
	operator string,
) {
	op2 := ctx.popDatum()
	op1 := ctx.popDatum()

	op1IsNodeset := isNodeset(op1)
	op2IsNodeset := isNodeset(op2)

	switch {
	case op1IsNodeset || op2IsNodeset:
		ctx.compareNodesetsAndPush(boolFn, litFn, numFn,
			operator, op1, op2)

	default:
		// Unlike equality operators ('=' and '!='), if neither operand is a
		// nodeset, then anything not a number is converted to a number and
		// the comparison operation is done on the 2 numbers.
		ctx.pushDatum(NewBoolDatum(numFn(op1, op2)))
	}
}

// Comparison (including relational operators) for a nodeset versus
// nodeset/number/literal/bool.
//
// If at least one element in the nodeset passes the relevant compare
// function then we push TRUE, otherwise we push FALSE.
//
func (ctx *context) compareWorker(ops1, ops2 []Datum, compareFn datumCompFn) {
	for _, op1 := range ops1 {
		for _, op2 := range ops2 {
			if compareFn(op1, op2) {
				ctx.pushDatum(NewBoolDatum(true))
				return
			}
		}
	}
	ctx.pushDatum(NewBoolDatum(false))
}

func (ctx *context) compareAndPushNodesets(
	ops1 []Datum,
	ops2 []Datum,
	boolCompare datumCompFn,
	litCompare datumCompFn,
	numCompare datumCompFn,
) {
	switch {
	case isNum(ops1[0]) || isNum(ops2[0]):
		ctx.compareWorker(ops1, ops2, numCompare)

	case isBool(ops1[0]) || isBool(ops2[0]):
		ctx.compareWorker(ops1, ops2, boolCompare)

	case isLiteral(ops1[0]) && isLiteral(ops2[0]):
		ctx.compareWorker(ops1, ops2, litCompare)

	default:
		panic(fmt.Sprintf("Cannot compare %s to %s",
			ops1[0].name(), ops2[0].name()))
	}
}

// validatePath - verify path in a must/when statement points to a valid node
//
// Check that the given path refers to a YANG node that could exist if
// configured.  If it does, add to the context's 'nonWarnings'.  If it does
// not, determine if the problem is a missing/wrong prefix (ie underlying
// path exists) or if the node really cannot exist ever.
//
// Additionally, if the path points to a non-presence container, we need to
// flag this, as such containers exist for validation even when they have
// no children, and this can be very confusing.  We therefore discourage their
// use by printing a warning, but it's non-fatal in case users do need this
// (in which case the warning will hopefully make them check it really is what
// they need!).
func (ctx *context) validatePath(
	pathElements []pathElem,
	refExpr string,
) bool {

	if len(pathElements) == 0 {
		ctx.execError("Cannot validate path with no elements", "")
		return false
	}
	if ctx.stackedNodesets != 0 {
		ctx.execError("Cannot validate path with stacked nodesets", "")
		return false
	}
	if len(pathElements) > 1 {
		if pathElements[len(pathElements)-1].baseString() == ":*" {
			pathElements = pathElements[:len(pathElements)-1]
		}
	}

	origDebug := ctx.debug
	origCtxB := ctx.b
	defer func() {
		ctx.debug = origDebug
		ctx.b = origCtxB
	}()
	ctx.debug = true
	ctx.b.Reset()

	ctx.addDebug(ctx.pfx + "----\n")
	ctx.formatAndAddDebug(
		"%sValidatePath:\t\tCtx: '%s'\n", ctx.pfx, ctx.node.XPath())

	return ctx.validatePathInternal(pathElements, refExpr)
}

func (ctx *context) validatePathInternal(
	pathElements []pathElem,
	refExpr string,
) bool {

	var startNodes = make([]xutils.XpathNode, 0)
	startNodes = append(startNodes, ctx.node)

	foundNodes := ctx.generateNodeSet(pathElements, startNodes,
		true /* match prefix */)
	debugOutput := ctx.b.String()

	searchPath := pathElemString(pathElements)
	retStatus := true
	if len(foundNodes) == 0 {
		warn := xutils.DoesntExist
		ctx.b.Reset()
		foundNodes := ctx.generateNodeSet(pathElements, startNodes,
			false /* retry, don't match prefix */)
		debugOutputNoPrefix := ctx.b.String()
		if len(foundNodes) != 0 {
			debugOutput = debugOutput +
				"\n\nIf we ignore prefixes, we now get:\n\n" +
				debugOutputNoPrefix
			warn = xutils.MissingOrWrongPrefix
		}
		ctx.res.warnings = append(ctx.res.warnings,
			xutils.NewWarning(
				warn, ctx.node.XPath().String(), refExpr, ctx.xpathStmtLoc,
				searchPath, debugOutput))

		retStatus = false
	} else {
		ctx.res.nonWarnings = append(ctx.res.nonWarnings,
			xutils.NewWarning(
				xutils.ValidPath, ctx.node.XPath().String(), refExpr,
				ctx.xpathStmtLoc, searchPath, ""))
	}

	if len(foundNodes) != 0 {
		for _, node := range foundNodes {
			if node.XIsNonPresCont() {
				ctx.res.warnings = append(ctx.res.warnings,
					xutils.NewWarning(
						xutils.RefNPContainer, ctx.node.XPath().String(),
						refExpr, ctx.xpathStmtLoc, searchPath, ""))
			}
		}
	}

	return retStatus
}

func (ctx *context) createNodeSet(pathElements []pathElem) []xutils.XpathNode {

	var startNodes = make([]xutils.XpathNode, 0)

	if (len(pathElements) == 0) && (ctx.stackedNodesets == 0) {
		ctx.execError("Cannot create nodeset without a path.", "")
		return startNodes
	}

	// Set our initial start point, the current node.
	ctx.addDebug(ctx.pfx + "----\n")

	if ctx.stackedNodesets > 0 {
		ctx.addDebug(ctx.pfx + "CreateNodeSet:\t\tUsing stacked nodeset:\n")
		ctx.stackedNodesets--
		startNodeset := ctx.popNodeSet("Stacked nodeset")
		ctx.addDebugNodeset(startNodeset)
		startNodes = append(startNodes, startNodeset...)
	} else {
		if ctx.debug {
			ctx.formatAndAddDebug(
				"%sCreateNodeSet:\t\tCtx: '%s'\n", ctx.pfx, ctx.node.XPath())
		}
		startNodes = append(startNodes, ctx.node)
	}

	return ctx.generateNodeSet(pathElements, startNodes, true /* match pfx */)
}

func (ctx *context) generateNodeSet(
	pathElements []pathElem,
	nodesToEval []xutils.XpathNode,
	matchPrefix bool,
) []xutils.XpathNode {

	var tmpEvalNodes = make([]xutils.XpathNode, 0)

	for _, pathElem := range pathElements {
		ctx.formatAndAddDebug("%s\tApply: %s\n", ctx.pfx, pathElem)
		for _, evalNode := range nodesToEval {
			// Each node may disappear, be replaced, or become multiple
			// nodes.
			newNodes, errStr := pathElem.applyToNode(
				evalNode, matchPrefix, ctx.filter)
			if errStr != "" {
				ctx.execError(errStr, "")
			}
			tmpEvalNodes = append(tmpEvalNodes, newNodes...)
		}
		if ctx.debug {
			ns := NewNodesetDatum(tmpEvalNodes)
			ctx.addDebug(ns.(nodesetDatum).Print(ctx.pfx))
		}
		nodesToEval = tmpEvalNodes
		tmpEvalNodes = []xutils.XpathNode{}

		// If we have navigated down a tree then back up, we may have a
		// scenario where 2 nodes have the same parent.  We could remove
		// duplicates here, but it does no harm to carry them to the
		// end, and it's only realistic for a set of nodes to multiply if
		// we have a test case that goes up and down the tree for no good
		// reason other than to exercise this corner case!
	}

	// Ensure final nodeset is unique.
	return xutils.RemoveDuplicateNodes(nodesToEval)
}

// Run - run the machine in the given context, feeding errors back to caller.
//
// Any panics while running the machine are caught and the error fed back
// in the result to the caller.
func (ctx *context) Run() (res *Result) {

	defer func() {
		if r := recover(); r != nil {
			ctx.res.runErr = fmt.Errorf("%s", r)
			res = ctx.res
		}
		ctx.saveDebug()
		if ctx.level == 0 {
			ctx.logDebug()
		}
	}()

	ctx.addDebugProgHeader(ctx.refExpr)

	for _, instr := range ctx.prog {
		ctx.addDebugInstrAndStack(instr.fnName)
		instr.fn(ctx)
		ctx.addDebug(ctx.pfx + "----\n")
	}

	return ctx.res
}
