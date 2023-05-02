// Copyright (c) 2017-2021, AT&T Intellectual Property. All rights reserved
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/pathutil"
	"github.com/iptecharch/yang-parser/xpath"
	"github.com/iptecharch/yang-parser/xpath/xutils"
)

/*
 * The typesystem for yang is horrbily complex in the way restrictions are applied to types
 * to satisfy this we create a golang type that reflects the yang type with a validation
 * method. We also define restrictions that are preparsed golang representations of the
 * restriction for faster runtime validation. To make matters worse the yang definition doesn't
 * lexically specify which type of restriction is allowed on which type instead these are semantic
 * restrictions that must be resolved at type creation time.
 */
var errInvalRestriction = errors.New("invalid restriction for type")

var fdtab = map[Fracdigit]Drb{
	1:  {-922337203685477580.8, 922337203685477580.7},
	2:  {-92233720368547758.08, 92233720368547758.07},
	3:  {-9223372036854775.808, 9223372036854775.807},
	4:  {-922337203685477.5808, 922337203685477.5807},
	5:  {-92233720368547.75808, 92233720368547.75807},
	6:  {-9223372036854.775808, 9223372036854.775807},
	7:  {-922337203685.4775808, 922337203685.4775807},
	8:  {-92233720368.54775808, 92233720368.54775807},
	9:  {-9223372036.854775808, 9223372036.854775807},
	10: {-922337203.6854775808, 922337203.6854775807},
	11: {-92233720.36854775808, 92233720.36854775807},
	12: {-9223372.036854775808, 9223372.036854775807},
	13: {-922337.2036854775808, 922337.2036854775807},
	14: {-92233.72036854775808, 92233.72036854775807},
	15: {-9223.372036854775808, 9223.372036854775807},
	16: {-922.3372036854775808, 922.3372036854775807},
	17: {-92.23372036854775808, 92.23372036854775807},
	18: {-9.223372036854775808, 9.223372036854775807},
}

type BitWidth int

const (
	BitWidth8  BitWidth = 8
	BitWidth16 BitWidth = 16
	BitWidth32 BitWidth = 32
	BitWidth64 BitWidth = 64
)

var inttab = map[BitWidth]Rb{
	BitWidth8:  {-128, 127},
	BitWidth16: {-32768, 32767},
	BitWidth32: {-2147483648, 2147483647},
	BitWidth64: {-9223372036854775808, 9223372036854775807},
}

var uinttab = map[BitWidth]Urb{
	BitWidth8:  {0, 255},
	BitWidth16: {0, 65535},
	BitWidth32: {0, 4294967295},
	BitWidth64: {0, 18446744073709551615},
}

func pathstr(path []string) string {
	var str string
	for _, v := range path {
		str += "/" + strings.Replace(url.QueryEscape(v), "+", "%20", -1)
	}
	return str
}

type Type interface {
	Validate(ctx ValidateCtx, path []string, s string) error
	Name() xml.Name
	errors() []string
	ytype()
	// Strings (possibly other types) may have an empty string as the default.
	// So, we need to explicitly note the presence / absence of a default
	Default() (string, bool)
	AllowedValues(ctxNode xutils.XpathNode, debug bool) (
		allowedValues []string, err error)
}

func genErrorString(t Type) string {
	var buf bytes.Buffer
	errstrs := t.errors()
	if len(errstrs) == 1 {
		return "Must have value " + errstrs[0]
	}
	buf.WriteString("Must have one of the following values: ")
	for i, estr := range errstrs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(estr)
	}
	return buf.String()

}

type Status int

const (
	Current = iota
	Deprecated
	Obsolete
)

func (s Status) String() string {
	switch s {
	case Current:
		return "Current"
	case Deprecated:
		return "Deprecated"
	case Obsolete:
		return "Obsolete"
	default:
		panic(fmt.Errorf("Unexpected status value %d", s))
	}
}

/* types */
type yrestrict struct {
}

func (yrestrict) errors() []string {
	return nil
}
func (yrestrict) restriction() {}

/*
 * 'def'ault can't be set directly for a type.  Instead it's either set on
 * a typedef or a leaf / choice statement.  However it is set, it ends up
 * being associated with the leaf / choice schema node via its attached
 * type node, so that is the logical place within the schema to store it.
 */
type ytyp struct {
	yrestrict
	name       xml.Name
	def        string
	hasDefault bool
}

func newType(name xml.Name, def string, hasDef bool) ytyp {
	return ytyp{
		name:       name,
		def:        def,
		hasDefault: hasDef,
	}
}

func (t *ytyp) Name() xml.Name {
	return t.name
}

func (t *ytyp) AllowedValues(
	ctxNode xutils.XpathNode,
	debug bool,
) ([]string, error) {
	return []string{}, nil
}

func (t *ytyp) Default() (string, bool) {
	if t.hasDefault {
		return t.def, true
	}

	return "", false
}

