// Copyright (c) 2017-2021, AT&T Intellectual Property.
// All rights reserved.
//
// Copyright (c) 2014-2017 by Brocade Communications Systems, Inc.
// All rights reserved.
//
// SPDX-License-Identifier: MPL-2.0

package schema

import (
	"encoding/xml"
	"errors"
	"fmt"
	"sort"

	"github.com/danos/mgmterror"
	"github.com/steiler/yang-parser/xpath"
	"github.com/steiler/yang-parser/xpath/xutils"
)

func cardinalityInRange(p xutils.PathType, min, max uint, len int) error {
	if max == ^uint(0) {
		if (min > 0) && (uint(len) < min) {
			path := p.String()
			err := mgmterror.NewTooFewElementsError(path)
			err.Path = path
			err.Message = fmt.Sprintf("Invalid number of nodes: must be at least %d", min)
			return err
		}
	} else if ((min > 0) && (uint(len) < min)) || ((max > 0) && (uint(len) > max)) {
		path := p.String()
		err := mgmterror.NewTooManyElementsError(path)
		err.Path = path
		err.Message = fmt.Sprintf("Invalid number of nodes: must be in the range %d to %d", min, max)
		return err
	}
	return nil
}

type Model interface {
	Tree
	EncodeXML(*xml.Encoder)
	Identifier() string
	Version() string
	Data() string
	Features() []string
	Rpcs() map[string]Rpc
	Notifications() map[string]Notification
	Deviations() []string
	isModel()
}

type schemaDetails struct {
	Identifier string `xml:"identifier"`
	Version    string `xml:"version"`
	Format     string `xml:"format"`
	Namespace  string `xml:"namespace"`
	Location   string `xml:"location"`
}

type model struct {
	Tree
	schema        schemaDetails
	data          string
	features      []string
	rpcs          map[string]Rpc
	notifications map[string]Notification
	deviations    []string
}

// Ensure that other schema types don't meet the interface
func (s *model) isModel() {}

// Compile time check that the concrete type meets the interface
var _ Model = (*model)(nil)

func (s *model) Identifier() string   { return s.schema.Identifier }
func (s *model) Version() string      { return s.schema.Version }
func (s *model) Format() string       { return s.schema.Format }
func (s *model) Namespace() string    { return s.schema.Namespace }
func (s *model) Location() string     { return s.schema.Location }
func (s *model) Data() string         { return s.data }
func (s *model) Features() []string   { return s.features }
func (s *model) Rpcs() map[string]Rpc { return s.rpcs }

func (s *model) Notifications() map[string]Notification {
	return s.notifications
}

func (s *model) Deviations() []string { return s.deviations }

func (s *model) EncodeXML(enc *xml.Encoder) {
	enc.EncodeElement(s.schema, xml.StartElement{Name: xml.Name{Local: "schema"}})
}

func NewModel(
	name, revision, namespace, data string,
	tree Tree,
	rpcs map[string]Rpc,
	features []string,
	notifications map[string]Notification,
	deviations []string,
) Model {

	return &model{
		Tree:          tree,
		schema:        schemaDetails{name, revision, "yang", namespace, "NETCONF"},
		data:          data,
		rpcs:          rpcs,
		features:      features,
		notifications: notifications,
		deviations:    deviations,
	}
}

type Submodule interface {
	EncodeXML(*xml.Encoder)
	Identifier() string
	Namespace() string
	Data() string
	isSubmodule()
}

type submodule struct {
	schema schemaDetails
	data   string
}

// Ensure that other schema types don't meet the interface
func (s *submodule) isSubmodule() {}

// Compile time check that the concrete type meets the interface
var _ Submodule = (*submodule)(nil)

func NewSubmodule(identifier, namespace, revision, data string) Submodule {
	return &submodule{
		schema: schemaDetails{
			identifier, revision, "yang", namespace, "NETCONF"},
		data: data}
}

func (s *submodule) Identifier() string { return s.schema.Identifier }
func (s *submodule) Namespace() string  { return s.schema.Namespace }
func (s *submodule) Data() string       { return s.data }

func (s *submodule) EncodeXML(enc *xml.Encoder) {
	enc.EncodeElement(s.schema,
		xml.StartElement{Name: xml.Name{Local: "schema"}})
}

type Rpc interface {
	Input() Tree
	Output() Tree
	isRpc()
}

type rpc struct {
	input  Tree
	output Tree
}

// Ensure that other schema types don't meet the interface
func (r *rpc) isRpc() {}

// Compile time check that the concrete type meets the interface
var _ Rpc = (*rpc)(nil)

func NewRpc(input, output Tree) Rpc {
	return &rpc{input, output}
}

func (r *rpc) Input() Tree  { return r.input }
func (r *rpc) Output() Tree { return r.output }

type Notification interface {
	Schema() Tree
	isNotification()
}

type notification struct {
	notification Tree
}

// Ensure that other chema types don't meet the interface
func (n *notification) isNotification() {}

// Compile time check that the concrete type meets the interface
var _ Notification = (*notification)(nil)

func NewNotification(n Tree) Notification {
	return &notification{notification: n}
}

func (n *notification) Schema() Tree { return n.notification }

type Tree interface {
	Node
	isTree()
}

type tree struct {
	*node
}

// Ensure that other schema types don't meet the interface
func (*tree) isTree() {}

// Compile time check that the concrete type meets the interface
var _ Tree = (*tree)(nil)

func (t *tree) Paths(s string) []string {
	out := make([]string, 0)
	keys := make([]string, 0, len(t.children))
	for k, _ := range t.children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c := t.children[k]
		cpaths := c.Paths("")
		out = append(out, cpaths...)
	}
	return out
}

