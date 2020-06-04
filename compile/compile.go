// Copyright (c) 2017-2021, AT&T Intellectual Property. All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package compile

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"
	"runtime"
	"strings"

	"github.com/danos/utils/tsort"
	"github.com/danos/yang/parse"
	"github.com/danos/yang/schema"
	"github.com/danos/yang/xpath"
	"github.com/danos/yang/xpath/grammars/expr"
	"github.com/danos/yang/xpath/grammars/leafref"
	"github.com/danos/yang/xpath/grammars/path_eval"
	"github.com/danos/yang/xpath/xutils"
)

const DefaultCapsLocation = "/config/features"

//TODO: We should make this able to be configured.
//We are currently under a crunch, so doing this for now.
//This location is non-user changeable so only non-shipped features
//can be toggled.
const SystemCapsLocation = "/opt/vyatta/etc/features"

const emptyDefault = ""

type Config struct {
	YangDir       string
	YangLocations YangLocator
	CapsLocation  string
	Features      FeaturesChecker
	SkipUnknown   bool
	Filter        SchemaFilter
	// Used for the likes of yangc to inject names of valid custom functions
	// that otherwise would not be visible to the compiler.  Only used during
	// path evaluation.
	UserFnCheckFn xpath.UserCustomFunctionCheckerFn
}

func (c *Config) features() FeaturesChecker {
	return MultiFeatureCheckers(
		FeaturesFromLocations(true, c.CapsLocation),
		c.Features)
}

func (c *Config) yangLocations() YangLocator {
	return YangLocations(YangDirs(c.YangDir), c.YangLocations)
}

type SchemaType int

const (
	SchemaBool SchemaType = iota
	SchemaEmpty
	SchemaEnumeration
	SchemaIdentity
	SchemaInstanceId
	SchemaNumber
	SchemaDecimal64
	SchemaString
	SchemaUnion
	SchemaBits
	SchemaLeafRef
)

var validRestrictionsType = map[SchemaType]map[parse.NodeType]struct{}{
	SchemaBool: {
		// None allowed
	},
	SchemaEmpty: {
		// None allowed
	},
	SchemaEnumeration: {
		parse.NodeEnum: struct{}{},
	},
	SchemaIdentity: {
		// None allowed
	},
	SchemaInstanceId: {
		parse.NodeRequireInstance: struct{}{},
	},
	SchemaNumber: {
		parse.NodeRange:         struct{}{},
		parse.NodeConfigdSyntax: struct{}{},
	},
	SchemaDecimal64: {
		parse.NodeFractionDigits: struct{}{},
		parse.NodeRange:          struct{}{},
		parse.NodeConfigdSyntax:  struct{}{},
	},
	SchemaString: {
		parse.NodeLength:        struct{}{},
		parse.NodePattern:       struct{}{},
		parse.NodeConfigdSyntax: struct{}{},
	},
	SchemaUnion: {
		parse.NodeTyp: struct{}{},
	},
	SchemaBits: {
		parse.NodeBit: struct{}{},
	},
	SchemaLeafRef: {
		parse.NodePath: struct{}{},
	},
}

type Compiler struct {
	modules          map[string]*parse.Module
	modnames         []string
	submodules       map[string]*parse.Module
	skipUnknown      bool
	verifiedFeatures featuresMap
	featuresChecker  FeaturesChecker
	identities       map[string]parse.Node
	generateWarnings bool
	filter           SchemaFilter
	extensions       Extensions
	deviations       map[string]map[string]struct{}
	warnings         []xutils.Warning
	// Custom Fns list passed in to avoid false positive errors in path
	// evaluation for configd:must statements when using tools that are run
	// without custom function plugins present (eg yangc / DRAM).
	userFnChecker xpath.UserCustomFunctionCheckerFn
}

const (
	dontGenWarnings = false
	genWarnings     = true
)

// Get the system features
//
// If the file location/yang_module_name/feature_name exists, then the
// feature is enabled, otherwise, its disabled.
func getSystemFeatures(location string, features map[string]bool) {
	if location == "" {
		// None defined
		return
	}
	fi, err := os.Stat(location)
	if err != nil {
		// features do not exist
		return
	}

	if fi.Mode().IsDir() {
		d, err := os.Open(location)
		if err != nil {
			return
		}
		defer d.Close()

		names, err := d.Readdir(0)
		if err != nil {
			return
		}

		for _, name := range names {
			if name.IsDir() {
				featDir, err := os.Open(location + "/" + name.Name())
				sysFeatures, err := featDir.Readdir(0)
				if err != nil {
					// Skip any problematic directories
					continue
				}
				for _, feat := range sysFeatures {
					if !feat.IsDir() {
						features[name.Name()+":"+feat.Name()] = true
					}
				}
				featDir.Close()
			}
		}
	}
}

func NewCompiler(
	extensions Extensions,
	modules map[string]*parse.Module,
	submodules map[string]*parse.Module,
	features FeaturesChecker,
	skipUnknown, generateWarnings bool,
	filter SchemaFilter,
) *Compiler {

	c := &Compiler{}
	c.modules = modules
	c.submodules = submodules
	c.skipUnknown = skipUnknown
	c.verifiedFeatures = newFeaturesMap()
	c.generateWarnings = generateWarnings
	c.extensions = extensions
	c.filter = filter
	c.deviations = make(map[string]map[string]struct{})
	c.featuresChecker = features

	if dlog, err := syslog.NewLogger(syslog.LOG_DEBUG, 0); err == nil {
		xpath.SetDebugLogger(dlog)
	}

	return c
}

func (c *Compiler) addCustomFnChecker(
	userFnChecker xpath.UserCustomFunctionCheckerFn) *Compiler {
	c.userFnChecker = userFnChecker
	return c
}

func (c *Compiler) addDeviation(target, source string) {
	if t, ok := c.deviations[target]; ok {
		if _, ok := t[source]; !ok {
			t[source] = struct{}{}
		}
	} else {
		d := make(map[string]struct{})
		d[source] = struct{}{}
		c.deviations[target] = d
	}
}

func (c *Compiler) getDeviations(mod string) []string {
	devs := make([]string, 0)

	if t, ok := c.deviations[mod]; ok {
		for d, _ := range t {
			devs = append(devs, d)
		}
	}
	return devs
}

func (c *Compiler) featureEnabled(feature string) bool {
	if c.featuresChecker == nil {
		return false
	}
	return c.featuresChecker.Status(feature) == ENABLED
}

func (c *Compiler) verifiedFeatureEnabled(feature string) bool {
	return c.verifiedFeatures.Status(feature) == ENABLED
}

func (c *Compiler) recover(errp *error) {
	e := recover()
	if e != nil {
		if _, ok := e.(runtime.Error); ok {
			panic(e)
		}
		*errp = e.(error)
	}
}

func (c *Compiler) error(n parse.Node, err error) {
	s, _ := n.ErrorContext()
	panic(fmt.Errorf("%s: %s", s, err))
}

func (c *Compiler) saveWarning(n parse.Node, warn xutils.Warning) {
	c.warnings = append(c.warnings, warn)
}

func (c *Compiler) getWarnings() []xutils.Warning { return c.warnings }

func (c *Compiler) printDebug(format string, params ...interface{}) {
	if compilerDebugEnabled {
		fmt.Printf(format, params...)
	}
}

var compilerDebugEnabled = false

func EnableCompilerDebug()  { compilerDebugEnabled = true }
func DisableCompilerDebug() { compilerDebugEnabled = false }

// First part of range boundary validation.  Create our range from the
// parsed range passed in, referring to the base_rb min/max (which may be
// the default if we aren't refining / changing via typedef.
// Validation here is to ensure that each individual range we parse is within
// a range in the base_rb range set, including allowance for two base ranges
// to be contiguous (where they contain only whole numbers, so not dec64!).
func (comp *Compiler) createRangeBdry(node parse.Node,
	base_rb schema.RangeBoundarySlicer,
	parsed_rbs parse.RangeArgBdrySlice) (
	rangeBdrySlice schema.RangeBoundarySlicer) {

	// base_min/max represent the overall min / max values within which all
	// ranges must fit.  There may be gaps within this, but we check for that
	// later.
	var err error
	var base_min, base_max interface{}
	base_min = base_rb.GetStart(0)
	base_max = base_rb.GetEnd(base_rb.Len() - 1)

	// Loop through all parsed ranges and add them to our range boundary slice.
	var start, end interface{}
	rangeBdrySlice = base_rb.Create(0, len(parsed_rbs))
	for _, parsedRangeBdry := range parsed_rbs {
		if parsedRangeBdry.Min {
			start = base_min
		} else {
			start, err = rangeBdrySlice.Parse(parsedRangeBdry.Start, 0, 64)
			if err != nil {
				comp.error(node, err)
			}
			if rangeBdrySlice.LessThan(start, base_min) {
				comp.error(node, errors.New(
					"derived type range must be restrictive"))
			}
		}
		if parsedRangeBdry.Max {
			end = base_max
		} else {
			end, err = rangeBdrySlice.Parse(parsedRangeBdry.End, 0, 64)
			if err != nil {
				comp.error(node, err)
			}
			if rangeBdrySlice.GreaterThan(end, base_max) {
				comp.error(node, errors.New(
					"derived type range must be restrictive"))
			}
		}

		// Now we have our start and end, and know they are within the
		// overall min/max range, we need to check they are also within
		// each 'sub'range if base_rb has multiple ranges.  Note that
		// uint and int types can have 2 contiguous ranges as these
		// types contain only whole numbers.  dec64 contains real
		// numbers so between any 2 ranges there are 'missing' numbers.
		var rangeMin, rangeMax interface{}
		var curStart, curEnd interface{}
		for index := 0; index < base_rb.Len(); index++ {
			curStart = base_rb.GetStart(index)
			curEnd = base_rb.GetEnd(index)
			// Only update rangeMin if new range is not contiguous.
			// NB: beware decimal64 where no ranges are contiguous!
			if (index == 0) || !rangeBdrySlice.Contiguous(rangeMax, curStart) {
				rangeMin = curStart
			}
			rangeMax = curEnd

			if (!rangeBdrySlice.LessThan(start, rangeMin)) &&
				(!rangeBdrySlice.GreaterThan(end, rangeMax)) {
				// Start is big enough and end small enough.  Note start
				// could be > end, but in that case, we will catch that
				// in the validation below where we check for end < start.
				break
			}

			if rangeBdrySlice.LessThan(start, rangeMin) {
				// We can assume base_rb has been validated, so if no match
				// here we have a less restrictive range.
				comp.error(node, errors.New("derived range must be restrictive"))
			}

			// No point doing anything if pend > rangeMax.  Either we'll loop
			// to the next range and it will fit, or we will keep looping.
			// Bear in mind that we've already checked against absolute
			// max of last entry in range so we can't exceed that.
		}
		rangeBdrySlice = rangeBdrySlice.Append(start, end)
	}

	comp.validateRangeBoundaries(rangeBdrySlice, node)
	return rangeBdrySlice
}

