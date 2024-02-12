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

// Copyright (c) 2017,2019, AT&T Intellectual Property.  All rights reserved.
//
// Copyright (c) 2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file provides test utilities for creating full schemas from snippets
// of YANG.  This allows tests to clearly specify the YANG that is being
// tested, without it being lost in the noise of the boilerplate YANG.
// Furthermore, it easily allows for multiple modules.

package testutils

import (
	"fmt"
)

const schemaImportTemplate = `
	import %s {
	    prefix %s;
    }
`

const schemaIncludeTemplate = `
	include %s;
`

const schemaModuleTemplate = `
module %s {
	namespace "urn:vyatta.com:test:%s";
	prefix %s;
    %s
    %s
	organization "Brocade Communications Systems, Inc.";
	contact
		"Brocade Communications Systems, Inc.
		 Postal: 130 Holger Way
		         San Jose, CA 95134
		 E-mail: support@Brocade.com
		 Web: www.brocade.com";
	revision 2014-12-29 {
		description "Test schema for configd";
	}
	%s
}
`

const schemaSubmoduleTemplate = `
submodule %s {
	belongs-to %s {
		prefix %s;
	}
	%s
	%s
}
`

// Used for creating tests with multiple modules without resorting to reading
// them in from file as this means you can't read the schema and the test
// together easily.
type TestSchema struct {
	Name          NameDef
	Imports       []NameDef
	Includes      []string
	BelongsTo     NameDef
	Prefix        string
	SchemaSnippet string
}

type NameDef struct {
	Namespace string
	Prefix    string
}

func NewTestSchema(namespace, prefix string) *TestSchema {
	return &TestSchema{Name: NameDef{Namespace: namespace, Prefix: prefix}}
}

func (ts *TestSchema) AddInclude(module string) *TestSchema {
	ts.Includes = append(ts.Includes, module)
	return ts
}

func (ts *TestSchema) AddBelongsTo(namespace, prefix string) *TestSchema {
	ts.BelongsTo.Namespace = namespace
	ts.BelongsTo.Prefix = prefix
	return ts
}

func (ts *TestSchema) AddSchemaSnippet(snippet string) *TestSchema {
	ts.SchemaSnippet = snippet
	return ts
}

func ConstructSchema(schemaDef TestSchema) (schema string) {
	var importStr, includeStr string

	for _, inc := range schemaDef.Includes {
		includeStr = includeStr + fmt.Sprintf(schemaIncludeTemplate, inc)
	}

	if schemaDef.BelongsTo.Namespace != "" {
		schema = fmt.Sprintf(schemaSubmoduleTemplate,
			schemaDef.Name.Namespace,
			schemaDef.BelongsTo.Namespace, schemaDef.BelongsTo.Prefix,
			includeStr, schemaDef.SchemaSnippet)
	} else {
		for _, imp := range schemaDef.Imports {
			importStr = importStr + fmt.Sprintf(schemaImportTemplate,
				imp.Namespace, imp.Prefix)
		}

		schema = fmt.Sprintf(schemaModuleTemplate,
			schemaDef.Name.Namespace, schemaDef.Name.Namespace,
			schemaDef.Name.Prefix, importStr, includeStr,
			schemaDef.SchemaSnippet)
	}

	return schema
}