func (t *tree) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		return nil
	}
	c, ok := t.node.children[p[0]]
	if !ok {
		return NewPathInvalidError(path, p[0])
	}
	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}

func (t *tree) HasPresence() bool {
	return true
}

func (t *tree) Child(s string) Node {
	return t.children[s]
}

func (t *tree) Descendant(path []string) Node {
	return t.descendant(t, path)
}

// Nothing of substance. Perhaps needs a different type...
func NewTree(children []Node) (Tree, error) {
	t := &tree{node: makenode()}
	t.config = true

	err := t.addChildren(children)
	if err != nil {
		return nil, err
	}

	return t, nil
}

type ModelSet interface {
	Tree
	Modules() map[string]Model
	Submodules() map[string]Submodule
	Rpcs() map[string]map[string]Rpc
	Notifications() map[string]map[string]Notification
	isModelSet()
	FindOrWalk(
		nodeToFind NodeSpec,
		actionFn ActionFnType,
		param interface{}) (Node, bool, []interface{})
}

type modelSet struct {
	tree
	modules       map[string]Model
	submodules    map[string]Submodule
	rpcs          map[string]map[string]Rpc
	notifications map[string]map[string]Notification
}

// Ensure that other schema types don't meet the interface
func (*modelSet) isModelSet() {}

// Compile time check that the concrete type meets the interface
var _ ModelSet = (*modelSet)(nil)

func (t *modelSet) Modules() map[string]Model        { return t.modules }
func (t *modelSet) Submodules() map[string]Submodule { return t.submodules }
func (t *modelSet) Rpcs() map[string]map[string]Rpc  { return t.rpcs }

func (t *modelSet) Notifications() map[string]map[string]Notification {
	return t.notifications
}

func NewModelSet(modules map[string]Model, submodules map[string]Submodule,
) (ModelSet, error) {
	ms := &modelSet{tree: tree{node: makenode()}}

	ms.config = true
	ms.submodules = submodules
	ms.modules = make(map[string]Model)
	ms.rpcs = make(map[string]map[string]Rpc)
	ms.notifications = make(map[string]map[string]Notification)

	// Merge the modules into a single tree
	for _, mod := range modules {
		ms.modules[mod.Identifier()] = mod
		err := ms.addChildren(mod.Children())
		if err != nil {
			return nil, err
		}

		ms.rpcs[mod.Namespace()] = mod.Rpcs()
		ms.notifications[mod.Namespace()] = mod.Notifications()
		for _, chs := range mod.Choices() {
			ms.addChoice(chs)
		}
	}

	return ms, nil
}

type Node interface {
	Child(string) Node
	descendant(Node, []string) Node
	Descendant([]string) Node
	Validate(ValidateCtx, []string, []string) error
	CheckCardinality(xutils.PathType, int) error
	Children() []Node
	OpdChildren() []Node
	Choices() []Node
	Name() string
	Namespace() string
	Module() string
	Submodule() string
	Paths(string) []string
	Whens() []WhenContext
	Musts() []MustContext
	Type() Type
	HasDefault() bool
	DefaultChild(name string) Node
	DefaultChildNames() []string
	DefaultChildren() []Node
	OrdBy() string
	HasPresence() bool
	Config() bool
	String() string
	Status() Status
	Description() string
	Repeatable() bool
	addParent(*node)
	addParentToChildren()
	Parent() *node
	addArgument(string)
	Arguments() []string
	Mandatory() bool
}

// WhenAndMustContext stores common context for When and Must machines.
type WhenAndMustContext struct {
	// Mach is the machine that evaulates the full XPATH logic.
	Mach *xpath.Machine

	// ErrMsg is the error message when the XPATH expression returns false.
	//
	// For 'must', this can be specified in YANG, or a default generated by
	// the compiler.  For 'when', it is always the latter.
	ErrMsg string

	// AppTag is used in the <rpc-error> error-app-tag field in errors sent
	// to NETCONF clients.
	AppTag string

	// PathEvalMach is the machine that checks paths in the XPATH expression.
	PathEvalMach *xpath.Machine
	Namespace    string
}

type WhenContext struct {
	WhenAndMustContext

	// RunAsParent indicates whether the machine is run on node or parent.
	//
	// 'when' statements can be added directly under an augment statement. In
	// this case we store the statement on each augmented child, but need to
	// run it using the context (current node) set to the parent node.
	RunAsParent bool
}

func NewWhenContext(
	mach, pathEvalMach *xpath.Machine,
	errMsg string,
	runAsParent bool,
	namespace string,
) WhenContext {
	return WhenContext{
		WhenAndMustContext: WhenAndMustContext{
			Mach:         mach,
			PathEvalMach: pathEvalMach,
			ErrMsg:       errMsg,
			Namespace:    namespace,
		},
		RunAsParent: runAsParent,
	}
}

type MustContext struct {
	WhenAndMustContext

	// Used to evaluate paths in the configd:must extension
	ExtPathEvalMach *xpath.Machine
}

func NewMustContext(
	mach, basePathEvalMach, extPathEvalMach *xpath.Machine,
	errMsg string,
	appTag string,
	namespace string,
) MustContext {
	return MustContext{
		WhenAndMustContext: WhenAndMustContext{
			Mach:         mach,
			PathEvalMach: basePathEvalMach,
			ErrMsg:       errMsg,
			AppTag:       appTag,
			Namespace:    namespace,
		},
		ExtPathEvalMach: extPathEvalMach,
	}
}