func (t *ytyp) String() string {
	if t.name.Space != "" {
		return t.name.Space + ":" + t.name.Local
	}
	return t.name.Local
}
func (ytyp) ytype() {}

type Binary interface {
	Type
	Length() *Length
	isBinary()
}

type binary struct {
	ytyp
	len *Length
}

// Ensure that other schema types don't meet the interface
func (*binary) isBinary() {}

// Compile time check that the concrete type meets the interface
var _ Binary = (*binary)(nil)

func (b *binary) Length() *Length { return b.len }

func (b *binary) Validate(ctx ValidateCtx, path []string, s string) error {
	return b.len.Validate(uint64(len(s)))
}
func NewBinary() Binary {
	return &binary{}
}

type Boolean interface {
	Type
	isBoolean()
}

type boolean struct {
	ytyp
}

// Ensure that other schema types don't meet the interface
func (*boolean) isBoolean() {}

// Compile time check that the concrete type meets the interface
var _ Boolean = (*boolean)(nil)

func (b *boolean) Validate(ctx ValidateCtx, path []string, s string) error {
	if s == "true" || s == "false" {
		return nil
	}
	return newInvalidValueError(path, genErrorString(b))
}

func (b *boolean) errors() []string {
	return []string{"true", "false"}
}

func NewBoolean(
	name xml.Name,
	def string,
	hasDef bool,
) Boolean {

	return &boolean{ytyp: newType(name, def, hasDef)}
}

type Decimal64 interface {
	Number
	Fd() Fracdigit
	Rbs() DrbSlice
	isDecimal64()
}

type decimal64 struct {
	ytyp
	fd     Fracdigit
	rbs    DrbSlice
	msg    string
	appTag string
}

// Ensure that other schema types don't meet the interface
func (*decimal64) isDecimal64() {}

// Compile time check that the concrete type meets the interface
var _ Decimal64 = (*decimal64)(nil)

func (d *decimal64) Fd() Fracdigit               { return d.fd }
func (d *decimal64) Rbs() DrbSlice               { return d.rbs }
func (d *decimal64) Ranges() RangeBoundarySlicer { return d.rbs }
func (d *decimal64) Msg() string                 { return d.msg }
func (d *decimal64) AppTag() string              { return d.appTag }
func (d *decimal64) BitWidth() BitWidth          { return BitWidth64 }

func NewDecimal64(
	name xml.Name,
	fd Fracdigit,
	rbs []Drb,
	msg string,
	appTag string,
	def string,
	hasDef bool,
) Decimal64 {
	if rbs == nil {
		rbs = make([]Drb, 0, 1)
		if fd > 0 {
			rbs = append(rbs, fdtab[fd])
		}
	}
	if appTag == "" {
		appTag = "range-violation"
	}
	return &decimal64{
		ytyp:   newType(name, def, hasDef),
		rbs:    rbs,
		msg:    msg,
		appTag: appTag,
		fd:     fd,
	}
}

func (d *decimal64) Validate(ctx ValidateCtx, path []string, s string) error {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		goto out
	}

	err = validateDecimal64String(s, int(d.fd))
	if err != nil {
		goto out
	}

	if len(d.rbs) == 0 {
		rb, ok := fdtab[d.fd]
		if ok {
			err = rb.Validate(f)
			if err != nil {
				goto out
			}
		}
		goto out
	}
	for _, r := range d.rbs {
		err = r.Validate(f)
		if err == nil {
			break
		}
	}
out:
	if err == nil {
		return nil
	}
	if d.msg != "" {
		return newInvalidValueErrorWithAppTag(path, d.msg, d.appTag)
	}
	switch v := err.(type) {
	case *strconv.NumError:
		if v.Err == strconv.ErrSyntax {
			return newInvalidValueErrorWithAppTag(path,
				fmt.Sprintf("%s is not a decimal64", s), d.appTag)
		}
		return newInvalidValueErrorWithAppTag(path, genErrorString(d), d.appTag)
	case *validateDecimal64Error:
		return newInvalidValueErrorWithAppTag(path, err.Error(), d.appTag)
	default:
		return newInvalidValueErrorWithAppTag(path, genErrorString(d), d.appTag)
	}
}

func (d *decimal64) errors() []string {
	out := make([]string, 0, len(d.rbs))
	for _, rb := range d.rbs {
		out = append(out, rb.Error())
	}
	return out
}

type Empty interface {
	Type
	isEmpty()
}

type empty struct {
	ytyp
}

// Ensure that other schema types don't meet the interface
func (*empty) isEmpty() {}

// Compile time check that the concrete type meets the interface
var _ Empty = (*empty)(nil)

func (*empty) Validate(ctx ValidateCtx, path []string, s string) error {
	if s != "" {
		if len(path) > 1 {
			return NewEmptyLeafValueError(s, path[:len(path)-1])
		}
		return NewEmptyLeafValueError(s, []string{})
	}
	return nil
}
func NewEmpty(name xml.Name, def string, hasDef bool) Empty {
	return &empty{ytyp: newType(name, def, hasDef)}
}

