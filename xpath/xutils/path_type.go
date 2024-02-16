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
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file implements PathType (useful for managing operations on paths) and
// MatchFilter() for filtering on XML names (part of paths).

package xutils

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type XFilter struct {
	name    xml.Name
	matchOn MatchType
}

type MatchType int

const (
	FullTree MatchType = iota
	ConfigOnly
	OpdOnly
)

func (xf XFilter) Name() xml.Name { return xf.name }
func (xf XFilter) MatchConfigOnly() bool {
	return xf.matchOn == ConfigOnly
}

func NewXFilter(name xml.Name, matchOn MatchType) XFilter {
	return XFilter{name: name, matchOn: matchOn}
}

func NewXFilterFullTree(name xml.Name) XFilter {
	return XFilter{name: name, matchOn: FullTree}
}

func NewXFilterConfigOnly(name xml.Name) XFilter {
	return XFilter{name: name, matchOn: ConfigOnly}
}

type XTarget struct {
	name       xml.Name
	targetType TargetType
}

// TargetType - node type of the target to be matched
// Subtly different to MatchType as here we want to specify the type of our
// node explicitly whereas for MatchType we want to specify a set of types to
// be matched.
type TargetType int

const (
	NotConfigOrOpdTarget TargetType = iota
	ConfigTarget
	OpdTarget
)

func NewXTarget(name xml.Name, targetType TargetType) XTarget {
	return XTarget{name: name, targetType: targetType}
}

func NewXConfigTarget(name xml.Name) XTarget {
	return XTarget{name: name, targetType: ConfigTarget}
}

func NewXNonConfigOrOpdTarget(name xml.Name) XTarget {
	return XTarget{name: name, targetType: NotConfigOrOpdTarget}
}

func (t XTarget) IsConfig() bool { return t.targetType == ConfigTarget }
func (t XTarget) IsOpd() bool    { return t.targetType == OpdTarget }
func (t XTarget) Name() xml.Name { return t.name }

// This is the global wildcard, representing all child nodes regardless of
// module.
var AllChildren = NewXFilterFullTree(xml.Name{Local: "*"})
var AllCfgChildren = NewXFilterConfigOnly(xml.Name{Local: "*"})

// Return true (match) if filter matches target or filter is AllChildren
// Note that filter may be '<namespaceName>:*'.  Note also that we ignore
// prefix as this is local to a single namespace, and is converted to a
// namespace name for global uniqueness *and* global consistency of naming.
//
// The filter may or may not have a prefix - if not then we only match on
// localPart.
//
// Additionally, if the node we are matching is not a config node, and the
// filter indicates we can only match config nodes, no match will be made.
func MatchFilter(filter XFilter, target XTarget) bool {
	if !target.IsConfig() && filter.MatchConfigOnly() {
		return false
	}
	switch {
	// DON'T REORDER - it matters
	case filter.Name() == AllChildren.name: // '*' (global wildcard)
		return true
	case filter.Name().Space == "": // 'bar' (unqualified name)
		return (filter.Name().Local == target.Name().Local)
	case filter.name.Local == "*": // 'foo:*' (wildcard within namespace)
		return (filter.Name().Space == target.Name().Space)
	default: // 'foo:bar' (fully-qualified target)
		return (filter.Name().Space == target.Name().Space) &&
			(filter.Name().Local == target.Name().Local)
	}
}

// The following functions are useful for converting XPATH path expressions
// that may contain prefixes and/or predicates into an absolute path without
// either prefixes or predicates for use in error messages and the like.

func GetAbsPath(expr string, curPath PathType) PathType {
	// First strip out any predicates as these are not needed, and may
	// contain '/' symbols which we want to ignore here.
	expr = removePredicatesFromPathString(expr)

	// Next strip out prefixes.
	expr = removePrefixesFromNonPredicatedPathString(expr)

	// Now deal with relative paths.  Count them up, remove them, and then
	// use the curPath minus the counted number of '../'s.
	steps := strings.Count(expr, "../")
	if steps == 0 {
		// Absolute path
		return NewPathType(expr)
	}

	if len(expr) <= steps*3 /* chars in '../' */ {
		return NewPathType(expr)
	}

	// Remove '../' from front and switch from string to PathType
	lrPath := NewPathType(expr[steps*3:])

	elemsToRemove := steps + 1 // Number of '../'s + 1 for leaf value
	curPathLen := len(curPath)
	if elemsToRemove < curPathLen {
		return append(curPath[:curPathLen-elemsToRemove], lrPath...)
	}

	// We don't seem to have a long enough curPath, so insert '(unknown)'
	// in front of what we do know!
	return append(NewPathType("(unknown)"), lrPath...)
}

func removePredicatesFromPathString(expr string) string {
	for i := 0; i < strings.Count(expr, "["); i++ {
		start := strings.Index(expr, "[")
		end := strings.Index(expr, "]")
		// Just in case, return path as it currently stands.  Better than nowt.
		if end < start {
			return expr
		}
		// If expression ends with predicate, take special care.
		if end >= len(expr)-1 {
			expr := expr[:start]
			return expr
		}
		expr = expr[:start] + expr[end+1:]
	}

	return expr
}

// Will not work well with predicates present, so if they are it just
// returns initial string.  Problem is that to deal with predicates the
// logic has to be much cleverer at locating the start of the prefix,
// whereas with no predicates we can just use '/' or start of path as
// indicator of possible prefix start.
func removePrefixesFromNonPredicatedPathString(expr string) string {
	if strings.Count(expr, ":") > 0 {
		subEntry := ""
		newEntry := ""
		for _, r := range expr {
			char := fmt.Sprintf("%c", r)
			if char == "/" { // Use subEntry
				newEntry += subEntry + "/"
				subEntry = ""
			} else if char == ":" { // Ignore prefix
				subEntry = ""
			} else { // Continue to build up subEntry
				subEntry += char
			}
		}
		newEntry += subEntry

		expr = newEntry
	}

	return expr
}

// PATHTYPE
//
// Useful to have this for manipulating paths.  Provides a
// comparison function and string function.
type PathType []string

func NewPathType(path string) PathType {
	// Strip leading and trailing WS
	path = strings.TrimSpace(path)
	if len(path) == 0 {
		return []string{}
	}

	var pt PathType

	// Initial '/' is noted.
	if path[0] == '/' {
		pt = append(pt, "/")
		if len(path) == 1 {
			return pt
		}
		path = path[1:]
	}

	// Split rest by '/' and add to slice
	pt = append(pt, strings.Split(path, "/")...)
	return pt
}

// Generate path string, separated by '/'.
func (p PathType) String() (pathStr string) {
	if len(p) == 0 {
		return ""
	}

	// If first element is '/' then we will get // at the front.  Otherwise
	// we don't want any leading /.
	if p[0] != "/" {
		pathStr = p[0]
	}
	p = p[1:]

	// Now we can iterate through remaining elements, if any...
	for _, elem := range p {
		pathStr = pathStr + "/" + elem
	}

	return pathStr
}

func (p PathType) SpacedString() string {
	if len(p) == 0 {
		return ""
	}

	// Ignore leading '/' in space-separated format.
	if p[0] == "/" {
		if len(p) == 1 {
			return ""
		}
		p = p[1:]
	}

	return strings.TrimSpace(strings.Join(p, " "))
}

func (p1 PathType) EqualTo(p2 PathType) bool {
	if len(p1) != len(p2) {
		return false
	}
	for index, elem1 := range p1 {
		if elem1 != p2[index] {
			return false
		}
	}
	return true
}