type node struct {
	name         xml.Name
	submodule    string
	children     map[string]Node
	defChildren  map[string]Node
	Desc         string
	Ref          string
	module       string
	config       bool
	status       Status
	parent       *node
	arguments    []string
	whenContexts []WhenContext
	mustContexts []MustContext
	choices      []Node
}

func (n *node) String() string {
	return n.name.Local
}

func (n *node) Status() Status {
	return n.status
}

func makenode() *node {
	n := &node{}
	n.children = make(map[string]Node)
	n.defChildren = make(map[string]Node)
	return n
}
func NewNode() *node {
	return makenode()
}

func (n *node) Type() Type {
	return nil
}
func (n *node) Whens() []WhenContext {
	return n.whenContexts
}
func (n *node) Musts() []MustContext {
	return n.mustContexts
}
func (n *node) addArgument(string) {}
func (n *node) Arguments() []string {
	return n.arguments
}
func (n *node) Paths(path string) []string {
	path = path + "/" + n.name.Local
	out := make([]string, 0)
	if len(n.children) == 0 {
		return []string{path}
	}
	keys := make([]string, 0, len(n.children))
	for k, _ := range n.children {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		c := n.children[k]
		cpaths := c.Paths(path)
		out = append(out, cpaths...)
	}
	return out
}

func (n *node) addChoice(ch Node) error {
	name := ch.Name()
	for _, cn := range n.choices {
		if cn.Name() == name {
			return errors.New("redefinition of name " + name)
		}
	}

	n.choices = append(n.choices, ch)
	return nil
}

func (n *node) addChild(ch Node) error {
	name := ch.Name()
	if _, exists := n.children[name]; exists {
		return errors.New("redefinition of name " + name)
	}
	n.children[name] = ch

	if ch.HasDefault() {
		n.defChildren[ch.Name()] = ch
	}

	return nil
}

type nodeFilter = func(n Node) bool

func anyNode(n Node) bool { return true }

func choiceNode(n Node) bool {
	_, ok := n.(Choice)
	return ok
}

func caseNode(n Node) bool {
	_, ok := n.(Case)
	return ok
}

// addAction allows additional/alternative actions to be performed
// when adding chilren nodes to a node. It takes the parent node, that
// children are being added to, and a child Node that may be
// added as a child to the parent.
// It returns true when when no other actions should be performed
// after this action.

type addAction = func(parent *node, child Node) (halt bool, err error)

func addChoiceToChoices(parent *node, child Node) (bool, error) {
	var err error
	if _, ok := child.(Choice); ok {
		err = parent.addChoice(child)
	}
	return false, err
}

func addToChoices(include nodeFilter) addAction {
	return func(parent *node, child Node) (bool, error) {
		if include != nil && include(child) {
			if err := parent.addChoice(child); err != nil {
				return true, err
			}
		}
		return false, nil
	}
}

func includeChildrenOf(include, exclude nodeFilter) addAction {
	return func(parent *node, child Node) (bool, error) {
		if include != nil && include(child) == false {
			return false, nil
		}
		for _, ch := range child.Children() {
			if exclude != nil && exclude(ch) == true {
				continue
			}
			if err := parent.addChild(ch); err != nil {
				return true, err
			}
		}
		return true, nil
	}
}

func addToChildrenExcluding(exclude nodeFilter) addAction {
	return func(parent *node, child Node) (bool, error) {
		if exclude != nil && exclude(child) {
			return false, nil
		}
		err := parent.addChild(child)
		return false, err
	}
}

// addChildrenWithActionChain performs the given actions in the order provided until
// one of the actions returns halt, at which point all other actions are ignored.
func (n *node) addChildrenWithActionChain(children []Node, actions ...addAction) error {
	for _, ch := range children {
		for _, action := range actions {
			halt, err := action(n, ch)
			if err != nil {
				return err
			}
			if halt {
				// no more processing for this child node
				break
			}
		}
	}

	return nil
}

func (n *node) addChildren(children []Node) error {
	return n.addChildrenWithActionChain(children,
		addChoiceToChoices,
		includeChildrenOf(choiceNode, caseNode),
		addToChildrenExcluding(choiceNode),
	)
}

func (n *node) Repeatable() bool {
	return false
}

func (n *node) addOpdChildren(children []Node) error {
	for _, ch := range children {
		if err := n.addChild(ch); err != nil {
			return err
		}
		if _, ok := ch.(OpdArgument); ok {
			n.addArgument(ch.Name())
		}
	}
	return nil
}

func (n *node) addParent(p *node) {
}

func (n *node) Parent() *node {
	return n.parent
}

func (n *node) removeChild(ch Node, children map[string]Node) error {
	name := ch.Name()
	if _, ok := children[name]; ok {
		delete(children, name)
	}
	return nil
}

func (n *node) Name() string {
	return n.name.Local
}

func (n *node) Namespace() string {
	return n.name.Space
}

func (n *node) Module() string {
	return n.module
}

func (n *node) Submodule() string {
	return n.submodule
}

func (n *node) DefaultChild(s string) Node {
	return n.defChildren[s]
}

func (n *node) descendant(spec Node, p []string) Node {
	if len(p) == 0 {
		return spec
	}
	c := spec.Child(p[0])
	if c == nil {
		return nil
	}
	return c.descendant(c, p[1:])
}

func genChildList(children map[string]Node) []Node {
	ch := make([]Node, 0, len(children))
	for _, v := range children {
		ch = append(ch, v)
	}

	return ch
}

func genSchemaChildList(children map[string]Node) []Node {
	ch := make([]Node, 0, len(children))
	for _, v := range children {
		switch v.(type) {

		case Choice:
			ch = append(ch, v)
		case Case:
			ch = append(ch, v)
		default:
			ch = append(ch, v)
		}
	}
	return ch

}

