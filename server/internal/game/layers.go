package game

import "sync"

// layers.go is the S16 continuous-effect engine skeleton (CR 613).
//
// Sub-PR 1 ships the types + the public recompute entry point as a
// no-op pass — there are no static abilities in the catalog yet, so
// every "recompute" just advances LastResolvedVersion to the current
// LayerVersion. The engine has shape; nothing is observable until
// sub-PR 3 (`Spec.Static` + Glorious Anthem) gives the recompute
// actual work.
//
// Layer ordering follows CR 613:
//   1. Copy effects (Clone) — deferred to S16.5.
//   2. Control-changing effects (Mind Control) — deferred to S17.
//   3. Text-changing effects — out of scope for the layer foundation.
//   4. Type-changing effects (Mycosynth Lattice) — sub-PR 4.
//   5. Color-changing effects — engine stub, no in-scope card.
//   6. Ability-changing effects (Lord of Atlantis grants) — sub-PR 4.
//   7. Power / toughness — sub-layers 7a (CDA, Tarmogoyf), 7b (set),
//      7c (modify, Glorious Anthem), 7d (counters, delegates to
//      CurrentPower/CurrentToughness), 7e (switch).
//
// Dependency detection (CR 613.8) is INTENTIONALLY skipped — pure
// timestamp ordering covers ~95% of real cards. ADR 0012 documents
// the trade-off + the S16.5 hand-off when a real card surfaces.

// Layer is one of CR 613's seven continuous-effect application
// stages. Layer7PT carries a SubLayer; the others ignore it.
type Layer int

const (
	Layer1Copy    Layer = 1
	Layer2Control Layer = 2
	Layer3Text    Layer = 3
	Layer4Type    Layer = 4
	Layer5Color   Layer = 5
	Layer6Ability Layer = 6
	Layer7PT      Layer = 7
)

// SubLayer scopes Layer7PT into CR 613.4's a–e fan-out:
//
//	7a — characteristic-defining abilities setting P/T (Tarmogoyf)
//	7b — effects setting P/T to specific values
//	7c — effects modifying P/T (Glorious Anthem +1/+1)
//	7d — counter modifications (+1/+1, -1/-1)
//	7e — switch effects (P/T swap)
//
// Sub-layer is ignored for layers 1-6.
type SubLayer int

const (
	SubLayer7A_CDA SubLayer = iota
	SubLayer7B_Set
	SubLayer7C_Modify
	SubLayer7D_Counters
	SubLayer7E_Switch
)

// ContinuousEffect is one continuous effect contributed by a static
// ability on a battlefield permanent. Sub-PR 3 introduces the first
// concrete implementation when `Spec.Static` lands; sub-PR 1 ships
// the interface so dependent code (layers.go, view.go, the engine
// loop) can compile against a stable shape.
//
// Layer + Timestamp drive the application order: lower layer first,
// ties broken by earlier timestamp. AppliesTo is evaluated per
// candidate permanent; Apply mutates the candidate's Characteristic
// in place. Apply receives the source card so "creatures you
// control" predicates can read source.Controller.
type ContinuousEffect interface {
	Layer() (Layer, SubLayer)
	Timestamp() int64
	AppliesTo(target *Card, g *Game) bool
	Apply(c *Characteristic, target *Card, g *Game)
}

// recomputeMu serializes layer-engine recomputes against each other.
// The Game's read lock is held by snapshot callers, so the recompute
// can't promote to a write lock — instead it serialises through this
// dedicated mutex and double-checks the version after acquire to
// avoid duplicate work. Sub-PR 1 ships the mutex but the recompute
// is a no-op; sub-PR 3 starts populating Card.effective behind it.
//
// Held only for the duration of a single recompute pass; never held
// across a snapshot or mutation.
type recomputeState struct {
	mu sync.Mutex
}

// RecomputeLayersIfStaleLocked is the public entry point the
// snapshot path calls before serialising views. Fast-path no-op when
// the version counters match. Safe to call under the game's read
// lock — version reads + writes are atomic; the recompute mutex
// serialises the body against duplicate work.
//
// Sub-PR 1 ships the no-op pass: increments the recompute counter
// (so the fast-path test can assert it), then advances
// lastResolvedVersion to the current layerVersion. Sub-PR 3 fills in
// the actual layer-application body.
func (g *Game) RecomputeLayersIfStaleLocked() {
	if g.layerVersion.Load() == g.lastResolvedVersion.Load() {
		return
	}
	g.recompute.mu.Lock()
	defer g.recompute.mu.Unlock()
	// Double-check after acquiring the recompute mutex — another
	// caller may have just resolved while we waited.
	if g.layerVersion.Load() == g.lastResolvedVersion.Load() {
		return
	}
	g.recomputeLayersLocked()
}

// recomputeLayersLocked is the unconditional recompute pass. Always
// runs end-to-end; advances lastResolvedVersion on completion.
// Sub-PR 1 is a no-op stub: there are no static abilities in the
// catalog, so the printed characteristics ARE the effective
// characteristics — Card.Effective() returns printedCharacteristic
// verbatim.
//
// Sub-PR 3 wires the body:
//  1. For each battlefield card, reset effective = printed.
//  2. Collect active continuous effects from every static ability
//     on the battlefield via the CatalogStaticAbilities hook.
//  3. For layer in 1..6 and sub-layer in 7a..7e: filter effects to
//     this (sub-)layer, sort by timestamp, apply each.
//  4. Layer 7d delegates to Card.CurrentPower / CurrentToughness
//     for counter math (the existing S13.2 code path).
//
// Caller must hold the recompute mutex (RecomputeLayersIfStaleLocked
// acquires it). The recompute counter bumps on every call so the
// fast-path test can assert it ran exactly once.
func (g *Game) recomputeLayersLocked() {
	g.recomputeCount.Add(1)
	g.lastResolvedVersion.Store(g.layerVersion.Load())
}

// BumpLayerVersionForTest bumps the layer-engine invalidation
// counter from outside the package so tests can prove the
// fast-path / recompute behaviour without wiring real listener
// events. Production callers use the listener pipeline (sub-PR 2).
//
// Atomic — safe to call under the game's read or write lock.
func (g *Game) BumpLayerVersionForTest() {
	g.layerVersion.Add(1)
}

// LayerRecomputeCountForTest returns the number of times the layer
// engine has run a full recompute pass since game start. Used by
// the fast-path regression test (exit criterion #10).
func (g *Game) LayerRecomputeCountForTest() uint64 {
	return g.recomputeCount.Load()
}
