// Copyright (c) 2017-2019 AT&T Intellectual Property
// All rights reserved.
//
// Copyright (c) 2015-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/danos/mgmterror"
	"github.com/danos/utils/exec"
	"github.com/danos/yang/data/datanode"
	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/xutils"
)

func init() {
	exec.NewExecError = func(path []string, err string) error {
		return mgmterror.NewExecError(path, err)
	}
}

// An optional type to be passed into validation to allow
// extensions to operate
type ValidateCtx interface {
	ErrorHelpText() []string
	// For validating paths that could exist, eg when making a NETCONF request
	// for the tree under a non-presence container, we need to allow a path
	// that stops at a non-presence container to be considered valid.  Likewise
	// list names w/o entry name, leaves w/o values.
	AllowIncompletePaths() bool
}

func skipCheck(c xnode, valType ValidationType) bool {
	switch valType {
	case DontValidate:
		return true
	case ValidateState:
		return c.schema().Config()
	case ValidateConfig:
		return !c.schema().Config()
	case ValidateAll:
		return false
	}
	return false
}

func checkWhenAndMusts(
	c xnode,
	debug bool,
	valType ValidationType,
) ([]*exec.Output, []error, bool) {

	if skipCheck(c, valType) {
		return nil, nil, true
	}

	outs, errs := make([]*exec.Output, 0), make([]error, 0)

	// In theory we can only have a single 'when', but in the case of an
	// augment with a when directly under it, we append the when to each
	// augmented child instead, running the when with the context of the
	// parent.
	for _, ctxt := range c.schema().Whens() {
		checkWhenMachineFn := func() ([]*exec.Output, []error, bool) {
			return checkMachine(c, ctxt.Mach, ctxt.RunAsParent, "when", debug,
				ctxt.ErrMsg)
		}
		outs, errs, _ = exec.AppendOutput(checkWhenMachineFn, outs, errs)
	}

	// If when fails, then the node effectively doesn't exist.  As such
	// we should not run the must checks on a node deemed unconfigurable.
	if len(errs) > 0 {
		return outs, errs, true
	}

	// For must, we validate all checks and report all errors.
	for _, ctxt := range c.schema().Musts() {
		checkMachineFn := func() ([]*exec.Output, []error, bool) {
			return checkMachine(c, ctxt.Mach, false, "must", debug,
				ctxt.ErrMsg)
		}
		outs, errs, _ = exec.AppendOutput(checkMachineFn, outs, errs)
	}

	// Additionally, must statements on non-presence child containers
	// of configured nodes must be evaluated.
	checkNPContMustsFn := func() ([]*exec.Output, []error, bool) {
		return checkNPContMusts(c, debug)
	}
	outs, errs, _ = exec.AppendOutput(checkNPContMustsFn, outs, errs)

	return outs, errs, len(errs) == 0

}

// checkNPContMusts - evaluate musts on unconfigured NP containers.
//
// Similarly to mandatory statements, we need to evaluate must statements
// on unconfigured non-presence (NP) container child nodes of configured nodes.
// As we need to check the tree recursively, exec.AppendOutput doesn't work
// as it uses pass by value.  So, this wrapper around the internal function
// allows us to accumulate results by passing by reference instead.
func checkNPContMusts(c xnode, debug bool) ([]*exec.Output, []error, bool) {

	outs, errs := make([]*exec.Output, 0), make([]error, 0)

	ret_outs, ret_errs, ret_status :=
		checkNPContMustsInternal(c, debug, &outs, &errs)
	return *ret_outs, *ret_errs, ret_status
}

func checkNPContMustsInternal(
	c xnode,
	debug bool,
	outs *[]*exec.Output,
	errs *[]error,
) (*[]*exec.Output, *[]error, bool) {

	npContChildNodes := getUnconfiguredNPContainerChildren(c)

	var parent *xdatanode
	if dn, ok := c.(*xdatanode); ok {
		parent = dn
	}
	for _, npContNode := range npContChildNodes {
		// Create XNode representing the child, including an ephemeral
		// datanode so we have a source node on which to run the must
		// statement(s).
		npContXNode := createEphemeralXNode(
			datanode.CreateDataNode(npContNode.Name(), nil, nil),
			npContNode,
			parent)

		for _, ctxt := range npContXNode.schema().Musts() {
			new_outs, new_errs, _ := checkMachine(
				npContXNode, ctxt.Mach, false, "must", debug, ctxt.ErrMsg)
			*outs = append(*outs, new_outs...)
			*errs = append(*errs, new_errs...)
		}

		// Check children ...
		checkNPContMustsInternal(npContXNode, debug, outs, errs)
	}

	return outs, errs, true
}