func (n *node) Choices() []Node {
	return n.choices
}

func (n *node) Children() []Node {
	return genChildList(n.children)
}

func (n *node) DefaultChildren() []Node {
	return genChildList(n.defChildren)
}

func (n *node) DefaultChildNames() []string {
	chs := make([]string, 0, len(n.defChildren))
	for _, v := range n.defChildren {
		chs = append(chs, v.Name())
	}
	return chs
}

func (n *node) OrdBy() string {
	return "system"
}

func (n *node) Config() bool {
	return n.config
}

func (n *node) Description() string {
	return n.Desc
}

func (n *node) HasPresence() bool { return false }

func (n *node) Mandatory() bool { return false }

func (n *node) Validate(ctx ValidateCtx, path []string, p []string) error {
	return nil
}

func (n *node) CheckCardinality(p xutils.PathType, len int) error {
	return nil
}

func (n *node) HasDefault() bool { return len(n.defChildren) > 0 }

func (n *node) OpdChildren() []Node {
	return genChildList(n.children)
}

type OpdCommand interface {
	Node
	OnEnter() string
	Privileged() bool
	Local() bool
	Secret() bool
	PassOpcArgs() bool
	isOpdCommand()
}

type opdCommand struct {
	*node
	onEnter     string
	priv        bool
	local       bool
	secret      bool
	repeatable  bool
	passOpcArgs bool
}

// Ensure that other schema types don't meet the interface
func (opdCommand) isOpdCommand() {}

// Compile time check that the concrete type meets the interface
var _ OpdCommand = (*opdCommand)(nil)

func (n *opdCommand) OnEnter() string {
	return n.onEnter
}
func (n *opdCommand) Privileged() bool {
	return n.priv
}
func (n *opdCommand) Local() bool {
	return n.local
}
func (n *opdCommand) Secret() bool {
	return n.secret
}

func (n *opdCommand) Repeatable() bool {
	return n.repeatable
}
func (n *opdCommand) PassOpcArgs() bool {
	return n.passOpcArgs
}
func NewOpdCommand(
	name, namespace, modulename, desc, ref, onenter string,
	priv bool,
	local bool,
	secret bool,
	repeatable bool,
	passOpcArgs bool,
	status Status,
	children []Node,
) (OpdCommand, error) {

	c := &opdCommand{node: makenode()}
	c.name.Space = namespace
	c.name.Local = name
	c.module = modulename
	c.Desc = desc
	c.Ref = ref
	c.onEnter = onenter
	c.priv = priv
	c.local = local
	c.secret = secret
	c.repeatable = repeatable
	c.passOpcArgs = passOpcArgs
	c.config = false
	c.status = status
	c.arguments = make([]string, 1)
	c.addArguments(children)
	if err := c.addChildren(children); err != nil {
		return nil, err
	}
	if repeatable {
		c.addParentToChildren()
	}
	return c, nil
}
func (n *node) addParentToChildren() {
	p := n.Parent()
	if p == nil {
		p = n
	}
	for _, ch := range n.children {
		if ch.Repeatable() {
			ch.addParent(p)
			ch.addParentToChildren()
		}
	}
}
func (n *opdCommand) addArgument(s string) {
	n.arguments[0] = s
}
func (n *opdCommand) addArguments(children []Node) {
	for _, ch := range children {
		if _, ok := ch.(OpdArgument); ok {
			n.addArgument(ch.Name())
		}
	}
}
func (n *opdCommand) Arguments() []string { return n.arguments }

func (n *opdCommand) addParent(p *node) {
	n.parent = p
}

func (n *opdCommand) Child(s string) Node {
	return n.getOpdChild(s)
}

func (n *opdCommand) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *opdCommand) HasPresence() bool {
	return n.OnEnter() != ""
}

func (n *opdCommand) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		return nil
	}

	c := n.Child(p[0])
	if c == nil {
		return NewInvalidPathError(path)
	}

	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}

func (n *opdCommand) HasDefault() bool {
	return false
}

type Container interface {
	Node
	Presence() bool
	isContainer()
}

type container struct {
	*node
	presence bool
}

// Ensure that other schema types don't meet the interface
func (*container) isContainer() {}

// Compile time check that the concrete type meets the interface
var _ Container = (*container)(nil)

func NewContainer(
	name, namespace, modulename, submodule, desc, ref string,
	presence, config bool,
	status Status,
	whens []WhenContext,
	musts []MustContext,
	children []Node,
) (Container, error) {

	c := &container{node: makenode()}
	c.name.Space = namespace
	c.name.Local = name
	c.module = modulename
	c.submodule = submodule
	c.Desc = desc
	c.Ref = ref
	c.presence = presence
	c.config = config
	c.status = status
	c.whenContexts = whens
	c.mustContexts = musts

	if err := c.addChildren(children); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *container) Presence() bool {
	return c.presence
}

func (n *container) Child(s string) Node {
	return n.children[s]
}

func (n *container) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *container) HasPresence() bool {
	return n.Presence()
}

func (n *container) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		// If we are validating a path provided by NETCONF to use in GetTree /
		// GetTreeFull(), it's fine for the path to end on a non-presence
		// container.  It's not fine when this is a set command path.
		if n.Presence() || ctx.AllowIncompletePaths() {
			return nil
		}
		return NewMissingChildError(path)
	}
	c, ok := n.node.children[p[0]]
	if !ok {
		return NewPathInvalidError(path, p[0])
	}
	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}

func (n *container) HasDefault() bool {
	return !n.Presence() && n.node.HasDefault()
}