// Sets of Range Boundaries have various requirements in terms of not
// overlapping, and each range starting with a higher starting value than
// the end of the previous one.  Using the RangeBoundarySlicer interface
// allows us to deal with all the different RangeBoundary (rb) types here.
func (comp *Compiler) validateRangeBoundaries(
	ranges schema.RangeBoundarySlicer,
	node parse.Node) {

	if ranges.LessThan(ranges.GetEnd(0), ranges.GetStart(0)) {
		comp.error(node, errors.New(
			"range start must be greater than or equal to range end"))
	}
	for i := 1; i < ranges.Len(); i++ {
		if ranges.LessThan(ranges.GetEnd(i), ranges.GetStart(i)) {
			comp.error(node, errors.New(
				"range start must be greater than or equal to range end"))
		}
		if ranges.GreaterThan(ranges.GetStart(i-1), ranges.GetStart(i)) {
			comp.error(node, fmt.Errorf(
				"ranges must be in ascending order: %s then %s",
				ranges.String(i-1), ranges.String(i)))
		}
		if !ranges.LessThan(ranges.GetEnd(i-1), ranges.GetStart(i)) {
			comp.error(node, errors.New("ranges must be disjoint"))
		}
	}
}

// Given a Node, get the module and Node that is being referenced
// The reference will be of the form [prefix:]name
// It is an implicit reference to the local module when the optional
// [prefix:] is absent
func (c *Compiler) getModuleAndReference(m, n parse.Node, targetType parse.NodeType) (parse.Node, parse.Node) {
	// Assume an implicit local module reference until
	// we learn otherwise.
	targetModule := m
	name := n.Argument().String()
	nameparts := strings.Split(name, ":")
	if len(nameparts) > 2 {
		// Can't have more than one ':'
		c.error(n, fmt.Errorf("Invalid %s name: %s", targetType.String(), name))
		return nil, nil
	}
	if len(nameparts) == 2 {
		// The feature reference includes an explicit module prefix
		var err error
		targetModule, err = n.GetModuleByPrefix(
			nameparts[0], c.modules, c.skipUnknown)
		if err != nil {
			c.error(n, err)
		}
		name = nameparts[1]
	}

	reference := targetModule.LookupChild(targetType, name)
	if reference == nil {
		if !c.skipUnknown {
			// Feature not found in specified module
			c.error(n, fmt.Errorf("%s not valid: %s", targetType.String(), n.Argument().String()))
			return nil, nil
		}
		var nc parse.NodeCardinality
		if c.extensions != nil {
			nc = c.extensions.NodeCardinality
		}
		reference = parse.NewFakeNodeByType(nc, targetType, name)
		targetModule.AddChildren(reference)
	}

	return targetModule, reference
}

// Verify a feature.
//
// Evaluate any features referenced by if-feature child nodes
// to determine if the feature is enabled or disabled
// Check for cyclic references - a reference back to ourselves via
// a chain of if-features.
func (c *Compiler) isFeatureValid(m parse.Node, n parse.Node, featTree map[string]bool) bool {
	var enabled bool

	// Build the <module-name>:<feature-name> for this feature
	featName := m.Name() + ":" + n.Name()

	enabled = c.featureEnabled(featName)

	// Check we have not already encountered this feature through an
	// if-feature chain of references
	if _, ok := featTree[featName]; ok {
		c.error(n, fmt.Errorf("Feature cyclic reference: %s", featName))
		return false
	}
	featTree[featName] = true

	// Verify each feature that this feature references via an if-feature
	for _, ifFeat := range n.ChildrenByType(parse.NodeIfFeature) {
		mod, feature := c.getModuleAndReference(m, ifFeat, parse.NodeFeature)
		c.assertReferenceStatus(n, feature, schema.Current)
		enabled = c.isFeatureValid(mod, feature, featTree) && enabled
	}

	// update the verified features
	c.verifiedFeatures.set(featName, enabled)
	return enabled
}

// Check a modules features, determining if they are enabled or not.
// Also catch any duplicate feature names within a module
// Filter out any features that do not appear in the yang
func (c *Compiler) checkFeatures() error {
	filteredFeatures := newFeaturesMap()
	for _, module := range c.modules {
		m := module.GetModule()
		dupChk := make(map[string]bool)
		for _, feat := range m.ChildrenByType(parse.NodeFeature) {
			if _, ok := dupChk[feat.Name()]; ok {
				// Already seen this feature
				c.error(feat, fmt.Errorf("Duplicate feature %s in module %s", feat.Name(), m.Name()))
			} else {
				dupChk[feat.Name()] = true
			}
			// verify feature and update its enabled/disabled
			// status
			filteredFeatures.set(m.Name()+":"+feat.Name(),
				c.isFeatureValid(m, feat, make(map[string]bool)))
		}
	}

	c.verifiedFeatures = filteredFeatures
	return nil
}

func (c *Compiler) getEnabledFeaturesForPrefix(name string) []string {
	var features []string

	prefix := name + ":"

	for featName, enabled := range c.verifiedFeatures.features {
		if enabled && strings.HasPrefix(featName, prefix) {
			features = append(features,
				strings.TrimPrefix(featName, prefix))
		}
	}
	return features
}

func (c *Compiler) identityCheckCyclicRef(name string, ids map[string]parse.Node, assigned map[string]bool) {
	if _, ok := assigned[name]; ok {
		c.error(ids[name], fmt.Errorf("Identity cyclic reference %s\n", name))
	}
	assigned[name] = true

	for _, nd := range ids[name].ChildrenByType(parse.NodeIdentity) {
		nm := nd.Root().Name() + ":" + nd.Name()
		c.identityCheckCyclicRef(nm, ids, assigned)
	}

}

func (c *Compiler) checkIdentities() error {
	ids := make(map[string]parse.Node)

	// Get all identities, check for duplicates
	for _, module := range c.modules {
		mod := module.GetModule()
		for _, ident := range mod.ChildrenByType(parse.NodeIdentity) {
			name := mod.Name() + ":" + ident.Name()
			if _, ok := ids[name]; ok {
				c.error(ident, fmt.Errorf("Duplicate identity %s in module %s", ident.Name(), mod.Name()))
			} else {
				ids[name] = ident
			}
		}
	}

	// Process derived identities, building
	// identity tree.
	for name, ident := range ids {
		for _, base := range ident.ChildrenByType(parse.NodeBase) {
			mod, tIdent := c.getModuleAndReference(ident.Root(), base, parse.NodeIdentity)
			tnm := mod.Name() + ":" + tIdent.Name()
			if _, ok := ids[tnm]; ok {
				tIdent.AddChildren(ident)
				c.assertReferenceStatus(ident, tIdent, schema.Current)
			} else {
				c.error(ident, fmt.Errorf("Can't find base identity %s for identity %s\n", base.Name(), name))
			}
		}
	}

	// Now we have an identity tree,
	// check there are no cyclic references
	for nme, _ := range ids {
		c.identityCheckCyclicRef(nme, ids, make(map[string]bool))
	}

	c.identities = ids
	return nil
}

func (c *Compiler) findMissingImportStatement(name string) parse.Node {
	for _, module := range c.modules {
		for _, ch := range module.GetModule().ChildrenByType(parse.NodeImport) {
			if ch.Name() == name {
				return ch
			}
		}
	}
	return nil
}

func (c *Compiler) ExpandModules() (err error) {

	defer c.recover(&err)
	//Attach submodules to modules
	for mn, subm := range c.submodules {
		belongs := subm.GetModule().ChildByType(parse.NodeBelongsTo).Name()
		mod, ok := c.modules[belongs]
		if !ok {
			c.error(subm.GetModule(),
				fmt.Errorf("submodule belongs to non-existent module %s", mn))
		}
		mod.GetSubmodules()[mn] = subm.GetModule()
	}

	//Process includes
	for _, module := range c.modules {
		r := module.GetModule()
		c.VerifyModuleIncludes(r, module.GetSubmodules())
		for _, s := range module.GetSubmodules() {
			c.ProcessSubmoduleIncludes(s, module.GetSubmodules())
		}
		c.ProcessModuleIncludes(r, module.GetSubmodules())
	}

	//Process imports
	g := tsort.New()
	for mn, module := range c.modules {
		r := module.GetModule()
		g.AddVertex(mn)
		for _, i := range r.ChildrenByType(parse.NodeImport) {
			g.AddEdge(mn, i.Name())
		}
	}
	c.modnames, err = g.Sort()
	if err != nil {
		panic(fmt.Errorf("import %s", err))
	}

	//Process features
	err = c.checkFeatures()
	if err != nil {
		panic(fmt.Errorf("feature %s", err))
	}

	err = c.checkIdentities()
	if err != nil {
		panic(fmt.Errorf("identity %s", err))
	}

	// Check for cycles in all groupings before applying
	for _, module := range c.modules {
		if err := c.validateModuleGroupings(module.GetModule()); err != nil {
			c.error(module.GetModule(), err)
		}
		for _, sm := range module.GetSubmodules() {
			if err := c.validateModuleGroupings(sm); err != nil {
				c.error(sm, err)
			}
		}
	}

	// Apply uses and augments
	for _, name := range c.modnames {
		module, ok := c.modules[name]
		if ok {
			c.expandModule(module)
		} else if !c.skipUnknown {
			i := c.findMissingImportStatement(name)
			c.error(i, fmt.Errorf("module not found"))
		}
	}

	// Apply deviations
	for _, name := range c.modnames {
		module, ok := c.modules[name]
		if ok {
			c.processDeviations(module)
		} else if !c.skipUnknown {
			i := c.findMissingImportStatement(name)
			c.error(i, fmt.Errorf("module not found"))
		}
	}

	return nil
}

func (c *Compiler) BuildModules() (modules map[string]schema.Model, err error) {

	defer c.recover(&err)

	modules = make(map[string]schema.Model)
	for _, name := range c.modnames {
		module, ok := c.modules[name]
		if ok {
			newModule := c.BuildModule(module, module.GetModule())
			modules[name] = newModule
		} else if !c.skipUnknown {
			panic(fmt.Errorf("required module %s was not found", name))
		}
	}
	return modules, nil
}

func (c *Compiler) VerifyModuleIncludes(m parse.Node, submodules map[string]parse.Node) {
	g := tsort.New()
	for _, i := range m.ChildrenByType(parse.NodeInclude) {
		g.AddEdge(m.Name(), i.Name())
	}
	for _, s := range submodules {
		for _, i := range s.ChildrenByType(parse.NodeInclude) {
			g.AddEdge(s.Name(), i.Name())
		}
	}
	_, err := g.Sort()
	if err != nil {
		c.error(m, err)
	}
}

func (c *Compiler) ProcessSubmoduleIncludes(m parse.Node, submodules map[string]parse.Node) {
	tenv := m.Tenv()
	genv := m.Genv()
	for _, i := range m.ChildrenByType(parse.NodeInclude) {
		smod, ok := submodules[i.Name()]
		if !ok {
			c.error(i, fmt.Errorf("unknown submodule %s", i.Name()))
		}
		for _, t := range smod.ChildrenByType(parse.NodeTypedef) {
			err := tenv.Put(t.Name(), t)
			if err != nil {
				c.error(t, err)
			}
		}
		for _, g := range smod.ChildrenByType(parse.NodeGrouping) {
			err := genv.Put(g.Name(), g)
			if err != nil {
				c.error(g, err)
			}
		}

		m.AddChildren(smod.ChildrenByType(parse.NodeImport)...)
	}
}

