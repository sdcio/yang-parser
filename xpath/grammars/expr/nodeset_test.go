// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// These tests verify nodeset functionality for XPATH, using the
// test-only XpathTestNode object to drive the XPATH code that
// operates on nodesets.

package expr

import (
	"fmt"
	"testing"

	. "github.com/iptecharch/yang-parser/xpath/xpathtest"
	"github.com/iptecharch/yang-parser/xpath/xutils"
)

// Make sure our test code is creating a vaguely sane tree!
func TestNodesetCreateTree(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "dataplane/name+dp0s2", "address@4321"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
			{"protocols", "mpls", "max-label+1000000"},
		})

	if err := xutils.ValidateTree(configTree); err != nil {
		t.Fatalf("ValidateTree failed: %s", err.Error())
	}
}

func TestParseEqualityForEmptyNodesets(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})

	// Empty nodeset comparison result always false
	checkBoolResultWithContext(t, "serial = cereal", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal = serial", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal != serial", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal = cereal", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal != cereal", false,
		configTree, xutils.PathType([]string{"/", "interface"}))

	// ... and as a double check where nodesets are not empty:
	checkBoolResultWithContext(t, "serial = serial", true,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "serial != serial", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
}

func TestParseEqualityForNodesets(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})

	// Same must match same!
	checkBoolResultWithContext(t, ".. = ..", true,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, ".. != ..", false,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))

	// Same must match same!
	checkBoolResultWithContext(t, ". = .", true,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, ". != .", false,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))

	// Nodes with value 2111 on both sides. = then !=
	checkBoolResultWithContext(t, "*/address = ../protocols/mpls/min-label",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "*/address != ../protocols/mpls/min-label",
		true, configTree, xutils.PathType([]string{"/", "interface"}))

	// address only gives dp0s1 node as we started at dataplane which means
	// we will have picked first dataplane node (dp0s1).
	checkBoolResultWithContext(t, "address != ../../protocols/mpls/min-label",
		true, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))

	// No match.
	checkBoolResultWithContext(t,
		"../serial/address = ../../protocols/mpls/min-label", false,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t,
		"../serial/address != ../../protocols/mpls/min-label", true,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
}

func TestParseEqualityNodesetVsNumber(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})

	checkBoolResultWithContext(t, "address = 2222",
		false, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, "2222 = */address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))

	// Yes, '2222 = */address' AND '2222 != */address' are both true as there
	// is at least one node in the nodeset for which the comparison operation
	// returns true.
	checkBoolResultWithContext(t, "2222 != */address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))

}

func TestParseEqualityNodesetVsLiteral(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})

	checkBoolResultWithContext(t, "../*/address = '5555'",
		true, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, "'5555' = ../*/address",
		true, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, "../*/address = '6666'",
		false, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))

	checkBoolResultWithContext(t, "../*/address != '5555'",
		true, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, "../serial/address != '5555'",
		false, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
}

func TestParseEqualityNodesetVsBoolean(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})

	checkBoolResultWithContext(t, "../*/address = true()",
		true, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, "../*/address = false()",
		false, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))

	// Important to check reverse order (nodeset second)
	checkBoolResultWithContext(t, "true() != ../*/address",
		false, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
	checkBoolResultWithContext(t, "false() != ../*/address",
		true, configTree, xutils.PathType([]string{"/", "interface", "dataplane"}))
}

func TestParseComparisonForEmptyNodesets(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})

	// Empty nodeset never equal to another (empty or not empty) nodeset
	checkBoolResultWithContext(t, "serial < cereal", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal > serial", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal >= serial", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal <= cereal", false,
		configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "cereal < cereal", false,
		configTree, xutils.PathType([]string{"/", "interface"}))

}

