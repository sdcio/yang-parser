// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// These tests verify functionality for parsing 'leafref' XPATH statements
// from YANG.

package leafref

import (
	"fmt"
	"testing"

	"github.com/danos/yang/xpath"
	. "github.com/danos/yang/xpath/grammars/lexertest"
	"github.com/danos/yang/xpath/xpathtest"
	"github.com/danos/yang/xpath/xutils"
)

// For non-predicate tests, we can reuse the same 'config' which gives us
// enough different node types to test things.
func getTestCfgTree(t *testing.T) *xpathtest.TNode {
	return xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interface", "dataplane/name+dp0s1"},
			{"interface", "dataplane/name+dp0s1", "address@1111"},
			{"interface", "dataplane/name+dp0s2", "address@2111"},
			{"interface", "dataplane/name+dp0s2", "address@2222"},
			{"interface", "serial/name+s1", "address@5555"},
			{"interface", "loopback/name+lo2"},
			{"protocols", "mpls", "min-label+2111"},
		})
}

func checkLeafrefNodeSetResult(
	t *testing.T,
	expr string,
	mapFn xpath.PfxMapFn,
	configTree *xpathtest.TNode,
	absStartPath xutils.PathType,
	expResult xpathtest.TNodeSet,
) {
	mach, err := NewLeafrefMachine(expr, mapFn)
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %s", expr, err.Error())
		return
	}
	CheckNodeSetResult(t, mach, configTree, absStartPath, expResult)
}

func checkLeafrefNodeSetResultWithDebug(
	t *testing.T,
	expr string,
	mapFn xpath.PfxMapFn,
	configTree *xpathtest.TNode,
	absStartPath xutils.PathType,
	expResult xpathtest.TNodeSet,
	expOut string,
) {
	mach, err := NewLeafrefMachine(expr, mapFn)
	if err != nil {
		t.Fatalf("Unexpected error parsing %s: %s", expr, err.Error())
		return
	}
	CheckNodeSetResultWithDebug(t, mach, configTree, absStartPath, expResult,
		expOut)
}

func checkParseError(t *testing.T, expr string, errMsgs []string) {
	_, err := NewLeafrefMachine(expr, nil)
	CheckParseError(t, expr, err, errMsgs)
}

func checkParseErrorWithMap(
	t *testing.T,
	expr string,
	errMsgs []string,
	mapFn xpath.PfxMapFn,
) {
	_, err := NewLeafrefMachine(expr, mapFn)
	CheckParseError(t, expr, err, errMsgs)
}

// Absolute or relative path

// Absolute path: 1*("/" (node-identifier *path-predicate))
//
// AbsolutePath:Root NodeIdentifier PathPredicate1Plus AbsolutePathStep
// 		|		Root NodeIdentifier PathPredicate1Plus
// 		|		Root NodeIdentifier AbsolutePathStep
// 		|		Root NodeIdentifier
// 		;
// AbsolutePathStep:
// 				'/' NodeIdentifier PathPredicate1Plus AbsolutePathStep
// 		|		'/' NodeIdentifier PathPredicate1Plus
// 		|		'/' NodeIdentifier AbsolutePathStep
// 		|		'/' NodeIdentifier
// 		;
//
// The following 5 expressions test the last 2 lines in each production and
// cover the different node types as well.
//
// /interface                   [container]
// /interface/dataplane         [list]
// /interface/dataplane/address [leaf-list]
// /interface/dataplane/name    [keynode]
// /protocols/mpls/min-label    [leaf]
//
func TestParseAbsolutePathContainer(t *testing.T) {
	checkLeafrefNodeSetResult(t, "/interface", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "protocols", "mpls"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTContainer(
				nil, xutils.PathType([]string{"/", "interface"}),
				"", "interface")}))
}

func TestParseAbsolutePathList(t *testing.T) {
	checkLeafrefNodeSetResult(t, "/interface/dataplane", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

func TestParseAbsolutePathLeafList(t *testing.T) {
	checkLeafrefNodeSetResult(t, "/interface/dataplane/address", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2222")}))
}

func TestParseAbsolutePathKeynode(t *testing.T) {
	checkLeafrefNodeSetResult(t, "/interface/dataplane/name", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeaf(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "name"}),
				"", "name", "dp0s1"),
			xpathtest.NewTLeaf(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "name"}),
				"", "name", "dp0s2")}))
}