func (c *Compiler) ProcessModuleIncludes(m parse.Node, submodules map[string]parse.Node) {
	tenv := m.Tenv()
	genv := m.Genv()
	for _, i := range m.ChildrenByType(parse.NodeInclude) {
		smod, ok := submodules[i.Name()]
		if !ok {
			c.error(i, fmt.Errorf("unknown submodule %s", i.Name()))
		}
		for _, t := range smod.ChildrenByType(parse.NodeTypedef) {
			err := tenv.Put(t.Name(), t)
			if err != nil {
				c.error(t, err)
			}
		}
		for _, g := range smod.ChildrenByType(parse.NodeGrouping) {
			err := genv.Put(g.Name(), g)
			if err != nil {
				c.error(g, err)
			}
		}
		m.AddChildren(smod.ChildrenByType(parse.NodeImport)...)
		m.AddChildren(smod.ChildrenByType(parse.NodeDataDef)...)
		m.AddChildren(smod.ChildrenByType(parse.NodeAugment)...)
	}
}

type inheritedFeatures struct {
	config      bool
	status      schema.Status
	onEnter     string
	priv        bool
	repeatable  bool
	passOpcArgs bool
}

func (c *Compiler) buildSchemaTree(m parse.Node, n parse.Node) schema.Tree {
	if n == nil {
		tree, _ := schema.NewTree(nil)
		return c.extendTree(nil, tree)
	}

	body := n.ChildrenByType(parse.NodeDataDef)
	inherited := inheritedFeatures{config: true, status: schema.Current}

	children := c.buildChildren(inherited, m, body)
	tree, err := schema.NewTree(children)
	if err != nil {
		c.error(m, err)
	}
	return c.extendTree(n, tree)
}

func (c *Compiler) BuildModule(module *parse.Module, m parse.Node) schema.Model {
	c.CheckChildren(m, m)
	rpcs := make(map[string]schema.Rpc)
	for _, r := range m.ChildrenByType(parse.NodeRpc) {
		input := r.ChildByType(parse.NodeInput)
		inputTree := c.buildSchemaTree(m, input)

		output := r.ChildByType(parse.NodeOutput)
		outputTree := c.buildSchemaTree(m, output)

		rpc := schema.NewRpc(inputTree, outputTree)
		rpcs[r.Name()] = c.extendRpc(r, rpc)
	}

	notifications := make(map[string]schema.Notification)
	for _, n := range m.ChildrenByType(parse.NodeNotification) {
		notificationTree := c.buildSchemaTree(m, n)
		notification := schema.NewNotification(notificationTree)
		notifications[n.Name()] = c.extendNotification(n, notification)
	}

	inherited := inheritedFeatures{config: true, status: schema.Current}
	children := c.buildChildren(inherited, m, m.ChildrenByType(parse.NodeDataDef))

	modTree, err := schema.NewTree(children)
	if err != nil {
		c.error(m, err)
	}

	extTree := c.extendTree(m, modTree)
	modSchema := schema.NewModel(
		// TODO - better api avoiding intermediate GetModule()?
		module.GetModule().Name(),
		module.GetModule().Revision(),
		module.GetModule().Ns(),
		module.GetTree().String(),
		extTree,
		rpcs,
		c.getEnabledFeaturesForPrefix(module.GetModule().Name()),
		notifications,
		c.getDeviations(module.GetModule().Name()), // replace with deviations
	)

	return c.extendModel(m, modSchema, extTree)
}

func (c *Compiler) IgnoreNode(node parse.Node, parentStatus schema.Status) bool {
	if node.NotSupported() {
		return true
	}

	for _, ifn := range node.ChildrenByType(parse.NodeIfFeature) {
		if !c.CheckIfFeature(ifn, c.getStatus(node, parentStatus)) {
			return true
		}
	}
	return false
}

func parseStatus(statusStatement parse.Node) schema.Status {

	statusString := statusStatement.ArgStatus()
	switch statusString {
	case "current":
		return schema.Current
	case "deprecated":
		return schema.Deprecated
	case "obsolete":
		return schema.Obsolete
	}
	panic(fmt.Errorf("Unexpected value for status: %s", statusString))
}

func (c *Compiler) getStatus(node parse.Node, inheritedStatus schema.Status) schema.Status {

	if statusStatement := node.ChildByType(parse.NodeStatus); statusStatement != nil {
		status := parseStatus(statusStatement)
		if status < inheritedStatus {
			c.error(statusStatement, fmt.Errorf("Cannot override status of parent"))
		}
		return status
	}

	return inheritedStatus
}

func (c *Compiler) getConfig(node parse.Node, inheritedConfig bool) bool {

	if configStatement := node.ChildByType(parse.NodeConfig); configStatement != nil {
		config := configStatement.ArgBool()
		if inheritedConfig == false && config == true {
			c.error(configStatement, fmt.Errorf("config true node can't have a config false parent"))
		}
		return config
	}
	return inheritedConfig
}

func (c *Compiler) getOnEnter(node parse.Node, inheritedOnEnter string) string {
	if onEnterStatement := node.ChildByType(parse.NodeOpdOnEnter); onEnterStatement != nil {
		return onEnterStatement.ArgString()
	}
	return inheritedOnEnter
}

func (c *Compiler) getPassOpcArgs(node parse.Node, inheritedPassOpcArgs bool) bool {
	if passOpcArgsStmt := node.ChildByType(parse.NodeOpdPassOpcArgs); passOpcArgsStmt != nil {
		return passOpcArgsStmt.ArgBool()
	}
	return inheritedPassOpcArgs
}

func (c *Compiler) getPriv(node parse.Node, inheritedPriv bool) bool {
	if privStatement := node.ChildByType(parse.NodeOpdPrivileged); privStatement != nil {
		return privStatement.ArgBool()
	}
	return inheritedPriv
}

func (c *Compiler) getRepeatable(node parse.Node, inheritedRepeatable bool) bool {
	if repeatableStatement := node.ChildByType(parse.NodeOpdRepeatable); repeatableStatement != nil {

		return repeatableStatement.ArgBool()
	}
	return inheritedRepeatable
}
func (c *Compiler) overrideInherited(
	inherited inheritedFeatures, dataDef parse.Node,
) inheritedFeatures {

	// Inherit from parent by default
	features := inherited

	features.status = c.getStatus(dataDef, inherited.status)
	features.config = c.getConfig(dataDef, inherited.config)
	if inhNd := dataDef.ChildByType(parse.NodeOpdInherit); inhNd != nil {
		features.onEnter = c.getOnEnter(inhNd, inherited.onEnter)
		features.priv = c.getPriv(inhNd, inherited.priv)
		features.passOpcArgs = c.getPassOpcArgs(inhNd, inherited.passOpcArgs)
	}
	features.repeatable = c.getRepeatable(dataDef, inherited.repeatable)
	return features
}

func (c *Compiler) buildChildren(inherited inheritedFeatures, m parse.Node, body []parse.Node) []schema.Node {

	var children []schema.Node

	for _, dataDef := range body {
		if c.IgnoreNode(dataDef, inherited.status) {
			continue
		}
		ch := c.BuildNode(inherited, m, dataDef, false)
		for _, sn := range ch {
			if c.filter != nil && !c.filter(sn) {
				continue
			}
			children = append(children, sn)
		}
	}

	return children
}

type key struct {
	schema.Leaf
}

func (*key) Mandatory() bool         { return false }
func (*key) Default() (string, bool) { return "", false }

func (c *Compiler) buildListChildren(
	keys []string,
	inherited inheritedFeatures,
	m parse.Node,
	body []parse.Node,
) []schema.Node {

	var children []schema.Node

	for _, dataDef := range body {
		if c.IgnoreNode(dataDef, inherited.status) {
			continue
		}
		var isKey bool
		for _, v := range keys {
			if dataDef.Name() == v {
				isKey = true
				if dataDef.Type() != parse.NodeLeaf {
					c.error(dataDef, fmt.Errorf("List key must be a leaf"))
				}
			}
		}
		ch := c.BuildNode(inherited, m, dataDef, isKey)
		for _, sn := range ch {
			if c.filter != nil && !c.filter(sn) {
				continue
			}
			children = append(children, sn)
		}
	}

	return children
}

func (c *Compiler) BuildNode(
	inherited inheritedFeatures,
	m parse.Node,
	n parse.Node,
	isKey bool,
) (retNodes []schema.Node) {

	features := c.overrideInherited(inherited, n)

	switch n.Type() {
	case parse.NodeContainer:
		retNodes = []schema.Node{c.BuildContainer(features, m, n)}
	case parse.NodeList:
		retNodes = []schema.Node{c.BuildList(features, m, n)}
	case parse.NodeLeafList:
		retNodes = []schema.Node{c.BuildLeafList(features, m, n)}
	case parse.NodeLeaf:
		retNodes = []schema.Node{c.BuildLeaf(features, m, n, isKey)}
	case parse.NodeChoice:
		retNodes = []schema.Node{c.BuildChoice(features, m, n)}
	case parse.NodeCase:
		retNodes = []schema.Node{c.BuildCase(features, m, n)}
	case parse.NodeOpdCommand:
		retNodes = []schema.Node{c.BuildOpdCommand(features, m, n)}
	case parse.NodeOpdOption:
		retNodes = []schema.Node{c.BuildOpdOption(features, m, n)}
	case parse.NodeOpdArgument:
		retNodes = []schema.Node{c.BuildOpdArgument(features, m, n)}
	default:
		retNodes = nil
	}

	return retNodes

}

func (c *Compiler) CheckChildren(m parse.Node, n parse.Node) {
	for _, ch := range n.Children() {
		switch ch.Type() {
		case parse.NodeList:
			c.CheckUniqueConstraint(m, ch)
		case parse.NodeUnknown:
			c.CheckUnknown(m, ch)
		}
		c.CheckChildren(m, ch)
	}
}

func xmlPathString(path []xml.Name) string {
	var buf = new(bytes.Buffer)
	var getxmlname = func(name xml.Name) string {
		if name.Space != "" {
			return fmt.Sprintf("%s:%s", name.Space, name.Local)
		} else {
			return name.Local
		}
	}
	if len(path) == 0 {
		return ""
	}
	fmt.Fprintf(buf, getxmlname(path[0]))
	for _, elem := range path[1:] {
		fmt.Fprintf(buf, "/%s", getxmlname(elem))
	}
	return buf.String()
}

// unique-arg cannot traverse a descedant list
// (see https://www.ietf.org/mail-archive/web/netmod/current/msg06386.html)
func (c *Compiler) CheckUniqueConstraint(m parse.Node, n parse.Node) {
	for _, uniq := range allUniques(n) {
		for _, path := range uniq {
			var child parse.Node = n
			for i, elem := range path {
				child = child.LookupChild(parse.NodeDataDef, elem.Local)
				if child == nil {
					c.error(n, fmt.Errorf("unknown descendant %s referenced in unique", xmlPathString(path)))
				}
				if child.Type() == parse.NodeList {
					c.error(n, fmt.Errorf("list descendant %s referenced in unique", xmlPathString(path)))
				}
				if i == len(path)-1 {
					if child.Type() != parse.NodeLeaf {
						c.error(n, fmt.Errorf("non leaf descendant %s referenced in unique", xmlPathString(path)))
					}
					// Check for empty leaf type in BuildList()
				}
			}
		}
	}
}

