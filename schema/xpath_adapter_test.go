// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"testing"

	"github.com/danos/yang/data/encoding"
	"github.com/danos/yang/schema"
	"github.com/danos/yang/testutils"
	"github.com/danos/yang/xpath/xutils"
)

func TestBasicDecodedJSON(t *testing.T) {
	const input_schema = `
module test-yang-schema-xpath {
	namespace "urn:vyatta.com:test:yang-schema-xpath";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for xpath adapter";
	}
  container testcontainer {
	leaf testboolean {
		type boolean;
		default false;
	}
	leaf teststring {
		type string;
	}
	leaf-list testleaflist {
		type string;
	}
	leaf-list testleaflistuser {
		type string;
		ordered-by user;
	}
	list testlist {
		key name;
		leaf name {
			type string;
		}
		leaf bar {
			type empty;
		}
	}
	list testlistuser {
		ordered-by "user";
		key name;
		leaf name {
			type string;
		}
		leaf bar {
			type empty;
		}
	}
  }
  container state {
	config false;
	leaf status {
		type string;
		default "foo";
	}
  }
}`

	const input_json = `
{"testcontainer":
    {"testboolean":false,
     "testleaflist":["foo",
                     "bar"],
     "testleaflistuser":["foo",
                         "bar"],
     "testlist":[{"name":"bar","bar":null},
                 {"name":"baz","bar":null},
                 {"name":"foo","bar":null}],
     "testlistuser":[{"name":"foo","bar":null},
                     {"name":"baz","bar":null},
                     {"name":"bar","bar":null}],
     "teststring":"foo"
    }
}`

	sn, err := testutils.GetFullSchema([]byte(input_schema))
	if err != nil {
		t.Fatalf("Failed to compile test schema: %s\n", err.Error())
	}

	dn, err := encoding.NewUnmarshaller(encoding.JSON).
		Unmarshal(sn, []byte(input_json))
	if err != nil {
		t.Fatalf("Failed to decode input JSON")
	}

	xn := schema.ConvertToXpathNode(dn, sn)

	expected := []*xChecker{
		xCheck("testcontainer", "", "/testcontainer",
			"/testcontainer",
			xCheck("testboolean", "false", "/testcontainer/testboolean",
				"/testcontainer/testboolean (false)"),
			xCheck("testleaflist", "bar", "/testcontainer/testleaflist",
				"/testcontainer/testleaflist (bar)"),
			xCheck("testleaflist", "foo", "/testcontainer/testleaflist",
				"/testcontainer/testleaflist (foo)"),
			xCheck("testleaflistuser", "foo", "/testcontainer/testleaflistuser",
				"/testcontainer/testleaflistuser (foo)"),
			xCheck("testleaflistuser", "bar", "/testcontainer/testleaflistuser",
				"/testcontainer/testleaflistuser (bar)"),
			xCheck("testlist", "bar", "/testcontainer/testlist",
				"/testcontainer/testlist[name='bar']",
				xCheck("bar", "", "/testcontainer/testlist/bar",
					"/testcontainer/testlist[name='bar']/bar ()"),
				xCheck("name", "bar", "/testcontainer/testlist/name",
					"/testcontainer/testlist[name='bar']/name (bar)"),
			),
			xCheck("testlist", "baz", "/testcontainer/testlist",
				"/testcontainer/testlist[name='baz']",
				xCheck("bar", "", "/testcontainer/testlist/bar",
					"/testcontainer/testlist[name='baz']/bar ()"),
				xCheck("name", "baz", "/testcontainer/testlist/name",
					"/testcontainer/testlist[name='baz']/name (baz)"),
			),
			xCheck("testlist", "foo", "/testcontainer/testlist",
				"/testcontainer/testlist[name='foo']",
				xCheck("bar", "", "/testcontainer/testlist/bar",
					"/testcontainer/testlist[name='foo']/bar ()"),
				xCheck("name", "foo", "/testcontainer/testlist/name",
					"/testcontainer/testlist[name='foo']/name (foo)"),
			),
			xCheck("testlistuser", "foo", "/testcontainer/testlistuser",
				"/testcontainer/testlistuser[name='foo']",
				xCheck("bar", "", "/testcontainer/testlistuser/bar",
					"/testcontainer/testlistuser[name='foo']/bar ()"),
				xCheck("name", "foo", "/testcontainer/testlistuser/name",
					"/testcontainer/testlistuser[name='foo']/name (foo)"),
			),
			xCheck("testlistuser", "baz", "/testcontainer/testlistuser",
				"/testcontainer/testlistuser[name='baz']",
				xCheck("bar", "", "/testcontainer/testlistuser/bar",
					"/testcontainer/testlistuser[name='baz']/bar ()"),
				xCheck("name", "baz", "/testcontainer/testlistuser/name",
					"/testcontainer/testlistuser[name='baz']/name (baz)"),
			),
			xCheck("testlistuser", "bar", "/testcontainer/testlistuser",
				"/testcontainer/testlistuser[name='bar']",
				xCheck("bar", "", "/testcontainer/testlistuser/bar",
					"/testcontainer/testlistuser[name='bar']/bar ()"),
				xCheck("name", "bar", "/testcontainer/testlistuser/name",
					"/testcontainer/testlistuser[name='bar']/name (bar)"),
			),
			xCheck("teststring", "foo", "/testcontainer/teststring",
				"/testcontainer/teststring (foo)"),
		),
	}

	checkAllChildren(t, xutils.AllChildren, xn, expected)
}