func TestParseComparisonNodesets(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s1", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@6666"},
			{"interface", "serial/name+s1", "address@111"},
			{"interface", "serial/name+s1", "address@1111"},
			{"interface", "serial/name+s1", "address@5555"},
			// At time of writing, predicate support didn't exist, and
			// in any case, this is not a test of predicates.  So, to
			// be able to get some single node nodesets, we'll have some
			// custom interface types
			{"interface", "loopback0/name+lo0", "address@10"},
			{"interface", "loopback1/name+lo1", "address@6666"},
			{"interface", "loopback2/name+lo2", "address@1111"},
		})

	// Range of values against nodeset with single or multiple values,
	// pass and fail case for each operator.  {} indicates range of values
	// being compared on each side of the operator, with *x* indicating that
	// a value meets the condition when not all values in {} meet it.

	// None of {1111, 2111, 6666} <= {10}
	checkBoolResultWithContext(t, "dataplane/address <= loopback0/address",
		false, configTree, xutils.PathType([]string{"/", "interface"}))
	// {*1111*, 2111, 6666} <= {1111}
	checkBoolResultWithContext(t, "dataplane/address <= loopback2/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))

	// None of {1111, 2111, 6666} < {10}
	checkBoolResultWithContext(t, "dataplane/address < loopback0/address",
		false, configTree, xutils.PathType([]string{"/", "interface"}))
	// {*1111*, *2111*, 6666} < {111, 1111, *5555*}
	checkBoolResultWithContext(t, "loopback1/address < dataplane/address",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	// All of {1111, 2111, 6666} > {10}
	checkBoolResultWithContext(t, "dataplane/address > loopback0/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	// None of {1111, 2111, 6666} > {6666}
	checkBoolResultWithContext(t, "dataplane/address > loopback1/address",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	// {1111, 2111, *6666*} >= {6666}
	checkBoolResultWithContext(t, "dataplane/address >= loopback1/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	// {10} is not >= {1111, 2111, 6666}
	checkBoolResultWithContext(t, "loopback0/address >= dataplane/address",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	// Range of values against range of values.  So long as one comparison
	// passes, result is true.  So, comparing the same 2 nodesets here gives
	// true in all 4 cases (-:
	checkBoolResultWithContext(t, "dataplane/address > serial/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address >= serial/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address < serial/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address <= serial/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))

	// Finally, let's do a comparison on the 'string-value' when this is
	// made up of a set of terminal node values.
	//
	// Compares 'dp0s111112111' > 's111111115555'
	//      and 'dp0s26666' > 's111111115555'
	checkBoolResultWithContext(t, "dataplane > serial",
		false, configTree, xutils.PathType([]string{"/", "interface"}))
}

func TestParseComparisonNodesetVsNumber(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s1", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@6666"},
		})

	checkBoolResultWithContext(t, "dataplane/address > 1234",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address > 6666",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "dataplane/address < 1234",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "6666 < dataplane/address",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "dataplane/address >= 6666",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address >= 6667",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "6666 <= dataplane/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address <= 1110",
		false, configTree, xutils.PathType([]string{"/", "interface"}))
}

func TestParseComparisonNodesetVsLiteral(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s1", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@6666"},
		})

	checkBoolResultWithContext(t, "'6667' > dataplane/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address > '9999'",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "dataplane/address >= '6666'",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address >= '6667'",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "dataplane/address < '1112'",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address < '1111'",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "dataplane/address <= '1111'",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address <= '1110'",
		false, configTree, xutils.PathType([]string{"/", "interface"}))
}

func TestParseComparisonNodesetVsBoolean(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1", "address@1111"},
		})

	checkBoolResultWithContext(t, "dataplane/address > false()",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address > true()",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "false() >= dataplane/address",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "false() < dataplane/address",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address < true()",
		false, configTree, xutils.PathType([]string{"/", "interface"}))

	checkBoolResultWithContext(t, "dataplane/address <= true()",
		true, configTree, xutils.PathType([]string{"/", "interface"}))
	checkBoolResultWithContext(t, "dataplane/address <= false()",
		false, configTree, xutils.PathType([]string{"/", "interface"}))
}

// Test starting with root node
func TestRootPath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
		})

	// Root container
	checkNodeSetResult(t, "/", nil,
		configTree, xutils.PathType([]string{"/"}),
		TNodeSet{
			NewTContainer(nil, xutils.PathType([]string{"/"}), "", "root"),
		})
}

// Try to go above root - ensure it fails correctly.
func TestInvalidPath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "dataplane/name+dp0s2", "address@2333"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
		})

	// Root container
	checkNodeSetResult(t, "../../interface", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{})
}

// Check we can correctly find all expected nodes for container, leaf,
// leaf-lists and lists.
//
// In all of these, we use absolute not relative path, so start path is
// irrelevant.
func TestContainerPath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "dataplane/name+dp0s2", "address@2333"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
		})

	// Root container
	checkNodeSetResult(t, "/", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTContainer(nil, xutils.PathType([]string{"/"}), "", "root"),
		})

	// Container, one level down
	checkNodeSetResult(t, "/interface", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTContainer(
				nil, xutils.PathType([]string{"/", "interface"}), "", "interface")})

	// Container as '.' (current)
	checkNodeSetResult(t, ".", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTContainer(nil, xutils.PathType([]string{"/", "interface"}),
				"", "interface")})
}

