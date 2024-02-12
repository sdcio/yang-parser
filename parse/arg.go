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

// Copyright (c) 2017-2020, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package parse

import (
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type HasArgument interface {
	Argument() Argument
	ArgBool() bool
	ArgDate() string
	ArgDescendantSchema() []xml.Name
	ArgFractionDigits() int
	ArgIdRef() xml.Name
	ArgId() string
	ArgInt() int
	ArgKey() []string
	ArgLength() []Lb
	ArgMax() uint
	ArgOrdBy() string
	ArgPattern() *regexp.Regexp
	ArgPrefix() string
	ArgRange() RangeArgBdrySlice
	ArgSchema() []xml.Name
	ArgStatus() string
	ArgString() string
	ArgUint() uint
	ArgUnique() [][]xml.Name
	ArgUri() string
	ArgWhen() string
	ArgMust() string
	ArgPath() string

	checkArgument() error
}

type hasArgument struct {
	arg Argument
}

func (h *hasArgument) Argument() Argument   { return h.arg }
func (h *hasArgument) checkArgument() error { return h.arg.Parse() }

func (h *hasArgument) ArgStatus() string               { return h.arg.(*StatusArg).String() }
func (h *hasArgument) ArgString() string               { return h.arg.(StringArg).String() }
func (h *hasArgument) ArgPrefix() string               { return h.arg.(*PrefixArg).String() }
func (h *hasArgument) ArgUri() string                  { return h.arg.(*UriArg).String() }
func (h *hasArgument) ArgDate() string                 { return h.arg.(*DateArg).String() }
func (h *hasArgument) ArgMax() uint                    { return h.arg.(*MaxValueArg).i.i }
func (h *hasArgument) ArgSchema() []xml.Name           { return h.arg.(SchemaArg).Path() }
func (h *hasArgument) ArgDescendantSchema() []xml.Name { return h.arg.(*DescendantSchemaArg).path }
func (h *hasArgument) ArgInt() int                     { return h.arg.(*IntArg).i }
func (h *hasArgument) ArgUint() uint                   { return h.arg.(*UintArg).i }
func (h *hasArgument) ArgKey() []string                { return h.arg.(*KeyArg).keys }
func (h *hasArgument) ArgOrdBy() string                { return h.arg.(*OrdByArg).String() }
func (h *hasArgument) ArgId() string                   { return h.arg.(*IdArg).String() }
func (h *hasArgument) ArgIdRef() xml.Name              { return h.arg.(*IdRefArg).name }
func (h *hasArgument) ArgBool() bool                   { return h.arg.(*BoolArg).b }
func (h *hasArgument) ArgUnique() [][]xml.Name         { return h.arg.(*UniqueArg).paths }
func (h *hasArgument) ArgPattern() *regexp.Regexp      { return h.arg.(*PatternArg).Regexp }
func (h *hasArgument) ArgRange() RangeArgBdrySlice     { return h.arg.(*RangeArg).rbs }
func (h *hasArgument) ArgLength() []Lb                 { return h.arg.(*LengthArg).lbs }
func (h *hasArgument) ArgFractionDigits() int          { return h.arg.(*FractionDigitsArg).fdigits }
func (h *hasArgument) ArgWhen() string                 { return h.arg.(StringArg).String() }
func (h *hasArgument) ArgMust() string                 { return h.arg.(StringArg).String() }
func (h *hasArgument) ArgPath() string                 { return h.arg.(StringArg).String() }

//Extra character classes from XML spec.
//translate them to normal character classes
var patternReplacements = map[string]string{
	"\\p{IsBasicLatin}": "[\\x{0000}-\\x{007F}]",
}

type Argument interface {
	String() string
	Parse() error
	Validate(string) bool
	argument()
}

type arg string

func (a arg) String() string { return string(a) }
func (a arg) Parse() error   { return nil }
func (a arg) Validate(s string) bool {
	return a.String() == s
}
func (arg) argument() {}

type StringArg struct {
	arg
}

type IdArg struct {
	arg
}

/*
 * Parse an identifier, validating it as specified in
 * RFC 6020; Sec 12 "YANG ABNF Grammar"
 *
 *    ;; An identifier MUST NOT start with (('X'|'x') ('M'|'m') ('L'|'l'))
 *    identifier          = (ALPHA / "_")
 *                          *(ALPHA / DIGIT / "_" / "-" / ".")
 */
func (a *IdArg) Parse() error {
	str := string(a.arg)
	ErrInval := errors.New("invalid identifier: " + str)

	if len(str) == 0 {
		return ErrInval
	}
	if len(str) >= 3 {
		if strings.ToUpper(str[:3]) == "XML" {
			return errors.New("invalid identifier," +
				" not allowed to start with xml: " + str)
		}
	}
	var r rune = rune(str[0])
	if !(r == '_' || unicode.IsLetter(r)) {
		return ErrInval
	}
	for i := 1; i < len(str); i++ {
		var r rune = rune(str[i])
		if !isAlphaNumeric(r) && r != '-' && r != '.' {
			return ErrInval
		}
	}
	return nil
}

type PrefixArg struct {
	arg
	id *IdArg
}

func (a *PrefixArg) Parse() error {
	a.id = &IdArg{a.arg}
	err := a.id.Parse()
	if err != nil {
		return errors.New("prefix: " + err.Error())
	}
	return nil
}

type IdRefArg struct {
	arg
	pfx  *PrefixArg
	id   *IdArg
	name xml.Name
}

func (a *IdRefArg) Parse() error {
	//[prefix ":"] id
	var err error

	parts := strings.Split(string(a.arg), ":")

	switch len(parts) {
	case 1:
		a.id = &IdArg{arg: arg(parts[0])}
		err = a.id.Parse()
		if err != nil {
			return errors.New("id-ref: " + err.Error())
		}
	case 2:
		a.pfx = &PrefixArg{arg: arg(parts[0])}
		err = a.pfx.Parse()
		if err != nil {
			return errors.New("id-ref: " + err.Error())
		}
		a.id = &IdArg{arg(parts[1])}
		err = a.id.Parse()
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid identifier reference")
	}
	if a.pfx != nil {
		a.name.Space = a.pfx.String()
	}
	a.name.Local = a.id.String()
	return nil
}

type UriArg struct {
	arg
	url *url.URL
}

func (a *UriArg) Parse() error {
	u, e := url.Parse(string(a.arg))
	if e != nil {
		return e
	}
	a.url = u
	return nil
}

type BoolArg struct {
	arg
	b bool
}

func (a *BoolArg) Parse() error {
	b, e := strconv.ParseBool(string(a.arg))
	if e != nil {
		return e
	}
	a.b = b
	return nil
}

type DateArg struct {
	arg
	year, month, day int
}

func (a *DateArg) Parse() error {
	//4DIGIT"-"2DIGIT"-"2DIGIT
	var err error
	var i int
	var str = string(a.arg)
	var ErrInval = errors.New("invalid date: " + str)

	if len(str) != 10 {
		return ErrInval
	}

	/* 4DIGIT */
	i = strings.Index(str, "-")
	if i != 4 {
		return ErrInval
	}
	y := str[:i]
	a.year, err = strconv.Atoi(y)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); ok {
			return errors.New(ErrInval.Error() + ": error parsing " +
				strconv.Quote(ne.Num) + ": " + ne.Err.Error())
		}
		return errors.New(ErrInval.Error() + ": " + err.Error())
	}

	/* "-" */
	str = str[i+1:]

	/* 2DIGIT */
	i = strings.Index(str, "-")
	if i != 2 {
		return ErrInval
	}
	m := str[:i]
	a.month, err = strconv.Atoi(m)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); ok {
			return errors.New(ErrInval.Error() + ": error parsing " +
				strconv.Quote(ne.Num) + ": " + ne.Err.Error())
		}
		return errors.New(ErrInval.Error() + ": " + err.Error())
	}

	/* "-" */
	str = str[i+1:]

	/* 2DIGIT */
	if len(str) != 2 {
		return ErrInval
	}
	a.day, err = strconv.Atoi(str)
	if err != nil {
		if ne, ok := err.(*strconv.NumError); ok {
			return errors.New(ErrInval.Error() + ": error parsing " +
				strconv.Quote(ne.Num) + ": " + ne.Err.Error())
		}
		return errors.New(ErrInval.Error() + ": " + err.Error())
	}

	return nil
}