func (c *Compiler) CheckUnknown(m parse.Node, n parse.Node) {
	//unknown nodes represent unregisted extensions, if the name is not registerd
	//and it is not defined by one of the imported yang files then bail out here
	name := n.Statement()
	nameparts := strings.Split(name, ":")
	if len(nameparts) > 2 {
		c.error(n, fmt.Errorf("invalid extension name %s", name))
	}
	if len(nameparts) == 2 {
		space, local := nameparts[0], nameparts[1]
		m, err := n.GetModuleByPrefix(space, c.modules, c.skipUnknown)
		if err != nil {
			c.error(n, err)
		}
		ext := m.LookupChild(parse.NodeExtension, local)
		if ext == nil {
			if c.skipUnknown {
				return
			}
			c.error(n, fmt.Errorf("unknown extension %s:%s", space, local))
		}
	} else {
		local := nameparts[0]
		ext := m.LookupChild(parse.NodeExtension, local)
		if ext == nil {
			c.error(n, fmt.Errorf("unknown extension %s", local))
		}
	}
}

// Verify that an if-feature reference is valid, and determine if the
// referenced feature is enabled
// A referenced feature takes the form [prefix:]feature-name
// If no prefix is present, it is an implicit reference to the
// local module.
func (c *Compiler) CheckIfFeature(n parse.Node, parentStatus schema.Status) bool {

	mod, feature := c.getModuleAndReference(n.Root(), n, parse.NodeFeature)

	c.assertReferenceStatus(n, feature, parentStatus)

	return c.verifiedFeatureEnabled(mod.Name() + ":" + feature.Name())
}

// Takes a parse.Node ErrorContext for a must / when node and extracts
// the file and line number.  Initial string is of the following format:
// '/tmp/tmpvy9n_g8i/yang/vyatta-protocols-ospfv3-v1.yang:1464:5:
//    must not(../../../if-loopback:tagnode) or (current() = 'point-to-point')'
func extractFileAndLineFromErrorContext(mustOrWhen parse.Node) string {
	fullLocStr, _ := mustOrWhen.ErrorContext()
	filePlusLine := strings.Join(strings.Split(fullLocStr, ":")[:2], ":")
	filePlusLineSlice := strings.Split(filePlusLine, "/")
	return filePlusLineSlice[len(filePlusLineSlice)-1]
}

func (c *Compiler) BuildWhens(n parse.Node) []schema.WhenContext {

	var whens []schema.WhenContext
	var pathEvalMachine *xpath.Machine
	var errPE error

	for _, when := range n.ChildrenByType(parse.NodeWhen) {
		mapFn := func(prefix string) (module string, err error) {
			return when.YangPrefixToNamespace(prefix, c.modules, c.skipUnknown)
		}

		whenMachine, errW := expr.NewExprMachine(when.ArgWhen(), mapFn)
		if errW != nil {
			c.error(n, errW)
		}
		errMsg := fmt.Sprintf("'when' condition is false: '%s'", when.ArgWhen())

		if c.generateWarnings {
			pathEvalMachine, errPE = path_eval.NewPathEvalMachine(
				when.ArgWhen(), mapFn, extractFileAndLineFromErrorContext(when))
			if errPE != nil {
				c.error(n, errPE)
			}
		}
		whenNs, err := when.YangPrefixToNamespace("", c.modules, c.skipUnknown)
		if err != nil {
			c.error(when, err)
		}
		whens = append(whens, schema.NewWhenContext(
			whenMachine, pathEvalMachine, errMsg, when.AddedByAugment(),
			whenNs))
	}

	return whens
}

func (c *Compiler) BuildMusts(n parse.Node) []schema.MustContext {

	var musts []schema.MustContext
	var basePathEvalMachine, extPathEvalMachine *xpath.Machine
	var errM error

	for _, must := range n.ChildrenByType(parse.NodeMust) {
		mapFn := func(prefix string) (module string, err error) {
			return must.YangPrefixToNamespace(prefix, c.modules, c.skipUnknown)
		}

		// Create machine for must statement.  Try extended must statement
		// first, if present.  Fall back silently to standard must on error.
		mustExpr := must.ArgMust()
		baseMustExpr := mustExpr
		extMustExpr := c.extendMust(n, must)
		var mustMachine *xpath.Machine

		if extMustExpr != "" {
			mustMachine, errM =
				expr.NewExprMachineWithCustomFunctions(extMustExpr, mapFn)
			if errM == nil {
				mustExpr = extMustExpr
			}
		}
		if mustMachine == nil {
			mustMachine, errM = expr.NewExprMachine(baseMustExpr, mapFn)
			if errM != nil {
				c.error(n, errM)
			}
			mustExpr = baseMustExpr
		}

		errMsg := must.Msg()
		if errMsg == "" {
			errMsg = fmt.Sprintf("'must' condition is false: '%s'", mustExpr)
		}

		// If not set, default is added when error occurs.
		appTag := must.AppTag()

		// We may have been asked to do extra validation.  This consists of
		// checking both must and configd:must compile (we only needed one to
		// work to pass above checks), and creating path evaluation machines
		// to run once full schema is compiled to ensure there are no
		// non-existent paths in our XPATH statements.  The latter is allowed,
		// but we try to prevent it in our own YANG files by doing this check.
		if c.generateWarnings {
			basePathEvalMachine = c.createPathEvalMachine(
				n, baseMustExpr, mapFn, must, false)
			extPathEvalMachine = c.createPathEvalMachine(
				n, extMustExpr, mapFn, must, true)
		}

		mustNs, err := must.YangPrefixToNamespace("", c.modules, c.skipUnknown)
		if err != nil {
			c.error(must, err)
		}
		musts = append(musts, schema.NewMustContext(
			mustMachine, basePathEvalMachine, extPathEvalMachine,
			errMsg, appTag, mustNs))
	}

	return musts
}

func (c *Compiler) createPathEvalMachine(
	n parse.Node,
	mustExpr string,
	mapFn func(prefix string) (module string, err error),
	must parse.Node,
	allowCustomFns bool,
) *xpath.Machine {

	if mustExpr == "" {
		return nil
	}

	var errPE error
	var warnType xutils.WarnType
	var pathEvalMachine *xpath.Machine

	if allowCustomFns {
		// On the vRouter, custom functions are loaded from plugins, and
		// are visible to standard configd:must compilation.  For tools such
		// as yangc (used by build-iso and DRAM), plugins may (build-iso) or
		// may not (DRAM) be installed.  In the latter case we can inject
		// the expected function names, but to be on the safe side, they are
		// only used for the path evaluation machine that does not actually
		// run the custom functions.  We will still detect any relevant
		// compilation errors but there's no chance we will try to run the
		// non-existent function implementations.
		pathEvalMachine, errPE = path_eval.NewPathEvalMachineWithCustomFns(
			mustExpr, mapFn, extractFileAndLineFromErrorContext(must),
			c.userFnChecker)
		warnType = xutils.ConfigdMustCompilerError
	} else {
		pathEvalMachine, errPE = path_eval.NewPathEvalMachine(
			mustExpr, mapFn, extractFileAndLineFromErrorContext(must))
		warnType = xutils.CompilerError
	}

	if errPE != nil {
		c.saveWarning(n, xutils.NewWarning(
			warnType,
			n.Name(), // Full path tricky to get w/o writing test-specific code
			mustExpr,
			extractFileAndLineFromErrorContext(must),
			"(n/a)", // testPath
			errPE.Error()))
	}

	return pathEvalMachine
}

func (c *Compiler) filterDisabledExtensions(n parse.Node) {

	rn := make([]parse.Node, 0)
	for _, en := range n.Children() {
		if !en.Type().IsExtensionNode() {
			continue
		}
		if c.IgnoreNode(en, schema.Current) {
			rn = append(rn, en)
		}
	}

	for _, r := range rn {
		n.ReplaceChild(r)
	}
}

func (c *Compiler) BuildContainer(features inheritedFeatures, m parse.Node, n parse.Node) schema.Node {
	con, err := schema.NewContainer(
		n.Name(),
		n.GetNodeNamespace(m, c.modules),
		n.GetNodeModulename(m),
		n.GetNodeSubmoduleName(),
		n.Desc(),
		n.Ref(),
		n.Presence(),
		features.config,
		features.status,
		c.BuildWhens(n),
		c.BuildMusts(n),
		c.buildChildren(features, m, n.ChildrenByType(parse.NodeDataDef)),
	)

	if err != nil {
		c.error(n, err)
	}

	c.filterDisabledExtensions(n)

	return c.extendContainer(n, con)
}

func allUniques(n parse.Node) [][][]xml.Name {
	uniqs := make([][][]xml.Name, 0, 0)
	for _, ch := range n.ChildrenByType(parse.NodeUnique) {
		uniqs = append(uniqs, ch.ArgUnique())
	}
	return uniqs
}

func (c *Compiler) BuildList(features inheritedFeatures, m parse.Node, n parse.Node) schema.Node {
	c.CheckMinMax(n, n.Min(), n.Max())

	children := c.buildListChildren(n.Keys(), features, m, n.ChildrenByType(parse.NodeDataDef))

	l, err := schema.NewList(
		n.Name(),
		n.GetNodeNamespace(m, c.modules),
		n.GetNodeModulename(m),
		n.GetNodeSubmoduleName(),
		n.Desc(),
		n.Ref(),
		n.OrdBy(),
		n.Min(),
		n.Max(),
		features.config,
		features.status,
		n.Keys(),
		allUniques(n),
		c.BuildWhens(n),
		c.BuildMusts(n),
		children,
	)

	if err != nil {
		c.error(n, err)
	}

	// Now that we have type information, verify unique-args don't
	// reference an empty leaf node. We have already done some
	// checking in CheckUniqueConstraint()
	//
	// Uniques is [][][]xml.Name : set of unique statement(s)
	//
	for _, uniq := range l.Uniques() {
		for _, path := range uniq {
			// If we wanted to check that ALL key(s) are not present in
			// the unique statement, then this is where to do it.  For
			// now, as this is only a pyang warning, we will not enforce
			// it as this might preclude us from using external YANG files.

			// Check we aren't referencing an empty node.
			var child schema.Node = l
			for i, elem := range path {
				for _, ch := range child.Children() {
					// Skip children until we find a match ...
					// We could optimise by exiting if we find no match at
					// one level.
					if ch.Name() != elem.Local {
						continue
					}

					// Matched path element, so change child to point to
					// one level deeper in the tree.  Next element in path
					// will now be compared to children at this level.
					child = ch

					// If we are at the end of the path, check that the
					// child is not an empty node as this is invalid.
					if i == len(path)-1 {
						if _, ok := child.Type().(schema.Empty); ok {
							c.error(n, fmt.Errorf(
								"empty leaf descendant %s referenced "+
									"in unique statement",
								xmlPathString(path)))
						}
					}
				}
			}
		}
	}

	c.filterDisabledExtensions(n)
	return c.extendList(n, l)
}