type Enumeration interface {
	Type
	Enums() []*Enum
	String() string
	isEnumeration()
}

type enumeration struct {
	ytyp
	enums []*Enum
}

// Ensure that other schema types don't meet the interface
func (*enumeration) isEnumeration() {}

// Compile time check that the concrete type meets the interface
var _ Enumeration = (*enumeration)(nil)

func (e *enumeration) Enums() []*Enum { return e.enums }

func (e *enumeration) String() string {
	var s string
	s = e.enums[0].Val
	for _, e := range e.enums[1:] {
		s = s + ", " + e.Val
	}
	return s
}
func (e *enumeration) Validate(ctx ValidateCtx, path []string, s string) error {
	for _, en := range e.enums {
		if en.Val == s {
			return nil
		}
	}
	return newInvalidValueError(path, genErrorString(e))
}

func (e *enumeration) errors() []string {
	out := make([]string, 0, len(e.enums))
	for _, enum := range e.enums {
		if enum.Status() == Obsolete {
			continue
		}
		out = append(out, enum.Val)
	}
	return out
}

func NewEnumeration(
	name xml.Name,
	enums []*Enum,
	def string,
	hasDef bool,
) Enumeration {
	if enums == nil {
		enums = make([]*Enum, 0)
	}

	return &enumeration{
		ytyp:  newType(name, def, hasDef),
		enums: enums,
	}
}

type Number interface {
	Type
	Ranges() RangeBoundarySlicer
	Msg() string
	AppTag() string
	BitWidth() BitWidth
}

type Integer interface {
	Number
	Rbs() RbSlice
	isInteger()
}

type integer struct {
	ytyp
	t      BitWidth
	rbs    RbSlice
	msg    string
	appTag string
}

// Ensure that other schema types don't meet the interface
func (*integer) isInteger() {}

// Compile time check that the concrete type meets the interface
var _ Integer = (*integer)(nil)

func (i *integer) Rbs() RbSlice                { return i.rbs }
func (i *integer) Ranges() RangeBoundarySlicer { return i.rbs }
func (i *integer) Msg() string                 { return i.msg }
func (i *integer) AppTag() string              { return i.appTag }
func (i *integer) BitWidth() BitWidth          { return i.t }

func (i *integer) Validate(ctx ValidateCtx, path []string, s string) error {
	var si int64
	var e error
	si, e = strconv.ParseInt(s, 10, int(i.t))
	if e != nil {
		goto out
	}
	if len(i.rbs) == 0 {
		rb, ok := inttab[i.t]
		if ok {
			e = rb.Validate(si)
			if e != nil {
				goto out
			}
		}
		goto out
	}
	for _, r := range i.rbs {
		e = r.Validate(si)
		if e == nil {
			break
		}
	}
out:
	if e == nil {
		return nil
	}
	if i.msg != "" {
		return newInvalidValueErrorWithAppTag(path, i.msg, i.appTag)
	}
	switch v := e.(type) {
	case *strconv.NumError:
		if v.Err == strconv.ErrSyntax {
			return newInvalidValueErrorWithAppTag(path,
				fmt.Sprintf("'%s' is not an int%d", s, i.t), i.appTag)
		}
		return newInvalidValueErrorWithAppTag(path, genErrorString(i), i.appTag)
	default:
		return newInvalidValueErrorWithAppTag(path, genErrorString(i), i.appTag)
	}
}

func (i *integer) errors() []string {
	out := make([]string, 0, len(i.Rbs()))
	for _, rb := range i.Rbs() {
		out = append(out, rb.Error())
	}
	return out
}

func NewInteger(
	bitSize BitWidth,
	name xml.Name,
	rbs []Rb,
	msg string,
	appTag string,
	def string,
	hasDef bool,
) Integer {

	if rbs == nil {
		rbs = make([]Rb, 0, 1)
		rbs = append(rbs, inttab[bitSize])
	}
	if appTag == "" {
		appTag = "range-violation"
	}
	return &integer{
		ytyp:   newType(name, def, hasDef),
		t:      bitSize,
		rbs:    rbs,
		msg:    msg,
		appTag: appTag,
	}
}

type Uinteger interface {
	Number
	Rbs() UrbSlice
	isUinteger()
}

type uinteger struct {
	ytyp
	t      BitWidth
	rbs    UrbSlice
	msg    string
	appTag string
}

// Ensure that other schema types don't meet the interface
func (*uinteger) isUinteger() {}

// Compile time check that the concrete type meets the interface
var _ Uinteger = (*uinteger)(nil)

func (i *uinteger) Rbs() UrbSlice               { return i.rbs }
func (i *uinteger) Ranges() RangeBoundarySlicer { return i.rbs }
func (i *uinteger) Msg() string                 { return i.msg }
func (i *uinteger) AppTag() string              { return i.appTag }
func (i *uinteger) BitWidth() BitWidth          { return i.t }

