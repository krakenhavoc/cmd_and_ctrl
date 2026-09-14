package game

import (
	"sort"

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
//   2. Control-changing effects (Mind Control) — S24. The output
//      lands in Characteristic.Controller and is materialised back
//      onto Card.Controller at the end of the pass; see
//      materialiseControlLocked for why that, and not an
//      EffectiveController() accessor plus a 385-site sweep.
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

	// RemovesAbilities reports whether this is a CR 613.1f
	// ability-removing effect. True means the engine empties the
	// target's ability list and stamps
	// Characteristic.AbilitiesRemoved BEFORE calling Apply, so an
	// effect that removes and grants in one breath ("has
	// indestructible, and it loses all OTHER abilities") only has to
	// append the part it keeps.
	//
	// It is a declaration rather than something Apply does for
	// itself because the removal is load-bearing outside layer 6:
	// the recompute has to know which permanents were silenced so it
	// can drop THEIR contributions from every other layer, and it
	// cannot learn that by watching an opaque closure mutate a
	// struct.
	RemovesAbilities() bool
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

	// RemovesAbilities declares this a CR 613.1f ability-removing
	// effect — Darksteel Mutation's "loses all other abilities",
	// Kenrith's Transformation's "loses all abilities", Song of the
	// Dryads' CR 305.7 land conversion. Only meaningful alongside
	// Layer6Ability.
	//
	// The engine empties Characteristic.Abilities and sets
	// Characteristic.AbilitiesRemoved on every target this applies
	// to, before Apply runs. Declare it rather than clearing the
	// slice by hand: clearing the slice removes KEYWORDS and leaves
	// every catalogued activated, triggered, mana, static and
	// replacement ability working, because those are read through
	// the Catalog* hooks at use time. Build one with
	// effects.LoseAllAbilities.
	RemovesAbilities bool
}

