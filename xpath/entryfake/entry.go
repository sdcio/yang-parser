// Package entryfake provides a minimal, in-memory implementation of
// xpath.Entry for use in tests. It mirrors just enough of how a real
// schema-backed Entry (e.g. data-server's tree adapter) represents leaf,
// container and leaf-list nodes to exercise the xpath engine's Entry-based
// evaluation path (as opposed to the classic xutils.XpathNode-based path
// exercised by the xpathtest package).
//
// In particular, a leaf-list here has no per-item node identity: GetValue()
// returns a single DatumSliceDatum bundling all of its values, exactly as
// data-server's yangParserEntryAdapter.valueToDatum does for a
// sdcpb.TypedValue_LeaflistVal.
package entryfake

import (
	"context"
	"fmt"

	sdcpb "github.com/sdcio/sdc-protos/sdcpb"
	"github.com/sdcio/yang-parser/xpath"
)

// Node is a small, mutable in-memory tree node used to build fixtures.
type Node struct {
	name     string
	children map[string]*Node
	// leafList holds this node's values when it represents a leaf-list.
	// A nil slice means "not a leaf-list".
	leafList []string
	// leaf holds this node's scalar value when it represents a leaf.
	// Only meaningful when isLeaf is true, so an empty-string leaf value
	// can be distinguished from "not a leaf".
	leaf   string
	isLeaf bool
}

// NewContainer creates a container node (e.g. a list instance or a plain
// container) with the given children.
func NewContainer(name string, children ...*Node) *Node {
	m := make(map[string]*Node, len(children))
	for _, c := range children {
		m[c.name] = c
	}
	return &Node{name: name, children: m}
}

// NewLeaf creates a scalar leaf node.
func NewLeaf(name, value string) *Node {
	return &Node{name: name, leaf: value, isLeaf: true}
}

// NewLeafList creates a leaf-list node carrying all of its values at once,
// matching how data-server's adapter models leaf-lists: one Entry, one
// TypedValue_LeaflistVal, no per-item node identity.
func NewLeafList(name string, values ...string) *Node {
	return &Node{name: name, leafList: values}
}

// entry adapts a Node into an xpath.Entry.
type entry struct {
	node *Node
}

// NewEntry wraps a Node as the root xpath.Entry for a machine run.
func NewEntry(root *Node) xpath.Entry {
	return &entry{node: root}
}

func (e *entry) GetValue() (xpath.Datum, error) {
	switch {
	case e.node.leafList != nil:
		datums := make([]xpath.Datum, 0, len(e.node.leafList))
		for _, v := range e.node.leafList {
			datums = append(datums, xpath.NewLiteralDatum(v))
		}
		return xpath.NewDatumSliceDatum(datums), nil
	case e.node.isLeaf:
		return xpath.NewLiteralDatum(e.node.leaf), nil
	default:
		// Plain container: mirrors yangParserEntryAdapter's non-list
		// container branch.
		return xpath.NewBoolDatum(true), nil
	}
}

func (e *entry) Navigate(path *sdcpb.Path) (xpath.Entry, error) {
	cur := e.node
	for _, elem := range path.GetElem() {
		child, ok := cur.children[elem.GetName()]
		if !ok {
			return nil, fmt.Errorf(
				"entryfake: no child %q under %q", elem.GetName(), cur.name)
		}
		cur = child
	}
	return &entry{node: cur}, nil
}

func (e *entry) Copy() xpath.Entry {
	return &entry{node: e.node}
}

func (e *entry) FollowLeafRef() (xpath.Entry, error) {
	return nil, fmt.Errorf("entryfake: FollowLeafRef not supported")
}

func (e *entry) GetSdcpbPath() *sdcpb.Path {
	return nil
}

func (e *entry) BreadthSearch(
	_ context.Context, _ *sdcpb.Path,
) ([]xpath.Entry, error) {
	return nil, fmt.Errorf("entryfake: BreadthSearch not supported")
}