func (c *Compiler) BuildLeafList(features inheritedFeatures, m parse.Node, n parse.Node) schema.Node {
	c.CheckMinMax(n, n.Min(), n.Max())
	typ := c.BuildType(n, n.ChildByType(parse.NodeTyp), emptyDefault, false, features.status)

	l := schema.NewLeafList(
		n.Name(),
		n.GetNodeNamespace(m, c.modules),
		n.GetNodeModulename(m),
		n.GetNodeSubmoduleName(),
		n.Desc(),
		n.Ref(),
		n.Def(),
		n.OrdBy(),
		n.Units(),
		n.Min(),
		n.Max(),
		typ,
		features.config,
		features.status,
		c.BuildWhens(n),
		c.BuildMusts(n),
	)

	c.filterDisabledExtensions(n)
	return c.extendLeafList(n, l)
}

func (c *Compiler) BuildOpdCommand(features inheritedFeatures, m parse.Node, n parse.Node) schema.Node {
	com, err := schema.NewOpdCommand(
		n.Name(),
		n.GetNodeNamespace(m, c.modules),
		n.GetNodeModulename(m),
		n.Desc(),
		n.Ref(),
		c.getOnEnter(n, features.onEnter),
		c.getPriv(n, features.priv),
		n.Local(),
		n.Secret(),
		c.getRepeatable(n, features.repeatable),
		c.getPassOpcArgs(n, features.passOpcArgs),
		features.status,
		c.buildChildren(features, m, n.ChildrenByType(parse.NodeOpdDef)),
	)

	if err != nil {
		c.error(n, err)
	}

	c.filterDisabledExtensions(n)
	return c.extendOpdCommand(n, com)
}

func (c *Compiler) BuildOpdOption(
	features inheritedFeatures,
	mod parse.Node,
	node parse.Node,
) schema.Node {

	mandatory := node.Mandatory()
	hasDef := node.HasDef()
	defVal := node.Def()

	var typ schema.Type
	tch := node.ChildByType(parse.NodeTyp)
	if tch != nil {
		typ = c.BuildType(
			node,
			tch,
			defVal,
			hasDef,
			features.status)
	} else {
		emptyName := xml.Name{Space: "builtin", Local: ""}
		typ = schema.NewEmpty(emptyName, "", false)
	}
	option, err := schema.NewOpdOption(
		node.Name(),
		node.GetNodeNamespace(mod, c.modules),
		node.GetNodeModulename(mod),
		node.Desc(),
		node.Ref(),
		node.Units(),
		c.getOnEnter(node, features.onEnter),
		c.getPriv(node, features.priv),
		node.Local(),
		node.Secret(),
		c.getRepeatable(node, features.repeatable),
		mandatory,
		c.getPassOpcArgs(node, features.passOpcArgs),
		typ,
		features.status,
		c.buildChildren(features, mod, node.ChildrenByType(parse.NodeOpdDef)),
	)
	if err != nil {
		c.error(node, err)
	}

	c.filterDisabledExtensions(node)
	return c.extendOpdOption(node, option)
}

func (c *Compiler) BuildOpdArgument(
	features inheritedFeatures,
	mod parse.Node,
	node parse.Node,
) schema.Node {

	hasDef := node.HasDef()
	defVal := node.Def()

	typ := c.BuildType(
		node,
		node.ChildByType(parse.NodeTyp),
		defVal,
		hasDef,
		features.status)

	option, err := schema.NewOpdArgument(
		node.Name(),
		node.GetNodeNamespace(mod, c.modules),
		node.GetNodeModulename(mod),
		node.Desc(),
		node.Ref(),
		node.Units(),
		c.getOnEnter(node, features.onEnter),
		c.getPriv(node, features.priv),
		node.Local(),
		node.Secret(),
		c.getRepeatable(node, features.repeatable),
		node.Mandatory(),
		c.getPassOpcArgs(node, features.passOpcArgs),
		typ,
		features.status,
		c.buildChildren(features, mod, node.ChildrenByType(parse.NodeOpdDef)),
	)
	if err != nil {
		c.error(node, err)
	}

	c.filterDisabledExtensions(node)
	return c.extendOpdArgument(node, option)
}

func (comp *Compiler) BuildLeaf(
	features inheritedFeatures,
	mod parse.Node,
	node parse.Node,
	isKey bool,
) schema.Node {

	mandatory := node.Mandatory()
	hasDef := node.HasDef()
	defVal := node.Def()

	typ := comp.BuildType(
		node,
		node.ChildByType(parse.NodeTyp),
		defVal,
		hasDef,
		features.status)

	if mandatory {
		if hasDef {
			comp.error(node, errors.New("Leaf cannot have default and be mandatory."))
		}
	}

	// Need to note if EXPLICIT default was set on leaf so when we check
	// mutual exclusivity of default vs mandatory, we can say 'ah, wasn't
	// set on LEAF, so can unset.'.
	leaf := schema.NewLeaf(
		node.Name(),
		node.GetNodeNamespace(mod, comp.modules),
		node.GetNodeModulename(mod),
		node.GetNodeSubmoduleName(),
		node.Desc(),
		node.Ref(),
		node.Units(),
		mandatory,
		typ,
		features.config,
		features.status,
		comp.BuildWhens(node),
		comp.BuildMusts(node),
	)

	if isKey {
		leaf = &key{leaf}
	}
	comp.filterDisabledExtensions(node)
	return comp.extendLeaf(node, leaf)
}

func (c *Compiler) checkChoiceDefaultCaseExists(sn schema.Node) error {
	choice, ok := sn.(schema.Choice)
	if !ok {
		return nil
	}
	if choice.DefaultCase() == "" ||
		c.filter != nil && !c.filter(choice) {
		return nil
	}
	for _, child := range choice.Choices() {

		if child.Name() == choice.DefaultCase() {
			return nil
		}
	}

	return fmt.Errorf("Choice default %s not found.", choice.DefaultCase())
}

func (c *Compiler) BuildChoice(features inheritedFeatures, m parse.Node, n parse.Node) schema.Node {
	children := c.buildChildren(features, m, n.ChildrenByType(parse.NodeDataDef))
	if n.HasDef() && n.Mandatory() {
		c.error(n, errors.New("Choice cannot have default and be mandatory."))
	}

	choice, err := schema.NewChoice(
		n.Name(),
		n.GetNodeNamespace(m, c.modules),
		n.GetNodeModulename(m),
		n.GetNodeSubmoduleName(),
		n.Def(),
		n.Desc(),
		n.Ref(),
		n.Mandatory(),
		features.config,
		features.status,
		c.BuildWhens(n),
		children,
	)

	if err != nil {
		c.error(n, err)
	}
	if err := c.checkChoiceDefaultCaseExists(choice); err != nil {
		c.error(n, err)
	}

	c.filterDisabledExtensions(n)
	nd := c.extendChoice(n, choice)
	return nd
}

func (c *Compiler) BuildCase(features inheritedFeatures, m parse.Node, n parse.Node) schema.Node {
	children := c.buildChildren(features, m, n.ChildrenByType(parse.NodeDataDef))

	// TODO: PAC: Namespace should be same as choice
	ycase, err := schema.NewCase(
		n.Name(),
		n.GetNodeNamespace(m, c.modules),
		n.GetNodeModulename(m),
		n.GetNodeSubmoduleName(),
		n.Desc(),
		n.Ref(),
		features.config,
		features.status,
		c.BuildWhens(n),
		children,
	)

	if err != nil {
		c.error(n, err)
	}

	c.filterDisabledExtensions(n)
	return c.extendCase(n, ycase)
}

func (c *Compiler) makeBoolean(
	tname xml.Name,
	node parse.Node,
	base schema.Boolean,
	def string,
	hasDef bool,
) schema.Type {

	c.validateRestrictions(node, base, SchemaBool)

	if base == nil {
		base = schema.NewBoolean(tname, "", false)
	}

	def, hasDef = c.getDefault(base, def, hasDef)

	return schema.NewBoolean(tname, def, hasDef)
}

func (comp *Compiler) makeDecimal64(
	name xml.Name,
	node parse.Node,
	base schema.Decimal64,
	def string,
	hasDef bool,
) schema.Type {

	comp.validateRestrictions(node, base, SchemaDecimal64)

	if base == nil {
		// Get the initial Rbs and Fd
		fd := schema.Fracdigit(node.FracDigit())
		if fd == 0 {
			comp.error(node, errors.New("missing fraction-digits"))
		}
		base = schema.NewDecimal64(name, fd, nil, "", "", "", false)
	}

	fd := base.Fd()
	rbs, msg, appTag := comp.getRangeBoundary(base, node)
	def, hasDef = comp.getDefault(base, def, hasDef)

	return schema.NewDecimal64(name, fd, rbs.(schema.DrbSlice),
		msg, appTag, def, hasDef)
}

func (c *Compiler) makeEmpty(
	tname xml.Name,
	node parse.Node,
	base schema.Empty,
	def string,
	hasDef bool,
) schema.Type {

	c.validateRestrictions(node, base, SchemaEmpty)

	if base == nil {
		base = schema.NewEmpty(tname, "", false)
	}

	def, hasDef = c.getDefault(base, def, hasDef)

	return schema.NewEmpty(tname, def, hasDef)
}

func (c *Compiler) getEnums(base schema.Enumeration, node parse.Node) []*schema.Enum {

	num_enums := node.ChildrenByType(parse.NodeEnum)
	if base != nil {
		if len(num_enums) > 0 {
			c.error(node, errors.New("cannot restrict predefined enumeration"))
		}
		return base.Enums()
	}

	if len(num_enums) == 0 {
		c.error(node, errors.New("enumeration requires at least one enum"))
	}

	enums := make([]*schema.Enum, 0, len(num_enums))
	for _, en := range node.ChildrenByType(parse.NodeEnum) {
		enum := schema.NewEnum(en.ArgString(), en.Desc(), en.Ref(),
			c.getStatus(en, schema.Current), en.Value())
		enums = append(enums, enum)
	}

	return enums
}

func (comp *Compiler) makeEnumeration(
	name xml.Name,
	node parse.Node,
	base schema.Enumeration,
	def string,
	hasDef bool,
) schema.Type {

	comp.validateRestrictions(node, base, SchemaEnumeration)

	enums := comp.getEnums(base, node)
	def, hasDef = comp.getDefault(base, def, hasDef)

	return schema.NewEnumeration(name, enums, def, hasDef)
}

func (c *Compiler) identityValues(cfgNode, node parse.Node, ident parse.Node, rt []*schema.Identity) []*schema.Identity {
	strp := cfgNode.GetNodeModulename(cfgNode.Root()) + ":"

	for _, id := range ident.ChildrenByType(parse.NodeIdentity) {
		nm := id.Root().Name() + ":" + id.Name()
		rname := strings.TrimPrefix(nm, strp)
		i := schema.NewIdentity(id.GetNodeModulename(id.Root()),
			id.GetNodeNamespace(id.Root(), c.modules),
			rname, id.Desc(), id.Ref(),
			c.getStatus(id, schema.Current), id.Name())
		rt = append(rt, i)
		n, _ := c.identities[nm]
		rt = c.identityValues(cfgNode, node, n, rt)
	}
	return rt
}

