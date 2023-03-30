// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"strings"
	"testing"

	"github.com/steiler/yang-parser/data/datanode"
	"github.com/steiler/yang-parser/data/encoding"
	"github.com/steiler/yang-parser/schema"
)

//
// HELPER FUNCTIONS
//

func getDataTreeWithFilterAsJSON(
	t *testing.T, input_schema, input_json string, filter schema.Filter,
) string {

	sn := getSchema(t, input_schema)
	dn := getOriginalDataTree(t, sn, input_json)

	filtered_tree := schema.FilterTree(sn, dn, filter)
	return string(encoding.ToJSON(sn, filtered_tree))
}

// FILTER TESTS
var filter = func(s schema.Node, d datanode.DataNode, children []datanode.DataNode) bool {
	if len(children) != 0 {
		return true
	}
	return strings.Contains(d.YangDataName(), "match")
}

func TestMatches(t *testing.T) {

	const input_json = `{"matchLeaf":true}`
	const input_schema = `
leaf matchLeaf {
	type boolean;
}`

	actual := getDataTreeWithFilterAsJSON(t, input_schema, input_json, filter)
	expect := `{"matchLeaf":true}`
	assertMatch(t, expect, actual)
}

func TestNoMatches(t *testing.T) {

	const input_json = `{"skipLeaf":true}`
	const input_schema = `
leaf skipLeaf {
	type boolean;
}`

	actual := getDataTreeWithFilterAsJSON(t, input_schema, input_json, filter)
	expect := `{}`
	assertMatch(t, expect, actual)
}

func TestContainer(t *testing.T) {
	const input_json = `{"matchContainer":{"matchLeaf":true,"skipLeaf":true}}`
	const input_schema = `
container matchContainer {
    leaf matchLeaf {
	    type boolean;
    }
    leaf skipLeaf {
	    type boolean;
    }
}`

	actual := getDataTreeWithFilterAsJSON(t, input_schema, input_json, filter)
	expect := `{"matchContainer":{"matchLeaf":true}}`
	assertMatch(t, expect, actual)
}

func TestDropEmptyContainer(t *testing.T) {
	const input_json = `{"matchContainer":{"skipLeaf":true}}`
	const input_schema = `
container matchContainer {
    leaf matchLeaf {
	    type boolean;
    }
    leaf skipLeaf {
	    type boolean;
    }
}`

	actual := getDataTreeWithFilterAsJSON(t, input_schema, input_json, filter)
	expect := `{}`
	assertMatch(t, expect, actual)
}

func TestKeepPresenceContainer(t *testing.T) {
	const input_json = `{"matchContainer":{"skipLeaf":true}}`
	const input_schema = `
container matchContainer {
    presence "totes!";
    leaf matchLeaf {
	    type boolean;
    }
    leaf skipLeaf {
	    type boolean;
    }
}`

	actual := getDataTreeWithFilterAsJSON(t, input_schema, input_json, filter)
	expect := `{"matchContainer":{}}`
	assertMatch(t, expect, actual)
}

func TestKeepListKeys(t *testing.T) {
	const input_json = `{"skipList":[{"skipLeaf":true,"matchLeaf":true}]}`
	const input_schema = `
list skipList {
    key skipLeaf;
    leaf matchLeaf {
	    type boolean;
    }
    leaf skipLeaf {
	    type boolean;
    }
}`

	actual := getDataTreeWithFilterAsJSON(t, input_schema, input_json, filter)
	expect := `{"skipList":[{"skipLeaf":true,"matchLeaf":true}]}`
	assertMatch(t, expect, actual)
}
