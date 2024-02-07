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