func (c *Compiler) getIdentities(cfgNode parse.Node, i schema.Identityref, node parse.Node, parentStatus schema.Status) []*schema.Identity {

	baseStmnt := node.ChildByType(parse.NodeBase)
	if i != nil {
		if baseStmnt != nil {
			c.error(node, errors.New("cannot restrict predefined identityref"))
		}
		return i.Identities()
	}

	if baseStmnt == nil {
		c.error(node, errors.New("cannot use identityref without a base"))
	}

	mod := node.Root()
	tm, ident := c.getModuleAndReference(mod, baseStmnt, parse.NodeIdentity)

	idid, _ := c.identities[tm.Name()+":"+ident.Name()]

	idents := make([]*schema.Identity, 0, 0)

	c.assertReferenceStatus(node, idid, parentStatus)
	node.AddChildren(ident)
	ids := c.identityValues(cfgNode, node, idid, idents)
	return ids
}

func (c *Compiler) makeIdentityRef(
	name xml.Name,
	cfgNode parse.Node,
	node parse.Node,
	base schema.Identityref,
	parentStatus schema.Status,
	def string,
	hasDef bool,
) schema.Type {

	c.validateRestrictions(node, base, SchemaIdentity)

	idents := c.getIdentities(cfgNode, base, node, parentStatus)
	def, hasDef = c.getDefault(base, def, hasDef)

	return schema.NewIdentityref(name, idents, def, hasDef)
}

func (c *Compiler) getRequire(base schema.InstanceId, node parse.Node) bool {

	if req_node := node.ChildByType(parse.NodeRequireInstance); req_node != nil {
		return req_node.ArgBool()
	}
	if base != nil {
		return base.Require()
	}
	return true
}

func (c *Compiler) makeInstanceId(
	name xml.Name,
	node parse.Node,
	base schema.InstanceId,
	def string,
	hasDef bool,
) schema.Type {

	c.validateRestrictions(node, base, SchemaInstanceId)

	require := c.getRequire(base, node)
	def, hasDef = c.getDefault(base, def, hasDef)

	return schema.NewInstanceId(name, require, def, hasDef)
}

func (c *Compiler) getPath(base schema.Leafref, node parse.Node) (mach *xpath.Machine) {

	path := node.Path()
	if base != nil {
		if path != "" {
			c.error(node, errors.New("cannot refine path"))
		}
		return base.Mach()
	}
	if path == "" {
		c.error(node, errors.New("missing path"))
	}

	mapFn := func(prefix string) (module string, err error) {
		return node.YangPrefixToNamespace(prefix, c.modules, c.skipUnknown)
	}
	mach, err := leafref.NewLeafrefMachine(path, mapFn)
	if err != nil {
		c.error(node, err)
	}
	return mach
}

func (c *Compiler) makeLeafref(
	node parse.Node,
	name xml.Name,
	base schema.Leafref,
	def string,
	hasDef bool,
) schema.Type {

	c.validateRestrictions(node, base, SchemaLeafRef)

	mach := c.getPath(base, node)
	def, hasDef = c.getDefault(base, def, hasDef)

	return schema.NewLeafref(name, mach, def, hasDef)
}

func (comp *Compiler) getBitSize(base schema.Number, node parse.Node, name xml.Name) schema.BitWidth {
	if base != nil {
		return base.BitWidth()
	}
	switch name.Local {
	case "int8", "uint8":
		return schema.BitWidth8
	case "int16", "uint16":
		return schema.BitWidth16
	case "int32", "uint32":
		return schema.BitWidth32
	case "int64", "uint64":
		return schema.BitWidth64
	default:
		comp.error(node, fmt.Errorf("Unrecognised integer type %s", name.Local))
		return 0
	}
}

func (comp *Compiler) getRangeBoundary(
	base schema.Number, node parse.Node,
) (rbs schema.RangeBoundarySlicer, msg, appTag string) {

	if base != nil {
		rbs, msg, appTag = base.Ranges(), base.Msg(), base.AppTag()
	}

	rng := node.ChildByType(parse.NodeRange)
	if rng == nil {
		return rbs, msg, appTag
	}

	rbs = comp.createRangeBdry(node, rbs, rng.ArgRange())
	return rbs, rng.Msg(), rng.AppTag()
}

func (comp *Compiler) makeInteger(
	name xml.Name,
	node parse.Node,
	base schema.Integer,
	def string,
	hasDef bool,
) schema.Type {

	comp.validateRestrictions(node, base, SchemaNumber)

	// Override or combine with local settings
	bits := comp.getBitSize(base, node, name)
	if base == nil {
		// Get the initial Rbs
		base = schema.NewInteger(bits, name, nil, "", "", "", false)
	}

	rbs, msg, appTag := comp.getRangeBoundary(base, node)
	def, hasDef = comp.getDefault(base, def, hasDef)

	return schema.NewInteger(bits, name, rbs.(schema.RbSlice),
		msg, appTag, def, hasDef)
}

func (comp *Compiler) makeUinteger(
	name xml.Name,
	node parse.Node,
	base schema.Number,
	def string,
	hasDef bool,
) schema.Type {

	comp.validateRestrictions(node, base, SchemaNumber)

	// Override or combine with local settings
	bitSize := comp.getBitSize(base, node, name)
	if base == nil {
		base = schema.NewUinteger(bitSize, name, nil, "", "", "", false)
	}

	rbs, msg, appTag := comp.getRangeBoundary(base, node)
	def, hasDef = comp.getDefault(base, def, hasDef)

	return schema.NewUinteger(bitSize, name, rbs.(schema.UrbSlice),
		msg, appTag, def, hasDef)
}

func (c *Compiler) getTypes(base schema.Union, cfgNode, node parse.Node, parentStatus schema.Status) []schema.Type {
	if base != nil {
		if len(node.ChildrenByType(parse.NodeTyp)) > 0 {
			c.error(node, errors.New("cannot restrict predefined union"))
		}
		return base.Typs()
	}

	num_types := len(node.ChildrenByType(parse.NodeTyp))

	if num_types == 0 {
		c.error(node, errors.New("union requires at least one type"))
	}

	types := make([]schema.Type, 0, num_types)
	for _, t := range node.ChildrenByType(parse.NodeTyp) {
		typ := c.BuildType(cfgNode, t, emptyDefault, false, parentStatus)
		types = append(types, typ)
	}

	return types
}

func (c *Compiler) makeUnion(
	name xml.Name,
	cfgNode parse.Node,
	node parse.Node,
	base schema.Union,
	parentStatus schema.Status,
	def string,
	hasDef bool,
) schema.Type {

	c.validateRestrictions(node, base, SchemaUnion)

	typs := c.getTypes(base, cfgNode, node, parentStatus)
	def, hasDef = c.getDefault(base, def, hasDef)

	return schema.NewUnion(name, typs, def, hasDef)
}

func (c *Compiler) getPatterns(base schema.String, n parse.Node) [][]schema.Pattern {

	pats := base.Pats()

	patterns := n.ChildrenByType(parse.NodePattern)
	ps := make([]schema.Pattern, 0, len(patterns))
	for _, p := range patterns {
		ps = append(ps, schema.Pattern{
			Pattern: p.Argument().String(),
			Regexp:  p.ArgPattern(),
			Msg:     p.Msg(),
			AppTag:  p.AppTag()})
	}

	return append(pats, ps)
}

func (c *Compiler) getPatternHelps(base schema.String, n parse.Node) [][]string {

	pathelps := base.PatHelps()

	patternhelps := n.ChildrenByType(parse.NodeConfigdPHelp)
	patternhelps = append(patternhelps,
		n.ChildrenByType(parse.NodeOpdPatternHelp)...)

	phs := make([]string, 0, len(patternhelps))
	for _, ph := range patternhelps {
		phs = append(phs, ph.ArgString())
	}

	return append(pathelps, phs)
}

func (c *Compiler) getLength(base schema.String, n parse.Node) *schema.Length {

	baseLen := base.Len()

	length := n.ChildByType(parse.NodeLength)
	if length == nil {
		return baseLen
	}
	plbs := length.ArgLength()
	if plbs == nil {
		return baseLen
	}

	// TODO - ideally we'd be using createRangeBdry() here.  However, strings
	//        use RangeLength not RangeArg and so it requires a bit more
	//        refactoring to incorporate strings within that function.
	var imin, imax uint64
	imin = baseLen.Lbs[0].Start
	imax = baseLen.Lbs[len(baseLen.Lbs)-1].End

	lbs := make(schema.LbSlice, 0, len(plbs))
	var lb schema.Lb
	//create schema length boundaries from parser values,
	for _, p := range plbs {
		if p.Min {
			lb.Start = imin
		} else {
			lb.Start = p.Start
			if p.Start < imin {
				c.error(n, errors.New(
					"derived type length must be restrictive"))
			}
		}
		if p.Max {
			lb.End = imax
		} else {
			lb.End = p.End
			if p.End > imax {
				c.error(n, errors.New(
					"derived type length must be restrictive"))
			}
		}

		var rangeMin, rangeMax uint64
		var curStart, curEnd uint64
		for index, length := range baseLen.Lbs {
			curStart = length.Start
			curEnd = length.End
			// Only update rangeMin if new range is not contiguous.
			// NB: beware decimal64 where no ranges are contiguous!
			if (index == 0) || ((rangeMax + 1) != curStart) {
				rangeMin = curStart
			}
			rangeMax = curEnd

			if (lb.Start >= rangeMin) && (lb.End <= rangeMax) {
				// Start is big enough and end small enough.  Note start
				// could be > end, but in that case, we will catch that
				// in the validation below where we check for end < start.
				break
			}

			if lb.Start < rangeMin {
				// We can assume base_rb has been validated, so if no match
				// here we have a less restrictive range.
				c.error(n, errors.New(
					"derived range must be restrictive"))
			}

			// No point doing anything if pend > rangeMax.  Either we'll loop
			// to the next range and it will fit, or we will keep looping.
			// Bear in mind that we've already checked against absolute
			// max of last entry in range so we can't exceed that.
		}

		lbs = append(lbs, lb)
	}
	//Validate disjointness and ordering
	var rangeBdrySlice schema.RangeBoundarySlicer
	rangeBdrySlice = lbs
	c.validateRangeBoundaries(rangeBdrySlice, n)

	return &schema.Length{
		Lbs:    lbs,
		Msg:    length.Msg(),
		AppTag: length.AppTag(),
	}
}

func (c *Compiler) getDefault(base schema.Type, def string, hasDef bool) (string, bool) {
	if base == nil || hasDef {
		return def, hasDef
	}
	return base.Default()
}