type YangVersionArg struct {
	arg
}

func (a *YangVersionArg) Parse() error {
	if a.arg == "1" {
		return nil
	}
	return errors.New("invalid yang-version: " + string(a.arg))
}

type EmptyArg struct {
	arg
}

func (a EmptyArg) Parse() error {
	if a.arg == "" {
		return nil
	}
	return errors.New("invalid argument: " + string(a.arg))
}

type KeyArg struct {
	arg
	keys []string
}

func (a *KeyArg) split() []string {
	var str = string(a.arg)
	var start, pos int
	strs := make([]string, 0)
	for pos = 0; pos < len(str); pos++ {
		if isSep(rune(str[pos])) {
			s := str[start:pos]
			if len(s) > 0 {
				strs = append(strs, s)
			}
			start = pos + 1
		}
	}
	s := str[start:pos]
	if len(s) > 0 {
		strs = append(strs, s)
	}
	return strs
}
func (a *KeyArg) Parse() error {
	strs := a.split()
	if len(strs) == 0 {
		return errors.New("invalid key argument: " + string(a.arg))
	}
	a.keys = strs
	return nil
}

type UintArg struct {
	arg
	i uint
}

func (a *UintArg) Parse() error {
	i, e := strconv.ParseUint(string(a.arg), 0, 32)
	if e != nil {
		return e
	}
	a.i = uint(i)
	return nil
}