// staticContinuousEffect is the internal `ContinuousEffect` adapter
// for a `StaticAbility` bound to a source card. The recompute pass
// builds one of these per (battlefield card, declared ability) pair
// and feeds them through the layer-ordering loop.
type staticContinuousEffect struct {
	ability   StaticAbility
	source    *Card
	timestamp int64

	// live marks a source that is a permanent on the battlefield
	// RIGHT NOW, as opposed to a turn-scoped static's last-known-
	// information copy. Only a live source can be silenced by a
	// CR 613.1f ability removal: a "until end of turn" effect
	// outlived its source by construction (CR 611.2b) and nothing on
	// the board can take it back.
	//
	// A bool rather than a nil-effective test, because a ScopedStatic
	// captured its Source by value while the card was on the
	// battlefield — layer cache pointer and all — and that stale
	// cache would otherwise silence the effect forever.
	live bool
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

func (e staticContinuousEffect) RemovesAbilities() bool {
	return e.ability.RemovesAbilities
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
// Caller must hold g.mu in write mode.
//
// The returned source pointers reference into g.Battlefield.Cards
// — safe for the duration of the recompute pass that holds the
// write lock; not safe to retain across mutations.
//
// `silenced` names permanents a CR 613.1f ability-removing effect
// applied to on the previous pass. A silenced source contributes
// NOTHING — not its layer-2 control change, not its layer-4 type
// change, not its anthem — because it has no abilities left to
// generate a continuous effect from (CR 613.1f + CR 113.3). Nil on
// the discovery pass, which is how the set is learned in the first
// place; see recomputeLayersLocked.
//
// Turn-scoped statics are never silenced. They have no battlefield
// source to take abilities away from: the effect outlived its source
// by construction (CR 611.2b), so nothing on the board can switch it
// off.
func (g *Game) activeStaticAbilitiesLocked(silenced map[uuid.UUID]bool) []ContinuousEffect {
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
		// `silenced`, not CatalogAbilityKey: the pass resets every
		// effective characteristic to printed before it gathers, so
		// there is nothing for the accessor to read yet. The set
		// carried in from the previous pass is the only thing that
		// knows.
		if silenced[src.InstanceID] {
			continue
		}
		abilities := CatalogStaticAbilities(CatalogKey(*src))
		if len(abilities) == 0 {
			continue
		}
		// CR 613.7d: an Equipment's or Aura's continuous effect
		// takes a NEW timestamp when it becomes attached, not the
		// one it got when it entered the battlefield. Unobservable
		// for every card in the S24 catalog (layer 6 grants and 7c
		// modifies are both commutative), and right by construction
		// for the first 7b "set" that meets a 7c "modify".
		ts := src.EnteredBattlefieldAt
		if src.AttachedAt != 0 {
			ts = src.AttachedAt
		}
		for _, ab := range abilities {
			out = append(out, staticContinuousEffect{
				ability:   ab,
				source:    src,
				timestamp: ts,
				live:      true,
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
		// CR 613.1f between two ability-removing effects. An effect
		// whose SOURCE has already been silenced earlier in this
		// bucket no longer exists to apply: two Song of the Dryads,
		// each enchanting the other, resolve to "the earlier one
		// wins" rather than to a paradox, because layer 6 is walked
		// in timestamp order and the earlier removal has already
		// taken the later one's abilities away by the time we get
		// here.
		//
		// Only checkable from layer 6 onwards — before it, nothing
		// has been silenced yet. That is what the second recompute
		// pass is for.
		if src := effectSourceLocked(eff); src != nil && src.HasLostAllAbilities() {
			continue
		}
		removes := eff.RemovesAbilities()
		for i := range g.Battlefield.Cards {
			target := &g.Battlefield.Cards[i]
			if !eff.AppliesTo(target, g) {
				continue
			}
			if removes {
				// Emptied BEFORE Apply so "has indestructible, and
				// it loses all other abilities" is one effect that
				// clears and then re-grants, in the order it is
				// printed — and so a card file cannot ship the
				// removal without the engine learning about it.
				target.effective.Abilities = nil
				target.effective.AbilitiesRemoved = true
			}
			eff.Apply(target.effective, target, g)
		}
	}
}

// effectSourceLocked returns the LIVE battlefield permanent an effect
// was gathered from, or nil when it has none — a turn-scoped static,
// or any future effect shape that is not sourced from a permanent.
// Only a live source can have had its abilities removed.
func effectSourceLocked(eff ContinuousEffect) *Card {
	sce, ok := eff.(staticContinuousEffect)
	if !ok || !sce.live {
		return nil
	}
	return sce.source
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

// RecomputeLayersIfStaleLocked is the entry point every mutator
// calls before it reads an effective characteristic, and the one
// ReadSnapshot calls on behalf of the snapshot path. Fast-path no-op
// when the version counters match.
//
// **Caller must hold the game's WRITE lock.** A recompute is a
// write, not a read: it reassigns Card.effective on every
// battlefield card and — since the S24 layer-2 control change —
// Card.Controller, Card.SummonedThisTurn, Card.AttackingTarget and
// Card.BlockingTarget on any permanent whose controller just moved.
//
// This used to advertise itself as read-lock-safe on the strength of
// a dedicated recompute mutex. That mutex serialised recomputes
// against EACH OTHER and did nothing about a reader holding only the
// shared read lock, so every concurrent read-lock holder raced the
// pass — `Game.AutoTapForCostExcluding` (lobby /autotap) and
// `Game.ControllerOfCard` (actions authorization) both read
// Card.Controller off the battlefield from their own goroutines. A
// torn 16-byte uuid.UUID there silently attributes a permanent to
// the wrong seat. The write lock is now the serialiser, so the
// separate mutex is gone; `ReadSnapshot` upgrades rather than
// recomputing in place. Pinned by layers_concurrency_test.go.
//
// Advances lastResolvedVersion to the current layerVersion on
// completion, and bumps the recompute counter so the fast-path test
// can assert it ran exactly once.
func (g *Game) RecomputeLayersIfStaleLocked() {
	if g.layerVersion.Load() == g.lastResolvedVersion.Load() {
		return
	}
	g.recomputeLayersLocked()
}

// recomputeLayersLocked is the unconditional recompute. Always runs
// end-to-end; advances lastResolvedVersion on completion.
//
// One layer pass (layerPassLocked) is:
//  1. For each battlefield card, reset effective = printed.
//  2. Collect active continuous effects from every static ability
//     on the battlefield via the CatalogStaticAbilities hook.
//  3. For each (Layer, SubLayer) bucket in CR 613 order: filter
//     effects to that bucket, sort by source timestamp ascending,
//     apply each effect across every battlefield card it AppliesTo.
//
// S24 runs that pass more than once when a CR 613.1f ability-removing
// effect is on the board; see the body.
//
// Caller must hold g.mu in write mode — the pass writes
// Card.effective on every battlefield card and, via
// materialiseControlLocked, Card.Controller on any permanent layer 2
// moved. The recompute counter bumps on every call so the fast-path
// test can assert it ran exactly once.
func (g *Game) recomputeLayersLocked() {
	g.recomputeCount.Add(1)
	// S24: the ability-removal fixed point. The first pass is the
	// discovery pass — it runs with nothing silenced and learns
	// which permanents a layer-6 ability-removing effect applied to.
	// If that set is empty, which it is on every board that has
	// never seen a Darksteel Mutation, we are done and this has cost
	// one extra battlefield walk.
	//
	// If it is not empty the pass runs again with the set honoured
	// at GATHER time, so a silenced permanent stops contributing to
	// layers 1-5 and 7 as well as to 6 — a Mind Control that became
	// a Forest has to stop stealing the creature, and layer 2 is
	// applied long before layer 6 could have told it to.
	//
	// This is the one CR 613.8 dependency the engine resolves, and
	// it resolves it by iterating to a fixed point rather than by
	// analysing the effects (ADR 0012 declined the analysis, and
	// still does). The cap is what makes that safe: a board that has
	// not settled after maxLayerPasses keeps the last pass's answer
	// rather than spinning. Reaching it needs a cycle of ability
	// removers that the within-layer-6 timestamp skip in
	// applyLayerLocked does not already break, which no catalogued
	// card can currently build.
	var silenced map[uuid.UUID]bool
	for pass := 0; pass < maxLayerPasses; pass++ {
		next := g.layerPassLocked(silenced)
		if sameCardSet(silenced, next) {
			break
		}
		silenced = next
	}
	g.materialiseControlLocked()
	g.lastResolvedVersion.Store(g.layerVersion.Load())
}

// maxLayerPasses bounds the ability-removal fixed point. Two passes
// is the answer for every shape the catalog can build today (one to
// discover, one to apply); the extra headroom is for a future card
// whose removal changes which OTHER removal applies.
const maxLayerPasses = 4

// sameCardSet reports whether two silenced sets name the same cards.
// nil and empty are the same set — the common board, where nothing
// has lost anything.
func sameCardSet(a, b map[uuid.UUID]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for id := range a {
		if !b[id] {
			return false
		}
	}
	return true
}

// layerPassLocked runs one complete CR 613 application over the
// battlefield with `silenced` held out of the gather, and returns
// the set of permanents an ability-removing effect applied to during
// it. Caller must hold g.mu in write mode.
func (g *Game) layerPassLocked(silenced map[uuid.UUID]bool) map[uuid.UUID]bool {
	if g.Battlefield != nil {
		for i := range g.Battlefield.Cards {
			c := &g.Battlefield.Cards[i]
			// S24 layer 2: capture the control baseline on the first
			// recompute after this permanent entered. Lazy rather
			// than stamped at every write site because all ~15 sites
			// that assign Card.Controller do so as a permanent
			// ENTERS, and every entry bumps the layer version — so
			// this runs before anything can observe a layer-2 value,
			// and MoveCard clearing the field on exit is what makes
			// a re-entry re-capture.
			if c.BaseController == uuid.Nil {
				c.BaseController = c.Controller
			}
			printed := c.printedCharacteristic()
			c.effective = &printed
		}
	}
	effects := g.activeStaticAbilitiesLocked(silenced)
	for _, b := range layerOrder {
		g.applyLayerLocked(effects, b.Layer, b.SubLayer, b.has7Sub)
	}
	return g.silencedSetLocked()
}

// silencedSetLocked reads back the permanents the pass just finished
// marked as having lost all their abilities. Returns nil rather than
// an empty map for the overwhelmingly common "nothing was silenced"
// board, so sameCardSet's cheap length compare ends the loop on the
// first pass.
func (g *Game) silencedSetLocked() map[uuid.UUID]bool {
	if g.Battlefield == nil {
		return nil
	}
	var out map[uuid.UUID]bool
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil || !c.effective.AbilitiesRemoved {
			continue
		}
		if out == nil {
			out = make(map[uuid.UUID]bool, 2)
		}
		out[c.InstanceID] = true
	}
	return out
}

// materialiseControlLocked copies layer 2's output back onto
// Card.Controller — the step that makes a control-changing effect
// visible to the rest of the engine.
//
// The alternative was a Card.EffectiveController() accessor and a
// sweep of every read site. There are ~385 of them, roughly 150 of
// which are `target.Controller == source.Controller` inside
// individual catalog card files; a partial sweep would have produced
// an engine where a Mind Controlled creature attacks for its new
// controller but is still pumped by its old one's Glorious Anthem.
// Materialising means every read site is right without being
// touched, and the same pattern ADR 0034 already uses for the flat
// printed face fields ("the MATERIALISATION of Faces[ActiveFace]").
//
// A control change is more than a field assignment, so the two
// CR consequences ride with it:
//
//   - CR 302.6 — the permanent has summoning sickness under its new
//     controller until their next untap step. It is set rather than
//     cleared, so a hasty creature is still hasty (HasSummoningSickness
//     reads the keyword at read time).
//   - CR 506.4 — a permanent that changes control is removed from
//     combat.
//
// Both fire only on an actual delta, so a recompute that changes
// nothing touches nothing.
//
// Caller holds g.mu in write mode. This is the step that made the
// layer engine's lock contract load-bearing: Card.Controller is read
// all over the engine under the READ lock, so writing it needs
// exclusivity, not the shared lock the recompute used to run under.
func (g *Game) materialiseControlLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil || c.effective.Controller == uuid.Nil {
			continue
		}
		if c.effective.Controller == c.Controller {
			continue
		}
		c.Controller = c.effective.Controller
		c.SummonedThisTurn = true
		c.AttackingTarget = uuid.Nil
		c.BlockingTarget = uuid.Nil
	}
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
