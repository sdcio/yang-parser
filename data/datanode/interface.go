// Copyright (c) 2019, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2015-2016 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package datanode

/*
 * A simple interface for presenting YANG data.
 *
 * The layout closely matches the YANG model, specifically the JSON encoding
 * of the document where lists are children of a list node and leaf-lists
 * exist as a list of children.
 */
type DataNode interface {

	// The name of the schema node this data represents
	YangDataName() string

	// The child nodes of this node. Either nodes in a container, or entries
	// in a list, including keys.  Note that NoSorting() guarantees we won't
	// waste time sorting the reply on each call (which is slow), though if
	// returning a cached list it may have already been sorted.
	YangDataChildren() []DataNode
	YangDataChildrenNoSorting() []DataNode

	// The values of a leaf or a leaf-list
	YangDataValues() []string
	YangDataValuesNoSorting() []string
}