func (i *uinteger) Validate(ctx ValidateCtx, path []string, s string) error {
	var ui uint64
	var e error
	ui, e = strconv.ParseUint(s, 10, int(i.t))
	if e != nil {
		goto out
	}
	if len(i.rbs) == 0 {
		rb, ok := uinttab[i.t]
		if ok {
			e = rb.Validate(ui)
			if e != nil {
				goto out
			}
		}
		goto out
	}
	for _, r := range i.rbs {
		e = r.Validate(ui)
		if e == nil {
			break
		}
	}
out:
	if e == nil {
		return nil
	}
	if i.Msg() != "" {
		return newInvalidValueErrorWithAppTag(path, i.Msg(), i.AppTag())
	}
	switch v := e.(type) {
	case *strconv.NumError:
		if v.Err == strconv.ErrSyntax {
			return newInvalidValueErrorWithAppTag(path,
				fmt.Sprintf("'%s' is not an uint%d", s, i.t), i.AppTag())
		}
		return newInvalidValueErrorWithAppTag(path,
			genErrorString(i), i.AppTag())
	default:
		return newInvalidValueErrorWithAppTag(path,
			genErrorString(i), i.AppTag())
	}
}

func (i *uinteger) errors() []string {
	out := make([]string, 0, len(i.Rbs()))
	for _, rb := range i.Rbs() {
		out = append(out, rb.Error())
	}
	return out
}

func NewUinteger(
	bitSize BitWidth,
	name xml.Name,
	rbs []Urb,
	msg string,
	appTag string,
	def string,
	hasDef bool,
) Uinteger {
	if rbs == nil {
		rbs = make([]Urb, 0, 1)
		rbs = append(rbs, uinttab[bitSize])
	}
	if appTag == "" {
		appTag = "range-violation"
	}
	return &uinteger{
		ytyp:   newType(name, def, hasDef),
		t:      bitSize,
		rbs:    rbs,
		msg:    msg,
		appTag: appTag,
	}
}

type String interface {
	Type
	Len() *Length
	Pats() [][]Pattern
	PatHelps() [][]string
	isString()
}

type ystring struct {
	ytyp
	len      *Length
	pats     [][]Pattern
	pathelps [][]string
}

// Ensure that other schema types don't meet the interface
func (*ystring) isString() {}

// Compile time check that the concrete type meets the interface
var _ String = (*ystring)(nil)

func (s *ystring) Len() *Length         { return s.len }
func (s *ystring) Pats() [][]Pattern    { return s.pats }
func (s *ystring) PatHelps() [][]string { return s.pathelps }

func (y *ystring) Validate(ctx ValidateCtx, path []string, s string) error {
	var err error
	err = y.len.Validate(uint64(len(s)))
	if err != nil {
		switch merr := err.(type) {
		case *mgmterror.InvalidValueApplicationError:
			merr.Path = pathutil.Pathstr(path)
			return merr
		default:
			return err
		}
	}

	//patterns must contain all subtype patterns leading up to to this type
	//every type can have or patterns. so we have a two demensional matrix of patterns.
	for _, ps := range y.pats {
		for _, p := range ps {
			err = p.Validate(s)
			if err != nil {
				switch merr := err.(type) {
				case *mgmterror.InvalidValueApplicationError:
					merr.Path = pathutil.Pathstr(path)
					return merr
				default:
					return err
				}
			}
		}
	}
	return nil
}

func (y *ystring) errors() []string {
	out := make([]string, 0, len(y.PatHelps()))
	for _, pathelps := range y.PatHelps() {
		for _, pathelp := range pathelps {
			out = append(out, pathelp)
		}
	}
	return out
}

func NewString(
	name xml.Name,
	pats [][]Pattern,
	pathelps [][]string,
	initlen *Length,
	def string,
	hasDef bool,
) String {
	r := uinttab[BitWidth32]
	if initlen == nil {
		initlen = &Length{
			Lbs: []Lb{
				Lb{
					Start: r.Start,
					End:   r.End,
				},
			},
		}
	}
	if pats == nil {
		pats = make([][]Pattern, 0)
	}
	if pathelps == nil {
		pathelps = make([][]string, 0)
	}

	return &ystring{
		ytyp:     newType(name, def, hasDef),
		pats:     pats,
		pathelps: pathelps,
		len:      initlen,
	}
}

type Union interface {
	Type
	Typs() []Type
	MatchType(ctx ValidateCtx, path []string, s string) Type
	isUnion()
}

type union struct {
	ytyp
	typs []Type
}

// Ensure that other schema types don't meet the interface
func (*union) isUnion() {}

// Compile time check that the concrete type meets the interface
var _ Union = (*union)(nil)

func (u *union) Typs() []Type { return u.typs }

