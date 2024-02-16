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
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package datanode

type datanode struct {
	name     string
	children []DataNode
	values   []string
}

func CreateDataNode(name string, children []DataNode, values []string) DataNode {
	return &datanode{name, children, values}
}

func (n *datanode) YangDataName() string {
	return n.name
}

func (n *datanode) YangDataChildren() []DataNode {
	return n.children
}

func (n *datanode) YangDataChildrenNoSorting() []DataNode {
	return n.children
}

func (n *datanode) YangDataValues() []string {
	return n.values
}

func (n *datanode) YangDataValuesNoSorting() []string {
	return n.values
}
