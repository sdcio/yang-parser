// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains tests on restrictions available to types

package compile_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/iptecharch/yang-parser/testutils"
)

func buildRestrictionSchema(typeType string, restriction RestType) []byte {

	extra := ""

	// decimal64 requires fraction-digits to test other restrictions
	if typeType == "decimal64" && restriction != frc {
		extra = string(frc) + ";"
	}

	// leafref requires path to test other restrictions
	if typeType == "leafref" && restriction != pth {
		extra = string(pth) + ";"
	}

	testLeaf := fmt.Sprintf(
		`leaf testLeaf {
			type %s {
                %s
				%s;
			}
		}`, typeType, extra, restriction)

	return bytes.NewBufferString(
		fmt.Sprintf(SchemaTemplate, testLeaf)).Bytes()
}

func checkInvalidRestriction(t *testing.T, typeType string, restriction RestType) {

	schema_text := buildRestrictionSchema(typeType, restriction)
	_, err := testutils.GetConfigSchema(schema_text)

	expected := fmt.Sprintf(
		"type %s: %s restriction is not valid for this type",
		typeType, restriction)
	if typeType == "union" {
		expected = fmt.Sprintf(
			"type union: cannot restrict %s of a union type - "+
				"restrictions must be applied to members instead",
			restriction)
	}

	assertErrorContains(t, err, expected)
}

func checkValidRestriction(t *testing.T, typeType string, restriction RestType) {

	schema_text := buildRestrictionSchema(typeType, restriction)
	_, err := testutils.GetConfigSchema(schema_text)

	text := fmt.Sprintf(
		"Testing type %s: %s restriction is valid for this type",
		typeType, restriction)

	assertSuccess(t, text, err)
}

type RestType string

const (
	lng RestType = "length 0..1"
	rng RestType = "range 0..1"
	frc RestType = "fraction-digits 2"
	pat RestType = "pattern [a-z]"
	enm RestType = "enum foo"
	typ RestType = "type int64"
	req RestType = "require-instance true"
	bit RestType = "bit foo"
	pth RestType = "path /foo/bar"
)

func checkRestrictions(t *testing.T, typeType string, valid []RestType) {
	here := struct{}{}
	invalid := map[RestType]struct{}{
		lng: here,
		rng: here,
		frc: here,
		pat: here,
		typ: here,
		req: here,
		bit: here,
		pth: here,
	}

	for _, v := range valid {
		checkValidRestriction(t, typeType, v)
		delete(invalid, v)
	}

	for v, _ := range invalid {
		checkInvalidRestriction(t, typeType, v)
	}
}

func TestRestrictions(t *testing.T) {
	checkRestrictions(t, "boolean", []RestType{})
	checkRestrictions(t, "empty", []RestType{})
	checkRestrictions(t, "enumeration", []RestType{enm})
	checkRestrictions(t, "identityref", []RestType{})
	checkRestrictions(t, "int16", []RestType{rng})
	checkRestrictions(t, "uint64", []RestType{rng})
	checkRestrictions(t, "decimal64", []RestType{frc, rng})
	checkRestrictions(t, "bits", []RestType{bit})
	checkRestrictions(t, "leafref", []RestType{pth})
	checkRestrictions(t, "union", []RestType{typ})
	checkRestrictions(t, "string", []RestType{lng, pat})
	checkRestrictions(t, "instance-identifier", []RestType{req})
}