// getUnconfiguredNPContainerChildren - filter children to return unconfigured
// non-presence container children only.
func getUnconfiguredNPContainerChildren(c xnode) map[string]Node {

	npContChildNodes := make(map[string]Node)

	for _, child := range c.schema().Children() {
		if v, ok := child.(Container); ok {
			if !v.Presence() {
				npContChildNodes[v.Name()] = child
			}
		}
	}

	if len(npContChildNodes) == 0 {
		return nil
	}

	// Remove configured children from the list of all non-presence container
	// children.
	configuredChildren := c.children()
	for key, _ := range npContChildNodes {
		for _, n := range configuredChildren {
			if key == n.YangDataName() {
				delete(npContChildNodes, key)
				break
			}
		}
	}

	return npContChildNodes
}

func checkLeafref(
	c xnode,
	lref Leafref,
	debug bool,
	valType ValidationType,
) ([]*exec.Output, []error, bool) {

	if skipCheck(c, valType) {
		return nil, nil, true
	}

	outs, errs := make([]*exec.Output, 0), make([]error, 0)

	if allowedValues, err := lref.AllowedValues(c, debug); err != nil {
		return outs, append(errs, err), false
	} else {
		for _, value := range allowedValues {
			if c.XValue() == value {
				return outs, errs, true
			}
		}
	}

	// TODO: do we check if require-instance=false somewhere?
	cerr := mgmterror.NewExecError(
		c.path(),
		fmt.Sprintf("The following path must exist:\n  [%s %s]",
			lref.GetAbsPath(c.path()).SpacedString(), c.XValue()))

	return outs, append(errs, cerr), false
}

// Helper function to get path to match node we're using as current context
// for the when statement.
func getPath(c xnode, runAsParent bool) []string {
	if runAsParent {
		return c.XParent().(xnode).path()
	}
	return c.path()
}

// Processing for when and must is identical, with same context
func checkMachine(
	c xnode,
	mach *xpath.Machine,
	runAsParent bool,
	checkName string,
	debug bool,
	errMsg string,
) ([]*exec.Output, []error, bool) {
	outs, errs := make([]*exec.Output, 0), make([]error, 0)

	if mach == nil {
		return outs, errs, true
	}

	var res *xpath.Result
	filter := xutils.FullTree
	if c.schema().Config() {
		filter = xutils.ConfigOnly
	}
	if runAsParent {
		res = xpath.NewCtxFromMach(mach, c.XParent()).
			SetDebug(debug).
			SetAccessibleTree(filter).
			Run()
	} else {
		res = xpath.NewCtxFromMach(mach, c).
			SetDebug(debug).
			SetAccessibleTree(filter).
			Run()
	}
	boolResult, err := res.GetBoolResult()
	if err != nil {
		// Machine failed to execute.
		return outs,
			append(errs, mgmterror.NewExecError(getPath(c, runAsParent),
				err.Error())),
			boolResult
	}

	if boolResult == false {
		return outs,
			append(errs, mgmterror.NewExecError(getPath(c, runAsParent),
				errMsg)),
			boolResult
	}

	return outs, errs, boolResult
}

func validateLeafSchema(
	c xnode,
	debug bool,
	valType ValidationType,
) ([]*exec.Output, []error, bool) {

	vals := c.YangDataValuesNoSorting()
	outs, errs := make([]*exec.Output, 0), make([]error, 0)

	err := c.schema().CheckCardinality(c.XPath(), len(vals))
	if err != nil {
		return outs, append(errs, err), false
	}

	children := c.children()
	for _, child := range children {
		checkWhenAndMustsFn := func() ([]*exec.Output, []error, bool) {
			return checkWhenAndMusts(child, debug, valType)
		}
		outs, errs, _ = exec.AppendOutput(checkWhenAndMustsFn, outs, errs)

		if lref, ok := child.schema().Type().(Leafref); ok {
			checkLeafrefFn := func() ([]*exec.Output, []error, bool) {
				return checkLeafref(child, lref, debug, valType)
			}
			outs, errs, _ = exec.AppendOutput(checkLeafrefFn, outs, errs)
		}
	}

	return outs, errs, len(errs) == 0

}

func appendMandatoryError(path []string, name string, errs []error) []error {
	err := mgmterror.NewExecError(path,
		fmt.Sprintf("Missing mandatory node %s", name))
	return append(errs, err)
}

