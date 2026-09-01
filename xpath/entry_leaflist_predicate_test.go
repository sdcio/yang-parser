package xpath_test

import (
	"context"
	"testing"

	"github.com/sdcio/yang-parser/xpath"
	"github.com/sdcio/yang-parser/xpath/entryfake"
	"github.com/sdcio/yang-parser/xpath/grammars/expr"
)

// Regression test for the SONiC PORT_LIST/adv_speeds must-statement crash:
// count(leaflist[text()='x']) used to fail with "Fn 'count' takes NODESET,
// not BOOL as arg 0" because Eq() collapsed the leaf-list/literal
// comparison straight to a Bool before count() ever saw it.
//
// This exercises the Entry-based evaluation path (xpath.Entry, driven by
// NewCtxFromCurrent) -- the path data-server's tree adapter actually uses,
// and where leaf-lists have no per-item node identity, only a single
// DatumSliceDatum bundling every value. This is deliberately not modeled
// via the xutils.XpathNode / xpathtest.CreateTree path used elsewhere in
// this package's tests, since that path gives each leaf-list value its own
// node and so never exercises Eq()'s DatumSlice-collapsing branch at all.
func runNum(t *testing.T, root *entryfake.Node, exprStr string) (float64, error) {
	t.Helper()
	mach, err := expr.NewExprMachine(exprStr, nil)
	if err != nil {
		t.Fatalf("failed to compile %q: %s", exprStr, err)
	}
	res := xpath.NewCtxFromCurrent(context.Background(), mach, entryfake.NewEntry(root)).
		EnableValidation().Run()
	return res.GetNumResult()
}

func runBool(t *testing.T, root *entryfake.Node, exprStr string) (bool, error) {
	t.Helper()
	mach, err := expr.NewExprMachine(exprStr, nil)
	if err != nil {
		t.Fatalf("failed to compile %q: %s", exprStr, err)
	}
	res := xpath.NewCtxFromCurrent(context.Background(), mach, entryfake.NewEntry(root)).
		EnableValidation().Run()
	return res.GetBoolResult()
}

func advSpeedsTree() *entryfake.Node {
	return entryfake.NewContainer("port",
		entryfake.NewLeafList("adv_speeds", "10000", "100000", "all"))
}

func TestCountLeafListPredicate_ExactlyOneMatch(t *testing.T) {
	got, err := runNum(t, advSpeedsTree(), "count(adv_speeds[text()='all'])")
	if err != nil {
		t.Fatalf("count(adv_speeds[text()='all']) returned error: %s", err)
	}
	if got != 1 {
		t.Errorf("count(adv_speeds[text()='all']) = %v, want 1", got)
	}
}

func TestCountLeafListPredicate_NoMatch(t *testing.T) {
	got, err := runNum(t, advSpeedsTree(), "count(adv_speeds[text()='missing'])")
	if err != nil {
		t.Fatalf("count(adv_speeds[text()='missing']) returned error: %s", err)
	}
	if got != 0 {
		t.Errorf("count(adv_speeds[text()='missing']) = %v, want 0", got)
	}
}

func TestCountLeafListPredicate_NotEqual(t *testing.T) {
	// Ne() gets the same DatumSlice generalization as Eq(): count() over
	// a leaf-list[.!=literal] predicate should return the number of
	// non-matching items, not crash the way count(leaflist[text()=...])
	// used to.
	got, err := runNum(t, advSpeedsTree(), "count(adv_speeds[text()!='all'])")
	if err != nil {
		t.Fatalf("count(adv_speeds[text()!='all']) returned error: %s", err)
	}
	if got != 2 {
		t.Errorf("count(adv_speeds[text()!='all']) = %v, want 2", got)
	}
}

func TestCountLeafListPredicate_Relational(t *testing.T) {
	// mtu_options: 1500, 9000, 9216 -- exercise all four relational
	// operators (Gt is covered above the other three here) since Eq/Ne
	// and relational operators are wired through separate dispatch
	// functions (popCompareEqualityAndPush vs popCompareRelationalAndPush)
	// that both route through compareDatumSlicesAndPush.
	tests := []struct {
		exprStr string
		want    float64
	}{
		{"count(mtu_options[text() > 2000])", 2},  // 9000, 9216
		{"count(mtu_options[text() >= 9000])", 2}, // 9000, 9216
		{"count(mtu_options[text() < 9000])", 1},  // 1500
		{"count(mtu_options[text() <= 9000])", 2}, // 1500, 9000
	}

	for _, tt := range tests {
		tree := entryfake.NewContainer("port",
			entryfake.NewLeafList("mtu_options", "1500", "9000", "9216"))

		got, err := runNum(t, tree, tt.exprStr)
		if err != nil {
			t.Fatalf("%s returned error: %s", tt.exprStr, err)
		}
		if got != tt.want {
			t.Errorf("%s = %v, want %v", tt.exprStr, got, tt.want)
		}
	}
}

