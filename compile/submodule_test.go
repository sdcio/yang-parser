// Copyright (c) 2017,2019 AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"testing"

	"github.com/steiler/yang-parser/testutils"
)

func TestSubmoduleSimple(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile {
		namespace "urn:vyatta.com:test:yang-compile";
		prefix test;

		include subone;

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}
	}`)

	submodule_text := bytes.NewBufferString(
		`submodule subone {
			belongs-to test-yang-compile { prefix test; }
			leaf one {
				type string;
			}
		}`)

	expected := NewLeafChecker("one")

	st, err := testutils.GetConfigSchema(
		module_text.Bytes(), submodule_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	if actual := findSchemaNodeInTree(t, st,
		[]string{"one"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestUsesGroupingInSubmodule(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile {
		namespace "urn:vyatta.com:test:yang-compile";
		prefix test;

		include subone;

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}
	}`)

	submodule_text := bytes.NewBufferString(
		`submodule subone {
			belongs-to test-yang-compile { prefix test; }
			grouping one {
				leaf one {
					type string;
				}
				container cont-one {
					leaf cont-leaf-one {
						type string;
					}
				}
			}
			uses one;
		}`)

	st, err := testutils.GetConfigSchema(
		module_text.Bytes(), submodule_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	expected := NewLeafChecker("one")
	if actual := findSchemaNodeInTree(t, st,
		[]string{"one"}); actual != nil {
		expected.check(t, actual)
	}

	expected = NewContainerChecker(
		"cont-one",
		[]NodeChecker{
			NewLeafChecker("cont-leaf-one"),
		})
	if actual := findSchemaNodeInTree(t, st,
		[]string{"cont-one"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestRefinesGroupingInSubmodule(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile {
			namespace "urn:vyatta.com:test:yang-compile";
			prefix test;

			include subone;

			organization "AT&T, Inc.";
			revision 2017-09-12 {
				description "Test schema";
			}
			grouping top-grouping {
				leaf top-grp-leaf {
					type string;
				}
			}
			uses top-grouping {
				refine top-grp-leaf {
					must "true()";
				}
			}
		}`)

	submodule_text := bytes.NewBufferString(
		`submodule subone {
			belongs-to test-yang-compile { prefix test; }
			grouping one-group-a {
				leaf one-leaf-a {
					type string;
				}
			}
			grouping one-group-b {
				leaf one-leaf-b {
					type string;
				}
			}
			uses one-group-a;
			uses one-group-b {
				refine one-leaf-b {
					must "true()";
				}
			}
		}`)

	expected := NewLeafChecker("one-leaf-b")

	st, err := testutils.GetConfigSchema(module_text.Bytes(), submodule_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}
	if actual := findSchemaNodeInTree(t, st,
		[]string{"one-leaf-b"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestUsingGroupingFromSubmodule(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile {
		namespace "urn:vyatta.com:test:yang-compile";
		prefix test;

		include subone;

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		uses one;
	}`)

	submodule_text := bytes.NewBufferString(
		`submodule subone {
			belongs-to test-yang-compile { prefix test; }
			grouping one {
				leaf one {
					type string;
				}
			}
		}`)

	expected := NewLeafChecker("one")

	st, err := testutils.GetConfigSchema(module_text.Bytes(), submodule_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}
	if actual := findSchemaNodeInTree(t, st,
		[]string{"one"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestUsingGroupingAcrossSubmodules(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile {
		namespace "urn:vyatta.com:test:yang-compile";
		prefix test;

		include subone;
		include subtwo;

		organization "Brocade Communications Systems, Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}
	}`)

	submodule1_text := bytes.NewBufferString(
		`submodule subone {
			belongs-to test-yang-compile { prefix test; }
			grouping one {
				leaf one {
					type string;
				}
			}
		}`)

	submodule2_text := bytes.NewBufferString(
		`submodule subtwo {
			belongs-to test-yang-compile { prefix test; }

			include subone;

			uses one;
		}`)

	expected := NewLeafChecker("one")

	st, err := testutils.GetConfigSchema(
		module_text.Bytes(),
		submodule1_text.Bytes(),
		submodule2_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}
	if actual := findSchemaNodeInTree(t, st,
		[]string{"one"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestAugmentAcrossSubmodules(t *testing.T) {
	module_text := bytes.NewBufferString(
		`module test-yang-compile {
		namespace "urn:vyatta.com:test:yang-compile";
		prefix test;

		include subone;
		include subtwo;

		organization "AT&T Inc.";
		revision 2017-08-29 {
			description "Test schema";
		}
	}`)

	submodule1_text := bytes.NewBufferString(
		`submodule subone {
			belongs-to test-yang-compile { prefix test; }
			container sub-one-container {
			}
		}`)

	submodule2_text := bytes.NewBufferString(
		`submodule subtwo {
			belongs-to test-yang-compile { prefix test; }

			include subone;

			augment /sub-one-container {
			    leaf sub-two-leaf {
					type string;
				}
			}
		}`)

	expected := NewLeafChecker("sub-two-leaf")

	st, err := testutils.GetConfigSchema(
		module_text.Bytes(),
		submodule1_text.Bytes(),
		submodule2_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}
	if actual := findSchemaNodeInTree(t, st,
		[]string{"sub-one-container", "sub-two-leaf"}); actual != nil {
		expected.check(t, actual)
	}
}
