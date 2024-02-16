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

package compile

import (
	"github.com/sdcio/yang-parser/schema"
)

type SchemaFilter func(sn schema.Node) bool

// Filter configuration nodes
func IsConfig(sn schema.Node) bool {
	return sn.Config()
}

// Filter opd: extension nodes
func IsOpd(sn schema.Node) bool {
	switch sn.(type) {
	case schema.OpdCommand, schema.OpdArgument, schema.OpdOption:
		return true

	}
	return false
}

// Filter operational state nodes, which are config false
func IsState(sn schema.Node) bool {
	return !IsConfig(sn) && !IsOpd(sn)
}

// Filter configuration and operational state nodes
func IsConfigOrState() SchemaFilter {
	return Include(IsConfig, IsState)
}

// Returns a filter which will include nodes that match a set of filters
func Include(filters ...SchemaFilter) SchemaFilter {
	return func(sn schema.Node) bool {
		for _, fltr := range filters {
			if fltr != nil && fltr(sn) {
				return true
			}
		}
		return false
	}
}

// Returns a filter which will exclude nodes that match a set of filters
func Exclude(filters ...SchemaFilter) SchemaFilter {
	return func(sn schema.Node) bool {
		for _, fltr := range filters {
			if fltr != nil && fltr(sn) {
				return false
			}
		}
		return true
	}
}

// Returns an operational state node filter if state is true
func IncludeState(state bool) SchemaFilter {
	if state {
		return IsState
	}
	return Exclude(IsState)
}
