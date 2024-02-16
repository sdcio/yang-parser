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

// Copyright (c) 2019-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema_test

import (
	"fmt"
	"testing"

	"github.com/sdcio/yang-parser/data/datanode"
	"github.com/sdcio/yang-parser/data/encoding"
	"github.com/sdcio/yang-parser/schema"
	"github.com/sdcio/yang-parser/testutils"
)

//
// HELPER FUNCTIONS
//

func getSchema(t *testing.T, input_schema string) schema.Node {
	const schema_template = `
module test-yang-schema {
	namespace "urn:vyatta.com:test:yang-schema";
	prefix test;
	organization "Brocade Communications Systems, Inc.";
	revision 2014-12-29 {
		description "Test schema for xpath adapter";
	}

    %s
}`

	module := []byte(fmt.Sprintf(schema_template, input_schema))

	sn, err := testutils.GetFullSchema(module)
	if err != nil {
		t.Fatalf("Failed to compile test schema: %s\n", err.Error())
	}

	return sn
}

func getOriginalDataTree(t *testing.T, sn schema.Node, input_json string) datanode.DataNode {

	dn, err := encoding.NewUnmarshaller(encoding.JSON).
		Unmarshal(sn, []byte(input_json))
	if err != nil {
		t.Fatalf("Failed to decode input JSON")
	}

	return dn
}

func getDataTreeWithDefaultsAsJSON(t *testing.T, input_schema, input_json string) string {

	sn := getSchema(t, input_schema)
	dn := getOriginalDataTree(t, sn, input_json)

	with_defaults := schema.AddDefaults(sn, dn)

	return string(encoding.ToJSON(sn, with_defaults))
}

func assertMatch(t *testing.T, expect, actual string) {

	if string(actual) != expect {
		t.Fatalf("Unexpected result:\n    expect: %s\n    actual: %s\n",
			expect, actual)
	}
}

// Leaf Tests
//   * Values are not added if there isn't a default
//   * Defaults are added
//   * Existing values are not overridden

func TestNoDefault(t *testing.T) {

	const input_json = `{}`
	const input_schema = `
leaf testboolean {
	type boolean;
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{}`
	assertMatch(t, expect, actual)
}

func TestAddLeafDefault(t *testing.T) {

	const input_json = `{}`
	const input_schema = `
leaf testboolean {
	type boolean;
	default false;
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testboolean":false}`
	assertMatch(t, expect, actual)
}

func TestAddSkipDefault(t *testing.T) {

	const input_json = `{"testboolean":true}`
	const input_schema = `
leaf testboolean {
	type boolean;
	default false;
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testboolean":true}`
	assertMatch(t, expect, actual)
}

// Container Tests
//   * Add default container and leaf
//   * Existing values are not overriden
//   * Presence containers with defaults are not added
//   * Defaults in existing presences containers are added

func TestAddContainerDefault(t *testing.T) {

	const input_json = `{}`
	const input_schema = `
container testcontainer {
	leaf testboolean {
		type boolean;
		default true;
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"testboolean":true}}`
	assertMatch(t, expect, actual)
}

func TestSkipContainerDefault(t *testing.T) {

	const input_json = `{"testcontainer":{"testboolean":false}}`
	const input_schema = `
container testcontainer {
	leaf testboolean {
		type boolean;
		default true;
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"testboolean":false}}`
	assertMatch(t, expect, actual)
}

func TestSkipPresenceContainerDefault(t *testing.T) {

	const input_json = `{}`
	const input_schema = `
container testcontainer {
	presence "optional";
	leaf testboolean {
		type boolean;
		default true;
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{}`
	assertMatch(t, expect, actual)
}

// Special: A presence container isn't default but can contain defaults
func TestPresenceContainerDefault(t *testing.T) {

	const input_json = `{"testcontainer":{"isset":true}}`
	const input_schema = `
container testcontainer {
	presence "optional";
	leaf isset {
		type boolean;
	}
	leaf testboolean {
		type boolean;
		default true;
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"isset":true,"testboolean":true}}`
	assertMatch(t, expect, actual)
}

func TestEmptyChoiceNoDefaultCase(t *testing.T) {

	const input_json = `{"testcontainer":{"testleaf":"data"}}`
	const input_schema = `
container testcontainer {
	leaf testleaf {
		type string;
	}
	leaf testdefault {
		type string;
		default "defaultstring";
	}
	choice foobar {
		case foo {
			leaf foo {
				type string;
				default foo;
			}
		}
		case bar {
			leaf bar {
				type string;
				default bar;
			}
		}
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"testleaf":"data","testdefault":"defaultstring"}}`
	assertMatch(t, expect, actual)
}

func TestEmptyChoiceDefaultCase(t *testing.T) {

	const input_json = `{"testcontainer":{}}`
	const input_schema = `
container testcontainer {
	presence "for testing";
	choice foobar {
		default foo;
		case foo {
			leaf foo {
				type string;
				default foo;
			}
		}
		case bar {
			leaf bar {
				type string;
				default bar;
			}
		}
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"foo":"foo"}}`
	assertMatch(t, expect, actual)
}

func TestEmptyChoiceDefaultNestedCase(t *testing.T) {

	const input_json = `{"testcontainer":{}}`
	const input_schema = `
container testcontainer {
	presence "";
	leaf testleaf {
		type string;
	}
	choice foobar {
		default foo;
		case foo {
			leaf foo {
				type string;
			}

			choice foobar {
				default foobar;
				case bar {
					leaf barleaf {
						type string;
						default "barleafvalue";
					}
				}
				case foobar {
					leaf defone {
						type string;
						default "defone value";
					}
				}
			}
		}
		case bar {
			leaf bar {
				type string;
				default bar;
			}
		}
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"defone":"defone value"}}`
	assertMatch(t, expect, actual)
}

func TestChoiceOverrideDefaultCase(t *testing.T) {

	const input_json = `{"testcontainer":{"baz":"data"}}`
	const input_schema = `
container testcontainer {
	leaf testleaf {
		type string;
	}
	choice foobar {
		default foo;
		case foo {
			leaf foo {
				type string;
				default foo;
			}
		}
		case bar {
			leaf baz {
				type string;
			}
			leaf bar {
				type string;
				default bar;
			}
		}
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"baz":"data","bar":"bar"}}`
	assertMatch(t, expect, actual)
}

func TestNestedChoiceOverrideDefaultCase(t *testing.T) {

	const input_json = `{"testcontainer":{"baz":"data"}}`
	const input_schema = `
container testcontainer {
	leaf testleaf {
		type string;
	}
	choice top {
		default default;
		case foobar {
			choice foobar {
				default foo;
				case foo {
					leaf foo {
						type string;
						default foo;
					}
				}
				case bar {
					leaf baz {
						type string;
					}
					leaf bar {
						type string;
						default bar;
					}
				}
			}
		}
		leaf default {
			type string;
			default default;
		}
	}
}`

	actual := getDataTreeWithDefaultsAsJSON(t, input_schema, input_json)
	expect := `{"testcontainer":{"baz":"data","bar":"bar"}}`
	assertMatch(t, expect, actual)
}
