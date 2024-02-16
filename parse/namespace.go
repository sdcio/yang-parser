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

// Copyright (c) 2017-2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse

import (
	"fmt"
)

type Namespace interface {
	GetNodeNamespace(mod Node, modules map[string]*Module) string
	GetNodeModulename(mod Node) string
	GetNodeSubmoduleName() string
	GetModuleByPrefix(
		pfx string,
		modules map[string]*Module,
		skipUnknown bool,
	) (Node, error)
	YangPrefixToNamespace(
		prefix string,
		modules map[string]*Module,
		skipUnknown bool,
	) (string, error)
}

func getPfxName(n Node, pfx string) (string, bool) {
	for _, i := range n.ChildrenByType(NodeImport) {
		if i.Prefix() == pfx {
			return i.Name(), true
		}
	}
	return "", false
}

func getSubmoduleNamespace(
	n Node,
	modules map[string]*Module,
) (string, error) {

	// Assumes we have already checked node is of type NodeSubmodule!
	belongs := n.ChildByType(NodeBelongsTo).Name()
	if mod, ok := modules[belongs]; ok {
		return mod.mod.Ns(), nil
	}

	return "", fmt.Errorf("Unable to get namespace for submodule %s.\n",
		n.Name())
}

func (n *node) UsesRoot() Node {
	if n.useTree == nil {
		return n.Root()
	}
	return n.useTree.Root
}

// Get the correct namespace for a node.
// Nodes defined in a group have a second root associated with them
// called UsesRoot. UsesRoot will be nil for any other node
// If defined, return the namespace of the UsesRoot to ensure that
// grouping derived nodes belong to the namespace where they are used.
// All other node types will use the namespace in which they are
// defined. Root() is the module in which they were defined.
func getNodeNamespaceInternal(
	n Node,
	modules map[string]*Module,
) (string, error) {

	if ur := n.UsesRoot(); ur != nil {
		// Return namespace of where node is used, as this was
		// originally from a grouping.

		if ur.Type() == NodeSubmodule {
			return getSubmoduleNamespace(ur, modules)
		}
		return ur.Ns(), nil
	} else if n.Root() != nil {
		// Return namespace in which node was defined.
		return n.Root().Ns(), nil
	}
	return "", fmt.Errorf("Unable to get namespace for %s.", n.Name())
}

func (n *node) GetNodeNamespace(
	m Node,
	modules map[string]*Module,
) string {
	if ns, err := getNodeNamespaceInternal(n, modules); err == nil {
		return ns
	}
	// Some unit tests, and some compiler/parser failure scenarios can
	// result in no Root() or UsesRoot() being defined, so return the name
	// of the module in which the node is used to preserve some sanity.
	// This path is only expected to be used in failure or unit
	// test scenarios.
	// Just in case 'm' is also nil, return an empty namespace in this case.
	if m == nil {
		return ""
	}
	return m.Ns()
}

// Get the correct module name for a node.
// Nodes defined in a group, have a second root associated with them
// called UsesRoot. UsesRoot will be nil for any other node.
// If defined, return the module name of the UsesRoot to ensure that
// grouping derived nodes belong to the module name where they are used.
// All other node types will use the module name in which they are
// defined. Root() is the module in which they were defined.
func getNodeModulenameInternal(n Node) (string, error) {
	if n.UsesRoot() != nil {
		// Return module name of where node is used, as this was
		// originally from a grouping
		return n.UsesRoot().Name(), nil
	} else if n.Root() != nil {
		// Return module name in which node was defined.
		return n.Root().Name(), nil
	}
	return "", fmt.Errorf("Unable to get module name for %s.", n.Name())
}

func (n *node) GetNodeModulename(mod Node) string {
	if ns, err := getNodeModulenameInternal(n); err == nil {
		return ns
	}
	// Some unit tests, and some compiler/parser failure scenarios can
	// result in no Root() or UsesRoot() being defined, return the name
	// of the module in which the node is used to preserve some sanity.
	// This path is only expected to be used in failure or unit
	// test scenarios
	return mod.Name()
}

func (n *node) GetNodeSubmoduleName() string {

	if ur := n.UsesRoot(); ur != nil {
		// Return namespace of where node is used, as this was
		// originally from a grouping.

		if ur.Type() == NodeSubmodule {
			return ur.Name()
		}
	} else if r := n.Root(); r != nil {
		if r.Type() == NodeSubmodule {
			return r.Name()
		}
	}
	return ""
}

func (n *node) GetModuleByPrefix(
	pfx string,
	modules map[string]*Module,
	skipUnknown bool,
) (Node, error) {

	root := n.Root()
	if pfx == "" || root.Prefix() == pfx {
		// The local prefix may be ommitted or used explicitly
		return root, nil
	}
	mname, ok := getPfxName(root, pfx)
	if !ok {
		if !skipUnknown {
			return nil, fmt.Errorf("unknown import %s", pfx)
		} else {
			return nil, nil
		}
	}

	r, ok := modules[mname]
	if !ok {
		if skipUnknown {
			r = createFakeModule(mname)
			modules[mname] = r
		} else {
			return nil, fmt.Errorf("unknown module %s", mname)
		}
	}
	mod, ok := r.tree.Root.(Node)
	if !ok {
		return nil, fmt.Errorf("invalid root")
	}
	return mod, nil
}

// We need to be able to map from prefix (local / module scope) to
// the namespace (global scope across all modules) when we come
// across any node name in the XPath statement.  For unprefixed
// names we add the local namespace explicitly.
func (n *node) YangPrefixToNamespace(
	prefix string,
	modules map[string]*Module,
	skipUnknown bool,
) (string, error) {

	// If no prefix was specified, we need to return the correct local
	// Namespace.  Otherwise map the prefix to the namespace.
	if prefix == "" {
		return getNodeNamespaceInternal(n, modules)
	}
	moduleNode, err := n.GetModuleByPrefix(prefix, modules, skipUnknown)
	if err != nil {
		return "", err
	}
	if moduleNode != nil {
		return moduleNode.Root().Ns(), nil
	}
	return "", fmt.Errorf("Unable to map prefix '%s' to namespace.",
		prefix)
}
