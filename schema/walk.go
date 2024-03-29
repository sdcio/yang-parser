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

// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

// Structures used to specify the properties of the node to be found during
// a walk.
type NodeProperty struct {
	NodeProp  string
	NodeValue string
}

type NodeSubSpec struct {
	Type       string
	Properties []NodeProperty
}

// For the likes of verifying default is NOT set on a node, we have the
// NotPresent options, which default to false (present) as that is the more
// commonly tested case and this avoids needing to specify these so often.
//
// Note that NotPresent means node either doesn't exist, or has a
// different value to that specified.
type NodeSpec struct {
	Path                  []string
	DefaultPropNotPresent bool // If true, we expect node to exist w/o property
	Statement             NodeSubSpec
	DataPropNotPresent    bool // If true, we expect node to exist w/o data
	Data                  NodeSubSpec
}

// ActionFnType - action to be carried out on each node during a walk.
// As the tree is walked, we carry out the actionFn on each node in turn.
// For a find operation, this function may return <done> as true, indicating
// no further walking is required.  At this point, <success> indicates whether
// the overall 'mission' of the walk was successful or not.  Otherwise, if a
// walk runs to completion (all nodes walked) we return true.
// <retSlice> is used to store any number of objects returned by each
// invocation of the actionFn, eg output(s) generated by each node.
type ActionFnType func(
	targetNode Node,
	parentNode *XNode,
	nodeToFind NodeSpec,
	path []string,
	param interface{},
) (done, success bool, retSlice []interface{})

func findOrWalkWorker(
	targetNode Node,
	parentNode *XNode,
	nodeToFind NodeSpec,
	path []string,
	actionFn ActionFnType,
	param interface{},
) (Node, bool, []interface{}) {
	var retSlice []interface{}

	if actionFn != nil {
		done, success, retInt :=
			actionFn(targetNode, parentNode, nodeToFind, path, param)
		retSlice = append(retSlice, retInt...)
		if done {
			return targetNode, success, retSlice
		}
	}

	// Top-level node has no name, so ignore it.  Additionally, it's
	// pointless having leading '/' - just adds to noise in test file -
	// so logic below ensure we start with a non-zero-length node name.
	if targetNode.Name() != "" {
		path = append(path, targetNode.Name())
	}

	for _, subnode := range targetNode.Children() {
		retNode, success, retInt := findOrWalkWorker(
			subnode, NewXNode(targetNode, parentNode), nodeToFind,
			path, actionFn, param)
		retSlice = append(retSlice, retInt...)
		if retNode != nil {
			return retNode, success, retSlice
		}
	}
	if len(retSlice) == 0 {
		return nil, true, nil
	}
	return nil, true, retSlice
}

// Top level call to walk tree, carrying out actionFn
//
// NB: nodeToFind and actionFn may be null, though it's pretty pointless
//     walking with no actionFn!
func (ms *modelSet) FindOrWalk(
	nodeToFind NodeSpec,
	actionFn ActionFnType,
	param interface{},
) (Node, bool, []interface{}) {

	path := make([]string, 0)
	return findOrWalkWorker(ms, nil, nodeToFind, path, actionFn, param)
}
