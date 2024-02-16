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
