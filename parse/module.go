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

// Copyright (c) 2018-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-15 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse

import (
	"fmt"
	"time"
)

func checkModule(n Node) error {
	const (
		HDR int = iota
		LINK
		META
		REV
		BODY
	)
	prev := HDR
	for _, c := range n.Children() {
		switch c.Type() {
		case NodeUnknown:
		case NodeYangVersion, NodeNamespace, NodePrefix, NodeBelongsTo:
			if prev != HDR {
				return fmt.Errorf("unexpected header statement %s", c)
			}
		case NodeImport, NodeInclude:
			if prev > LINK {
				return fmt.Errorf("unexpected linkage statement %s", c)
			}
			prev = LINK
		case NodeOrganization, NodeContact, NodeDescription, NodeReference:
			if prev > META {
				return fmt.Errorf("unexpected meta statement %s", c)
			}
			prev = META
		case NodeRevision:
			if prev > REV {
				return fmt.Errorf("unexpected revision statement %s", c)
			}
			prev = REV
		default:
			if prev > BODY {
				return fmt.Errorf("unexpected body statement %s", c)
			}
			prev = BODY
		}
	}
	return nil
}

func checkRevisionOrder(n Node) error {
	const dateSuffix = "T00:00:00Z"
	rev, _ := time.Parse(time.RFC3339, "9999-12-31T23:59:59Z")
	for _, c := range n.Children() {
		switch c.Type() {
		case NodeRevision:
			date := c.ArgDate()
			fullDate := date + dateSuffix
			thisRev, err := time.Parse(time.RFC3339, fullDate)
			if err != nil {
				return fmt.Errorf("invalid revision date %s", date)
			}
			if thisRev.After(rev) {
				return fmt.Errorf("revision block out of order %s", date)
			} else if thisRev == rev {
				return fmt.Errorf("duplicated revision date %s", date)
			}
			rev = thisRev
		}
	}
	return nil
}

type Module struct {
	tree  *Tree
	mod   Node
	smods map[string]Node
}

func (m *Module) GetTree() *Tree                 { return m.tree }
func (m *Module) GetModule() Node                { return m.mod }
func (m *Module) GetSubmodules() map[string]Node { return m.smods }

func (m *Module) Imports() (imports map[string]string) {
	imports = make(map[string]string)
	for _, i := range m.mod.ChildrenByType(NodeImport) {
		imports[i.Name()] = i.ChildByType(NodePrefix).Name()
	}
	return imports
}

func (m *Module) Prefix() string {
	pfxNode := m.mod.ChildByType(NodePrefix)
	if pfxNode != nil {
		return pfxNode.Name()
	}
	return ""
}

func GetModules(trees map[string]*Tree) map[string]*Module {
	mods, _ := GetModulesAndSubmodules(trees)
	return mods
}

func GetModulesAndSubmodules(
	mods map[string]*Tree,
) (
	map[string]*Module,
	map[string]*Module,
) {
	modules := make(map[string]*Module)
	submodules := make(map[string]*Module, 0)
	for mn, m := range mods {
		switch m.Root.Type() {
		case NodeModule:
			modules[mn] = &Module{
				mod:   m.Root,
				tree:  m,
				smods: make(map[string]Node),
			}
		case NodeSubmodule:
			submodules[mn] = &Module{
				mod:   m.Root,
				tree:  m,
				smods: nil,
			}
		default:
			continue
		}
	}

	return modules, submodules
}

func createFakeModule(name string) *Module {
	fakeModule := fmt.Sprintf(`module %s {
	namespace "urn:vyatta.com:fake:%s";
	prefix fake;

	organization "Brocade Communications Systems, Inc.";
	contact "Brocade, Fake Set";
	revision 2015-08-27 {
		description "Initial revision.";
	}
}`, name, name)

	t, err := Parse(name, fakeModule, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse fake module: %s", err.Error()))
	}

	return &Module{
		mod:   t.Root,
		tree:  t,
		smods: make(map[string]Node),
	}
}
