// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile_test

import (
	"testing"

	. "github.com/danos/yang/testutils"
)

var featurePassTests = []TestCase{
	{
		Description: "Feature: Semi-colon terminated",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		`,
		ExpResult: true,
	},
	{
		Description: "Feature: With Description",
		Template:    BlankTemplate,
		Schema: `feature testfeature {
			description "Test feature";
		}
		`,
		ExpResult: true,
	},
	{
		Description: "Feature: With Description and if-features",
		Template:    BlankTemplate,
		Schema: `feature featureone;
		feature featuretwo;
		feature testfeature {
			if-feature featureone;
			description "Test feature";
			if-feature featuretwo;
		}
		`,
		ExpResult: true,
	},
}

var featureFailTests = []TestCase{
	{
		Description: "Invalid feature identifier; blank",
		Template:    BlankTemplate,
		Schema: `feature featureone;
		feature;
		`,
		ExpResult: false,
		ExpErrMsg: "invalid identifier:",
	},
	{
		Description: "Invalid feature identifier",
		Template:    BlankTemplate,
		Schema: `feature featureone;
		feature feature*one;
		`,
		ExpResult: false,
		ExpErrMsg: "invalid identifier: feature*one",
	},
	{
		Description: "Duplicate feature in a module",
		Template:    BlankTemplate,
		Schema: `feature featureone;
		feature featureone;
		`,
		ExpResult: false,
		ExpErrMsg: "Duplicate feature featureone",
	},
	{
		Description: "Feature cyclic reference via if-features (two features)",
		Template:    BlankTemplate,
		Schema: `feature featureone {
			if-feature featuretwo;
		}
		feature featuretwo {
			if-feature featureone;
		}
		`,
		ExpResult: false,
		ExpErrMsg: "Feature cyclic reference: test-yang-compile:featureone",
	},
	{
		Description: "Feature cyclic reference via if-features (four features)",
		Template:    BlankTemplate,
		Schema: `feature featureone {
			if-feature featuretwo;
		}
		feature featuretwo {
			if-feature featurethree;
		}
		feature featurethree {
			if-feature featurefour;
		}
		feature featurefour {
			if-feature featureone;
		}
		`,
		ExpResult: false,
		ExpErrMsg: "Feature cyclic reference: test-yang-compile:featureone",
	},
}

var ifFeaturePassTests = []TestCase{
	{
		Description: "Implicit local if-feature feature reference",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature {
			if-feature testfeature;
		}`,
		ExpResult: true,
	},
	{
		Description: "explicit local if-feature feature reference",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature {
			if-feature test:testfeature;
		}`,
		ExpResult: true,
	},
	{
		Description: "if-feature supported in container",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature;
		container testcontainer {
			description "Test container";
			if-feature testfeature;
			if-feature secondtestfeature;
		}`,
		ExpResult: true,
	},
	{
		Description: "if-feature supported in leaf",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature;
		container testcontainer {
			leaf testleaf {
				type string;
				if-feature testfeature;
				if-feature secondtestfeature;
			}
		}`,
		ExpResult: true,
	},
	{
		Description: "if-feature supported in leaf-list",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature;
		container testcontainer {
			leaf-list testleaf {
				type string;
				if-feature testfeature;
				if-feature secondtestfeature;
			}
		}`,
		ExpResult: true,
	},
	{
		Description: "if-feature supported in list",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature;
		container testcontainer {
			list testlist {
				key testkey;
				leaf testkey {
					type string;
				}
				if-feature testfeature;
				if-feature secondtestfeature;
			}
		}`,
		ExpResult: true,
	},
}

var ifFeatureFailTests = []TestCase{
	{
		Description: "if-feature invalid identifier",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature {
			if-feature;
		}`,
		ExpResult: false,
		ExpErrMsg: "invalid identifier",
	},
	{
		Description: "if-feature invalid identifier; bad prefix",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature {
			if-feature :testfeature;
		}`,
		ExpResult: false,
		ExpErrMsg: "invalid identifier",
	},
	{
		Description: "if-feature of non existent feature, " +
			"implicit local reference",
		Template: BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature {
			if-feature thirdtestfeature;
		}`,
		ExpResult: false,
		ExpErrMsg: "feature not valid: thirdtestfeature",
	},
	{
		Description: "if-feature of non existent feature, " +
			"explicit local reference",
		Template: BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature {
			if-feature test:thirdtestfeature;
		}`,
		ExpResult: false,
		ExpErrMsg: "feature not valid: test:thirdtestfeature",
	},
	{
		Description: "if-feature of non existent feature, remote reference",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature {
			if-feature bad-module:thirdtestfeature;
		}`,
		ExpResult: false,
		ExpErrMsg: "if-feature bad-module:thirdtestfeature: " +
			"unknown import bad-module",
	},
	{
		Description: "if-feature not allowed as a module substatement",
		Template:    BlankTemplate,
		Schema: `feature testfeature;
		feature secondtestfeature;
		if-feature testfeature;
		`,
		ExpResult: false,
		ExpErrMsg: "cardinality mismatch: invalid substatement 'if-feature'",
	},
}

func TestFeaturePass(t *testing.T) {
	runTestCases(t, featurePassTests)
}

func TestFeatureRejects(t *testing.T) {
	runTestCases(t, featureFailTests)
}

func TestIfFeaturePass(t *testing.T) {
	runTestCases(t, ifFeaturePassTests)
}

func TestIfFeatureRejects(t *testing.T) {
	runTestCases(t, ifFeatureFailTests)
}