// Check for mandatory children nodes, for each one found, create an error
func hasMandatoryChildren(sn Node, path []string, errs []error) []error {
	path = append(path, sn.Name())
	for _, csn := range sn.Children() {
		switch v := csn.(type) {
		case Leaf:
			if v.Mandatory() {
				errs = appendMandatoryError(path, v.Name(), errs)
			}
		case List:
			if v.Limit().Min > 0 {
				errs = appendMandatoryError(path, v.Name(), errs)
			}
		case LeafList:
			if v.Limit().Min > 0 {
				errs = appendMandatoryError(path, v.Name(), errs)
			}
		case Container:
			if !v.Presence() {
				errs = hasMandatoryChildren(csn, path, errs)
			}
		}
	}
	return errs
}

func checkMandatory(c xnode, valType ValidationType,
) ([]*exec.Output, []error, bool) {

	if skipCheck(c, valType) {
		return nil, nil, true
	}

	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	mandNodes := make(map[string]Node)
	for _, n := range c.schema().Children() {
		switch v := n.(type) {
		case Leaf:
			if v.Mandatory() {
				mandNodes[v.Name()] = n
			}
		case List:
			if v.Limit().Min > 0 {
				mandNodes[v.Name()] = n
			}
		case LeafList:
			if v.Limit().Min > 0 {
				mandNodes[v.Name()] = n
			}
		case Container:
			if !v.Presence() {
				// non-presence container is potentially
				// a mandatory node. Check later for
				// mandatory children.
				mandNodes[v.Name()] = n
			}
		}
	}

	children := c.children()
	for k, nd := range mandNodes {
		found := false
		for _, n := range children {
			if found = (k == n.YangDataName()); found {
				break
			}
		}
		if !found {
			if _, ok := nd.(Container); ok {
				// Non-presence container found, so
				// check for mandatory children.
				errs = hasMandatoryChildren(nd, c.path(), errs)
			} else {
				errs = appendMandatoryError(c.path(), k, errs)
			}
		}
	}
	return outs, errs, len(errs) == 0
}

func resolveDescendant(c xnode, path []xml.Name) string {

	if len(path) == 0 {
		return ""
	}
	hd, tl := path[0], path[1:]
	for _, ch := range c.children() {
		if ch.YangDataName() != hd.Local {
			continue
		}
		csn := c.schema().Child(ch.YangDataName())
		switch csn.(type) {
		case Container:
			return resolveDescendant(ch, tl)
		case Leaf:
			// Compiler enforces non-empty leaf reference
			return ch.YangDataValuesNoSorting()[0]
		default:
			return ""
		}
	}
	return ""
}

// If, and only if, the given config node contains ALL sub-nodes listed in
// the unique statement (uniques), return a string containing the value
// for each sub-node, separated by the 'middle dot' character.
//
// If any sub-node is not present, return an empty string. While this
// might seem unintuitive, this is what the RFC specifies.
func getUniqueKey(c xnode, uniques [][]xml.Name) string {

	var outs []string
	for _, uniq := range uniques {
		desc := resolveDescendant(c, uniq)
		if desc == "" {
			return ""
		}
		outs = append(outs, desc)
	}
	//use middle dot (U+00B7) to join strings so we don't have
	//problems with string values.
	return strings.Join(outs, "·")
}

func xmlPathToPath(path []xml.Name) []string {
	var p []string
	for _, e := range path {
		p = append(p, e.Local)
	}
	return p
}

func xmlPathJoin(path []xml.Name) string {
	var buf bytes.Buffer
	for i, e := range path {
		buf.WriteString(e.Local)
		if i != len(path)-1 {
			buf.WriteByte(' ')
		}
	}
	return buf.String()
}

func uniqueString(c xnode, uniques [][]xml.Name) string {

	var buf bytes.Buffer
	// Unique is kind of strange, if one supplies a space separated list
	// of paths it ensures that the concatenation of those resolved paths
	// (values at the end) is unique, but not the individual values
	// themselves. This allows one to do things like specify an ipaddress /
	// port pair that must be unique but where either the ip or port could
	// be reused.
	for i, uniq := range uniques {
		buf.WriteByte('[')
		buf.WriteString(xmlPathJoin(uniq))
		buf.WriteByte(' ')
		desc := resolveDescendant(c, uniq)
		buf.WriteString(desc)
		buf.WriteByte(']')
		if i != len(uniques)-1 {
			buf.WriteString(", ")
		}
	}
	return buf.String()
}