type Limit struct {
	Min uint
	Max uint
}

type List interface {
	Node
	Limit() Limit
	Keys() []string
	Uniques() [][][]xml.Name
	isList()
}

type list struct {
	*node
	orderedBy string
	limit     Limit
	keys      []string
	uniques   [][][]xml.Name

	entry *listEntry
}

// Ensure that other schema types don't meet the interface
func (*list) isList() {}

// Compile time check that the concrete type meets the interface
var _ List = (*list)(nil)

func NewList(
	name, namespace, modulename, submodule, desc, ref, orderedby string,
	min, max uint,
	config bool,
	status Status,
	keys []string,
	uniques [][][]xml.Name,
	whens []WhenContext,
	musts []MustContext,
	children []Node,
) (List, error) {

	if orderedby == "" {
		orderedby = "system"
	}

	l := &list{node: makenode()}
	l.name.Local = name
	l.name.Space = namespace
	l.module = modulename
	l.submodule = submodule
	l.Desc = desc
	l.Ref = ref
	l.config = config
	l.status = status
	l.orderedBy = orderedby
	l.limit = Limit{min, max}
	l.keys = keys
	l.uniques = uniques
	l.whenContexts = whens
	l.mustContexts = musts

	l.entry = &listEntry{l.node, l}

	if err := l.addChildren(children); err != nil {
		return nil, err
	}
	return l, nil
}

func (l *list) Limit() Limit            { return l.limit }
func (l *list) Keys() []string          { return l.keys }
func (l *list) Uniques() [][][]xml.Name { return l.uniques }

func (n *list) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *list) OrdBy() string {
	return n.orderedBy
}

func (n *list) Type() Type {
	return n.children[n.Keys()[0]].Type()
}
func (n *list) CheckCardinality(p xutils.PathType, len int) error {
	return cardinalityInRange(p, n.Limit().Min, n.Limit().Max, len)
}

func (n *list) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		if ctx.AllowIncompletePaths() {
			return nil
		}
		return NewMissingValueError(path)
	}
	//TODO(jhs): Multipart keys
	key := n.Keys()[0]
	k := n.children[key]

	if err := k.Validate(ctx, path, []string{p[0]}); err != nil {
		return err
	}

	path = append(path, p[0])
	p = p[1:]

	if len(p) == 0 {
		return nil
	}
	c, ok := n.children[p[0]]
	if !ok {
		return NewPathInvalidError(path, p[0])
	}
	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}

func (n *list) HasDefault() bool { return false }

func (n *list) Child(name string) Node {
	return n.entry
}

func (n *list) DefaultChildren() []Node {
	return nil
}

func (n *list) DefaultChildNames() []string {
	return nil
}

func (n *list) DefaultChild(s string) Node {
	return nil
}

type ListEntry interface {
	Node
	Keys() []string
	isListEntry() // Ensure List doesn't meet ListEntry interface
}

type listEntry struct {
	*node
	list *list
}

func (n *listEntry) isListEntry() {}

func (n *listEntry) Keys() []string { return n.list.Keys() }
func (n *listEntry) Type() Type {
	return n.list.Type()
}

func (n *listEntry) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *listEntry) Child(name string) Node {
	return n.children[name]
}

func (n *listEntry) HasPresence() bool {
	return true
}
func (n *listEntry) HasDefault() bool { return false }

// Nodes within the listEntry are always "ordered-by system" because
// only the listEntries themselves should ever be "ordered-by
// user".
func (n *listEntry) OrdBy() string {
	return "system"
}

func (n *listEntry) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		if ctx.AllowIncompletePaths() {
			return nil
		}
		return NewMissingValueError(path)
	}

	c, ok := n.children[p[0]]
	if !ok {
		return NewInvalidPathError(append(path, p[0]))
	}
	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}

type Leaf interface {
	Node
	Default() (string, bool)
	isLeaf()
}

type leaf struct {
	*node
	units     string
	mandatory bool
	typ       Type
}

// Compile time check that the concrete type meets the interface
var _ Leaf = (*leaf)(nil)

// Ensure that other schema types don't meet the interface
func (l *leaf) isLeaf() {}

func (l *leaf) Mandatory() bool { return l.mandatory }

func NewLeaf(
	name, namespace, modulename, submodule, desc, ref, units string,
	mandatory bool,
	typ Type,
	config bool,
	status Status,
	whens []WhenContext,
	musts []MustContext,
) Leaf {
	l := &leaf{node: makenode()}
	l.name.Local = name
	l.name.Space = namespace
	l.module = modulename
	l.submodule = submodule
	l.Desc = desc
	l.Ref = ref
	l.units = units
	l.mandatory = mandatory
	l.typ = typ
	l.config = config
	l.status = status
	l.whenContexts = whens
	l.mustContexts = musts
	return l
}
func (n *leaf) Descendant(path []string) Node {
	return n.descendant(n, path)
}
func (n *leaf) HasPresence() bool {
	_, ok := n.Type().(Empty)
	return ok
}

func (n *leaf) HasDefault() bool {
	_, hasDefault := n.Default()
	return hasDefault
}
func (n *leaf) Type() Type {
	return n.typ
}
func (n *leaf) Default() (string, bool) {
	if n.Mandatory() {
		return "", false
	}
	return n.Type().Default()
}

func (n *leaf) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		if _, ok := n.Type().(Empty); ok {
			return nil
		} else if ctx.AllowIncompletePaths() {
			return nil
		} else {
			return NewMissingValueError(path)
		}
	}
	h, t := p[0], p[1:]
	path = append(path, h)
	// There should be nothing after the value
	if len(t) != 0 {
		return NewPathInvalidError(path, p[1])
	}
	return n.Type().Validate(ctx, path, h)
}