type IntArg struct {
	arg
	i int
}

func (a *IntArg) Parse() error {
	i, e := strconv.ParseInt(string(a.arg), 0, 32)
	if e != nil {
		return e
	}
	a.i = int(i)
	return nil
}

type StatusArg struct {
	arg
}

func (a *StatusArg) Parse() error {
	str := string(a.arg)
	if str == "current" || str == "obsolete" || str == "deprecated" {
		return nil
	}
	return errors.New("invalid status argument: " + string(a.arg))
}

func (a *StatusArg) String() string {
	return string(a.arg)
}

type PatternArg struct {
	arg
	*regexp.Regexp
}

// Go's regexp engine doesn't anchor regexps to start and end of line, whereas
// YANG pattern statements use XSD regexps that do this implicitly.  So, we
// need to explicitly anchor pattern regexps to get the correct behaviour.
// As these patterns may be branched (contain '|' (or)), we need to
// parenthesise the pattern before anchoring it.
func (a *PatternArg) Parse() error {
	s := a.String()
	for k, v := range patternReplacements {
		s = strings.Replace(s, k, v, -1)
	}
	s = "^(" + s + ")$"
	re, err := regexp.Compile(s)
	if err != nil {
		return err
	}
	a.Regexp = re
	return nil
}
func (a *PatternArg) String() string {
	//Fix ambiguous selector, call String() on
	//arg not Regexp
	return a.arg.String()
}
func (a *PatternArg) Validate(s string) bool {
	return a.Match([]byte(s))
}

type OrdByArg struct {
	arg
}

func (a *OrdByArg) Parse() error {
	str := string(a.arg)
	if str == "system" || str == "user" {
		return nil
	}
	return errors.New("invalid argument: " + string(a.arg))
}

type DeviateArg struct {
	arg
}

