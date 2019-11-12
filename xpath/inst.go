// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This machine implements the 'Inst' (instruction) object which represents
// an XPATH instruction that can be run as part of an XPATH machine.

package xpath

// INST
//
// Set of instructions created for later execution.
type instFunc func(*context)

type Inst struct {
	fn         instFunc
	fnName     string // for debug mostly
	subMachine string // debug string for sub-machine, if present.
	count      int
}

func newInst(fn instFunc, fnName string) Inst {
	return Inst{fn: fn, fnName: fnName}
}

func newInstWithSubMachine(fn instFunc, fnName, subMachine string) Inst {
	return Inst{fn: fn, fnName: fnName, subMachine: subMachine}
}

func (i Inst) String() string {
	return i.fnName
}
