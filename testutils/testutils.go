// Copyright (c) 2017,2019 AT&T Intellectual Property
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// Utilities used by configd unit tests.

package testutils

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/sdcio/yang-parser/compile"
	"github.com/sdcio/yang-parser/parse"
	"github.com/sdcio/yang-parser/schema"
	"github.com/sdcio/yang-parser/xpath/xutils"
)

// Generic Test Case description allowing for cases where expected result
// may be acceptance or rejection.  In the latter case we verify the error
// generated against the expected error (empty expected message causes test
// to fail).  For tests we expect to pass, we can provide a set of node(s)
// with associated property/ies to be validated, or we can specify 'nil'.
//
// Schemas can be provided in 2 different ways:
//
// (a) Single schema + associated 'template'.  Useful when testing a small
//
//	snippet of schema in a single module context where you don't want
//	to keep duplicating YANG across multiple tests.
//
//	=> Use Schema + Template fields
//
// (b) Multiple schemas.  Useful for testing multiple modules.  Each
//
//	TestSchema needs to provide enough YANG to be pasted into a
//	generic template that contains a module, prefix and imports, but
//	nothing else.
//
//	=> Use Schemas
type TestCase struct {
	Description     string            // What we are testing
	Template        string            // Template for schema
	Schema          string            // Schema snippet under test
	Schemas         []TestSchema      // Alternative to schema/template
	ExpResult       bool              // true = schema must be accepted
	ExpErrMsg       string            // Can't use empty string.  Use either
	ExpErrs         []string          // ErrMsg or Errs.
	NodesToValidate []schema.NodeSpec // Nil if no nodes to validate
}

func LogStack(t *testing.T) {
	LogStackInternal(t, false)
}

func LogStackFatal(t *testing.T) {
	LogStackInternal(t, true)
}

func LogStackInternal(t *testing.T, fatal bool) {
	stack := make([]byte, 4096)
	runtime.Stack(stack, false)
	if fatal {
		t.Fatalf("%s", stack)
	} else {
		t.Logf("%s", stack)
	}
}

func nilExtCardinality(ntype parse.NodeType) map[parse.NodeType]parse.Cardinality {
	return map[parse.NodeType]parse.Cardinality{}
}

// Create ModelSet structure from multiple buffers, each buffer
// represents a single yang module.
func getSchema(getFullSchema, skipUnknown bool, bufs ...[]byte,
) (schema.ModelSet, error) {
	ms, _, err := getSchemaWithWarns(getFullSchema, skipUnknown, bufs...)
	return ms, err
}

func getSchemaWithWarns(getFullSchema, skipUnknown bool, bufs ...[]byte,
) (schema.ModelSet, []xutils.Warning, error) {

	const name = "schema"
	modules := make(map[string]*parse.Tree)
	for index, b := range bufs {
		t, err := parse.Parse(name+strconv.Itoa(index), string(b),
			nilExtCardinality)
		if err != nil {
			return nil, nil, err
		}
		mod := t.Root.Argument().String()
		modules[mod] = t
	}
	st, warns, err := compile.CompileModulesWithWarnings(
		nil, modules, "", skipUnknown,
		compile.Include(compile.IsConfig,
			compile.IncludeState(getFullSchema)))
	if err != nil {
		return nil, warns, err
	}
	return st, warns, nil
}

func GetConfigSchema(buf ...[]byte) (schema.ModelSet, error) {
	return getSchema(false, false, buf...)
}

func GetConfigSchemaWithWarns(buf ...[]byte,
) (schema.ModelSet, []xutils.Warning, error) {
	return getSchemaWithWarns(false, false, buf...)
}

func GetConfigSchemaSkipUnknown(buf ...[]byte) (schema.ModelSet, error) {
	return getSchema(false, true, buf...)
}

func GetFullSchema(buf ...[]byte) (schema.ModelSet, error) {
	return getSchema(true, false, buf...)
}

// Set of helper functions to produce correctly formatted config and that
// isolates test code that relies on correctly formatted config from subsequent
// format changes so that we are testing content not format.
func Prefix(entry, pfx string) string {
	tmp := strings.Replace(entry, "\n", "\n"+pfx, -1)
	return pfx + tmp[:len(tmp)-len(pfx)]
}

func Tab(entry string) string {
	return Prefix(entry, "\t")
}

func Add(entry string) string {
	return Prefix(entry, "+")
}

func Rem(entry string) string {
	return Prefix(entry, "-")
}

// Initially the +/- for changed lines get added right in front of the
// element being changed.  This function pulls them to the front of the line
// and inserts a leading space on unchanged lines.  Completely blank lines
// (other than leading tabs) do NOT get a leading space.
func FormatAsDiff(entry string) (diffs string) {
	lines := strings.Split(entry, "\n")
	for _, line := range lines {
		trimmed := strings.Trim(line, "\t")
		if len(trimmed) > 0 {
			if trimmed[0] == '+' || trimmed[0] == '-' {
				// Iteratively move + or - ahead of tabs
				for line[0] == '\t' {
					line = strings.Replace(line, "\t+", "+\t", 1)
					line = strings.Replace(line, "\t-", "-\t", 1)
				}
			} else {
				line = " " + line
			}
		}
		diffs += line + "\n"
	}
	return diffs
}

// ListEntries and Containers are handled exactly the same way.
func contOrListEntry(name string, entries []string) (retStr string) {
	retStr = name
	if len(entries) == 0 {
		return retStr + "\n"
	}

	retStr += " {\n"
	for _, entry := range entries {
		retStr += Tab(entry)
	}
	retStr += "}\n"
	return retStr
}

func Root(rootEntries ...string) (rootStr string) {
	return strings.Join(rootEntries, "")
}

func Cont(name string, contEntries ...string) (contStr string) {
	return contOrListEntry(name, contEntries)
}

func ListEntry(name string, leaves ...string) (listEntryStr string) {
	return contOrListEntry(name, leaves)
}

func listOrLeafList(name string, entries []string) (retStr string) {
	for _, entry := range entries {
		// Deal with +/- prefix
		if entry[0] == '+' || entry[0] == '-' {
			retStr += fmt.Sprintf("%c%s %s", entry[0], name, entry[1:])
		} else {
			retStr += name + " " + entry
		}
	}
	return retStr
}

func List(name string, listEntries ...string) (listStr string) {
	return listOrLeafList(name, listEntries)
}

func LeafList(name string, leafListEntries ...string) (leafListStr string) {
	return listOrLeafList(name, leafListEntries)
}

func LeafListEntry(name string) string { return name + "\n" }

func Leaf(name, value string) string { return name + " " + value + "\n" }

func EmptyLeaf(name string) string { return name + "\n" }
