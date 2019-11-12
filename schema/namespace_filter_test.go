// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"strings"
	"testing"

	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/schema"
	"github.com/danos/yang/testutils"
)

//
// HELPER FUNCTIONS
//

func getDataTreeWithFilter(
	t *testing.T, sn schema.Node, input_json string, filter schema.Filter,
) datanode.DataNode {

	dn := getOriginalDataTree(t, sn, input_json)

	return schema.FilterTree(sn, dn, filter)
}

func namespaceFilter(namespace string) schema.Filter {
	return func(s schema.Node, d datanode.DataNode, children []datanode.DataNode) bool {
		if len(children) != 0 {
			return true
		}
		return strings.Contains(s.Namespace(), namespace)
	}
}

const schemaModuleA = `
module test-configd-schema-A {
	namespace "urn:vyatta.com:test:configd-schema:a";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for configd";
	}
	container containerA {
		leaf leafA {
			type string;
		}
	}
	list listA {
		key name;
		leaf name {
			type string;
		}
		leaf leafA {
			type string;
		}
	}
}
`

const schemaModuleB = `
module test-configd-schema-B {
	namespace "urn:vyatta.com:test:configd-schema:b";
	prefix test;

	import test-configd-schema-A {
		prefix modA;
	}

	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for configd";
	}

	container containerB {
		leaf leafB {
			type string;
		}
	}
	list listB {
		key name;
		leaf name {
			type string;
		}
		leaf leafB {
			type string;
		}
	}

	augment /modA:containerA {
		leaf leafB {
			type string;
		}
	}
	augment /modA:listA {
		leaf leafB {
			type string;
		}
	}
}
`

func findNode(node datanode.DataNode, name string) datanode.DataNode {
	for _, cn := range node.YangDataChildrenNoSorting() {
		if cn.YangDataName() == name {
			return cn
		}
	}
	return nil
}

func assertExists(t *testing.T, actual datanode.DataNode, spath string) {
	path := strings.Split(spath, "/")

	for _, elem := range path {
		actual = findNode(actual, elem)
		if actual == nil {
			t.Errorf("Missing expected node '%s' in path %s", elem, spath)
			return
		}
	}
}

func assertMissing(t *testing.T, actual datanode.DataNode, spath string) {
	path := strings.Split(spath, "/")

	for _, elem := range path {
		actual = findNode(actual, elem)
		if actual == nil {
			return
		}
	}
	t.Errorf("Unexpected path found: %s", spath)
}

func TestNameSpaceFilter(t *testing.T) {
	schema_text_a := []byte(schemaModuleA)
	schema_text_b := []byte(schemaModuleB)

	ms, err := testutils.GetConfigSchema(schema_text_a, schema_text_b)
	if err != nil {
		t.Fatalf("Unexpected compilation failure:\n  %s\n\n", err.Error())
	}

	const input_json = `{
	"containerA":{"leafA":"skip","leafB":"match"},
	"listA":[{"name":"match","leafA":"skip","leafB":"match"},
	         {"name":"skip","leafA":"skip"}],
	"containerB":{"leafB":"match"},
	"listB":[{"name":"match","leafB":"match"}]
}`

	// Due to the use of maps in decoding JSON, order is random
	actual := getDataTreeWithFilter(
		t, ms, input_json, namespaceFilter("urn:vyatta.com:test:configd-schema:b"))

	assertMissing(t, actual, "containerA/leafA")
	assertExists(t, actual, "containerA/leafB")

	assertExists(t, actual, "listA/match/name")
	assertMissing(t, actual, "listA/match/leafA")
	assertExists(t, actual, "listA/match/leafB")
	assertMissing(t, actual, "listA/skip/name")
	assertMissing(t, actual, "listA/skip/leafA")

	assertExists(t, actual, "containerB/leafB")
	assertExists(t, actual, "listB/match/name")
	assertExists(t, actual, "listB/match/leafB")
}