func TestListPath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s2"},
			{"interface", "loopback/name+lo2"},
			{"interface", "serial/name+s1"},
		})

	// List with one entry
	checkNodeSetResult(t, "/interface/loopback", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "loopback"}),
				"", "loopback",
				"name", "lo2")})

	// List with 2 entries
	checkNodeSetResult(t, "/interface/dataplane", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")})

	// Wildcard list access
	checkNodeSetResult(t, "/interface/*", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane",
				"name", "dp0s1"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane",
				"name", "dp0s2"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "loopback"}),
				"", "loopback",
				"name", "lo2"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "serial"}),
				"", "serial",
				"name", "s1")})

	// Cannot directly reference list members.
	checkNodeSetResult(t, "/interface/dataplane/dp0s1", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{})
}

func TestLeafListPath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
		})

	checkNodeSetResult(t, "/interface/*/address", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1111"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2111"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2222"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "serial", "address"}),
				"", "address", "5555")})
}

func TestLeafPath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"protocols", "mpls", "min-label+16"},
			{"protocols", "mpls", "max-label+1000000"},
		})

	checkNodeSetResult(t, "/protocols/mpls/min-label", nil,
		configTree, xutils.PathType([]string{"/", "protocols"}),
		TNodeSet{
			NewTLeaf(
				nil, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}),
				"", "min-label", "16")})

	checkNodeSetResult(t, "/protocols/mpls/*", nil,
		configTree, xutils.PathType([]string{"/", "protocols"}),
		TNodeSet{
			NewTLeaf(
				nil, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}),
				"", "min-label", "16"),
			NewTLeaf(
				nil, xutils.PathType([]string{"/", "protocols", "mpls", "max-label"}),
				"", "max-label", "1000000")})
}

func TestEmptyLeafPath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"protocols", "mpls", "debug%"},
		})

	checkNodeSetResult(t, "/protocols/mpls/debug", nil,
		configTree, xutils.PathType([]string{"/", "protocols"}),
		TNodeSet{
			NewTEmptyLeaf(
				nil, xutils.PathType([]string{"/", "protocols", "mpls", "debug"}),
				"", "debug")})
}

func TestRelativePath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "dataplane/name+dp0s2", "address@2333"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
			{"protocols", "mpls", "max-label+1000000"},
		})

	// Simple reference
	checkNodeSetResult(t, "serial", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "serial"}),
				"", "serial", "name", "s1")})

	// '.'
	checkNodeSetResult(t, "./dataplane", nil,
		configTree, xutils.PathType([]string{"/", "interface"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane",
				"name", "dp0s1"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")})

	// '..'
	checkNodeSetResult(t, "../serial", nil,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "serial"}),
				"", "serial", "name", "s1")})

	// Roller-coaster ... checks that when we go up to a parent, we
	// need to weed out duplicate nodes.
	checkNodeSetResult(t, "../dataplane/../serial/./../loopback", nil,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "loopback"}),
				"", "loopback", "name", "lo2")})
}

func TestRelativePathErrors(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "loopback/name+lo2"},
			{"interface", "serial/name+s1"},
			{"protocols", "mpls", "min-label+16"},
		})

	// Not accessible from here ...
	checkNodeSetResult(t, "dataplane", nil,
		configTree, xutils.PathType([]string{"interface", "dataplane"}),
		TNodeSet{})

	// This won't return anything as dp0s1 is a leaf value here.
	checkNodeSetResult(t, "../dataplane/dp0s1", nil,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}),
		TNodeSet{})

	// This won't parse, as names cannot begin with numbers.
	expErrMsgs := []string{
		"Failed to compile '/protocols/mpls/min-label/16'",
		"Parse Error: syntax error",
		"Got to approx [X] in '/protocols/mpls/min-label/16 [X] '"}
	checkParseError(t, "/protocols/mpls/min-label/16", expErrMsgs)
}

