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
// Copyright (c) 2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

// This file contains the Warning object used to store XPATH compiler
// warnings

package xutils

import (
	"fmt"
	"strings"
)

type Warning struct {
	warnTyp   WarnType
	startNode string // Node XPATH statement is associated with
	xpathStmt string // Full XPATH statement (without must/when)
	xpathLoc  string // Original 'Module:Line' XPATH location (eg in grouping)
	testPath  string // Fully-qualified path being validated.
	debugStr  string // Output from running PathEval machine
}

type WarnType int

const (
	ValidPath WarnType = iota
	DoesntExist
	MissingOrWrongPrefix
	MustOnNPContainer
	MustOnNPContWithNPChild
	RefNPContainer
	CompilerError
	ConfigdMustCompilerError
)

func (w WarnType) String() string {
	switch w {
	case ValidPath:
		return "is valid"
	case DoesntExist:
		return "doesn't exist"
	case MissingOrWrongPrefix:
		return "has missing / wrong prefix(es)"
	case MustOnNPContainer:
		return "is on non-presence container"
	case MustOnNPContWithNPChild:
		return "is on non-presence container WITH non-presence child container"
	case RefNPContainer:
		return "references non-presence container"
	case CompilerError:
		return "compilation failed"
	case ConfigdMustCompilerError:
		return "compilation failed (configd:must)"
	}
	return "(undefined)"
}

func NewWarning(
	warning WarnType,
	startNode, xpathStmt, xpathLoc, testPath, debugStr string,
) Warning {
	return Warning{
		startNode: startNode,
		xpathStmt: xpathStmt,
		xpathLoc:  xpathLoc,
		testPath:  testPath,
		debugStr:  debugStr,
		warnTyp:   warning,
	}
}

const (
	StripPrefix     = true
	DontStripPrefix = false
)

// Uniquely identifies what is being tested, ie module:line source of statement,
// followed by specific path being tested.  Note that testPath can be stripped
// of its prefix if we are wanting to filter out false positives.
func (w Warning) GetUniqueString(stripPrefix bool) string {
	if stripPrefix {
		return w.xpathLoc + ":" +
			removePrefixesFromNonPredicatedPathString(w.testPath)
	}
	return w.xpathLoc + w.testPath
}

func (w Warning) GetType() WarnType {
	return w.warnTyp
}

func (w Warning) String() string {
	return fmt.Sprintf(
		"Node:\t\t%s\n"+
			"Xpath:\t\t'%s'\n"+
			"XpathLoc:\t%s\n"+
			"TestPath:\t%s\n"+
			"Warning:\t%s\n"+
			"Dbg:\n%s\n",
		w.startNode, w.xpathStmt, w.xpathLoc, w.testPath, w.warnTyp, w.debugStr)
}

// 'match' returns no error if all non-zero-length expWarn fields match <w>.
func (w Warning) Match(expWarn Warning) error {
	return w.matchInternal(expWarn, true)
}

// Less strict on debug string matching - look for substring not exact match
func (w Warning) MatchDebugContains(expWarn Warning) error {
	return w.matchInternal(expWarn, false)
}

func (w Warning) matchInternal(expWarn Warning, exactMatch bool) error {
	if expWarn.startNode != "" && expWarn.startNode != w.startNode {
		return fmt.Errorf("Start node values don't match")
	}
	if expWarn.xpathStmt != "" && expWarn.xpathStmt != w.xpathStmt {
		return fmt.Errorf("Xpath statements don't match")
	}
	if expWarn.xpathLoc != "" && expWarn.xpathLoc != w.xpathLoc {
		return fmt.Errorf("Xpath locations don't match")
	}
	if expWarn.testPath != "" && expWarn.testPath != w.testPath {
		return fmt.Errorf("Warning text doesn't match")
	}
	if expWarn.warnTyp != w.warnTyp {
		return fmt.Errorf("Warning type doesn't match")
	}

	if expWarn.debugStr != "" {
		if exactMatch {
			if expWarn.debugStr != w.debugStr {
				return fmt.Errorf("Debug strings don't match")
			}
		} else {
			if !strings.Contains(w.debugStr, expWarn.debugStr) {
				return fmt.Errorf("Debug string doesn't contain expected text")
			}
		}
	}

	return nil
}

func RemoveNPContainerWarnings(warns []Warning) []Warning {
	var filteredWarnings []Warning
	for _, warn := range warns {
		switch warn.warnTyp {
		case MustOnNPContainer, MustOnNPContWithNPChild, RefNPContainer:
			continue
		default:
			filteredWarnings = append(filteredWarnings, warn)
		}
	}
	return filteredWarnings
}