func (n *leaf) DefaultChildNames() []string {
	def, hasDefault := n.Default()
	if !hasDefault {
		return nil
	}
	return []string{def}
}

func (n *leaf) DefaultChild(name string) Node {
	def, hasDefault := n.Default()
	if hasDefault && def == name {
		return &leafValue{Node: n, name: name}
	}
	return nil
}

func (n *leaf) Child(name string) Node {
	return &leafValue{Node: n, name: name}
}

type leafValue struct {
	Node
	name string
}

type LeafValue interface {
	Node
	IsLeafValue()
}

func (n *leafValue) IsLeafValue() {}

func (n *leafValue) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *leafValue) Child(name string) Node {
	return nil
}
func (n *leafValue) DefaultChild(name string) Node {
	return nil
}

func (n *leafValue) HasPresence() bool {
	return true
}

func (n *leafValue) Status() Status {
	switch t := n.Type().(type) {
	case Enumeration:
		for _, e := range t.Enums() {
			if e.String() == n.name {
				return e.Status()
			}
		}
		return Current
	// TODO: bits and identity not supported yet
	// case Bits:
	// for _, b := range t.Bits() {
	// 	if b.String() == n.name {
	// 		return b.Status()
	// 	}
	// }
	// 	return Current
	// case Identity:
	// 	return Current
	default:
		return n.Node.Status()
	}
}

type OpdArgument interface {
	Node
	Default() (string, bool)
	OnEnter() string
	Privileged() bool
	Local() bool
	Secret() bool
	PassOpcArgs() bool
	isOpdArgument()
}

type opdArgument struct {
	*node
	units       string
	mandatory   bool
	typ         Type
	opdHelp     string
	opdAllowed  string
	onEnter     string
	priv        bool
	local       bool
	secret      bool
	repeatable  bool
	passOpcArgs bool
}

// Compile time check that the concrete type meets the interface
var _ OpdArgument = (*opdArgument)(nil)

func (opdArgument) isOpdArgument() {}

func (n *opdArgument) OnEnter() string {
	return n.onEnter
}
func (n *opdArgument) Privileged() bool {
	return n.priv
}
func (n *opdArgument) Local() bool {
	return n.local
}
func (n *opdArgument) Secret() bool {
	return n.secret
}

func (n *opdArgument) Repeatable() bool {
	return n.repeatable
}
func (n *opdArgument) PassOpcArgs() bool {
	return n.passOpcArgs
}
func (n *opdArgument) addArgument(s string) {
	n.arguments[0] = s
}
func (n *opdArgument) addArguments(children []Node) {
	for _, ch := range children {
		if _, ok := ch.(OpdArgument); ok {
			n.addArgument(ch.Name())
		}
	}
}
func (n *opdArgument) Mandatory() bool { return n.mandatory }

func (n *opdArgument) Arguments() []string { return n.arguments }
func NewOpdArgument(
	name, namespace, modulename, desc, ref, units, onenter string,
	priv bool,
	local bool,
	secret bool,
	repeatable bool,
	mandatory bool,
	passOpcArgs bool,
	typ Type,
	status Status,
	children []Node,
) (OpdArgument, error) {
	o := &opdArgument{node: makenode()}
	o.name.Local = name
	o.name.Space = namespace
	o.module = modulename
	o.Desc = desc
	o.Ref = ref
	o.units = units
	o.onEnter = onenter
	o.priv = priv
	o.local = local
	o.secret = secret
	o.repeatable = repeatable
	o.mandatory = mandatory
	o.passOpcArgs = passOpcArgs
	o.typ = typ
	o.config = false
	o.status = status
	o.arguments = make([]string, 1)
	o.addArguments(children)
	if err := o.addChildren(children); err != nil {
		return nil, err
	}
	if repeatable {
		o.addParentToChildren()
	}
	return o, nil
}
func (n *opdArgument) addParent(p *node) {
	n.parent = p
}

func (n *opdArgument) Descendant(path []string) Node {
	return n.descendant(n, path)
}
func (n *opdArgument) HasPresence() bool {
	return n.OnEnter() != ""
}

func (n *opdArgument) HasDefault() bool {
	_, hasDefault := n.Default()
	return hasDefault
}
func (n *opdArgument) Type() Type {
	return n.typ
}
func (n *opdArgument) Default() (string, bool) {
	if n.Mandatory() {
		return "", false
	}
	return n.Type().Default()
}

// An opdArgument differs from all the other nodes in that the value to be
// validated is NOT p[0], but path[len(path)-1]
func (n *opdArgument) Validate(ctx ValidateCtx, path []string, p []string) error {
	if _, ok := n.Type().(Empty); !ok {
		err := n.Type().Validate(ctx, path, path[len(path)-1])
		if err != nil {
			if _, ok := err.(mgmterror.Formattable); ok {
				return err
			}
			return newInvalidValueError(path, err.Error())
		}
	}
	if len(p) == 0 {
		return nil
	}
	c := n.Child(p[0])
	if c == nil {
		return NewInvalidPathError(path)
	}
	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}

func (n *opdArgument) DefaultChildNames() []string {
	return nil
}

func (n *opdArgument) DefaultChild(name string) Node {
	return nil
}

func (n *node) getOpdChild(name string) Node {
	children := n.children

	if len(children) < 1 && n.Parent() != nil {
		n = n.Parent()
		children = n.children
	}

	ch := children[name]

	if ch == nil {
		// if present, and allowed, then try for argument
		if len(n.arguments) > 0 {
			ch = children[n.arguments[0]]
		}
	}
	return ch
}
func (n *opdArgument) Child(name string) Node {
	return n.getOpdChild(name)
}
func (n *opdArgument) Children() []Node {
	return genChildList(n.children)
}