func TestParseAbsolutePathLeaf(t *testing.T) {
	checkLeafrefNodeSetResult(t, "/protocols/mpls/min-label", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeaf(
				nil, xutils.PathType([]string{"/", "protocols", "mpls", "min-label"}),
				"", "min-label", "2111")}))
}

func TestParseAbsolutePathNoMatch(t *testing.T) {
	checkLeafrefNodeSetResult(t, "/protocols/mpls/max-label", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "protocols"}),
		xpathtest.TNodeSet{})
}

// RelativePath
//
// RelativePath:DotDot '/' DescendantPath RelativePath
// 		|		DotDot '/' DescendantPath
// 		;
// DescendantPath:
//              NodeIdentifier PathPredicate1Plus AbsolutePathStep
// 		|		NodeIdentifier AbsolutePathStep
// 		|		NodeIdentifier
// 		;
//
// The following tests 'RelativePath' and the last 2 lines of 'DescendantPath'
// if we start at /interface/serial/
//
// ../dataplane
// ../dataplane/address
// ../../protocols

func TestRelativePathList(t *testing.T) {
	checkLeafrefNodeSetResult(t, "../dataplane", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "interface", "serial"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s1"),
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{"/", "interface", "dataplane"}),
				"", "dataplane", "name", "dp0s2")}))
}

func TestRelativePathMultipleIdentifiers(t *testing.T) {
	checkLeafrefNodeSetResult(t, "../dataplane/address", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "interface", "serial"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2222")}))
}

func TestRelativePathMultipleDotDots(t *testing.T) {
	checkLeafrefNodeSetResult(t, "../../protocols", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "interface", "serial"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTContainer(
				nil, xutils.PathType([]string{"/", "protocols"}),
				"", "protocols")}))
}

func TestRelativePathNoMatch(t *testing.T) {
	checkLeafrefNodeSetResult(t, "../../protocols/ospf", nil,
		getTestCfgTree(t), xutils.PathType([]string{"/", "interface", "serial"}),
		xpathtest.TNodeSet{})
}

// Basic path errors
func TestPathStartWithIdentifier(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile 'dataplane",
		"Parse Error: syntax error",
		"Got to approx [X] in 'dataplane [X] "}
	checkParseError(t, "dataplane", expErrMsgs)
}

func TestPathWithSingleDot(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile './dataplane",
		"Parse Error: syntax error",
		"Got to approx [X] in './ [X] dataplane",
		"Lexer Error: '.' is not a valid token."}
	checkParseError(t, "./dataplane", expErrMsgs)
}

func TestRelativePathMoreDotDots(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '../dataplane/../../protocols/mpls'",
		"Parse Error: syntax error",
		"Got to approx [X] in '../dataplane/.. [X] /../protocols/mpls'"}
	checkParseError(t, "../dataplane/../../protocols/mpls", expErrMsgs)
}

// PrefixedNames
func nodesetPfxMapFn(prefix string) (module string, err error) {
	switch prefix {
	case "xpath", "":
		return xpathtest.TestModule, nil
	default:
		return "",
			fmt.Errorf("Unable to locate module for prefix '%s'", prefix)
	}
}

func TestParsePrefixedName(t *testing.T) {
	checkLeafrefNodeSetResult(t, "../../xpath:protocols", nodesetPfxMapFn,
		getTestCfgTree(t), xutils.PathType([]string{"/", "interface", "serial"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTContainer(
				nil, xutils.PathType([]string{"/", "protocols"}),
				"", "protocols")}))
}

func TestParsePrefixedNames(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../xpath:interface/dataplane/xpath:address", nodesetPfxMapFn,
		getTestCfgTree(t), xutils.PathType([]string{"/", "interface", "serial"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "1111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{"/", "interface", "dataplane", "address"}),
				"", "address", "2222")}))
}

func TestParseUnknownPrefix(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '../unknown:protocols'",
		"Parse Error: syntax error",
		"Got to approx [X] in '../unknown:protocols [X] '",
		"Lexer Error: Unable to locate module for prefix 'unknown'"}
	checkParseErrorWithMap(t, "../unknown:protocols", expErrMsgs,
		nodesetPfxMapFn)
}

