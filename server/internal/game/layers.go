package game

import (
	"sort"
	"sync"

	"github.com/google/uuid"
)

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

// StaticAbility is the declarative shape catalog cards use to
// contribute continuous effects to the layer engine. Lives in the
// `game` package so card files (`server/internal/cards/effects/*`)
// can reference it via the existing `game` import without going
// through a separate adapter type — the package already exposes
// Card / Game / Characteristic / the Layer enums.
//
// Each `StaticAbility` becomes one `ContinuousEffect` per source
// card on the battlefield (via the CatalogStaticAbilities hook),
// bound to the source's `EnteredBattlefieldAt` timestamp. AppliesTo
// is evaluated per battlefield card on every recompute pass; Apply
// mutates the candidate's Characteristic in place.
//
// Apply receives the source card so predicates like "creatures you
// control" can read `source.Controller`, and self-exclusion
// predicates ("other Merfolk") can compare instance IDs.
//
// Added in S16 sub-PR 3.
type StaticAbility struct {
	Layer     Layer
	SubLayer  SubLayer
	AppliesTo func(target *Card, g *Game, source *Card) bool
	Apply     func(c *Characteristic, target *Card, g *Game, source *Card)
}

// staticContinuousEffect is the internal `ContinuousEffect` adapter
// for a `StaticAbility` bound to a source card. The recompute pass
// builds one of these per (battlefield card, declared ability) pair
// and feeds them through the layer-ordering loop.
type staticContinuousEffect struct {
	ability   StaticAbility
	source    *Card
	timestamp int64
}

func (e staticContinuousEffect) Layer() (Layer, SubLayer) {
	return e.ability.Layer, e.ability.SubLayer
}

func (e staticContinuousEffect) Timestamp() int64 { return e.timestamp }

func (e staticContinuousEffect) AppliesTo(target *Card, g *Game) bool {
	if e.ability.AppliesTo == nil {
		return false
	}
	return e.ability.AppliesTo(target, g, e.source)
}

func (e staticContinuousEffect) Apply(c *Characteristic, target *Card, g *Game) {
	if e.ability.Apply == nil {
		return
	}
	e.ability.Apply(c, target, g, e.source)
}

// activeStaticAbilitiesLocked collects every continuous effect in
// play from its two sources and returns them as
// `ContinuousEffect`s bound to source pointers + timestamps:
//
//  1. Battlefield permanents — walks g.Battlefield and looks each
//     card's catalog static abilities up via the
//     CatalogStaticAbilities hook. These live exactly as long as
//     the source permanent does (CR 113.6).
//  2. Turn-scoped statics — the S32 floating "until end of turn"
//     registry (turn_scoped_statics.go), which has no battlefield
//     source and expires on a clock instead (CR 514.2).
//
// Caller must hold either g.mu (write) or g.recompute.mu (the
// recompute serialisation mutex used during snapshot).
//
// The returned source pointers reference into g.Battlefield.Cards
// — safe for the duration of the recompute pass that holds the
// recompute mutex; not safe to retain across mutations.
func (g *Game) activeStaticAbilitiesLocked() []ContinuousEffect {
	// S32: floating "until end of turn" effects first. They are
	// gathered unconditionally — they outlive their source card, so
	// neither an empty battlefield nor a missing catalog hook can
	// switch them off. Order within this slice is irrelevant: the
	// per-bucket sort in applyLayerLocked re-orders everything by
	// timestamp (CR 613.7) before applying.
	out := g.turnScopedContinuousEffectsLocked()
	if g.Battlefield == nil || CatalogStaticAbilities == nil {
		return out
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		abilities := CatalogStaticAbilities(CatalogKey(*src))
		if len(abilities) == 0 {
			continue
		}
		for _, ab := range abilities {
			out = append(out, staticContinuousEffect{
				ability:   ab,
				source:    src,
				timestamp: src.EnteredBattlefieldAt,
			})
		}
	}
	return out
}

// layerOrder lists the (Layer, SubLayer) buckets in the order
// CR 613 applies them. Layers 1-6 each contribute one bucket;
// layer 7 fans out into 7a..7e. Effects are filtered to the bucket
// they belong to, sorted by timestamp, then applied in turn.
var layerOrder = []struct {
	Layer    Layer
	SubLayer SubLayer
	// has7Sub is true for 7a..7e buckets so the bucket-filter
	// matches on both Layer and SubLayer; false elsewhere.
	has7Sub bool
}{
	{Layer: Layer1Copy},
	{Layer: Layer2Control},
	{Layer: Layer3Text},
	{Layer: Layer4Type},
	{Layer: Layer5Color},
	{Layer: Layer6Ability},
	{Layer: Layer7PT, SubLayer: SubLayer7A_CDA, has7Sub: true},
	{Layer: Layer7PT, SubLayer: SubLayer7B_Set, has7Sub: true},
	{Layer: Layer7PT, SubLayer: SubLayer7C_Modify, has7Sub: true},
	{Layer: Layer7PT, SubLayer: SubLayer7D_Counters, has7Sub: true},
	{Layer: Layer7PT, SubLayer: SubLayer7E_Switch, has7Sub: true},
}