func (a *DeviateArg) Parse() error {
	str := string(a.arg)
	if str == "add" || str == "delete" || str == "replace" || str == "not-supported" {
		return nil
	}
	return errors.New("invalid argument: " + string(a.arg))
}

type SchemaArg interface {
	Argument
	Path() []xml.Name
}

type AbsoluteSchemaArg struct {
	arg
	ids  []*IdRefArg
	path []xml.Name
}

func (a *AbsoluteSchemaArg) Parse() error {
	strs := strings.Split(string(a.arg), "/")
	if len(strs) < 2 {
		return errors.New("invalid argument: " + string(a.arg))
	}
	if strs[0] != "" {
		return errors.New("invalid argument: " + string(a.arg) + " expected root")
	}
	strs = strs[1:]
	a.ids = make([]*IdRefArg, 0, len(strs))
	for _, v := range strs {
		i := &IdRefArg{arg: arg(v)}
		e := i.Parse()
		if e != nil {
			return e
		}
		a.ids = append(a.ids, i)
	}
	a.path = make([]xml.Name, 0, len(a.ids))
	for _, id := range a.ids {
		a.path = append(a.path, id.name)
	}
	return nil
}

func (a *AbsoluteSchemaArg) Path() []xml.Name {
	return a.path
}

type DescendantSchemaArg struct {
	arg
	ids  []*IdRefArg
	path []xml.Name
}

func (a *DescendantSchemaArg) Parse() error {
	strs := strings.Split(string(a.arg), "/")
	if len(strs) < 1 {
		return errors.New("invalid argument: " + string(a.arg))
	}
	if strs[0] == "" {
		return errors.New("invalid argument: " + string(a.arg) + " unexpected root")
	}
	a.ids = make([]*IdRefArg, 0, len(strs))
	for _, v := range strs {
		i := &IdRefArg{arg: arg(v)}
		e := i.Parse()
		if e != nil {
			return e
		}
		a.ids = append(a.ids, i)
	}
	a.path = make([]xml.Name, 0, len(a.ids))
	for _, id := range a.ids {
		a.path = append(a.path, id.name)
	}
	return nil
}

func (a *DescendantSchemaArg) Path() []xml.Name {
	return a.path
}

type MaxValueArg struct {
	arg
	i *UintArg
}

func (a *MaxValueArg) Parse() error {
	var i *UintArg
	if a.arg == "unbounded" {
		i = &UintArg{i: ^uint(0)}
	} else {
		i = &UintArg{arg: a.arg}
		e := i.Parse()
		if e != nil {
			return e
		}
	}
	a.i = i
	return nil
}

type argRb struct {
	Min, Max   bool
	Start, End string
}

type RangeArgBdrySlice []argRb

type RangeArg struct {
	arg
	rbs RangeArgBdrySlice
}

func (a *RangeArg) Parse() error {
	str := string(a.arg)
	ErrInval := errors.New("invalid argument: " + str)

	/* collapse string */
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, "\n", "", -1)

	/* range-part *(optsep "|" optsep range-part) */
	rparts := strings.Split(str, "|")
	a.rbs = make([]argRb, 0, len(rparts))
	for _, v := range rparts {
		/* range-boundary [optsep ".." optsep range-boundary] */
		var r argRb
		rbs := strings.Split(v, "..")
		switch len(rbs) {
		case 1:
			switch rbs[0] {
			case "max":
				r.Max = true
			case "min":
				r.Min = true
			default:
				r.Start = rbs[0]
				r.End = rbs[0]
			}
		case 2:
			switch rbs[0] {
			case "min":
				r.Min = true
			default:
				r.Start = rbs[0]
			}
			switch rbs[1] {
			case "max":
				r.Max = true
			default:
				r.End = rbs[1]
			}
		default:
			return ErrInval
		}
		a.rbs = append(a.rbs, r)
	}
	return nil
}

type Lb struct {
	Min, Max   bool
	Start, End uint64
}