// Check the removeDuplicateNodes function correctly identifies identical
// nodes.
func TestDeduplicationDirect(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "dataplane/name+dp0s3", "address@1234"},
			{"interface", "loopback/name+lo2"},
			{"interface", "serial/name+s1"},
			{"protocols", "mpls", "min-label+16"},
		})

	// Nodes 1a and 2a are identical, matching the interface container
	node1a := configTree.FindFirstNode(xutils.PathType([]string{"/", "interface"}))
	node1b := configTree.FindFirstNode(xutils.PathType([]string{"/", "interface"}))

	// Node 2 is the sole serial list entry.
	node2 := configTree.FindFirstNode(
		xutils.PathType([]string{"/", "interface", "serial"}))

	// Check parent and child generate different unique IDs when de-duped.
	dedupeNodes := xutils.RemoveDuplicateNodes(
		[]xutils.XpathNode{node1a, node2})
	if len(dedupeNodes) != 2 {
		t.Fatalf("De-duplication failed (parent vs child).\n")
	}

	// Check identical nodes match
	dedupeNodes = xutils.RemoveDuplicateNodes(
		[]xutils.XpathNode{node1a, node1b})
	if len(dedupeNodes) != 1 {
		t.Fatalf("De-duplication failed (same nodes).\n")
	}

}

// Check the remove duplicate functionality at a higher level by using an
// XPath expression to generate a set of duplicate nodes.
func TestDeduplicationWithXpath(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1", "address@1234"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "loopback/name+lo2"},
			{"interface", "serial/name+s1"},
			{"protocols", "mpls", "min-label+16"},
		})

	// This generates 256 nodes before they are correctly reduced back down
	// to 4 (-:
	checkNodeSetResult(t, "*/../*/../*/../*", nil,
		configTree, xutils.PathType([]string{"interface"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "loopback"}),
				"", "loopback", "name", "lo2"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "serial"}),
				"", "serial", "name", "s1")})

	// Both dataplane interface address nodes have same value and same path
	// as list entry paths don't include keys.
	checkNodeSetResult(t, "*/address", nil,
		configTree, xutils.PathType([]string{"interface"}),
		TNodeSet{
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1234"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1234")})
}

func TestPathUnion(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
		})

	// Simple union
	checkNodeSetResult(t, "../../interface/serial | "+
		"../../interface/dataplane", nil,
		configTree, xutils.PathType([]string{"/", "protocols", "mpls"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "serial"}),
				"", "serial", "name", "s1"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")})

	// Combine nodes at different levels of the tree with 2 '|' statements
	checkNodeSetResult(t, "../../interface/serial | "+
		"../../interface/dataplane/address | min-label", nil,
		configTree, xutils.PathType([]string{"/", "protocols", "mpls"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "serial"}),
				"", "serial", "name", "s1"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1234"),
			NewTLeaf(
				nil, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}),
				"", "min-label", "16")})
}

func nodesetPfxMapFn(prefix string) (module string, err error) {
	switch prefix {
	case "xpath", "":
		return TestModule, nil
	default:
		return "",
			fmt.Errorf("Unable to locate module for prefix '%s'", prefix)
	}
}

// Check qualified names (good and bad case)
func TestPathQualifiedNames(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s2", "address@1234"},
			{"interface", "serial/name+s1"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+16"},
		})

	// xpath will map to TestModule, as will blank prefix
	checkNodeSetResult(
		t, "../xpath:interface/dataplane/xpath:address", nodesetPfxMapFn,
		configTree, xutils.PathType([]string{"protocols"}),
		TNodeSet{
			NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interface", "dataplane", "address"}),
				"", "address", "1234")})

	// Now check we get lexer error with other-pfx.
	expErrMsgs := []string{
		"Parse Error: syntax error",
		"approx [X] in '../xpath:interface/dataplane/other-pfx:address [X] '",
		"Lexer Error: Unable to locate module for prefix 'other-pfx'"}
	checkParseErrorWithMap(
		t, "../xpath:interface/dataplane/other-pfx:address",
		expErrMsgs, nodesetPfxMapFn)
}

// Filter expressions can be simple (function returning nodeset) or compound,
// which is a basic expression followed by a relative path.
func TestPathFilterExpr(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label", "16"},
		})

	// Basic filter expression
	checkNodeSetResult(t, "current()",
		nil, configTree, xutils.PathType([]string{"interface", "dataplane"}),
		TNodeSet{
			NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane",
				"name", "dp0s1")})

	// Compound filter expression
	checkNodeSetResult(t, "current()/*/address",
		nil, configTree, xutils.PathType([]string{"interface"}),
		TNodeSet{
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1111"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2111"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2222"),
			NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "serial", "address"}),
				"", "address", "5555")})
}