func (u *union) MatchType(ctx ValidateCtx, path []string, s string) Type {
	for _, t := range u.typs {
		err := t.Validate(ctx, path, s)
		if err == nil {
			if u, ok := t.(Union); ok {
				// a type within a union matched
				// get the non-union base type
				return u.MatchType(ctx, path, s)
			}
			return t
		}
	}

	return nil
}

func (u *union) Validate(ctx ValidateCtx, path []string, s string) error {
	var err error
	var matched bool
	for _, t := range u.typs {
		err = t.Validate(ctx, path, s)
		if err == nil {
			matched = true
			break
		}
	}
	if !matched {
		return newInvalidValueError(path, genErrorString(u))
	}
	if err == nil {
		return nil
	}
	return newInvalidValueError(path, genErrorString(u))
}
func (u *union) errors() []string {
	var out []string
	for _, t := range u.typs {
		for _, estr := range t.errors() {
			out = append(out, estr)
		}
	}
	return out
}
func NewUnion(
	name xml.Name,
	typs []Type,
	def string,
	hasDef bool,
) Union {

	if typs == nil {
		typs = make([]Type, 0)
	}

	return &union{
		ytyp: newType(name, def, hasDef),
		typs: typs,
	}
}

type Identityref interface {
	Type
	Identities() []*Identity
	String() string
	isIdentityRef()
}

type identityref struct {
	ytyp
	identities []*Identity
}

// Ensure that other schema types don't meet the interface
func (*identityref) isIdentityRef() {}

// Compile time check that the concrete type meets the interface
var _ Identityref = (*identityref)(nil)

func (i *identityref) Identities() []*Identity { return i.identities }

func (i *identityref) String() string {
	var s string
	s = i.identities[0].Val
	for _, id := range i.identities[1:] {
		s = s + ", " + id.Val
	}
	return s
}
func (i *identityref) Validate(ctx ValidateCtx, path []string, s string) error {
	for _, id := range i.identities {
		if id.Val == s {
			return nil
		}
	}
	return newInvalidValueError(path, genErrorString(i))
}

func (i *identityref) errors() []string {
	out := make([]string, 0, len(i.identities))
	for _, id := range i.identities {
		if id.Status() == Obsolete {
			continue
		}
		out = append(out, id.Val)
	}
	return out
}

func NewIdentityref(
	name xml.Name,
	ids []*Identity,
	def string,
	hasDef bool,
) Identityref {
	if ids == nil {
		ids = make([]*Identity, 0)
	}

	return &identityref{
		ytyp:       newType(name, def, hasDef),
		identities: ids,
	}
}

type InstanceId interface {
	Type
	Require() bool
	isInstanceId()
}

type instanceId struct {
	ytyp
	require bool
}

// Ensure that other schema types don't meet the interface
func (*instanceId) isInstanceId() {}

// Compile time check that the concrete type meets the interface
var _ InstanceId = (*instanceId)(nil)

func (i *instanceId) Require() bool { return i.require }

func (i *instanceId) Validate(ctx ValidateCtx, path []string, s string) error {
	// TODO Ensure valid XPATH with limited predicates
	return nil
}

func NewInstanceId(
	name xml.Name,
	require bool,
	def string,
	hasDef bool,
) InstanceId {
	return &instanceId{
		ytyp:    newType(name, def, hasDef),
		require: require,
	}
}

type Leafref interface {
	Type
	Mach() *xpath.Machine
	GetAbsPath(xutils.PathType) xutils.PathType
	isLeafref()
}

type leafref struct {
	ytyp
	mach *xpath.Machine
}

// Ensure that other schema types don't meet the interface
func (*leafref) isLeafref() {}

// Compile time check that the concrete type meets the interface
var _ Leafref = (*leafref)(nil)

func (l *leafref) Mach() *xpath.Machine { return l.mach }

func (i *leafref) Validate(ctx ValidateCtx, path []string, s string) error {
	// Validation done at compile stage
	return nil
}

func NewLeafref(
	name xml.Name,
	mach *xpath.Machine,
	def string,
	hasDef bool,
) Leafref {
	return &leafref{
		ytyp: newType(name, def, hasDef),
		mach: mach,
	}
}

func (lr *leafref) AllowedValues(
	ctxNode xutils.XpathNode,
	debug bool,
) (allowedValues []string, err error) {
	return lr.mach.AllowedValues(ctxNode, debug)
}

// Get the absolute path pointed to by the leafref.  Initially for pretty-
// printing, it may prove useful when validating leafref paths at compile
// time in due course ...
//
// For now curPath has leaf value included.
func (lr *leafref) GetAbsPath(
	curPath xutils.PathType,
) xutils.PathType {
	return xutils.GetAbsPath(lr.mach.GetExpr(), curPath)
}

type Bit struct {
	Name   string
	Desc   string
	Ref    string
	status Status
	Pos    int32
}

func (b *Bit) String() string {
	return b.Name
}

func (b *Bit) Status() Status {
	return b.status
}

