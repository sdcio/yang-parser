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
// Copyright (c) 2016,2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package xpathtest

import (
	"strings"
)

// This file contains strings used in printing out the operation of a
// machine.  The definitions are separate to those in production code
// so if the format changes, we are made aware of it.  Otherwise it is
// possible that 2 strings could get swapped in production code and we
// would not spot it in test code.
//
// Extracting these strings has the benefit that they appear
// in black in emacs versus the brown for 'text in quotes' so you can see
// the details that change in each line more clearly.  More to the point,
// all the generic stuff can now be easily changed in a single place.
//
// In some cases, leading or trailing symbols are explicitly stated so it
// is easier to construct the expOut strings.
//
// T = Tab
// B = Brace ('{')
//
const (
	Brk         = "----\n"
	CrtNS       = "CreateNodeSet:\t\t"
	FiltNS      = "FilterNodeSet:\t\t"
	IgnPfxs     = "\n\nIf we ignore prefixes, we now get:\n\n"
	InstELP     = "Instr:\tevalLocPath"
	InstELPPS   = "Instr:\tevalLocPath(PredStart)"
	InstESM     = "Instr:\tevalSubMachine"
	InstLrefEq  = "Instr:\tlrefEquals"
	InstLrefPE  = "Instr:\tlrefPredEnd"
	InstLrefPS  = "Instr:\tlrefPredStart"
	InstNtPsh_B = "Instr:\tnameTestPush\t{"
	InstNumPsh  = "Instr:\tnumpush\t\t"
	InstPoPsh   = "Instr:\tpathOperPush\t"
	InstStore   = "Instr:\tstore"
	ModName     = "xpathNodeTestModule"
	PredMatch   = "Pred:\tMATCH\n"
	PredNoMatch = "Pred:\tNo Match\n"
	Run         = "Run\t"
	Stack       = "Stack:\t"
	StNS        = "Stack:\tNODESET\t\t"
	StNT_B      = "Stack:\tNAMETEST\t{"
	StNum       = "Stack:\tNUMBER\t\t"
	StPO        = "Stack:\tPATHOPER\t"
	Tab3        = "\t\t\t"
	T_ApNS      = "\tApply: NODESET\t\t"
	T_ApNT_B    = "\tApply: NAMETEST\t{"
	T_ApPO      = "\tApply: PATHOPER\t"
	T_PO_T      = "\tPATHOPER\t"
	T_NS_T      = "\tNODESET\t\t"
	T_NT_TB     = "\tNAMETEST\t{"
	T_UNS       = "Using stacked nodeset:\n"
	ValPath     = "ValidatePath:\t\tCtx: "
)

func Indent(input string) string {
	lines := strings.Split(input, "\n")
	var output string
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		output = output + "\t" + line + "\n"
	}
	return output
}