type LengthArg struct {
	arg
	lbs []Lb
}

func (a *LengthArg) Parse() error {
	str := string(a.arg)
	ErrInval := errors.New("invalid argument: " + str)

	/* collapse string */
	str = strings.Replace(str, " ", "", -1)
	str = strings.Replace(str, "\t", "", -1)
	str = strings.Replace(str, "\n", "", -1)

	/* length-part *(optsep "|" optsep length-part) */
	lparts := strings.Split(str, "|")
	a.lbs = make([]Lb, 0, len(lparts))
	for _, v := range lparts {
		/* length-boundary [optsep ".." optsep length-boundary] */
		var l Lb
		var i uint64
		var e error
		bs := strings.Split(v, "..")
		switch len(bs) {
		case 1:
			switch bs[0] {
			case "max":
				l.Max = true
			case "min":
				l.Min = true
			default:
				i, e := strconv.ParseUint(bs[0], 0, 64)
				if e != nil {
					return e
				}
				l.Start = i
				l.End = i
			}
		case 2:
			switch bs[0] {
			case "min":
				l.Min = true
			default:
				i, e = strconv.ParseUint(bs[0], 0, 64)
				if e != nil {
					return e
				}
				l.Start = i
			}
			switch bs[1] {
			case "max":
				l.Max = true
			default:
				i, e = strconv.ParseUint(bs[1], 0, 64)
				if e != nil {
					return e
				}
				l.End = i
			}
		default:
			return ErrInval
		}
		a.lbs = append(a.lbs, l)
	}
	return nil
}

type UniqueArg struct {
	arg
	args  []*DescendantSchemaArg
	paths [][]xml.Name
}

func (a *UniqueArg) split() ([]*DescendantSchemaArg, error) {
	var str = string(a.arg)
	var start, pos int
	args := make([]*DescendantSchemaArg, 0)
	for pos = 0; pos < len(str); pos++ {
		if isSep(rune(str[pos])) {
			s := str[start:pos]
			if len(s) > 0 {
				a := &DescendantSchemaArg{arg: arg(s)}
				if err := a.Parse(); err != nil {
					return nil, err
				}
				args = append(args, a)
			}
			start = pos + 1
		}
	}
	s := str[start:pos]
	if len(s) > 0 {
		a := &DescendantSchemaArg{arg: arg(s)}
		if err := a.Parse(); err != nil {
			return nil, err
		}
		args = append(args, a)
	}
	return args, nil
}
func (a *UniqueArg) Parse() error {
	args, err := a.split()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return errors.New("invalid argument: " + string(a.arg))
	}
	a.args = args
	a.paths = make([][]xml.Name, 0, len(a.args))
	for _, arg := range a.args {
		a.paths = append(a.paths, arg.path)
	}
	return nil
}

type FractionDigitsArg struct {
	arg
	fdigits int
}

func (a *FractionDigitsArg) Parse() error {
	var err error
	var str = string(a.arg)
	var ErrInval = errors.New("invalid argument: " + str)
	switch len(str) {
	case 1:
		fallthrough
	case 2:
		a.fdigits, err = strconv.Atoi(str)
		if err != nil {
			return errors.New(ErrInval.Error() + ": " + err.Error())
		}
	default:
		return ErrInval
	}
	if a.fdigits < 1 || a.fdigits > 18 {
		return ErrInval
	}
	return nil
}