func uniquePaths(c xnode, uniques [][]xml.Name) [][]string {
	paths := make([][]string, len(uniques))
	for i, uniq := range uniques {
		paths[i] = xmlPathToPath(uniq)
		paths[i] = append(paths[i], resolveDescendant(c, uniq))
	}
	return paths
}

func checkUnique(c xnode, valType ValidationType,
) ([]*exec.Output, []error, bool) {

	if skipCheck(c, valType) {
		return nil, nil, true
	}

	outs, errs := make([]*exec.Output, 0), make([]error, 0)
	sch := c.schema().(List)
	for _, u := range sch.Uniques() {
		m := make(map[string][]xnode)
		for _, key := range c.children() {
			k := getUniqueKey(key, u)
			if k == "" {
				// We skip entries that don't have all the nodes present
				continue
			}
			m[k] = append(m[k], key)
		}
		for _, ks := range m {
			if len(ks) < 2 {
				continue
			}
			keys := make([]string, 0, len(ks))
			for _, k := range ks {
				keys = append(keys, k.YangDataName())
			}
			setStr := "path"
			if len(u) > 1 {
				setStr = "set of paths"
			}
			err := mgmterror.NewExecError(
				c.path(),
				fmt.Sprintf(
					"The following %s must be unique:\n\n"+
						"  %s\n\nbut is defined in the following set "+
						"of keys:\n\n  %s",
					setStr,
					uniqueString(ks[0], u),
					keys,
				))
			errs = append(errs, err)
		}
	}
	return outs, errs, len(errs) == 0
}

func validateListSchema(
	c xnode,
	debug bool,
	valType ValidationType,
) ([]*exec.Output, []error, bool) {

	children := c.children()
	outs, errs := make([]*exec.Output, 0), make([]error, 0)

	err := c.schema().CheckCardinality(c.XPath(), len(children))
	if err != nil {
		return outs, append(errs, err), false
	}
	checkUniqueFn := func() ([]*exec.Output, []error, bool) {
		return checkUnique(c, valType)
	}
	outs, errs, _ = exec.AppendOutput(checkUniqueFn, outs, errs)
	for _, n := range children {
		validateSchemaFn := func() ([]*exec.Output, []error, bool) {
			return validateSchema(n, debug, valType)
		}
		outs, errs, _ = exec.AppendOutput(validateSchemaFn, outs, errs)
	}
	return outs, errs, len(errs) == 0
}

func ValidateSchema(
	sn Node, c datanode.DataNode, debug bool,
) ([]*exec.Output, []error, bool) {

	xnode := createXNode(c, sn, nil)
	return validateSchema(xnode, debug, ValidateAll)
}

func validateSchema(
	c xnode,
	debug bool,
	valType ValidationType,
) ([]*exec.Output, []error, bool) {

	switch c.schema().(type) {
	case Leaf, LeafList:
		return validateLeafSchema(c, debug, valType)
	case List:
		return validateListSchema(c, debug, valType)
	}

	checkWhenAndMustsFn := func() ([]*exec.Output, []error, bool) {
		return checkWhenAndMusts(c, debug, valType)
	}
	outs, errs, _ := checkMandatory(c, valType)
	outs, errs, _ = exec.AppendOutput(checkWhenAndMustsFn, outs, errs)

	for _, n := range c.children() {
		validateSchemaFn := func() ([]*exec.Output, []error, bool) {
			return validateSchema(n, debug, valType)
		}
		outs, errs, _ = exec.AppendOutput(validateSchemaFn, outs, errs)
	}

	return outs, errs, len(errs) == 0
}

type ValidationType int

const (
	ValidateAll ValidationType = iota
	DontValidate
	ValidateState
	ValidateConfig
)

type SchemaValidator struct {
	xn      xnode
	debug   bool
	valType ValidationType
}

func NewSchemaValidator(sn Node, c datanode.DataNode) *SchemaValidator {

	return &SchemaValidator{
		xn:      createXNode(c, sn, nil),
		debug:   false,
		valType: ValidateAll}
}

func (sv *SchemaValidator) SetValidation(
	valType ValidationType,
) *SchemaValidator {

	sv.valType = valType
	return sv
}

func (sv *SchemaValidator) EnableDebug() *SchemaValidator {
	sv.debug = true
	return sv
}

func (sv *SchemaValidator) SetDebug(debug bool) *SchemaValidator {
	sv.debug = debug
	return sv
}

func (sv *SchemaValidator) Validate() ([]*exec.Output, []error, bool) {
	return validateSchema(sv.xn, sv.debug, sv.valType)
}