func (comp *Compiler) makeString(
	node parse.Node,
	name xml.Name,
	base schema.String,
	def string,
	hasDef bool,
) schema.Type {

	comp.validateRestrictions(node, base, SchemaString)

	if base == nil {
		base = schema.NewString(name, nil, nil, nil, "", false)
	}

	// Override or combine with local settings
	pats := comp.getPatterns(base, node)
	pathelps := comp.getPatternHelps(base, node)
	length := comp.getLength(base, node)
	def, hasDef = comp.getDefault(base, def, hasDef)

	// Make the type
	return schema.NewString(name, pats, pathelps, length, def, hasDef)
}

func (c *Compiler) makeBits(n parse.Node, b schema.Bits) schema.Type {
	c.validateRestrictions(n, b, SchemaBits)
	//TODO(jhs): Implement bits
	if b == nil {
		return schema.NewBits(nil)
	}
	return b
}

func (c *Compiler) validateRestrictions(n parse.Node, typ schema.Type, schemaType SchemaType) {
	var msg string

	switch schemaType {
	case SchemaUnion:
		msg = "cannot restrict %s of a union type - restrictions must be applied to members instead"
	default:
		msg = "%s restriction is not valid for this type"
	}

	supp := validRestrictionsType[schemaType]

	for _, ch := range n.Children() {
		if !ch.Type().IsTypeRestriction() {
			continue
		}
		if _, ok := supp[ch.Type()]; !ok {
			c.error(n, fmt.Errorf(msg, ch.String()))
		}
	}
}

func (c *Compiler) makeBuiltinType(cfgNode, n parse.Node, typeName string, def string, hasDef bool, parentStatus schema.Status) schema.Type {

	tname := xml.Name{Space: "builtin", Local: typeName}

	var typ schema.Type
	switch typeName {
	case "binary":
		c.error(n, errors.New("unsupported builtin type "+typeName))
	case "bits":
		typ = c.makeBits(n, nil)
	case "boolean":
		typ = c.makeBoolean(tname, n, nil, def, hasDef)
	case "decimal64":
		typ = c.makeDecimal64(tname, n, nil, def, hasDef)
	case "empty":
		typ = c.makeEmpty(tname, n, nil, def, hasDef)
	case "enumeration":
		typ = c.makeEnumeration(tname, n, nil, def, hasDef)
	case "identityref":
		typ = c.makeIdentityRef(tname, cfgNode, n, nil, parentStatus, def, hasDef)
	case "instance-identifier":
		typ = c.makeInstanceId(tname, n, nil, def, hasDef)
	case "int8", "int16", "int32", "int64":
		typ = c.makeInteger(tname, n, nil, def, hasDef)
	case "leafref":
		typ = c.makeLeafref(n, tname, nil, def, hasDef)
	case "string":
		typ = c.makeString(n, tname, nil, def, hasDef)
	case "uint8", "uint16", "uint32", "uint64":
		typ = c.makeUinteger(tname, n, nil, def, hasDef)
	case "union":
		typ = c.makeUnion(tname, cfgNode, n, nil, parentStatus, def, hasDef)
	}

	// TODO make part of creating the type once CheckSyntax is out of the picture
	c.validateDefault(n, typ)

	return typ
}

func (c *Compiler) refineType(cfgNode, n parse.Node, tname xml.Name, typ schema.Type, def string, hasDef bool, parentStatus schema.Status) schema.Type {
	switch t := typ.(type) {
	case schema.Boolean:
		typ = c.makeBoolean(tname, n, t, def, hasDef)
	case schema.Decimal64:
		typ = c.makeDecimal64(tname, n, t, def, hasDef)
	case schema.Empty:
		typ = c.makeEmpty(tname, n, t, def, hasDef)
	case schema.Enumeration:
		typ = c.makeEnumeration(tname, n, t, def, hasDef)
	case schema.InstanceId:
		typ = c.makeInstanceId(tname, n, t, def, hasDef)
	case schema.Integer:
		typ = c.makeInteger(tname, n, t, def, hasDef)
	case schema.Uinteger:
		typ = c.makeUinteger(tname, n, t, def, hasDef)
	case schema.Union:
		typ = c.makeUnion(tname, cfgNode, n, t, parentStatus, def, hasDef)
	case schema.String:
		typ = c.makeString(n, tname, t, def, hasDef)
	case schema.Leafref:
		typ = c.makeLeafref(n, tname, t, def, hasDef)
	case schema.Identityref:
		typ = c.makeIdentityRef(tname, cfgNode, n, t, parentStatus, def, hasDef)
	default:
		c.error(n, errors.New("cannot modify type"))
	}

	c.validateDefault(n, typ)
	return typ
}

func (c *Compiler) validateDefault(node parse.Node, t schema.Type) {

	if defVal, defSet := t.Default(); defSet {
		// By using the existing Validate() function, we need not encode any
		// knowledge of whether an empty string is a valid default for the
		// type here.
		if err := t.Validate(nil, []string{}, defVal); err != nil {
			if err != nil {
				c.error(node, fmt.Errorf("Invalid default '%s' for %s: %s\n",
					defVal, t.Name(), err))
			}
		}
	}
}

// This recursively works down a chain of type / typedefs until it gets to
// the base type.  It then builds the base type, checking range.  Then
// we work back up, refining the type as we go if it's a typedef type.
//
// We also set the default, if specified, and pass it up.  At each level,
// a new default overrides an existing one, otherwise we inherit the default,
// if it exists.
//
// Cannot have default and mandatory - unclear how we will check if default
// is inherited.
func (c *Compiler) BuildType(
	cfgNode parse.Node,
	typ parse.Node,
	def string,
	hasDef bool,
	parentStatus schema.Status,
) schema.Type {

	baseType, tname, done := c.BuildBaseType(cfgNode, typ, def, hasDef, parentStatus)
	if done {
		c.filterDisabledExtensions(typ)
		return c.extendType(typ, nil, baseType)
	}

	// Having constructed the underlying type, we can now add the likes
	// of range / length etc.
	t := c.refineType(cfgNode, typ, tname, baseType, def, hasDef, parentStatus)
	c.filterDisabledExtensions(typ)
	return c.extendType(typ, baseType, t)
}

func (c *Compiler) BuildBaseType(
	cfgNode parse.Node,
	typ parse.Node,
	def string,
	hasDef bool,
	parentStatus schema.Status,
) (schema.Type, xml.Name, bool) {

	//recursively build type into its base components
	//no runtime lookup
	var refType parse.Node
	var ok bool
	tname := typ.ArgIdRef()
	var typeName string
	if tname.Space != "" {
		refMod, err := typ.GetModuleByPrefix(
			tname.Space, c.modules, c.skipUnknown)
		if err != nil {
			c.error(typ, err)
		}
		tname.Space = refMod.Name()
		refType, ok = refMod.LookupType(tname.Local)
		typeName = tname.Space + ":" + tname.Local
	} else {
		typeName = tname.Local
		tname.Space = typ.Root().Name()
		refType, ok = typ.LookupType(tname.Local)
	}
	if !ok {
		if c.skipUnknown {
			return c.makeString(typ, xml.Name{Space: "builtin", Local: "string"}, nil, def, hasDef), tname, true
		}
		c.error(typ, fmt.Errorf("unknown type %s", typeName))
	}

	if refType == nil {
		return c.makeBuiltinType(cfgNode, typ, tname.Local, def, hasDef, parentStatus), tname, true
	}
	c.assertReferenceStatus(typ, refType, parentStatus)

	typ2 := refType.ChildByType(parse.NodeTyp)
	tdef := refType.Def()
	thasdef := refType.HasDef()
	return c.BuildType(cfgNode, typ2, tdef, thasdef, schema.Current), tname, false
}

func (c *Compiler) CheckMinMax(n parse.Node, min, max uint) {
	if max == 0 {
		c.error(n, errors.New("max-elements must be greater than 0"))
	} else if min > max {
		c.error(n, errors.New("min-elements must be less than max-elements"))
	}
}

func YangModulesFromDir(dir string) ([]string, error) {
	fi, err := os.Stat(dir)
	if err != nil {
		return nil, err
	}
	if !fi.Mode().IsDir() {
		return nil, errors.New("Not a directory")
	}
	d, err := os.Open(dir)
	if err != nil {
		return nil, err
	}
	names, err := d.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	fnames := make([]string, 0)
	for _, name := range names {
		if !strings.HasSuffix(name, ".yang") {
			continue
		}
		fname := dir + "/" + name
		fnames = append(fnames, fname)
	}
	return fnames, nil
}

func ParseModuleDir(dir string, extCard parse.NodeCardinality) (map[string]*parse.Tree, error) {
	return ParseYang(extCard, YangDirs(dir))
}

func ParseModules(extCard parse.NodeCardinality, list ...string) (map[string]*parse.Tree, error) {
	modules := make(map[string]*parse.Tree)
	stringInterner := parse.NewStringInterner()
	argInterner := parse.NewArgInterner()
	for _, fname := range list {
		text, err := ioutil.ReadFile(fname)
		if err != nil && err != io.EOF {
			return nil, err
		}
		t, err := parse.ParseWithInterners(
			fname, string(text), extCard, stringInterner, argInterner)
		if err != nil {
			return nil, err
		}
		mod := t.Root.Argument().String()
		if n, ok := modules[mod]; ok {
			return nil, errors.New("module " + mod + " is already defined by file " + n.ParseName)
		}
		modules[mod] = t
	}
	return modules, nil
}

func ParseYang(extCard parse.NodeCardinality, locator YangLocator) (map[string]*parse.Tree, error) {
	yangfiles, err := locator()
	if err != nil {
		return nil, err
	}
	return ParseModules(extCard, yangfiles...)
}

func getMods(extensions Extensions, cfg *Config,
) (map[string]*parse.Tree, error,
) {
	var ext_card func(parse.NodeType) map[parse.NodeType]parse.Cardinality = nil
	if extensions != nil {
		ext_card = extensions.NodeCardinality
	}
	return ParseYang(ext_card,
		YangLocations(YangDirs(cfg.YangDir), cfg.YangLocations))
}

type YangLocator func() ([]string, error)

func YangDirs(dirs ...string) YangLocator {
	return func() ([]string, error) {
		y := make([]string, 0)
		for _, d := range dirs {
			if d == "" {
				continue
			}
			f, err := YangModulesFromDir(d)
			if err != nil {
				continue
			}
			y = append(y, f...)
		}
		return y, nil
	}
}

func YangFiles(files ...string) YangLocator {
	return func() ([]string, error) {
		y := make([]string, 0)
		for _, f := range files {
			if f == "" {
				continue
			}
			y = append(y, f)
		}
		return y, nil
	}
}

func YangLocations(locations ...YangLocator) YangLocator {
	return func() ([]string, error) {
		y := make([]string, 0)
		for _, l := range locations {
			if l == nil {
				continue
			}
			f, err := l()
			if err != nil {
				return nil, err
			}
			y = append(y, f...)
		}
		return y, nil
	}
}

func CompileDir(extensions Extensions, cfg *Config,
) (schema.ModelSet, error,
) {
	if mods, err := getMods(extensions, cfg); err == nil {
		modules, submodules := parse.GetModulesAndSubmodules(mods)
		ms, _, err := compileInternal(extensions, modules, submodules,
			cfg.features(), cfg.SkipUnknown, dontGenWarnings, cfg.Filter,
			cfg.UserFnCheckFn)
		return ms, err
	} else {
		return nil, err
	}
}

