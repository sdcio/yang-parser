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