func TestXChildrenFiltering(t *testing.T) {
	const input_schema = `
		module test-yang-schema-xpath {
		namespace "urn:vyatta.com:test:yang-schema-xpath";
		prefix test;
		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test XChildren filters correctly";
		}
		container testcontainer {
			leaf teststring {
				type string;
			}
		}

		container state {
			config false;
			leaf status {
				type string;
			}
		}
	}`

	const input_json = `
		{"testcontainer":{"teststring":"foo"},
		"state":{"status":"bar"}}`

	sn, err := testutils.GetFullSchema([]byte(input_schema))
	if err != nil {
		t.Fatalf("Failed to compile test schema: %s\n", err.Error())
	}

	dn, err := encoding.NewUnmarshaller(encoding.JSON).
		Unmarshal(sn, []byte(input_json))
	if err != nil {
		t.Fatalf("Failed to decode input JSON")
	}

	xn := schema.ConvertToXpathNode(dn, sn)

	expCfgOnly := []*xChecker{
		xCheck("testcontainer", "", "/testcontainer",
			"/testcontainer",
			xCheck("teststring", "foo", "/testcontainer/teststring",
				"/testcontainer/teststring (foo)"),
		),
	}
	expCfgAndState := []*xChecker{
		xCheck("state", "", "/state",
			"/state",
			xCheck("status", "bar", "/state/status",
				"/state/status (bar)"),
		),
		xCheck("testcontainer", "", "/testcontainer",
			"/testcontainer",
			xCheck("teststring", "foo", "/testcontainer/teststring",
				"/testcontainer/teststring (foo)"),
		),
	}

	checkAllChildren(t, xutils.AllCfgChildren, xn, expCfgOnly)
	checkAllChildren(t, xutils.AllChildren, xn, expCfgAndState)
}

func checkAllChildren(
	t *testing.T,
	filter xutils.XFilter,
	actual xutils.XpathNode,
	expectedChildren []*xChecker,
) {
	for i, xcn := range actual.XChildren(filter) {
		if i >= len(expectedChildren) {
			t.Errorf("Unexpected child found: %s\n", xcn.XName())
			continue
		}
		p := xcn.XParent()
		if p != actual {
			t.Errorf("XParent doesn't match expected parent: %s vs %s\n",
				p.XName(), actual.XName())
		}
		expectedChildren[i].check(t, filter, xcn)
	}
}

type xChecker struct {
	name       string
	value      string
	path       string
	nodeString string
	children   []*xChecker
}

func xCheck(
	name, value, path, nodeString string,
	children ...*xChecker,
) *xChecker {

	return &xChecker{
		name:       name,
		value:      value,
		path:       path,
		nodeString: nodeString,
		children:   children}
}

func (expect *xChecker) check(
	t *testing.T,
	filter xutils.XFilter,
	actual xutils.XpathNode,
) {
	if actual.XName() != expect.name {
		t.Fatalf("Name does not match for %s/%s:\n  expect %s\n  actual %s",
			expect.path, expect.value, expect.name,
			actual.XName())
	}
	if actual.XValue() != expect.value {
		t.Fatalf(
			"Value does not match for %s/%s:\n  expect %s\n  actual %s",
			expect.path, expect.value, expect.value,
			actual.XValue())
	}
	if actual.XPath().String() != expect.path {
		t.Fatalf("Path does not match for %s/%s:\n  expect %s\n  actual %s",
			expect.path, expect.value, expect.path,
			actual.XPath())
	}
	if xutils.NodeString(actual) != expect.nodeString {
		t.Fatalf("NodeString does not match for %s/%s:\n"+
			"  expect %s\n  actual %s",
			expect.path, expect.value, expect.nodeString,
			xutils.NodeString(actual))
	}
	checkAllChildren(t, filter, actual, expect.children)
}