func CompileDirWithWarnings(extensions Extensions, cfg *Config,
) (schema.ModelSet, []xutils.Warning, error,
) {
	if mods, err := getMods(extensions, cfg); err == nil {
		modules, submodules := parse.GetModulesAndSubmodules(mods)
		return compileInternal(extensions, modules, submodules,
			cfg.features(), cfg.SkipUnknown, genWarnings, cfg.Filter,
			cfg.UserFnCheckFn)
	} else {
		return nil, nil, err
	}
}

func CompileDirKeepMods(extensions Extensions, cfg *Config,
) (schema.ModelSet, error, map[string]*parse.Module, map[string]*parse.Module,
) {
	if mods, err := getMods(extensions, cfg); err == nil {
		modules, submodules := parse.GetModulesAndSubmodules(mods)
		st, _, retErr := compileInternal(extensions, modules, submodules,
			cfg.features(), cfg.SkipUnknown,
			dontGenWarnings, cfg.Filter, cfg.UserFnCheckFn)

		return st, retErr, modules, submodules
	} else {
		return nil, err, nil, nil
	}
}

func CompileModules(
	extensions Extensions,
	mods map[string]*parse.Tree,
	features string,
	skipUnknown bool,
	filter SchemaFilter,
) (schema.ModelSet, error) {

	modules, submodules := parse.GetModulesAndSubmodules(mods)

	ms, _, err := compileInternal(extensions, modules, submodules,
		FeaturesFromLocations(true, features), skipUnknown, dontGenWarnings,
		filter, nil)
	return ms, err
}

func CompileModulesWithWarnings(
	extensions Extensions,
	mods map[string]*parse.Tree,
	features string,
	skipUnknown bool,
	filter SchemaFilter,
) (schema.ModelSet, []xutils.Warning, error) {

	modules, submodules := parse.GetModulesAndSubmodules(mods)

	return compileInternal(extensions, modules, submodules,
		FeaturesFromLocations(true, features), skipUnknown, genWarnings,
		filter, nil)
}

func CompileModulesWithWarningsAndCustomFunctions(
	extensions Extensions,
	mods map[string]*parse.Tree,
	features string,
	skipUnknown bool,
	filter SchemaFilter,
	userFnChecker xpath.UserCustomFunctionCheckerFn,
) (schema.ModelSet, []xutils.Warning, error) {

	modules, submodules := parse.GetModulesAndSubmodules(mods)

	return compileInternal(extensions, modules, submodules,
		FeaturesFromLocations(true, features), skipUnknown, genWarnings,
		filter, userFnChecker)
}

func CompileParseTrees(
	extensions Extensions,
	mods map[string]*parse.Tree,
	features FeaturesChecker,
	skipUnknown bool,
	filter SchemaFilter,
) (schema.ModelSet, error) {

	modules, submodules := parse.GetModulesAndSubmodules(mods)

	ms, _, err := compileInternal(extensions, modules, submodules,
		features, skipUnknown, dontGenWarnings, filter, nil)
	return ms, err
}

func convertSubmodules(
	modules map[string]schema.Model,
	submods map[string]*parse.Module,
) map[string]schema.Submodule {

	convertedSubmods := make(map[string]schema.Submodule, len(submods))
	for name, submod := range submods {
		belongsTo := submod.GetModule().ChildByType(parse.NodeBelongsTo).Name()
		if mod, ok := modules[belongsTo]; ok {
			convertedSubmods[name] = schema.NewSubmodule(
				name, mod.Namespace(), mod.Version(), submod.GetTree().String())
		}
	}
	return convertedSubmods
}

func compileInternal(
	extensions Extensions,
	modules map[string]*parse.Module,
	submodules map[string]*parse.Module,
	features FeaturesChecker,
	skipUnknown,
	generateWarnings bool,
	filter SchemaFilter,
	userFnChecker xpath.UserCustomFunctionCheckerFn,
) (schema.ModelSet, []xutils.Warning, error) {

	c := NewCompiler(extensions, modules, submodules, features,
		skipUnknown, generateWarnings, filter).
		addCustomFnChecker(userFnChecker)

	err := c.ExpandModules()
	if err != nil {
		return nil, nil, err
	}

	moduleSchemas, err := c.BuildModules()
	if err != nil {
		return nil, nil, err
	}

	ms, err := schema.NewModelSet(moduleSchemas,
		convertSubmodules(moduleSchemas, submodules))
	if err != nil {
		return nil, nil, err
	}

	var warns []xutils.Warning
	if generateWarnings {
		var dummyNS schema.NodeSpec
		validXpathStmts := make(map[string]bool)

		_, _, intfWarns := ms.FindOrWalk(
			dummyNS, nodePathEvaluate, &validXpathStmts)
		for _, warn := range intfWarns {
			warns = append(warns, warn.(xutils.Warning))
		}
		warns = filterOutSometimesValidPaths(warns, validXpathStmts)

		// Add in any warnings created during YANG compilation
		warns = append(warns, c.getWarnings()...)
	}

	retMs, retErr := c.extendModelSet(ms)
	return retMs, warns, retErr
}

// Where groupings are shared (eg between non-VRF and VRF files), a path may
// be valid for the non-VRF case, and not for the VRF case.  In the latter,
// this part of the XPATH statement obviously needs to be carefully considered
// so it doesn't cause the whole expression to fail.  It is however perfectly
// valid, so we filter these out to avoid noise in the validation output.
func filterOutSometimesValidPaths(
	warns []xutils.Warning,
	validXpathStmts map[string]bool,
) []xutils.Warning {

	var filteredWarns []xutils.Warning

	for _, warn := range warns {
		_, ok := validXpathStmts[warn.GetUniqueString(xutils.StripPrefix)]
		if ok && ((warn.GetType() == xutils.DoesntExist) ||
			(warn.GetType() == xutils.MissingOrWrongPrefix)) {
			continue
		}
		filteredWarns = append(filteredWarns, warn)
	}

	return filteredWarns
}

func runPathEval(
	pathEvalMach *xpath.Machine,
	targetNode schema.Node,
	parentNode *schema.XNode,
	param interface{},
) []xutils.Warning {

	var warnings []xutils.Warning

	res := xpath.NewCtxFromMach(
		pathEvalMach, schema.NewXNode(targetNode, parentNode)).
		Run()
	newWarnings := res.GetWarnings()
	if len(newWarnings) > 0 {
		warnings = append(warnings, newWarnings...)
	}
	newNonWarnings := res.GetNonWarnings()
	validMap := param.(*map[string]bool)
	for _, nonWarn := range newNonWarnings {
		(*validMap)[nonWarn.GetUniqueString(xutils.StripPrefix)] = true
	}

	return warnings
}

func genWarningForMustOrWhenOnNPCont(
	targetNode schema.Node,
	path []string,
	mach *xpath.Machine,
	warnType xutils.WarnType,
) xutils.Warning {

	pt := xutils.PathType{}
	pt = append(pt, path...)
	pt = append(pt, targetNode.Name())

	return xutils.NewWarning(
		warnType,
		"/"+pt.String(),
		mach.GetExpr(),
		mach.GetLocation(),
		"(n/a)",
		"" /* No debug */)
}

func nodeHasVisibleDefaultOrMandatory(n schema.Node) bool {
	if n.HasDefault() {
		return true
	}
	if n.Mandatory() {
		return true
	}
	if nodeIsPresenceContainer(n) {
		return false
	}
	if nodeIsList(n) {
		return false
	}

	for _, child := range n.Children() {
		if nodeHasVisibleDefaultOrMandatory(child) {
			return true
		}
	}
	return false
}

func nodeIsPresenceContainer(n schema.Node) bool {
	_, ok := n.(schema.Container)
	return ok && n.HasPresence()
}

func nodeIsList(n schema.Node) bool {
	_, ok := n.(schema.List)
	return ok
}

func nodeHasChildNPContainer(n schema.Node) bool {
	for _, child := range n.Children() {
		if _, ok := child.(schema.Container); ok {
			if !child.HasPresence() {
				return true
			}
		}
	}
	return false
}

func checkIfNodeIsNPContWithoutDefaults(targetNode schema.Node,
) (bool, xutils.WarnType) {

	if _, ok := targetNode.(schema.Container); !ok {
		return false, xutils.ValidPath
	}

	if targetNode.HasPresence() {
		return false, xutils.ValidPath
	}

	// Does it have a default or mandatory underneath?  If not, we need to
	// generate a warning.  Exact type depends on whether we have a child NP
	// container (in which case node will always be instantiated and we are
	// more likely to have a problem).
	if !nodeHasVisibleDefaultOrMandatory(targetNode) {
		if nodeHasChildNPContainer(targetNode) {
			return true, xutils.MustOnNPContWithNPChild
		}
		return true, xutils.MustOnNPContainer
	}

	return false, xutils.ValidPath
}

// Path Evaluation is a one-off process, done at compile time.  Once the
// machine is run, we no longer need it, so can free up the resources.
func nodePathEvaluate(
	targetNode schema.Node,
	parentNode *schema.XNode,
	nodeToFind schema.NodeSpec,
	path []string,
	param interface{},
) (bool, bool, []interface{}) {

	musts := targetNode.Musts()
	whens := targetNode.Whens()
	if len(musts) == 0 && len(whens) == 0 {
		return false, true, nil
	}

	var warnings []xutils.Warning

	genMustOrWhenOnNPContWarning, warnType :=
		checkIfNodeIsNPContWithoutDefaults(targetNode)

	for _, must := range musts {
		if must.PathEvalMach != nil {
			addWarnings(&warnings,
				must.PathEvalMach, genMustOrWhenOnNPContWarning, warnType,
				targetNode, parentNode, path, param)
			must.PathEvalMach = nil
		}

		if must.ExtPathEvalMach != nil {
			addWarnings(&warnings,
				must.ExtPathEvalMach, genMustOrWhenOnNPContWarning, warnType,
				targetNode, parentNode, path, param)
			must.ExtPathEvalMach = nil
		}
	}
	for _, when := range whens {
		if when.PathEvalMach != nil {
			addWarnings(&warnings,
				when.PathEvalMach, genMustOrWhenOnNPContWarning, warnType,
				targetNode, parentNode, path, param)
			when.PathEvalMach = nil
		}
	}

	intfWarnings := make([]interface{}, 0, len(warnings))
	for _, warn := range warnings {
		intfWarnings = append(intfWarnings, warn)
	}
	return false, true, intfWarnings
}

func addWarnings(
	warnings *[]xutils.Warning,
	mach *xpath.Machine,
	genNPContWarning bool,
	warnType xutils.WarnType,
	targetNode schema.Node,
	parentNode *schema.XNode,
	path []string,
	param interface{},
) {
	*warnings = append(*warnings, runPathEval(
		mach, targetNode, parentNode, param)...)

	if genNPContWarning {
		*warnings = append(*warnings,
			genWarningForMustOrWhenOnNPCont(
				targetNode, path, mach, warnType))
	}
}