type OpdOption interface {
	Node
	Default() (string, bool)
	OnEnter() string
	Privileged() bool
	Local() bool
	Secret() bool
	PassOpcArgs() bool
	isOpdOption()
}

type opdOption struct {
	*node
	units       string
	mandatory   bool
	typ         Type
	opdHelp     string
	opdAllowed  string
	onEnter     string
	priv        bool
	local       bool
	secret      bool
	repeatable  bool
	passOpcArgs bool
}

// Compile time check that the concrete type meets the interface
var _ OpdOption = (*opdOption)(nil)

func (opdOption) isOpdOption() {}

func (n *opdOption) Mandatory() bool     { return n.mandatory }
func (n *opdOption) Arguments() []string { return n.arguments }

func (n *opdOption) OnEnter() string {
	return n.onEnter
}
func (n *opdOption) Privileged() bool {
	return n.priv
}
func (n *opdOption) Local() bool {
	return n.local
}
func (n *opdOption) Secret() bool {
	return n.secret
}

func (n *opdOption) Repeatable() bool {
	return n.repeatable
}
func (n *opdOption) PassOpcArgs() bool {
	return n.passOpcArgs
}
func (n *opdOption) addArgument(s string) {
	n.arguments[0] = s
}
func (n *opdOption) addArguments(children []Node) {
	for _, ch := range children {
		if _, ok := ch.(OpdArgument); ok {
			n.addArgument(ch.Name())
		}
	}
}
func NewOpdOption(
	name, namespace, modulename, desc, ref, units, onenter string,
	priv bool,
	local bool,
	secret bool,
	repeatable bool,
	mandatory bool,
	passOpcArgs bool,
	typ Type,
	status Status,
	children []Node,
) (OpdOption, error) {
	o := &opdOption{node: makenode()}
	o.name.Local = name
	o.name.Space = namespace
	o.module = modulename
	o.Desc = desc
	o.Ref = ref
	o.units = units
	o.onEnter = onenter
	o.priv = priv
	o.local = local
	o.secret = secret
	o.repeatable = repeatable
	o.mandatory = mandatory
	o.passOpcArgs = passOpcArgs
	o.typ = typ
	o.config = false
	o.status = status
	o.arguments = make([]string, 1)
	o.addArguments(children)
	if err := o.addChildren(children); err != nil {
		return nil, err
	}

	if repeatable {
		o.addParentToChildren()
	}
	return o, nil
}
func (n *opdOption) addParent(p *node) {
	n.parent = p
}

func (n *opdOption) Descendant(path []string) Node {
	return n.descendant(n, path)
}
func (n *opdOption) HasPresence() bool {
	if n.OnEnter() == "" {
		return false
	}
	_, ok := n.Type().(Empty)
	return ok
}

func (n *opdOption) HasDefault() bool {
	_, hasDefault := n.Default()
	return hasDefault
}
func (n *opdOption) Type() Type {
	return n.typ
}
func (n *opdOption) Default() (string, bool) {
	if n.Mandatory() {
		return "", false
	}
	return n.Type().Default()
}

func (n *opdOption) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		if _, ok := n.Type().(Empty); ok {
			return nil
		} else {
			return NewMissingValueError(path)
		}
	}
	h, t := p[0], p[1:]
	if _, ok := n.Type().(Empty); !ok {
		path = append(path, h)
		err := n.Type().Validate(ctx, path, h)
		if err != nil {
			if _, ok := err.(mgmterror.Formattable); ok {
				return err
			}
			err = newInvalidValueError(path, err.Error())
			return err
		}
	} else {
		// Type is Empty
		c := n.Child(p[0])
		if c == nil {
			return NewInvalidPathError(path)

		}
		path = append(path, p[0])
		return c.Validate(ctx, path, p[1:])
	}
	if len(t) == 0 {
		return nil
	}
	c := n.getOpdChild(t[0])
	if c == nil {
		return NewInvalidPathError(append(path, t[0]))
	}
	path = append(path, t[0])
	return c.Validate(ctx, path, t[1:])
}

func (n *opdOption) DefaultChildNames() []string {
	return nil
}

func (n *opdOption) DefaultChild(name string) Node {
	return nil
}

func (n *opdOption) Child(name string) Node {
	if _, ok := n.Type().(Empty); ok {
		return n.getOpdChild(name)
	}
	return &opdOptionValue{n.node, n}
}
func (n *opdOption) Children() []Node {
	return genChildList(n.children)
}

func (n *opdOption) OpdChildren() []Node {
	if len(n.children) > 0 {
		return genChildList(n.children)
	}

	if n.parent != nil {
		return genChildList(n.parent.children)
	}

	return genChildList(n.children)
}

type opdOptionValue struct {
	*node
	opdOption *opdOption
}

type OpdOptionValue interface {
	Node
	IsOpdOptionValue()
}

func (n *opdOptionValue) IsOpdOptionValue() {}

func (n *opdOptionValue) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *opdOptionValue) Child(name string) Node {
	if len(n.children) > 0 {
		ch := n.children[name]
		if ch == nil {
			ch = n.children[n.arguments[0]]
		}
		return ch
	}

	return n.opdOption.getOpdChild(name)
}
func (n *opdOptionValue) Children() []Node {
	if len(n.children) > 0 {
		return genChildList(n.children)
	}
	return genChildList(n.opdOption.children)

}
func (n *opdOptionValue) DefaultChildNames() []string {
	return nil
}
func (n *opdOptionValue) DefaultChild(name string) Node {
	return nil
}