// inBucket reports whether the effect belongs to the given layer
// bucket. Layer matches always; sub-layer matches only when the
// bucket is one of the 7a..7e fan-outs (layers 1-6 ignore
// sub-layer entirely).
func inBucket(eff ContinuousEffect, l Layer, sub SubLayer, has7Sub bool) bool {
	effLayer, effSub := eff.Layer()
	if effLayer != l {
		return false
	}
	if !has7Sub {
		return true
	}
	return effSub == sub
}

// applyLayerLocked filters effects to one (sub-)layer bucket,
// sorts by timestamp, and applies each effect's Apply across every
// battlefield card it AppliesTo. Stable sort so same-timestamp
// effects fall in their gather order — matters for test
// determinism and for catalog cards with multiple statics from a
// single source.
func (g *Game) applyLayerLocked(effects []ContinuousEffect, l Layer, sub SubLayer, has7Sub bool) {
	bucket := make([]ContinuousEffect, 0, len(effects))
	for _, eff := range effects {
		if inBucket(eff, l, sub, has7Sub) {
			bucket = append(bucket, eff)
		}
	}
	if len(bucket) == 0 {
		return
	}
	sort.SliceStable(bucket, func(i, j int) bool {
		return bucket[i].Timestamp() < bucket[j].Timestamp()
	})
	for _, eff := range bucket {
		for i := range g.Battlefield.Cards {
			target := &g.Battlefield.Cards[i]
			if !eff.AppliesTo(target, g) {
				continue
			}
			eff.Apply(target.effective, target, g)
		}
	}
}

// findCardOnBattlefield returns the index of cardID in the
// battlefield zone, or -1 if not present. Used by snapshot-time
// helpers that need to walk per-card without re-iterating.
func findCardOnBattlefield(g *Game, cardID uuid.UUID) int {
	if g.Battlefield == nil {
		return -1
	}
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == cardID {
			return i
		}
	}
	return -1
}

// DistinctCardTypesInAllGraveyards counts the unique card types
// across every player's graveyard. Drives Tarmogoyf-style CDAs
// (S16 sub-PR 5) — Tarmogoyf's power is the number of distinct
// card types among cards in all graveyards.
//
// Reads from each card's printedCharacteristic (the layer engine
// hasn't recomputed graveyard cards — they're not on battlefield),
// so the count reflects the printed types regardless of any
// continuous effect. Card-type set is whatever shows up in the
// "Types" slice — Artifact, Battle, Creature, Enchantment, Instant,
// Land, Planeswalker, Sorcery, Tribal/Kindred plus any future
// additions; "Token" is treated as a supertype here so a token's
// "Creature" type still counts.
//
// Caller may hold either lock (read or write). Walks every player
// graveyard once; cheap relative to a full game state.
func DistinctCardTypesInAllGraveyards(g *Game) int {
	if g == nil {
		return 0
	}
	seen := map[string]struct{}{}
	for _, p := range g.Seats {
		if p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			pc := c.printedCharacteristic()
			for _, t := range pc.Types {
				seen[t] = struct{}{}
			}
		}
	}
	return len(seen)
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
//
// S16 sub-PR 3 body:
//  1. For each battlefield card, reset effective = printed.
//  2. Collect active continuous effects from every static ability
//     on the battlefield via the CatalogStaticAbilities hook.
//  3. For each (Layer, SubLayer) bucket in CR 613 order: filter
//     effects to that bucket, sort by source timestamp ascending,
//     apply each effect across every battlefield card it AppliesTo.
//
// Caller must hold the recompute mutex (RecomputeLayersIfStaleLocked
// acquires it). The recompute counter bumps on every call so the
// fast-path test can assert it ran exactly once.
func (g *Game) recomputeLayersLocked() {
	g.recomputeCount.Add(1)
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			printed := g.Battlefield.Cards[i].printedCharacteristic()
			g.Battlefield.Cards[i].effective = &printed
		}
	}
	effects := g.activeStaticAbilitiesLocked()
	for _, b := range layerOrder {
		g.applyLayerLocked(effects, b.Layer, b.SubLayer, b.has7Sub)
	}
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

// CommanderIdentityForTest is the exported wrapper around the
// unexported commanderIdentityFor for cross-package tests
// (effects/tarmogoyf_test.go's regression guard for the S15→S16
// proxy replacement). Reads the commander's effective colors via
// the layer engine, falling back to the printed-cost proxy when
// effective is empty.
func CommanderIdentityForTest(p *Player) []string {
	return commanderIdentityFor(p)
}
