// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse_test

import (
	"strings"
	"testing"

	"github.com/iptecharch/yang-parser/parse"
)

func mkModule(revs string) string {
	const moduleHdr = `
	namespace "urn:vyatta.com:mgmt:parse-test";
	prefix module-test;

	organization "Brocade Communications Systems, Inc.";
	contact "Brocade Communications Systems, Inc.
		 Postal: 130 Holger Way
			 San Jose, CA 95134
		 E-mail: support@Brocade.com
		 Web: www.brocade.com";
`
	return "module module-test {" + moduleHdr + revs + "}"
}

func mkRevision(date string) string {
	return `	revision ` + date + ` {
		description "revision";
	}
`
}

func verifyParseError(t *testing.T, mod, exp string) {
	_, err := parse.Parse("verifyParseError", mod, nil)
	if err == nil {
		t.Errorf("Unexpected successful parse")
	} else if !strings.Contains(err.Error(), exp) {
		t.Errorf("Unexpected parse error. Expected [%s], Got [%s]",
			exp, err.Error())
	}
}

func TestMissingRevision(t *testing.T) {
	_, err := parse.Parse("TestMissingRevision", mkModule(""), nil)
	if err != nil {
		t.Errorf("Unable to parse basic module (%s)", err.Error())
	}
}

func TestInvalidRevision(t *testing.T) {
	var date = "15-12-16"
	var exp = "invalid date: " + date
	rev := mkRevision(date)
	mod := mkModule(rev)
	verifyParseError(t, mod, exp)
}

func TestDuplicateRevision(t *testing.T) {
	var date = "2015-12-16"
	var exp = "duplicated revision date " + date
	rev1 := mkRevision(date)
	rev2 := mkRevision(date)
	mod := mkModule(rev1 + rev2)
	verifyParseError(t, mod, exp)
}

func TestBadOrderRevision(t *testing.T) {
	var date1 = "2014-12-16"
	var date2 = "2015-12-16"
	var exp = "revision block out of order " + date2
	rev1 := mkRevision(date1)
	rev2 := mkRevision(date2)
	mod := mkModule(rev1 + rev2)
	verifyParseError(t, mod, exp)
}

func TestGoodOrderRevision(t *testing.T) {
	var date1 = "2015-12-16"
	var date2 = "2014-12-16"
	rev1 := mkRevision(date1)
	rev2 := mkRevision(date2)
	mod := mkModule(rev1 + rev2)
	_, err := parse.Parse("TestGoodOrderRevision", mod, nil)
	if err != nil {
		t.Errorf("Unable to parse good revision order (%s)", err.Error())
	}
}
