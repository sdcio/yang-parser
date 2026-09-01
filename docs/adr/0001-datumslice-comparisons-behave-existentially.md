# Comparisons against a leaf-list's DatumSlice keep per-item filter semantics instead of collapsing to Bool

`count(leaflist[text()='x'])` crashed ("Fn 'count' takes NODESET, not BOOL") because `Eq()` collapsed a leaf-list's bundled values (a `DatumSliceDatum`) straight to a `Bool` whenever it appeared as an operand, before `count()` ever saw the result.

## Why

Leaf-lists have no per-item node identity in this engine's `Entry` abstraction (unlike list instances, which get one `Entry`/`XpathNode` per instance) — `GetValue()` returns a single `DatumSliceDatum` bundling every value. The engine also has no generic "filter this nodeset by an arbitrary predicate" mechanism; the only existing predicate machinery is the list-key-equality shortcut (`predicatePathElemStack`/`XListKeyMatches`), which doesn't apply to leaf-lists. `Eq()`'s `DatumSlice` branch is therefore the *only* implementation of leaf-list-predicate filtering in the engine, and it existed purely to produce a boolean.

## Decision

`Eq()`, `Ne()`, `Lt()`, `Le()`, `Gt()`, `Ge()` now all treat a `DatumSlice` operand the same way: keep every item for which the comparison holds against the other operand, and push the filtered subset as a new `DatumSliceDatum` (see `compareDatumSlicesAndPush` in `context.go`), instead of collapsing to a `Bool`. `count()`'s declared argument type was widened to accept a `DatumSlice` as well as a real `Nodeset` (its function body already had an unreachable branch for this).

Boolean-context callers are unaffected: `DatumSliceDatum.Boolean()` already returns "non-empty", which is exactly the existential match XPath expects when a node-set (or equivalent) is reduced to a boolean — so `adv_speeds[text()='all']` used directly in a `must` statement keeps giving the same answer it always did.

## Considered options

**Rejected — give leaf-list items real per-item `Entry`/`XpathNode` identity in the calling adapter (e.g. data-server's tree), so leaf-list predicates go through the same machinery as list predicates.** This doesn't actually solve the problem: the engine's predicate machinery is a list-key-equality shortcut, not a general per-node predicate evaluator, so a real per-item node still couldn't be filtered by an arbitrary expression like `text()='x'`. It would also be a much larger, cross-cutting change to the calling adapter's tree model, for no additional correctness.

**Rejected — teach only `count()` to tolerate the collapsed `Bool`.** Narrower, but leaves every other `Nodeset`/`DatumSlice`-aware consumer (and `Ne`/relational operators specifically) with the same gap; the next caller to hit it would need the identical fix again.

## Consequences

This changes the result *type* of `Eq`/`Ne`/relational comparisons whenever a `DatumSlice` operand is involved (previously always `Bool`, now a `DatumSliceDatum`) — call sites that only ever reduce the result to a boolean are unaffected, but any future code inspecting the raw `Datum` type after such a comparison needs to expect a `DatumSliceDatum`, not a `Bool`.
