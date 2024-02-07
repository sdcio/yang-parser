// Copyright (c) 2018-2019, AT&T Intellectual Property Inc.
// All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse_test

import (
	"strings"
	"testing"

	. "github.com/sdcio/yang-parser/parse"
)

func getYangModuleText(yang_subtree string) string {
	return `module parse-test {
    namespace "urn:vyatta.com:mgmt:parse-test";          // Child 0
    prefix parse-test;                                   // Child 1

    organization "Brocade Communications Systems, Inc."; // Child 2
	contact                                              // Child 3
		"Brocade Communications Systems, Inc.
		 Postal: 130 Holger Way
			     San Jose, CA 95134
		 E-mail: support@Brocade.com
		 Web: www.brocade.com";

	revision 2015-04-10 {                                // Child 4
		description "Initial revision.";
	}
` + yang_subtree + `                                     // Child 5
}`
}

func createParseTreeFromYang(t *testing.T, yang_subtree string) *Tree {

	module_text := getYangModuleText(yang_subtree)
	tree, err := Parse("TestModule", module_text, nil)
	if err != nil {
		t.Errorf("Unexpected Parse Error - %s", err)
		t.FailNow()
	}

	return tree
}

func getParseErrorFromYang(t *testing.T, yang_subtree string) error {

	module_text := getYangModuleText(yang_subtree)
	_, err := Parse("TestModule", module_text, nil)
	if err == nil {
		t.Errorf("Unexpected Parse Success")
		t.FailNow()
	}

	return err
}

type NodeChecker interface {
	check(t *testing.T, node Node)
}

func verifyExpectedPass(
	t *testing.T,
	subtree string,
	topNodeName string,
	expected NodeChecker,
) {
	tree := createParseTreeFromYang(t, subtree)
	actual := findNode(tree, topNodeName)
	if actual == nil {
		t.Errorf("Failed to find expected schema node\n")
		return
	}

	expected.check(t, actual)
}

func verifyExpectedFail(
	t *testing.T,
	subtree string,
	expected string,
) {
	err := getParseErrorFromYang(t, subtree)

	if !strings.Contains(err.Error(), expected) {
		t.Errorf("Expected %s, Got %s", expected, err.Error())
	}
}

func checkError(t *testing.T, err error, expected string) {
	actual := err.Error()
	if !strings.Contains(actual, expected) {
		t.Errorf("Expected %s, Got %s", expected, err.Error())
	}
}

func findNode(tree *Tree, name string) Node {
	for _, v := range tree.Root.Children() {
		if strings.Contains(v.String(), name) {
			return v
		}
	}
	return nil
}

type LeafNodeChecker struct {
	Name   string
	Typ    string
	Status string
}

func (expected LeafNodeChecker) check(t *testing.T, actual Node) {
	if actual.Type() != NodeLeaf {
		t.Errorf("Unexpected Node Type: %s", actual.Type())
	}

	if actual.Name() != expected.Name {
		t.Errorf("Unexpected leaf name: Expected %s, Got %s",
			expected.Name, actual.Name())
	}

	if expected.Typ != "" {
		typ := actual.ChildByType(NodeTyp)
		tname := typ.ArgIdRef()
		if tname.Local != expected.Typ {
			t.Errorf("Unexpected Typedef type:\n  Exp: %s\n  Got: %s\n",
				expected.Typ, tname.Local)
		}
	}
	if expected.Status != "" {
		node := actual.ChildByType(NodeStatus)
		actual := node.ArgStatus()
		if expected.Status != actual {
			t.Errorf("Unexpected Status:\n  Exp: %s\n  Got: %s\n",
				expected.Status, actual)
		}
	}
}

type DeviationNodeChecker struct {
	Target string
	Body   []DeviateNodeChecker
}

func (expected DeviationNodeChecker) check(
	t *testing.T,
	actual Node,
) {
	if actual.Type() != NodeDeviation {
		t.Errorf("Unexpected Node Type: %s", actual.Type())
	}

	actualBody := actual.ChildrenByType(NodeDeviate)
	if len(actualBody) != len(expected.Body) {
		t.Errorf("Unexpected number of definitions: Expected %d, Got %d",
			len(expected.Body), len(actualBody))
		return
	}

	for i, _ := range expected.Body {
		expected.Body[i].check(t, actualBody[i])
	}
}

type DeviateNodeChecker struct {
	Type NodeType
}

func (expected DeviateNodeChecker) check(
	t *testing.T,
	actual Node,
) {
	if actual.Type() != expected.Type {
		t.Errorf("Unexpected Deviate type: Expecting %s Got %s", expected.Type, actual.Type())
	}
}

