// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains the 'machine' object used to construct 'machines'
// which consist of a set of instructions (the 'inst' type).

package xpath

import (
	"fmt"

	"github.com/sdcio/yang-parser/xpath/xutils"
)

// RESULT
//
// Wrapper around the raw result of the XPath expression, so we can keep it
// in its native type, but convert on request to other types.
type Result struct {
	value       Datum
	runErr      error            // Error when running machine
	output      string           // Debug output showing stack and instructions.
	warnings    []xutils.Warning // Warnings from running pathEval machines.
	nonWarnings []xutils.Warning // Valid paths from running pathEval machines.
}

func NewResult() *Result {
	return &Result{}
}

func (res *Result) save(d Datum) {
	res.value = d
}

func (res *Result) PrintResult() string {
	if res.runErr != nil {
		return fmt.Sprintf("Failed to run: %s\n", res.runErr.Error())
	}

	switch res.value.(type) {
	case boolDatum:
		if val, err := res.GetBoolResult(); err == nil {
			return fmt.Sprintf("BOOLEAN:\t%t\n", val)
		}
	case numDatum:
		if val, err := res.GetNumResult(); err == nil {
			return fmt.Sprintf("NUMBER:\t%v\n", val)
		}
	case litDatum:
		if lit, err := res.GetLiteralResult(); err == nil {
			return fmt.Sprintf("LITERAL:\t%s\n", lit)
		}
	case nodesetDatum:
		if ns, err := res.GetNodeSetResult(); err == nil {
			return fmt.Sprintf("NODESET: %v\n", ns)
		}
	}
	return "Unable to print result!"
}

func (res *Result) GetDebugOutput() string {
	return res.output
}

func (res *Result) GetNumResult() (float64, error) {
	if res.runErr != nil {
		return 0, res.runErr
	}

	if res.value == nil {
		return 0, fmt.Errorf("No result to return for number.")
	}

	return res.value.Number("GetNumResult"), nil
}

func (res *Result) IsNumber() bool { return isNum(res.value) }

func (res *Result) GetBoolResult() (bool, error) {
	if res.runErr != nil {
		return false, res.runErr
	}

	if res.value == nil {
		return false, fmt.Errorf("No result to return for boolean.")
	}

	return res.value.Boolean("GetBoolResult"), nil
}

func (res *Result) GetLiteralResult() (string, error) {
	if res.runErr != nil {
		return "", res.runErr
	}

	if res.value == nil {
		return "", fmt.Errorf("No result to return for literal.")
	}

	return res.value.Literal("GetLiteralResult"), nil
}

func (res *Result) GetNodeSetResult() ([]xutils.XpathNode, error) {
	if res.runErr != nil {
		return []xutils.XpathNode{}, res.runErr
	}

	if res.value == nil {
		return nil, fmt.Errorf("No result to return for nodeset.")
	}

	return res.value.Nodeset("GetNodesetResult"), nil
}

func (res *Result) GetError() error {
	return res.runErr
}

func (res *Result) GetWarnings() []xutils.Warning {
	return res.warnings
}

func (res *Result) GetNonWarnings() []xutils.Warning {
	return res.nonWarnings
}

// MACHINE
//
// Object used to encapsulate execution of an expression.
// Ideally all functions above would be methods on machine, but until / unless
// we can get Go's YACC implementation to allow parameters to be passed into
// exprParse, it's hard and we have to rely on a single machine running at one
// time then store the results out of the parser into the machine.
type PfxMapFn func(prefix string) (namespace string, err error)

type Machine struct {
	refExpr  string // For reference - our expression being evaluated.
	location string // Module:line XPATH expression is defined
	name     string // For debug only
	prog     []Inst // Actual set of operands / operations to run
}

func NewMachine(expr string, prog []Inst, name string) *Machine {
	return &Machine{refExpr: expr, prog: prog, name: name}
}

func NewMachineWithLocation(
	expr, location string,
	prog []Inst,
	name string,
) *Machine {
	return &Machine{refExpr: expr, location: location, prog: prog, name: name}
}

func (mach *Machine) GetExpr() string     { return mach.refExpr }
func (mach *Machine) GetLocation() string { return mach.location }

// Functions for executing the machine and managing it at a high level.

// Wrapper around a machine that will return the set of values.  If empty,
// must return []string{} not nil or TmplGetAllowed will barf with the likes
// of 'wrong return type for TmplGetAllowed got <nil> expecting []string'.
func (mach *Machine) AllowedValues(
	ctxNode xutils.XpathNode,
	debug bool,
) ([]string, error) {
	var err error
	allowedValues := []string{}

	res := NewCtxFromMach(mach, ctxNode).
		SetDebug(debug).
		AccessibleTreeConfigOnly().
		Run()
	nodeset, err := res.GetNodeSetResult()
	if err != nil {
		return nil, err
	}

	for _, node := range nodeset {
		allowedValues = append(allowedValues, node.XValue())
	}
	return allowedValues, nil
}

// Useful for debugging machines when they don't quite work as expected!
func (mach *Machine) PrintMachine() string {
	if mach == nil {
		return "No machine to print!"
	}

	return GetProgListing(mach.prog, 0)
}