func NewBit(name, desc, ref string, status Status, pos int32) *Bit {
	return &Bit{
		Name:   name,
		Desc:   desc,
		Ref:    ref,
		status: status,
		Pos:    pos,
	}
}

type Bits interface {
	Type
	Bits() []*Bit
	isBits()
}

type bits struct {
	ytyp
	Bs []*Bit
}

// Ensure that other schema types don't meet the interface
func (*bits) isBits() {}

// Compile time check that the concrete type meets the interface
var _ Bits = (*bits)(nil)

func (b *bits) Validate(ctx ValidateCtx, path []string, s string) error {
	return nil
}

func (b *bits) Bits() []*Bit {
	return b.Bs
}

func NewBits(bs []*Bit) Bits {
	if bs == nil {
		return &bits{Bs: make([]*Bit, 0, 1)}
	}
	return &bits{Bs: bs}
}

/* Restrictions */
type Restriction interface {
	restriction()
}

type Fracdigit int

func (Fracdigit) restriction() {}

type Pattern struct {
	Pattern string
	*regexp.Regexp
	Msg    string
	AppTag string
}

func (p Pattern) String() string {
	return p.Pattern
}

func (p Pattern) shortString() string {
	str := p.Pattern
	if len(str) > 15 {
		return fmt.Sprintf("%.12s...", str)
	}
	return str
}

func (p Pattern) Validate(s string) error {
	if p.MatchString(s) {
		return nil
	}
	merr := mgmterror.NewInvalidValueApplicationError()
	merr.Message = p.message()
	if p.AppTag == "" {
		merr.AppTag = "pattern-violation"
	} else {
		merr.AppTag = p.AppTag
	}
	merr.Info = append(merr.Info, *mgmterror.NewMgmtErrorInfoTag(
		mgmterror.VyattaNamespace, "pattern", p.String()))
	if p.Msg != "" {
		merr.Info = append(merr.Info, *mgmterror.NewMgmtErrorInfoTag(
			mgmterror.VyattaNamespace, "message", p.Msg))
	}
	return merr
}

func (p Pattern) message() string {
	if p.Msg == "" {
		return "Does not match pattern " + p.shortString()
	}
	return p.Msg
}

func (p Pattern) restriction() {}
func NewPattern(re *regexp.Regexp) Pattern {
	return Pattern{Regexp: re}
}

// Interface for ALL range boundary (<n>rb) types that allows us to have
// generic handling routines in the compiler for range creation and validation
// etc.
type RangeBoundarySlicer interface {
	// Return length of RangeBoundary
	Len() int

	// Pretty-print function for %s
	String(i int) string

	// Return start/end value at given position in the slice
	GetStart(i int) interface{}
	GetEnd(i int) interface{}

	// For ranges with discrete values (ie int/uint/string), if one range
	// ends at 'x' and the following one starts at 'x + 1', then the two
	// ranges may be deemed to be contiguous, and a derived type or
	// refined range may span the two.
	//
	// NB: decimal64 covers real rather than whole numbers and so no two
	// ranges can ever be considered contiguous.
	Contiguous(lower, higher interface{}) bool

	// Return a RangeBoundarySlicer object with given entries and capacity.
	Create(entries, capacity int) RangeBoundarySlicer

	// Parse the string to obtain the numeric values for the range(s)
	// specified.
	Parse(start string, base int, bitSize int) (interface{}, error)

	// Create a new slice element with start and end and append it to the
	// slice.
	Append(start, end interface{}) RangeBoundarySlicer

	// These two don't really belong here - a type switch function would
	// perhaps be more appropriate - as they only use the RangeBoundary to
	// implicitly tell what underlying type to use.
	LessThan(first, second interface{}) bool
	GreaterThan(first, second interface{}) bool
}

// int range boundary (Rb)
type Rb struct {
	Start, End int64
}

func (r Rb) Validate(i int64) error {
	if i < r.Start || i > r.End {
		return r
	}
	return nil
}
func (r Rb) String() string {
	if r.Start != r.End {
		return fmt.Sprintf("%d..%d", r.Start, r.End)
	}
	return fmt.Sprintf("%d", r.Start)
}
func (r Rb) Error() string {
	if r.Start != r.End {
		return fmt.Sprintf("between %d and %d", r.Start, r.End)
	}
	return fmt.Sprintf("equal to %d", r.Start)
}
func (Rb) restriction() {}
func NewRangeBoundary(start, end int64) Rb {
	return Rb{Start: start, End: end}
}

// Rb RangeBoundarySlicer interface implementation
type RbSlice []Rb

func (rangeBdry RbSlice) String(i int) string { return rangeBdry[i].String() }

func (rangeBdry RbSlice) Len() int { return len(rangeBdry) }

func (rangeBdry RbSlice) GetStart(i int) interface{} {
	return rangeBdry[i].Start
}

func (rangeBdry RbSlice) GetEnd(i int) interface{} {
	return rangeBdry[i].End
}

