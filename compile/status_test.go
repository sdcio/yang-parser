// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/steiler/yang-parser/schema"
	"github.com/steiler/yang-parser/testutils"
)

func CheckStatus(expect schema.Status) checkFn {
	return func(t *testing.T, actual schema.Node) {
		if actual.Status() != expect {
			t.Errorf("Config mismatch for %s\n    expect: %s\n    actual: %s\n",
				actual.Name(), expect.String(), actual.Status().String())
		}
	}
}

func ExpectSuccess(t *testing.T, expected NodeChecker, input string) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate, input))

	actual := getSchemaNodeFromPath(t, schema_text, []string{expected.Name})

	expected.check(t, actual)
}

func ExpectFailure(t *testing.T, expected string, input string) {

	schema_text := bytes.NewBufferString(fmt.Sprintf(
		SchemaTemplate, input))

	_, err := testutils.GetConfigSchema(schema_text.Bytes())

	assertErrorContains(t, err, expected)
}

func TestStatusOnLeaf(t *testing.T) {

	input := `container c1 {
			leaf def {
				type string;
			}
			leaf cur {
				type string;
				status current;
			}
			leaf dep {
				type string;
				status deprecated;
			}
			leaf obs {
				type string;
				status obsolete;
			}
		}`

	expected := NewContainerChecker(
		"c1",
		[]NodeChecker{
			NewLeafChecker("def",
				CheckStatus(schema.Current)),
			NewLeafChecker("cur",
				CheckStatus(schema.Current)),
			NewLeafChecker("dep",
				CheckStatus(schema.Deprecated)),
			NewLeafChecker("obs",
				CheckStatus(schema.Obsolete)),
		})

	ExpectSuccess(t, expected, input)
}

func TestInheritedDeprecatedStatus(t *testing.T) {

	input := `container dep {
			status deprecated;
			leaf def {
				type string;
			}
			leaf dep {
				type string;
				status deprecated;
			}
			leaf obs {
				type string;
				status obsolete;
			}}`

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("def",
				CheckStatus(schema.Deprecated)),
			NewLeafChecker("dep",
				CheckStatus(schema.Deprecated)),
			NewLeafChecker("obs",
				CheckStatus(schema.Obsolete)),
		})

	ExpectSuccess(t, expected, input)
}

func TestInheritedObsolete(t *testing.T) {

	input := `container obs {
			status obsolete;
			leaf def {
				type string;
			}
			leaf obs {
				type string;
				status obsolete;
			}}`

	expected := NewContainerChecker(
		"obs",
		[]NodeChecker{
			NewLeafChecker("def",
				CheckStatus(schema.Obsolete)),
			NewLeafChecker("obs",
				CheckStatus(schema.Obsolete)),
		})

	ExpectSuccess(t, expected, input)
}

func TestFailedOverride(t *testing.T) {

	input := `container obs {
			status obsolete;
			leaf def {
				type string;
			}
			leaf obs {
				type string;
				status deprecated;
			}}`

	expected := "status deprecated: Cannot override status of parent"

	ExpectFailure(t, expected, input)
}