// From RFC 6020: example of predicate usage
// Our test config here comes from RFC 6020, specifically from the section
// on leafrefs (unsurprisingly).
//
// We use the predicate part here.
//
//   container interfaces {
//     list interface {
//       key "name";
//       leaf name {
//         type string;
//       }
//       list address {
//         key "ip";
//         leaf ip {
//           type yang:ip-address;
//         }
//         leaf protocol {
//           type string;
//         }
//       }
//       leaf notKey { # Dummy, just for test.
//         type string;
//       }
//     }
//     leaf mgmt-interface { # Not used for this test.
//       type leafref {
//         path "../interface/name";
//       }
//     }
//
//     container default-address {
//       leaf ifname {
//         type leafref {
//           path "../../interface/name";
//         }
//       }
//       leaf protocol {
//         type leafref {
//           path "../../interface/address/protocols";
//         }
//       }
//       leaf address {
//         type leafref {
//           path "../../interface[name = current()/../ifname]"
//              + "/address/ip";
//         }
//       }
//     }
//   }
func getPredicateTestCfgTree(t *testing.T) *xpathtest.TNode {
	return xpathtest.CreateTree(t,
		[]xutils.PathType{
			{"interfaces", "interface/name+dp0s1", "address/ip+1111"},
			{"interfaces", "interface/name+dp0s1", "notKey+AAAA"},
			{"interfaces", "interface/name+dp0s2", "address/ip+2111"},
			{"interfaces", "interface/name+dp0s2", "address/ip+2222"},
			{"interfaces", "interface/name+dp0s2", "address/ip+3333",
				"protocol+tcp"},
			{"interfaces", "interface/name+dp0s2", "address/ip+3333",
				"protocol+udp"},
			{"interfaces", "interface/name+s1", "address@5555"},
			{"interfaces", "interface/name+lo2", "address@9999"},
			{"interfaces", "default-address", "ifname+dp0s2"},
			{"interfaces", "default-address", "badifname+dp0s66"},
			{"interfaces", "default-address", "address@6666"},
			{"interfaces", "default-address", "ip+3333"},
		})
}

// AbsolutePath ending in predicate
func TestAbsolutePathEndsWithPredicates(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "/interfaces/interface[name = current()/../ifname]",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTListEntry(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface"}),
				"", "interface", "name", "dp0s2")}))
}

// Absolute path with predicate in middle of expression
func TestAbsolutePathWithPredicatesMidway(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "/interfaces/interface[name = current()/../ifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2222"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "3333")}))
}

// Absolute path, no match on interface name
func TestAbsolutePathNoMatchOnPredicate(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "/interfaces/interface[name = current()/../badifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet{})
}

// RelativePath with predicates
func TestRelativePathWithPredicates(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../interface[name = current()/../ifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2222"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "3333")}))
}

// 2 predicates operating on same nodeset.  Note that as we only currently
// support a single key, the test uses the same key name twice.
func TestRelativePathConsecutivePredicates(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../interface[name = current()/../ifname]"+
			"[name = current()/../ifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2111"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "2222"),
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "ip"}),
				"", "ip", "3333")}))
}
func TestRelativePathPredicateNoMatchFirstKey(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../interface[name2 = current()/../ifname]"+
			"[name = current()/../ifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet{})
}

func TestRelativePathPredicateNoMatchSecondKey(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../interface[name = current()/../ifname]"+
			"[name2 = current()/../ifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet{})
}

func TestRelativePathMultiplePredicates(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../interface[name = current()/../ifname]"+
			"/address[ip = current()/../ip]/protocol",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "protocol"}),
				"", "protocol", "udp")}))
}

func TestParsePredicatePrefixedKeyName(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../interface[xpath:name = current()/../ifname]"+
			"/address[ip = current()/../ip]/protocol",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "protocol"}),
				"", "protocol", "udp")}))
}

func TestParsePredicatePrefixedKeyValue(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "../../interface[name = current()/../xpath:ifname]"+
			"/address[ip = current()/../ip]/protocol",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet([]*xpathtest.TNode{
			xpathtest.NewTLeafList(
				nil, xutils.PathType([]string{
					"/", "interfaces", "interface", "address", "protocol"}),
				"", "protocol", "udp")}))
}