func (rangeBdry RbSlice) LessThan(first, second interface{}) bool {
	return (first.(int64) < second.(int64))
}

func (rangeBdry RbSlice) GreaterThan(first, second interface{}) bool {
	return (first.(int64) > second.(int64))
}

func (rangeBdry RbSlice) Contiguous(lower, higher interface{}) bool {
	return ((lower.(int64) + 1) == higher.(int64))
}

func (rangeBdry RbSlice) Create(entries, capacity int) RangeBoundarySlicer {
	return (make(RbSlice, entries, capacity))
}

func (rangeBdry RbSlice) Parse(start string, base int, bitSize int) (interface{}, error) {
	return strconv.ParseInt(start, base, bitSize)
}

func (rangeBdry RbSlice) Append(start, end interface{}) RangeBoundarySlicer {
	return append(rangeBdry, Rb{Start: start.(int64), End: end.(int64)})
}

// uint range boundary (Urb)
type Urb struct {
	Start, End uint64
}

func (r Urb) Validate(i uint64) error {
	if i < r.Start || i > r.End {
		return r
	}
	return nil
}
func (r Urb) String() string {
	if r.Start != r.End {
		return fmt.Sprintf("%d..%d", r.Start, r.End)
	}
	return fmt.Sprintf("%d", r.Start)
}
func (r Urb) Error() string {
	if r.Start != r.End {
		return fmt.Sprintf("between %d and %d", r.Start, r.End)
	}
	return fmt.Sprintf("equal to %d", r.Start)
}
func (Urb) restriction() {}
func NewUnsignedRangeBoundary(start, end uint64) Urb {
	return Urb{Start: start, End: end}
}

// Urb RangeBoundarySlicer interface implementation
type UrbSlice []Urb

func (rangeBdry UrbSlice) String(i int) string { return rangeBdry[i].String() }

func (rangeBdry UrbSlice) Len() int { return len(rangeBdry) }

func (rangeBdry UrbSlice) GetStart(i int) interface{} {
	return rangeBdry[i].Start
}

func (rangeBdry UrbSlice) GetEnd(i int) interface{} {
	return rangeBdry[i].End
}

func (rangeBdry UrbSlice) LessThan(first, second interface{}) bool {
	return (first.(uint64) < second.(uint64))
}

func (rangeBdry UrbSlice) GreaterThan(first, second interface{}) bool {
	return (first.(uint64) > second.(uint64))
}

func (rangeBdry UrbSlice) Contiguous(lower, higher interface{}) bool {
	return ((lower.(uint64) + 1) == higher.(uint64))
}

func (rangeBdry UrbSlice) Create(entries, capacity int) RangeBoundarySlicer {
	return (make(UrbSlice, entries, capacity))
}

func (rangeBdry UrbSlice) Parse(start string, base int, bitSize int) (interface{}, error) {
	return strconv.ParseUint(start, base, bitSize)
}

func (rangeBdry UrbSlice) Append(start, end interface{}) RangeBoundarySlicer {
	return append(rangeBdry, Urb{Start: start.(uint64), End: end.(uint64)})
}

// decimal64 range boundary (Drb)
type Drb struct {
	Start, End float64
}

func (r Drb) Validate(i float64) error {
	if i < r.Start || i > r.End {
		return r
	}
	return nil
}
func (r Drb) String() string {
	if r.Start != r.End {
		return fmt.Sprintf("%f..%f", r.Start, r.End)
	}
	return fmt.Sprintf("%f", r.Start)
}
func (r Drb) Error() string {
	if r.Start != r.End {
		return fmt.Sprintf("between %f and %f", r.Start, r.End)
	}
	return fmt.Sprintf("equal to %f", r.Start)
}
func (Drb) restriction() {}

func NewDecimalRangeBoundary(start, end float64) Drb {
	return Drb{Start: start, End: end}
}

// To allow us to provide common handling for different RangeBoundary types
// it is useful to create an explicit slice type for each RB type as we can
// then create functions that operate on such slices.  We use these functions
// from the common handling functions using the magic of interface{}.
type DrbSlice []Drb

func (rangeBdry DrbSlice) String(i int) string { return rangeBdry[i].String() }

func (rangeBdry DrbSlice) Len() int { return len(rangeBdry) }

func (rangeBdry DrbSlice) GetStart(i int) interface{} {
	return rangeBdry[i].Start
}

func (rangeBdry DrbSlice) GetEnd(i int) interface{} {
	return rangeBdry[i].End
}

func (rangeBdry DrbSlice) LessThan(first, second interface{}) bool {
	return (first.(float64) < second.(float64))
}

func (rangeBdry DrbSlice) GreaterThan(first, second interface{}) bool {
	return (first.(float64) > second.(float64))
}

func (rangeBdry DrbSlice) Contiguous(lower, higher interface{}) bool {
	return false
}