func TestDeprecatedGroupingStatusOutwithModule(t *testing.T) {

	module1_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test1;

		organization "AT&T Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		grouping dep {
			status deprecated;
			container dep {
				leaf test {
					type empty;
				}
				leaf test2 {
					status deprecated;
					type empty;
				}
			}
		}
	}`)

	module2_text := bytes.NewBufferString(
		`module test-yang-compile2 {
		namespace "urn:vyatta.com:test:yang-compile2";
		prefix test;

		import test-yang-compile1 {
			prefix mod1;
		}

		organization "AT&T Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		uses mod1:dep;
	}`)

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("test",
				CheckStatus(schema.Current)),
			NewLeafChecker("test2",
				CheckStatus(schema.Deprecated)),
		})

	st, err := testutils.GetConfigSchema(module1_text.Bytes(), module2_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	if actual := findSchemaNodeInTree(t, st, []string{"dep"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestDeprecatedUsesIsInherited(t *testing.T) {

	input :=
		`grouping dep {
			container dep {
				leaf test {
					type empty;
				}
			}
		}

		uses dep {
			status deprecated;
		}`

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("test",
				CheckStatus(schema.Deprecated)),
		})

	ExpectSuccess(t, expected, input)
}

func TestDeprecatedGroupingFromDeprecatedUsesWithinModule(t *testing.T) {

	input :=
		`grouping dep {
			status deprecated;
			container dep {
				leaf test {
					type empty;
				}
			}
		}

		uses dep {
			status deprecated;
		}`

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("test",
				CheckStatus(schema.Deprecated)),
		})

	ExpectSuccess(t, expected, input)
}

func TestDeprecatedGroupingFromInheritedDeprecatedUsesWithinModule(t *testing.T) {

	input :=
		`grouping dep {
			status deprecated;
			container dep {
				leaf test {
					type empty;
				}
			}
		}

		container top {
			status deprecated;
			uses dep;
		}`

	expected := NewContainerChecker(
		"top",
		[]NodeChecker{
			NewContainerChecker(
				"dep",
				[]NodeChecker{
					NewLeafChecker("test",
						CheckStatus(schema.Deprecated)),
				})})

	ExpectSuccess(t, expected, input)
}

func TestDeprecatedGroupingStatusWithinModule(t *testing.T) {

	input :=
		`grouping dep {
			status deprecated;
			container dep {
				leaf test {
					type empty;
				}
			}
		}

		uses dep;`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestDeprecatedAugmentIsInherited(t *testing.T) {

	input :=
		`container dep;
		augment "/dep" {
			status deprecated;
			leaf test {
				type empty;
			}

		}`

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("test",
				CheckStatus(schema.Deprecated)),
		})

	ExpectSuccess(t, expected, input)
}

func TestDeprecatedAugmentReferenceOutwithModule(t *testing.T) {

	module1_text := bytes.NewBufferString(
		`module test-yang-compile1 {
		namespace "urn:vyatta.com:test:yang-compile1";
		prefix test1;

		organization "AT&T Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		container dep {
			status deprecated;
		}
	}`)

	module2_text := bytes.NewBufferString(
		`module test-yang-compile2 {
		namespace "urn:vyatta.com:test:yang-compile2";
		prefix test;

		import test-yang-compile1 {
			prefix mod1;
		}

		organization "AT&T Inc.";
		revision 2014-12-29 {
			description "Test schema";
		}

		augment "/mod1:dep" {
			leaf test {
				type empty;
			}
		}
	}`)

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("test",
				CheckStatus(schema.Deprecated)),
		})

	st, err := testutils.GetConfigSchema(module1_text.Bytes(), module2_text.Bytes())
	if err != nil {
		t.Fatalf("Unexpected error %s", err.Error())
	}

	if actual := findSchemaNodeInTree(t, st, []string{"dep"}); actual != nil {
		expected.check(t, actual)
	}
}

func TestAugmentReferenceDeprecatedWithinModule(t *testing.T) {

	input :=
		`container dep {
			status deprecated;
		}

		augment "/dep" {
			leaf test {
				type empty;
			}
		}`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestAugmentReferenceInheritedDeprecatedWithinModule(t *testing.T) {

	input :=
		`container dep {
			status deprecated;
			container dep {
			}
		}

		augment "/dep/dep" {
			leaf test {
				type empty;
			}
		}`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestDeprecatedAugmentReferenceDeprecatedWithinModule(t *testing.T) {

	input :=
		`container dep {
			status deprecated;
		}

		augment "/dep" {
			status deprecated;
			leaf test {
				type empty;
			}
		}`

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("test",
				CheckStatus(schema.Deprecated)),
		})

	ExpectSuccess(t, expected, input)
}

func TestInheritedDeprecatedAugmentReferenceDeprecatedWithinModule(t *testing.T) {

	input :=
		`grouping depgroup {
			container dep {
				status deprecated;
			}
		}

		uses depgroup {
			status deprecated;
			augment "dep" {
				leaf test {
					type empty;
				}
			}
		}`

	expected := NewContainerChecker(
		"dep",
		[]NodeChecker{
			NewLeafChecker("test",
				CheckStatus(schema.Deprecated)),
		})

	ExpectSuccess(t, expected, input)
}

func TestReferenceDeprecatedTypedef(t *testing.T) {

	input :=
		`typedef dep {
			type string;
			status deprecated;
		}

		leaf test {
			type dep;
		}`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestDeprecatedReferenceDeprecatedTypedef(t *testing.T) {

	input :=
		`typedef dep {
			type string;
			status deprecated;
		}

		leaf test {
			status deprecated;
			type dep;
		}`

	ExpectSuccess(t, NodeChecker{"test", nil}, input)
}

func TestLoseInheritedDeprecatedReferenceDeprecatedTypedef(t *testing.T) {

	input :=
		`typedef inner {
			type string;
			status deprecated;
		}

		typedef dep {
			type inner;
		}

		leaf test {
			status deprecated;
			type dep;
		}`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestUnionInheritedDeprecatedReferenceDeprecatedTypedef(t *testing.T) {

	input :=
		`typedef dep {
			type string;
			status deprecated;
		}

		leaf test {
			status deprecated;
			type union {
				type dep;
			}
		}`

	ExpectSuccess(t, NodeChecker{"test", nil}, input)
}

func TestReferenceDeprecatedIfFeature(t *testing.T) {

	input :=
		`feature dep {
			status deprecated;
		}
		leaf test {
			type string;
			if-feature dep;
		}`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestDeprecatedReferenceDeprecatedIfFeature(t *testing.T) {

	input :=
		`feature dep {
			status deprecated;
		}
		leaf dummy {
			type empty;
			description "Required so we have a node to check";
		}
		leaf test {
			type string;
			if-feature dep;
			description "Compiles, but feature is disabled";
			status deprecated;
		}`

	ExpectSuccess(t, NodeChecker{"dummy", nil}, input)
}

func TestFeatureReferenceDeprecatedIfFeature(t *testing.T) {

	input :=
		`feature dep {
			status deprecated;
		}
		feature feat1 {
			if-feature dep;
		}`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestDeprecatedFeatureReferenceDeprecatedIfFeature(t *testing.T) {

	input :=
		`feature dep {
			status deprecated;
		}
		feature feat1 {
			status deprecated;
			if-feature dep;
		}
		leaf dummy {
			type empty;
		}`

	ExpectSuccess(t, NodeChecker{"dummy", nil}, input)
}

func TestIdentityReferenceDeprecatedIdentity(t *testing.T) {

	input :=
		`identity foo {
			status obsolete;
		}
		identity bar {
			base foo;
		}`

	expected := "Current node cannot reference Obsolete node within same module"

	ExpectFailure(t, expected, input)
}
func TestIdentityReferenceObsoleteIdentity(t *testing.T) {

	input :=
		`identity foo {
			status deprecated;
		}
		identity bar {
			base foo;
		}`

	expected := "Current node cannot reference Deprecated node within same module"

	ExpectFailure(t, expected, input)
}

func TestDeprecatedIdentityReferenceObsoleteIdentity(t *testing.T) {

	input :=
		`identity foo {
			status obsolete;
		}
		identity bar {
			base foo;
			status deprecated;
		}`

	expected := "Deprecated node cannot reference Obsolete node within same module"

	ExpectFailure(t, expected, input)
}

func TestValidIdentityrefSuccess(t *testing.T) {

	input :=
		`identity foo {
			status current;
		}
		identity bar {
			base foo;
			status deprecated;
		}
		identity foobar {
			base bar;
			status obsolete;
		}
		leaf current {
			type identityref {
				base foo;
			}
		}
		leaf deprecated {
			status deprecated;
			type identityref {
				base bar;
			}
		}
		leaf obsolete {
			status obsolete;
			type identityref {
				base foobar;
			}
		}`

	ExpectSuccess(t, NodeChecker{"current", nil}, input)
}

func TestIdentityrefReferenceDeprecated(t *testing.T) {

	input :=
		`identity foo {
			status deprecated;
		}
		identity bar {
			base foo;
			status deprecated;
		}
		identity foobar {
			base bar;
			status obsolete;
		}
		leaf dummy {
			type identityref {
				base foo;
			}
		}`

	expected := "Current node cannot reference Deprecated node within same module"
	ExpectFailure(t, expected, input)
}

func TestInvalidCurrentIdentityrefReferenceObsoleteIdentity(t *testing.T) {

	input :=
		`identity foo {
			status obsolete;
		}
		identity bar {
			base foo;
			status obsolete;
		}
		identity foobar {
			base bar;
			status obsolete;
		}
		leaf dummy {
			type identityref {
				base foo;
			}
		}`

	expected := "Current node cannot reference Obsolete node within same module"
	ExpectFailure(t, expected, input)
}

func TestInvalidDeprecatedIdentityrefReferenceObsoleteIdentity(t *testing.T) {

	input :=
		`identity foo {
			status obsolete;
		}
		identity bar {
			base foo;
			status obsolete;
		}
		identity foobar {
			base bar;
			status obsolete;
		}
		typedef newtype {
			type identityref {
				base foo;
			}
			status obsolete;
		}
		container levelone {
			status deprecated;

			container leveltwo {
				leaf dummy {
					// status inherited from levelone
					type newtype;
				}
			}
		}`

	expected := "Deprecated node cannot reference Obsolete node within same module"
	ExpectFailure(t, expected, input)
}