func (n *opdOptionValue) Type() Type {
	return n.opdOption.Type()
}
func (n *opdOptionValue) HasPresence() bool {
	return n.opdOption.OnEnter() != ""
}

type LeafList interface {
	Node
	Limit() Limit
	isLeafList()
}

type leafList struct {
	*node
	def       string
	units     string
	limit     Limit
	orderedBy string
	typ       Type
}

func (l *leafList) isLeafList() {}

func NewLeafList(
	name, namespace, modulename, submodule,
	desc, ref, def, orderedby, units string,
	min, max uint,
	typ Type,
	config bool,
	status Status,
	whens []WhenContext,
	musts []MustContext,
) LeafList {
	if orderedby == "" {
		orderedby = "system"
	}

	l := &leafList{node: makenode()}
	l.name.Local = name
	l.name.Space = namespace
	l.module = modulename
	l.submodule = submodule
	l.Desc = desc
	l.Ref = ref
	l.def = def
	l.orderedBy = orderedby
	l.units = units
	l.limit.Min = min
	l.limit.Max = max
	l.typ = typ
	l.config = config
	l.status = status
	l.whenContexts = whens
	l.mustContexts = musts
	return l
}

func (l *leafList) Limit() Limit { return l.limit }

func (n *leafList) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *leafList) OrdBy() string {
	return n.orderedBy
}

func (n *leafList) DefaultChildNames() []string {
	return nil
}

func (n *leafList) DefaultChild(name string) Node {
	return nil
}

func (n *leafList) HasDefault() bool {
	return false
}
func (n *leafList) Type() Type {
	return n.typ
}
func (n *leafList) CheckCardinality(p xutils.PathType, len int) error {
	return cardinalityInRange(p, n.limit.Min, n.limit.Max, len)
}
func (n *leafList) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		if ctx.AllowIncompletePaths() {
			return nil
		}
		return NewMissingValueError(path)
	}
	h, t := p[0], p[1:]
	path = append(path, h)
	// There should be nothing after the value
	if len(t) != 0 {
		return NewPathInvalidError(path, p[1])
	}
	return n.typ.Validate(ctx, path, h)
}

func (n *leafList) Child(name string) Node {
	return &leafValue{Node: n, name: name}
}

type Choice interface {
	Node
	DefaultCase() string
	isChoice()
}

type choice struct {
	*node
	mandatory bool
	def       string
}

// Ensure that other schema types don't meet the interface
func (*choice) isChoice() {}

// Compile time check that the concrete type meets the interface
var _ Choice = (*choice)(nil)

func NewChoice(
	name, namespace, modulename, submodulename, def, desc, ref string,
	mandatory, config bool,
	status Status,
	whens []WhenContext,
	children []Node,
) (Choice, error) {
	c := &choice{node: makenode()}
	c.name.Local = name
	c.name.Space = namespace
	c.module = modulename
	c.submodule = submodulename
	c.def = def
	c.Desc = desc
	c.Ref = ref
	c.mandatory = mandatory
	c.config = config
	c.status = status
	c.whenContexts = whens

	if err := c.addChildrenWithActionChain(children,
		addToChoices(anyNode),
		includeChildrenOf(caseNode, choiceNode),
		addToChildrenExcluding(caseNode)); err != nil {
		return nil, err
	}
	return c, nil
}

func (n *choice) Child(s string) Node {
	return n.children[s]
}

func (n *choice) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *choice) HasDefault() bool {
	return n.def != ""
}
func (n *choice) DefaultChildNames() []string {
	return nil
}

func (n *choice) DefaultCase() string {
	return n.def
}

func (n *choice) Mandatory() bool {
	return n.mandatory
}

func (n *choice) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		return errors.New("choice requires argument")
	}
	path = append(path, p[0])
	c, ok := n.node.children[p[0]]
	if !ok {
		return NewInvalidPathError(path)
	}
	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}

type Case interface {
	Node
	isCase()
}

type ycase struct {
	*node
	mandatory bool
}

// Ensure that other schema types don't meet the interface
func (*ycase) isCase() {}

// Compile time check that the concrete type meets the interface
var _ Case = (*ycase)(nil)

func NewCase(
	name, namespace, modulename, submodule, desc, ref string,
	config bool,
	status Status,
	whens []WhenContext,
	children []Node,
) (Case, error) {
	c := &ycase{node: makenode()}
	c.name.Local = name
	c.name.Space = namespace
	c.module = modulename
	c.submodule = submodule
	c.Desc = desc
	c.Ref = ref
	c.config = config
	c.status = status
	c.whenContexts = whens

	if err := c.addChildrenWithActionChain(children,
		addToChoices(anyNode),
		includeChildrenOf(choiceNode, caseNode),
		addToChildrenExcluding(choiceNode)); err != nil {
		return nil, err
	}
	return c, nil
}

func (n *ycase) Child(s string) Node {
	return n.children[s]
}

func (n *ycase) Descendant(path []string) Node {
	return n.descendant(n, path)
}

func (n *ycase) HasDefault() bool {
	return false
}
func (n *ycase) DefaultChildNames() []string {
	return []string{}
}

func (n *ycase) Validate(ctx ValidateCtx, path []string, p []string) error {
	if len(p) == 0 {
		return errors.New("choice requires argument")
	}
	path = append(path, p[0])
	c, ok := n.node.children[p[0]]
	if !ok {
		return NewInvalidPathError(path)
	}
	path = append(path, p[0])
	return c.Validate(ctx, path, p[1:])
}