func getArgByType(ntype NodeType, a string, interner *ArgInterner) (out Argument) {
	defer func() {
		out = interner.Intern(ntype, out)
	}()
	switch ntype {
	// String Arguments
	case NodeConfigdHelp, NodeConfigdValidate, NodeConfigdSyntax, NodeConfigdAllowed,
		NodeConfigdBegin, NodeConfigdEnd, NodeConfigdCreate, NodeConfigdMust,
		NodeConfigdDelete, NodeConfigdUpdate, NodeConfigdSubst, NodeConfigdErrMsg,
		NodeConfigdPHelp, NodeConfigdCallRpc, NodeConfigdGetState, NodeUnknown,
		NodeOpdOnEnter, NodeOpdHelp, NodeOpdAllowed, NodeOpdInherit, NodeOpdPatternHelp,
		NodeErrorMessage, NodeReference, NodeDefault, NodePresence, NodeWhen, NodeErrorAppTag,
		NodeEnum, NodeMust, NodeContact, NodeDescription, NodeOrganization, NodeUnits,
		NodePath, NodeConfigdNormalize, NodeConfigdDeferActions:
		return StringArg{arg(a)}

	// Uint Arguments
	case NodePosition, NodeMinElements, NodeConfigdPriority:
		return &UintArg{arg: arg(a)}

	// Int Arguments
	case NodeValue:
		return &IntArg{arg: arg(a)}

	case NodeMaxElements:
		return &MaxValueArg{arg: arg(a)}

	// Boolean Arguments
	case NodeConfigdSecret, NodeYinElement, NodeRequireInstance, NodeConfig, NodeMandatory,
		NodeOpdRepeatable, NodeOpdPrivileged, NodeOpdLocal, NodeOpdSecret, NodeOpdPassOpcArgs:
		return &BoolArg{arg: arg(a)}

	// ID Arguments
	case NodeGrouping, NodeList, NodeChoice, NodeCase, NodeAnyxml, NodeContainer,
		NodeLeaf, NodeLeafList, NodeExtension, NodeArgument, NodeIdentity, NodeFeature,
		NodeRpc, NodeNotification, NodeBit, NodeTypedef, NodeModule, NodeSubmodule,
		NodeOpdCommand, NodeOpdOption, NodeOpdArgument,
		NodeImport, NodeInclude, NodeBelongsTo:
		return &IdArg{arg: arg(a)}

	// ID Ref Arguments
	case NodeBase, NodeIfFeature, NodeUses, NodeTyp:
		return &IdRefArg{arg: arg(a)}

	// Date Arguments
	case NodeRevision, NodeRevisionDate:
		return &DateArg{arg: arg(a)}

	// Empty Arguments
	case NodeInput, NodeOutput:
		return &EmptyArg{arg: arg(a)}

	// Specialist Arguments
	case NodeYangVersion:
		return &YangVersionArg{arg: arg(a)}
	case NodeNamespace:
		return &UriArg{arg: arg(a)}
	case NodeKey:
		return &KeyArg{arg: arg(a)}
	case NodeStatus:
		return &StatusArg{arg: arg(a)}
	case NodeOrderedBy:
		return &OrdByArg{arg: arg(a)}
	case NodeDeviate, NodeDeviateNotSupported, NodeDeviateAdd,
		NodeDeviateDelete, NodeDeviateReplace:
		return &DeviateArg{arg: arg(a)}
	case NodeDeviation:
		return &AbsoluteSchemaArg{arg: arg(a)}
	case NodeRefine:
		return &DescendantSchemaArg{arg: arg(a)}
	case NodeUnique:
		return &UniqueArg{arg: arg(a)}
	case NodePattern:
		return &PatternArg{arg: arg(a)}
	case NodePrefix:
		return &PrefixArg{arg: arg(a)}
	case NodeAugment, NodeOpdAugment:
		var sa SchemaArg
		sa = &AbsoluteSchemaArg{arg: arg(a)}
		if err := sa.Parse(); err != nil {
			sa = &DescendantSchemaArg{arg: arg(a)}
		}
		return sa
	case NodeRange:
		return &RangeArg{arg: arg(a)}
	case NodeLength:
		return &LengthArg{arg: arg(a)}
	case NodeFractionDigits:
		return &FractionDigitsArg{arg: arg(a)}
	default:
		panic(fmt.Errorf("Unexpected type %s", nodeNames[ntype]))
	}
}