// Predicate error cases - make sure the parser is rejecting everything as
// expected so we don't get any weird constructs making it past here.

func TestParseIsolatedPredicate(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '[foo]'",
		"Parse Error: syntax error",
		"Got to approx [X] in '[ [X] foo]'"}
	checkParseErrorWithMap(t, "[foo]", expErrMsgs,
		nodesetPfxMapFn)
}

func TestParsePredicateMissingCurrent(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '../dataplane[name = ../ifname]'",
		"Parse Error: syntax error",
		"Got to approx [X] in '../dataplane[name = .. [X] /ifname]'"}
	checkParseErrorWithMap(t, "../dataplane[name = ../ifname]",
		expErrMsgs, nodesetPfxMapFn)
}

func TestParsePredicateEndsWithCurrent(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '../dataplane[name = current()]'",
		"Parse Error: syntax error",
		"Got to approx [X] in '../dataplane[name = current()] [X] '"}
	checkParseErrorWithMap(t, "../dataplane[name = current()]",
		expErrMsgs, nodesetPfxMapFn)
}

func TestParsePredicateOtherFunction(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '../dataplane[name = false()/../ifname]'",
		"Parse Error: syntax error",
		"Got to approx [X] in '../dataplane[name = false [X] ()/../ifname]'",
		"Lexer Error: Function 'false' is not valid here."}
	checkParseErrorWithMap(t, "../dataplane[name = false()/../ifname]",
		expErrMsgs, nodesetPfxMapFn)
}

func TestParseRelPathEndingInPredicate(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '../dataplane[name = current()/../ifname]'",
		"Parse Error: syntax error",
		"Got to approx [X] in '../dataplane[name = current()/../ifname] [X] '"}
	checkParseErrorWithMap(t, "../dataplane[name = current()/../ifname]",
		expErrMsgs, nodesetPfxMapFn)
}

// KeyName is not a key but is a valid path.  Ensure we get an empty nodeset.
func TestParsePredicateNonKeyName(t *testing.T) {
	checkLeafrefNodeSetResult(
		t, "/interfaces/interface[notKey = current()/../badifname]/address/ip",
		nodesetPfxMapFn,
		getPredicateTestCfgTree(t), xutils.PathType([]string{
			"/", "interfaces", "default-address", "address"}),
		xpathtest.TNodeSet{})
}

func TestParsePredicateUnknownPrefixKeyName(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '/interfaces/interface[unknown:name = " +
			"current()/../ifname]/address/ip'",
		"Parse Error: syntax error",
		"Got to approx [X] in '/interfaces/interface[unknown:name [X]  " +
			"= current()/../ifname]/address/ip'",
		"Lexer Error: Unable to locate module for prefix 'unknown'"}

	checkParseErrorWithMap(t, "/interfaces/interface[unknown:name = "+
		"current()/../ifname]/address/ip",
		expErrMsgs, nodesetPfxMapFn)
}

func TestParsePredicateUnknownPrefixKeyValue(t *testing.T) {
	expErrMsgs := []string{
		"Failed to compile '/interfaces/interface[name = " +
			"current()/../unknown:ifname]/address/ip'",
		"Parse Error: syntax error",
		"Got to approx [X] in '/interfaces/interface[name " +
			"= current()/../unknown:ifname [X] ]/address/ip'",
		"Lexer Error: Unable to locate module for prefix 'unknown'"}

	checkParseErrorWithMap(t, "/interfaces/interface[name = "+
		"current()/../unknown:ifname]/address/ip",
		expErrMsgs, nodesetPfxMapFn)
}

func TestStillToDo(t *testing.T) {
	t.Skipf("Still to do ...")
	// LRefPredEnd: 0 or 2 leaf values TEST (program_test?)
	// Check final nodeset is all leaves, not other node types?
	// How do we determine 'current' in 2nd part of predicate: is this a
	// safe way to do it, and do we have quick UT to validate it?

	// Need to doc in README
	// At any '[', we eval loc path so consume any stacked nodeset.  This
	// means that when we evaluate name = <x>, EvalLocPath on <x> has no
	// stacked nodeset and thus uses the context node as the start point.
}