func (rangeBdry DrbSlice) Create(entries, capacity int) RangeBoundarySlicer {
	return (make(DrbSlice, entries, capacity))
}

func (rangeBdry DrbSlice) Parse(start string, base int, bitSize int) (interface{}, error) {
	return strconv.ParseFloat(start, bitSize)
}

func (rangeBdry DrbSlice) Append(start, end interface{}) RangeBoundarySlicer {
	return append(rangeBdry, Drb{Start: start.(float64), End: end.(float64)})
}

// string length (range) boundary (Lb)
type Lb struct {
	yrestrict
	Start, End uint64
}

func (l *Lb) Validate(i uint64) error {
	if i < l.Start || i > l.End {
		return l
	}
	return nil
}
func (l *Lb) String() string {
	if l.Start != l.End {
		return fmt.Sprintf("%d..%d", l.Start, l.End)
	}
	return fmt.Sprintf("%d", l.Start)
}
func (l *Lb) Error() string {
	if l.Start != l.End {
		return fmt.Sprintf("have length between %d and %d characters",
			l.Start, l.End)
	}
	return fmt.Sprintf("have length of %d characters", l.Start)
}
func NewLengthBoundary(start, end uint64) Lb {
	return Lb{Start: start, End: end}
}

// Lb RangeBoundarySlicer interface implementation
type LbSlice []Lb

func (rangeBdry LbSlice) String(i int) string { return rangeBdry[i].String() }

func (rangeBdry LbSlice) Len() int { return len(rangeBdry) }

func (rangeBdry LbSlice) GetStart(i int) interface{} {
	return rangeBdry[i].Start
}

func (rangeBdry LbSlice) GetEnd(i int) interface{} {
	return rangeBdry[i].End
}

func (rangeBdry LbSlice) LessThan(first, second interface{}) bool {
	return (first.(uint64) < second.(uint64))
}

func (rangeBdry LbSlice) GreaterThan(first, second interface{}) bool {
	return (first.(uint64) > second.(uint64))
}

func (rangeBdry LbSlice) Contiguous(lower, higher interface{}) bool {
	return ((lower.(uint64) + 1) == higher.(uint64))
}

func (rangeBdry LbSlice) Create(entries, capacity int) RangeBoundarySlicer {
	return (make(LbSlice, entries, capacity))
}

func (rangeBdry LbSlice) Parse(start string, base int, bitSize int) (interface{}, error) {
	return start, nil
}

func (rangeBdry LbSlice) Append(start, end interface{}) RangeBoundarySlicer {
	return append(rangeBdry, Lb{Start: start.(uint64), End: end.(uint64)})
}

// Length
type Length struct {
	yrestrict
	Lbs    LbSlice
	Msg    string
	AppTag string
}

func (l *Length) Validate(i uint64) error {
	var err error
	if l == nil {
		return nil
	}
	for _, lb := range l.Lbs {
		err = lb.Validate(i)
		if err == nil {
			break
		}
	}

	if err == nil {
		return nil
	}
	errMsg := l.Msg
	if errMsg == "" {
		errMsg = l.errMsg()
	}
	merr := mgmterror.NewInvalidValueApplicationError()
	merr.Message = errMsg
	if l.AppTag == "" {
		merr.AppTag = "length-violation"
	} else {
		merr.AppTag = l.AppTag
	}

	// While this is a lame info tag, its presence indicates this
	// is a length error
	merr.Info = append(merr.Info, *mgmterror.NewMgmtErrorInfoTag(
		mgmterror.VyattaNamespace, "length", "Invalid length"))
	return merr
}

func (l *Length) errors() []string {
	out := make([]string, 0, len(l.Lbs))
	for _, lb := range l.Lbs {
		out = append(out, lb.Error())
	}
	return out
}

func (l *Length) errMsg() string {
	var buf bytes.Buffer
	errstrs := l.errors()
	if len(errstrs) == 1 {
		return "Must " + errstrs[0]
	}
	buf.WriteString("Must be one of the following: ")
	for i, estr := range errstrs {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(estr)
	}
	return buf.String()
}

type Identity struct {
	yrestrict
	Val       string
	Desc      string
	Ref       string
	status    Status
	Value     string
	Module    string
	Namespace string
}

func NewIdentity(mod, namespace, val, desc, ref string, status Status, value string) *Identity {
	return &Identity{Module: mod, Namespace: namespace, Val: val, Desc: desc, Ref: ref, status: status, Value: value}
}

func (i *Identity) String() string {
	return i.Val
}

func (i *Identity) Status() Status {
	return i.status
}

type Enum struct {
	yrestrict
	Val    string
	Desc   string
	Ref    string
	status Status
	Value  int
}

func NewEnum(val, desc, ref string, status Status, value int) *Enum {
	return &Enum{Val: val, Desc: desc, Ref: ref, status: status, Value: value}
}

func (e *Enum) String() string {
	return e.Val
}

func (e *Enum) Status() Status {
	return e.status
}
