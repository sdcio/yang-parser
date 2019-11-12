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