// Proof of concept for VDR team to demonstrate ability to dynamically control
// number of list entries based on a separate 'max values' element in YANG.
func TestPathVplane(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"controller", "max-vplanes+2"},
			{"controller", "vplanes/name+vplane1"},
			{"controller", "vplanes/name+vplane2"},
			{"controller", "vplanes/name+vplane3"},
		})

	// First let's verify we get the expected values for the constituent
	// parts of our putative 'must' statement
	checkNumResultWithContext(t, "count(../vplanes)", 3,
		configTree, xutils.PathType([]string{"/", "controller", "vplanes"}))
	checkNumResultWithContext(t, "../max-vplanes", 2,
		configTree, xutils.PathType([]string{"/", "controller", "vplanes"}))

	// OK: let's check that "must '...'" fails as we have too many list entries
	checkBoolResultWithContext(t, "count(../vplanes) < ../max-vplanes", false,
		configTree, xutils.PathType([]string{"/", "controller", "vplanes"}))
}

func TestNodesetOrder(t *testing.T) {
	t.Skipf("Nodeset ordering")

	// Order: user vs system (list / leaf-list, and leaves separately)
	// ... and just generally getting a consistent document order.
}

func TestDoubleSlashPath(t *testing.T) {
	t.Skipf("Double slash path")
	//checkNodeSetResult(t, "/foo//bah/humbug", TNodeSet{}) // Absolute
	//checkNodeSetResult(t, "foo/bar//banana", TNodeSet{})  // Relative
}

const Nm_S1 = "[name='s1']"
const Nm_D1_CR = "[name='dp0s1']\n"
const Nm_D3_CR = "[name='dp0s3']\n"

func TestStackPrint(t *testing.T) {
	configTree := CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})

	expOut := "Run\t'substring('3456',1,2) + ../serial/name' on:\n" +
		"\t/interface/dataplane" + Nm_D1_CR +
		"----\n" +
		"Instr:	litpush		'3456'\n" +
		"Stack:	(empty)\n" +
		"----\n" +
		"Instr:	numpush		1\n" +
		"Stack:	LITERAL		3456\n" +
		"----\n" +
		"Instr:	numpush		2\n" +
		"Stack:	NUMBER		1\n" +
		"	LITERAL		3456\n" +
		"----\n" +
		"Instr:	bltin		substring()\n" +
		"Stack:	NUMBER		2\n" +
		"	NUMBER		1\n" +
		"	LITERAL		3456\n" +
		"----\n" +
		"Instr:	pathOperPush	..\n" +
		"Stack:	LITERAL		34\n" +
		"----\n" +
		"Instr:	nameTestPush\t{" + TestModule + " serial}\n" +
		"Stack:	PATHOPER	..\n" +
		"	LITERAL		34\n" +
		"----\n" +
		"Instr:	nameTestPush\t{" + TestModule + " name}\n" +
		"Stack:	NAMETEST\t{" + TestModule + " serial}\n" +
		"	PATHOPER	..\n" +
		"	LITERAL		34\n" +
		"----\n" +
		"Instr:	evalLocPath\n" +
		"Stack:	NAMETEST\t{" + TestModule + " name}\n" +
		"	NAMETEST\t{" + TestModule + " serial}\n" +
		"	PATHOPER	..\n" +
		"	LITERAL		34\n" +
		"----\n" +
		"CreateNodeSet:		Ctx: '/interface/dataplane'\n" +
		"\tApply: PATHOPER	..\n" +
		"\t\t\t/interface\n" +
		"\tApply: NAMETEST	{xpathNodeTestModule serial}\n" +
		"\t\t\t/interface/serial" + Nm_S1 + "\n" +
		"\tApply: NAMETEST	{xpathNodeTestModule name}\n" +
		"\t\t\t/interface/serial" + Nm_S1 + "/name (s1)\n" +
		"----\n" +
		"Instr:	add\n" +
		"Stack:	NODESET		/interface/serial" + Nm_S1 + "/name (s1)\n" +
		"	LITERAL		34\n" +
		"----\n" +
		"Instr:	store\n" +
		"Stack:	NUMBER		NaN\n" +
		"----\n"

	// Expression purely used to exercise stack print code - it does not
	// make any kind of sense!!!
	checkBoolResultWithContextDebugAndMap(
		t, "substring('3456',1,2) + ../serial/name", true,
		configTree, xutils.PathType([]string{"/", "interface", "dataplane"}),
		nodesetPfxMapFn,
		expOut)
}
