// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This test verifies that the debug dump of a machine is correct, and that
// the dump of a machine when run is also correct.

package leafref

import (
	"testing"

	. "github.com/danos/yang/xpath/xpathtest"
	"github.com/danos/yang/xpath/xutils"
)

// Check all valid options in a machine are printed correctly.
// Ensures this function keeps working in case it is needed for debug!
func TestMachineInstructionPrint(t *testing.T) {
	testMachine, _ := NewLeafrefMachine(
		"/interfaces/interface[name = current()/../ifname]/address/ip",
		nil)

	machineString := testMachine.PrintMachine()

	expectedString :=
		"--- machine start ---\n" +
			"pathOperPush	/ (2f)\n" +
			"nameTestPush	{ interfaces}\n" +
			"nameTestPush	{ interface}\n" +
			"lrefPredStart\n" +
			"nameTestPush	{ name}\n" +
			"lrefEquals\n" +
			"pathOperPush	. (2e)\n" +
			"pathOperPush	..\n" +
			"nameTestPush	{ ifname}\n" +
			"lrefPredEnd\n" +
			"nameTestPush	{ address}\n" +
			"nameTestPush	{ ip}\n" +
			"evalLocPath\n" +
			"store\n" +
			"---- machine end ----\n"

	if machineString != expectedString {
		t.Errorf("Expected:\n%s\n---\nGot:\n%s\n---\n",
			expectedString, machineString)
	}
}

// Definitions local to this file, but common enough to extract.
const Int2 = "/interfaces/interface"

func TestMachineExecutionPrint(t *testing.T) {
	expOut :=
		"Run\t'" + Int2 + "[name = " +
			"current()/../ifname]/address/ip' on:\n" +
			"\t/interfaces/default-address/address (6666)\n" +
			Brk +
			InstPoPsh + "/ (2f)\n" +
			Stack + "(empty)\n" +
			Brk +
			InstNtPsh_B + ModName + " interfaces}\n" +
			StPO + "/ (2f)\n" +
			Brk +
			InstNtPsh_B + ModName + " interface}\n" +
			StNT_B + ModName + " interfaces}\n" +
			T_PO_T + "/ (2f)\n" +
			Brk +
			InstLrefPS + "\n" +
			StNT_B + ModName + " interface}\n" +
			T_NT_TB + ModName + " interfaces}\n" +
			T_PO_T + "/ (2f)\n" +
			Brk +
			CrtNS + "Ctx: '/interfaces/default-address/address'\n" +
			T_ApPO + "/ (2f)\n" +
			Tab3 + "(root)\n" +
			T_ApNT_B + ModName + " interfaces}\n" +
			Tab3 + "/interfaces\n" +
			T_ApNT_B + ModName + " interface}\n" +
			Tab3 + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			InstNtPsh_B + ModName + " name}\n" +
			StNS + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			InstLrefEq + "\n" +
			StNT_B + ModName + " name}\n" +
			T_NS_T + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			InstPoPsh + ". (2e)\n" +
			StNT_B + ModName + " name}\n" +
			T_NS_T + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			InstPoPsh + "..\n" +
			StPO + ". (2e)\n" +
			T_NT_TB + ModName + " name}\n" +
			T_NS_T + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			InstNtPsh_B + ModName + " ifname}\n" +
			StPO + "..\n" +
			T_PO_T + ". (2e)\n" +
			T_NT_TB + ModName + " name}\n" +
			T_NS_T + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			InstLrefPE + "\n" +
			StNT_B + ModName + " ifname}\n" +
			T_PO_T + "..\n" +
			T_PO_T + ". (2e)\n" +
			T_NT_TB + ModName + " name}\n" +
			T_NS_T + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			CrtNS + "Ctx: '/interfaces/default-address/address'\n" +
			T_ApPO + ". (2e)\n" +
			Tab3 + "/interfaces/default-address/address (6666)\n" +
			T_ApPO + "..\n" +
			Tab3 + "/interfaces/default-address\n" +
			T_ApNT_B + ModName + " ifname}\n" +
			Tab3 + "/interfaces/default-address/ifname (dp0s2)\n" +
			Brk +
			FiltNS + "[{" + ModName + " name} = dp0s2]\n" +
			Tab3 + Int2 + "[name='dp0s1']\n" +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			Tab3 + Int2 + "[name='s1']\n" +
			Tab3 + Int2 + "[name='lo2']\n" +
			Brk +
			InstNtPsh_B + ModName + " address}\n" +
			StNS + Int2 + "[name='dp0s2']\n" +
			Brk +
			InstNtPsh_B + ModName + " ip}\n" +
			StNT_B + ModName + " address}\n" +
			T_NS_T + Int2 + "[name='dp0s2']\n" +
			Brk +
			InstELP + "\n" +
			StNT_B + ModName + " ip}\n" +
			T_NT_TB + ModName + " address}\n" +
			T_NS_T + Int2 + "[name='dp0s2']\n" +
			Brk +
			CrtNS + T_UNS +
			Tab3 + Int2 + "[name='dp0s2']\n" +
			T_ApNT_B + ModName + " address}\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='2111']\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='2222']\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='3333']\n" +
			T_ApNT_B + ModName + " ip}\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='2111']/ip (2111)\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='2222']/ip (2222)\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='3333']/ip (3333)\n" +
			Brk +
			InstStore + "\n" +
			StNS + Int2 + "[name='dp0s2']/address[ip='2111']/ip (2111)\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='2222']/ip (2222)\n" +
			Tab3 + Int2 + "[name='dp0s2']/address[ip='3333']/ip (3333)\n" +
			Brk

	checkLeafrefNodeSetResultWithDebug(
		t, "/interfaces/interface[name = current()/../ifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		TNodeSet{
			NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2111"),
			NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2222"),
			NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "3333")},
		expOut)

}