// Regression test for the SONiC PORT_LIST/adv_speeds must-statement in its
// actual field form:
//
//	count(adv_speeds[text()='all']) = 0 or count(adv_speeds) = 1
//
// This is a distinct bug from the one at the top of this file: even once
// count()/Eq() correctly handle a single leaf-list predicate, evaluating
// TWO independent path references to the same leaf-list within one
// expression used to corrupt the second one. EvalLocPath's "the predicate
// already produced the value, skip re-resolving the path" fast path
// (ctx.previousPredicateRequiresELP) left ctx.actualPathStack's frame for
// the abandoned "adv_speeds[...]" path unpopped, so the next bare
// "adv_speeds" reference appended onto it instead of starting fresh,
// producing a bogus two-element path that failed to navigate -- silently
// leaving count()'s argument unresolved and falling back to whatever
// Bool happened to be left on the stack from the first comparison.
func TestCountLeafListPredicate_CombinedWithBarePathReference(t *testing.T) {
	tests := []struct {
		name string
		tree *entryfake.Node
		want bool
	}{
		{
			name: "only 'all' present",
			tree: entryfake.NewContainer("port", entryfake.NewLeafList("adv_speeds", "all")),
			want: true,
		},
		{
			name: "'all' plus another value",
			tree: entryfake.NewContainer("port", entryfake.NewLeafList("adv_speeds", "all", "10000")),
			want: false,
		},
		{
			name: "single non-'all' value",
			tree: entryfake.NewContainer("port", entryfake.NewLeafList("adv_speeds", "10000")),
			want: true,
		},
		{
			name: "unset leaf-list",
			tree: entryfake.NewContainer("port", entryfake.NewLeafList("adv_speeds")),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runBool(t, tt.tree, "count(adv_speeds[text()='all']) = 0 or count(adv_speeds) = 1")
			if err != nil {
				t.Fatalf("returned error: %s", err)
			}
			if got != tt.want {
				t.Errorf("= %v, want %v", got, tt.want)
			}
		})
	}
}

// A bare, non-predicate leaf-list equality (no enclosing '[...]') must
// still behave the same existential way it did before the Eq() rewrite:
// ctx.predicateCount == 0 already forces the equality-check branch
// regardless of ctx.isLeafListFilter, so this is a pure backward-
// compatibility check on the isDS-outside-a-predicate case.
func TestLeafListEquality_BareNonPredicateUnaffected(t *testing.T) {
	tree := advSpeedsTree()

	if got, err := runBool(t, tree, "adv_speeds = 'all'"); err != nil {
		t.Fatalf("adv_speeds = 'all' returned error: %s", err)
	} else if !got {
		t.Errorf("adv_speeds = 'all' = %v, want true", got)
	}

	if got, err := runBool(t, tree, "adv_speeds = 'missing'"); err != nil {
		t.Fatalf("adv_speeds = 'missing' returned error: %s", err)
	} else if got {
		t.Errorf("adv_speeds = 'missing' = %v, want false", got)
	}
}

// Boolean-context usage (the pre-existing, already-working case) must keep
// giving the same existential-match answer once Eq()/count() are fixed.
func TestLeafListPredicate_BooleanContextUnaffected(t *testing.T) {
	tree := advSpeedsTree()

	if got, err := runBool(t, tree, "adv_speeds[text()='all']"); err != nil {
		t.Fatalf("adv_speeds[text()='all'] returned error: %s", err)
	} else if !got {
		t.Errorf("adv_speeds[text()='all'] = %v, want true", got)
	}

	if got, err := runBool(t, tree, "adv_speeds[text()='missing']"); err != nil {
		t.Fatalf("adv_speeds[text()='missing'] returned error: %s", err)
	} else if got {
		t.Errorf("adv_speeds[text()='missing'] = %v, want false", got)
	}
}