type ContainerNodeChecker struct {
	Name string
	Body []LeafNodeChecker // Need a DataDef node checker !!!
}

func (expected ContainerNodeChecker) check(
	t *testing.T,
	actual Node,
) {
	if actual.Type() != NodeContainer {
		t.Errorf("Unexpected Node Type: %s", actual.Type())
	}

	if actual.Name() != expected.Name {
		t.Errorf("Unexpected case name: Expected %s, Got %s",
			expected.Name, actual.Name())
	}

	actualBody := actual.ChildrenByType(NodeDataDef)
	if len(actualBody) != len(expected.Body) {
		t.Errorf("Unexpected number of definitions: Expected %d, Got %d",
			len(expected.Body), len(actualBody))
		return
	}

	for i, _ := range expected.Body {
		expected.Body[i].check(t, actualBody[i])
	}
}

type CaseNodeChecker struct {
	Name string
	Body []LeafNodeChecker
}

func (expected CaseNodeChecker) check(t *testing.T, actual Node) {
	if actual.Name() != expected.Name {
		t.Errorf("Unexpected case name: Expected %s, Got %s",
			expected.Name, actual.Name())
	}

	actualBody := actual.ChildrenByType(NodeDataDef)
	if len(actualBody) != len(expected.Body) {
		t.Errorf("Unexpected number of definitions: Expected %d, Got %d",
			len(expected.Body), len(actualBody))
		return
	}

	for i, _ := range expected.Body {
		expected.Body[i].check(t, actualBody[i])
	}
}

type ChoiceNodeChecker struct {
	Name  string
	Cases []CaseNodeChecker
}

func (expected ChoiceNodeChecker) check(t *testing.T, actual Node) {
	if actual.Type() != NodeChoice {
		t.Errorf("Unexpected Node Type: %s", actual.Type())
	}

	actualCases := actual.ChildrenByType(NodeCase)
	if len(actualCases) != len(expected.Cases) {
		t.Errorf("Unexpected number of cases: Expected %d, Got %d",
			len(expected.Cases), len(actualCases))
		return
	}

	for i, _ := range expected.Cases {
		expected.Cases[i].check(t, actualCases[i])
	}
}

type InputNodeChecker struct {
	Body []LeafNodeChecker
}

func (expected InputNodeChecker) check(t *testing.T, actual Node) {
	if actual == nil {
		t.Errorf("Missing input node")
	}

	actualBody := actual.ChildrenByType(NodeDataDef)
	if len(actualBody) != len(expected.Body) {
		t.Errorf("Unexpected number of definitions: Expected %d, Got %d",
			len(expected.Body), len(actualBody))
		return
	}

	for i, _ := range expected.Body {
		expected.Body[i].check(t, actualBody[i])
	}
}

type OutputNodeChecker struct {
	Body []LeafNodeChecker
}

func (expected OutputNodeChecker) check(t *testing.T, actual Node) {
	if actual == nil {
		t.Errorf("Missing output node")
	}

	actualBody := actual.ChildrenByType(NodeDataDef)
	if len(actualBody) != len(expected.Body) {
		t.Errorf("Unexpected number of definitions: Expected %d, Got %d",
			len(expected.Body), len(actualBody))
		return
	}

	for i, _ := range expected.Body {
		expected.Body[i].check(t, actualBody[i])
	}
}

type RpcNodeChecker struct {
	Name   string
	Input  InputNodeChecker
	Output OutputNodeChecker
}

func (expected RpcNodeChecker) check(t *testing.T, actual Node) {
	if actual.Type() != NodeRpc {
		t.Errorf("Unexpected Node Type: %s", actual.Type())
		return
	}

	if actual.Name() != expected.Name {
		t.Errorf("Unexpected RPC name:\n  Exp: %s\n  Got: %s\n",
			expected.Name, actual.Name())
	}
	expected.Input.check(t, actual.ChildByType(NodeInput))
	expected.Output.check(t, actual.ChildByType(NodeOutput))
}

type TypeDefNodeChecker struct {
	Name      string
	Typ       string
	Normalize string
}

func (expected TypeDefNodeChecker) check(t *testing.T, actual Node) {
	if actual.Type() != NodeTypedef {
		t.Errorf("Unexpected Node Type: %s", actual.Type())
		return
	}

	if actual.Name() != expected.Name {
		t.Errorf("Unexpected Typedef name:\n  Exp: %s\n  Got: %s\n",
			expected.Name, actual.Name())
	}

	typ := actual.ChildByType(NodeTyp)
	tname := typ.ArgIdRef()
	if tname.Local != expected.Typ {
		t.Errorf("Unexpected Typedef type:\n  Exp: %s\n  Got: %s\n",
			expected.Typ, tname.Local)
	}
}
