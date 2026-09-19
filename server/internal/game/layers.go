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
//   1. Copy effects (Clone) — landed in S16.5 as a rewrite of the
//      copiable-value baseline (Card.PrintedSelf), not as an effect
//      in this bucket; ADR 0043. The Layer1Copy bucket stays for
//      copies with a duration (Mirage Mirror, Cytoshape), which aren't
//      supported yet.
//   2. Control-changing effects (Mind Control) — S24. The output
//      lands in Characteristic.Controller and is materialised back
//      onto Card.Controller at the end of the pass; see
//      materialiseControlLocked for why that, and not an
//      EffectiveController() accessor plus a 385-site sweep.
//   3. Text-changing effects — out of scope for the layer foundation.
//   4. Type-changing effects (Mycosynth Lattice) — sub-PR 4.
//   5. Color-changing effects — engine stub in S16; first used by
//      Kenrith's Transformation and Song of the Dryads (ADR 0046).
//   6. Ability-changing effects (Lord of Atlantis grants) — sub-PR 4.
//   7. Power / toughness — sub-layers 7a (CDA, Tarmogoyf), 7b (set),
//      7c (modify, Glorious Anthem), 7d (counters, delegates to
//      CurrentPower/CurrentToughness), 7e (switch).
//
// Dependency ordering (CR 613.8) lives in layer_dependency.go: a
// bucket with two or more effects is ordered by dependency rather
// than by timestamp alone, decided by trial application rather than
// by a declaration on every card (ADR 0067). ADR 0012 and ADR 0043 §5
// deferred it on the grounds that no catalog pair could produce a
// dependency; that stopped being true on 2026-09-13, when Urborg,
// Tomb of Yawgmoth met Song of the Dryads and Arixmethes, and
// Maskwood Nexus met crew, The Warring Triad and Arixmethes.
//
// Ability removal (CR 613.1f) is applied in the layer it happens in
// and reaches forwards only: a source silenced in layer 6 still
// applied in layers 1-5, and an effect that already started applying
// carries on into the later layers (CR 613.6). That is why one pass
// is enough — see recomputeLayersLocked.

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
	// itself because the removal is load-bearing outside its own
	// Apply: the pass has to know which permanents were silenced and
	// IN WHICH LAYER, so it can stop their effects from starting in
	// the layers that come after (CR 613.6), and it cannot learn that
	// by watching an opaque closure mutate a struct.
	RemovesAbilities() bool

	// ContinuesAfterRemoval declares this effect part of a
	// continuous effect that STARTS in an earlier layer — the
	// layer-7b "with base power and toughness 0/1" half of "is a 0/1
	// Insect with indestructible and loses all other abilities".
	//
	// CR 613.6: an effect that has started to apply keeps applying in
	// every later layer "even if the ability generating the effect is
	// removed during this process". The engine models one printed
	// sentence as one StaticAbility per layer, so it cannot see on
	// its own that the layer-7b half and the layer-4 half are the
	// same effect — this says so. False, the default, is right for a
	// standalone ability (an anthem, a keyword grant, the printed
	// keyword synth), which simply stops existing when its source is
	// silenced in an earlier layer.
	ContinuesAfterRemoval() bool
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
	//
	// Layer 6 is the usual home, but not the only one: CR 305.7's
	// "an effect that sets a land's subtype to a basic land type
	// removes the abilities generated from its rules text" is part
	// of the TYPE change and belongs in layer 4 (Song of the Dryads,
	// Magus of the Moon — effects.SetsBasicLandType). Declaring it in
	// the layer the rules put it in is what makes a layer-6 grant
	// survive it, which is the whole of that rule's last sentence.
	RemovesAbilities bool

	// ContinuesAfterRemoval declares this static the later-layer half
	// of a continuous effect that starts in an earlier layer, so
	// CR 613.6 keeps it applying after its source has been silenced.
	// See ContinuousEffect.ContinuesAfterRemoval; ADR 0067 §2.
	ContinuesAfterRemoval bool
	// DependsOnHandSize declares that this ability's OUTPUT changes
	// when somebody's hand does — Psychosis Crawler's "power and
	// toughness are each equal to the number of cards in your hand".
	//
	// It is an INVALIDATION hint and nothing else: the layer engine
	// reads it to decide whether a hand-only zone move (a draw, a
	// discard, a card put into hand) has to drop the cached
	// resolution. See layerVersionBump.OnEvent.
	//
	// Declaring it is how a card whose value is not on the
	// battlefield stays correct between recomputes. The default —
	// false — is right for every ability whose inputs are permanents,
	// counters, tap state or the turn, which is all of them but one,
	// and that is why the hand is not on the unconditional bump list:
	// a hand change is the most frequent event in the game and
	// invalidating on it universally makes the recompute much hotter
	// for every table that has no such card in play.
	DependsOnHandSize bool

	// ActiveWhen is the CR 716 / 719 / 721 / 709.5 designation gate:
	// this static exists only while its source permanent has the
	// designation named — level N or greater, solved, N or more
	// charge counters, that door unlocked. The zero value is "no
	// gate", which is every static in the catalog but a handful.
	//
	// Evaluated in StaticAbilitiesForCard and nowhere else, so a
	// gated-off static is never gathered, never sorted into a bucket
	// and never applied. See designations.go and ADR 0071.
	ActiveWhen Designation
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

func (e staticContinuousEffect) ContinuesAfterRemoval() bool {
	return e.ability.ContinuesAfterRemoval
}

func (e staticContinuousEffect) Apply(c *Characteristic, target *Card, g *Game) {
	if e.ability.Apply == nil {
		return
	}
	e.ability.Apply(c, target, g, e.source)
}

// activeStaticAbilitiesLocked collects every continuous effect in
// play from its three sources and returns them as
// `ContinuousEffect`s bound to source pointers + timestamps:
//
//  1. Battlefield permanents — walks g.Battlefield and looks each
//     card's catalog static abilities up via the
//     CatalogStaticAbilities hook. These live exactly as long as
//     the source permanent does (CR 113.6).
//  2. Scoped statics — the floating continuous-effect registry
//     (scoped_statics.go), whose entries have no battlefield source
//     and end on a duration instead (CR 611.2).
//  3. Emblems — the S40 command-zone objects (emblem.go, #623),
//     whose abilities function where they are (CR 114.3) and which
//     nothing can remove short of their owner leaving the game.
//
// Caller must hold g.mu in write mode.
//
// The returned source pointers reference into g.Battlefield.Cards
// — safe for the duration of the recompute pass that holds the
// write lock; not safe to retain across mutations.
//
// EVERY battlefield permanent contributes, including one a CR 613.1f
// ability-removing effect is about to silence. Removal happens in a
// layer, and a layer reaches forwards only (CR 613.6): the gather
// cannot be where silencing is decided, because by the time the
// removal applies the earlier layers have already run and the rules
// do not take them back. applyBucketLocked owns the silencing, per
// layer, and ADR 0067 §2 has the rule.
//
// Until #669 this dropped a silenced source from every layer, which
// silently deleted its layer-2 control change and its layer-4 type
// change as well as its abilities — the opposite of Magus of the
// Moon's 2021-03-19 ruling, where a silenced Magus keeps turning
// nonbasic lands into Mountains.
//
// Scoped statics and emblems are never silenced. Neither has a
// battlefield source to take abilities away from: a scoped static
// outlived its source by construction (CR 611.2b), and nothing in the
// game can name an emblem at all (CR 114), so nothing on the board
// can switch either off.
func (g *Game) activeStaticAbilitiesLocked() []ContinuousEffect {
	// S32/S38: floating effects with a duration first. They are
	// gathered unconditionally — they outlive their source card, so
	// neither an empty battlefield nor a missing catalog hook can
	// switch them off. Order within this slice is irrelevant: the
	// per-bucket sort in applyLayerLocked re-orders everything by
	// timestamp (CR 613.7) before applying.
	out := g.scopedContinuousEffectsLocked()
	// #623 / CR 114.3: an emblem's abilities function in the command
	// zone. One more source list into the same gather, never a second
	// pass — see emblem.go.
	out = append(out, g.emblemContinuousEffectsLocked()...)
	if g.Battlefield == nil || CatalogStaticAbilities == nil {
		return out
	}
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		// StaticAbilitiesForCard, not the raw hook: it is the ONE
		// place an object is turned into its statics, so a static
		// gated on a designation the permanent does not have (a
		// level-3 anthem on a level-1 Class, a 7+ P/T set on a
		// Spacecraft with five charge counters) is never gathered at
		// all. ADR 0071.
		//
		// It reads CatalogKey, not CatalogAbilityKey, and
		// deliberately: the pass resets every effective characteristic
		// to printed before it gathers, so there is nothing for the
		// removal accessor to read yet — and nothing for it to say,
		// because a removal that has not been applied in this pass has
		// not happened.
		abilities := StaticAbilitiesForCard(*src)
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
// sorts by timestamp, and hands the bucket to applyBucketLocked.
// Stable sort so same-timestamp effects fall in their gather order —
// matters for test determinism and for catalog cards with multiple
// statics from a single source.
//
// `bucketIndex` is the effect's position in layerOrder, which is what
// "earlier layer" and "later layer" mean everywhere below. CR 613.6
// is a statement about that ordering and nothing else.
func (g *Game) applyLayerLocked(effects []ContinuousEffect, l Layer, sub SubLayer, has7Sub bool, bucketIndex int, st *layerPassState) {
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
	g.applyBucketLocked(bucket, l, bucketIndex, st)
}

// applyOneEffectLocked applies one continuous effect across every
// battlefield card it AppliesTo, honouring CR 613.6's rule about a
// source whose abilities have already been removed in this pass.
//
// The silencing rule, once, for every layer (ADR 0067 §2):
//
//   - Before the removal's own bucket, nothing is silenced — the
//     removal has not happened, and CR 613.6 will not let it reach
//     back. This is the branch that was missing (#669): a Magus of
//     the Moon under a Kenrith's Transformation keeps turning
//     nonbasic lands into Mountains, and every layer 1-5 effect of
//     any silenced source keeps applying.
//   - IN the removal's bucket, timestamp order decides, which is
//     what `HasLostAllAbilities` already answers: two Song of the
//     Dryads each enchanting the other settle on "the earlier one
//     wins" rather than on a paradox.
//   - AFTER it, the source's effects stop — unless this effect is the
//     later-layer half of one that already started applying to this
//     object, which CR 613.6 says carries on regardless
//     (ContinuesAfterRemoval; Humility's own base 1/1 in layer 7b
//     after it has taken its own abilities away in layer 6).
func (g *Game) applyOneEffectLocked(eff ContinuousEffect, bucketIndex int, st *layerPassState) {
	// #690: an effect in the 7a or 7b bucket DEFINES the object's
	// power and toughness, and the object it applied to stops being
	// a `*` whose stats the engine cannot compute. One bool per
	// effect, read per object it applies to; see
	// Characteristic.PTDefined and Card.ToughnessIsKnown.
	definesPT := definesPowerAndToughness(eff)
	// The fast path, and it is the board almost every recompute runs
	// on: nothing in play can remove an ability, so CR 613.6 has
	// nothing to say, no source can be silenced, and the whole of the
	// bookkeeping below is provably dead. `st.track` is decided once
	// per pass from the gathered effects, so this is one bool test
	// per effect rather than per (effect, object) — which matters,
	// because this loop is the hottest in the engine.
	if !st.track {
		for i := range g.Battlefield.Cards {
			target := &g.Battlefield.Cards[i]
			if !eff.AppliesTo(target, g) {
				continue
			}
			eff.Apply(target.effective, target, g)
			if definesPT {
				target.effective.PTDefined = true
			}
		}
		return
	}
	src := effectSourceLocked(eff)
	silenced := src != nil && src.HasLostAllAbilities()
	removes := eff.RemovesAbilities()
	continues := eff.ContinuesAfterRemoval()
	for i := range g.Battlefield.Cards {
		target := &g.Battlefield.Cards[i]
		if !eff.AppliesTo(target, g) {
			continue
		}
		if silenced && !st.stillApplies(src, target, bucketIndex, continues) {
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
			st.recordRemoval(target, bucketIndex)
		}
		eff.Apply(target.effective, target, g)
		if definesPT {
			target.effective.PTDefined = true
		}
		st.recordApplied(src, target)
	}
}

// definesPowerAndToughness reports whether an effect SETS power and
// toughness rather than moving them: CR 613.4a's
// characteristic-defining abilities (7a) and CR 613.4b's "is a 1/1"
// effects (7b). 7c modifies, 7d counts counters and 7e switches —
// all three need a number to already be there, so none of them tells
// the engine what the number is.
//
// The one reader is Characteristic.PTDefined, which the toughness
// state-based action consults through Card.ToughnessIsKnown (#690).
func definesPowerAndToughness(eff ContinuousEffect) bool {
	l, sub := eff.Layer()
	return l == Layer7PT && (sub == SubLayer7A_CDA || sub == SubLayer7B_Set)
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
// One pass is all of it. S24 used to run the pass repeatedly to a
// fixed point so it could learn which permanents to hold OUT of the
// gather; #669 deleted that, because holding a silenced source out of
// the gather is what broke CR 613.6 in the first place. A removal is
// applied in its own layer and reaches forwards only, so the single
// walk of layerOrder already knows everything the second walk used to
// discover.
//
// Caller must hold g.mu in write mode — the pass writes
// Card.effective on every battlefield card and, via
// materialiseControlLocked, Card.Controller on any permanent layer 2
// moved. The recompute counter bumps on every call so the fast-path
// test can assert it ran exactly once.
func (g *Game) recomputeLayersLocked() {
	g.recomputeCount.Add(1)
	// S38 (ADR 0063): a "for as long as ~" duration is a condition
	// the board can falsify at any moment — the source dies, it is
	// flickered into a new object (CR 400.7), its controller changes
	// — and every one of those bumps the layer version, so the top of
	// the recompute is the one place that is guaranteed to run
	// afterwards. The sweep bumps the version again when it drops
	// anything, and the store at the end of this function picks that
	// up, so a pass that ends an effect settles in one go rather than
	// looping.
	g.ClearExpiredScopedStaticsLocked()
	g.layerPassLocked()
	changed := g.materialiseControlLocked()
	g.lastResolvedVersion.Store(g.layerVersion.Load())
	// #930: the control deltas are EMITTED here, after the store, and
	// not from inside the walk that found them. EmitEvent dispatches
	// to the listeners synchronously — the trigger harvester among
	// them — and a harvested trigger reads the board, which can send
	// it back through RecomputeLayersIfStaleLocked. Emitting before
	// the store would make that a re-entrant recompute in the middle
	// of a pass whose results are half-written; emitting after it
	// finds the cache clean and returns, and anything a listener
	// invalidates is picked up by the next pass exactly as any other
	// mutation is. See emitControlChangesLocked.
	g.emitControlChangesLocked(changed)
}

// layerPassLocked runs one complete CR 613 application over the
// battlefield. Caller must hold g.mu in write mode.
func (g *Game) layerPassLocked() {
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
	effects := g.activeStaticAbilitiesLocked()
	st := newLayerPassState(effects)
	for i, b := range layerOrder {
		g.applyLayerLocked(effects, b.Layer, b.SubLayer, b.has7Sub, i, st)
	}
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
//     combat. Declaration AND announcement, through
//     `removeFromCombatLocked`: clearing `AttackingTarget` alone left
//     the creature marked as already-announced for this combat, so
//     its next attack declaration fired no trigger (#871).
//
// Both fire only on an actual delta, so a recompute that changes
// nothing touches nothing. The third consequence is the EVENT
// (#930), and it is the one thing this step does not do itself: the
// deltas are returned by value for `recomputeLayersLocked` to emit
// once the pass is over, because emitting mid-walk would dispatch
// listeners against a half-applied board and hand a *Card into a
// harvester whose Build may reallocate the battlefield slice — the
// hazard `commitAttackDeclarationLocked` collects by value to avoid.
//
// Caller holds g.mu in write mode. This is the step that made the
// layer engine's lock contract load-bearing: Card.Controller is read
// all over the engine under the READ lock, so writing it needs
// exclusivity, not the shared lock the recompute used to run under.
func (g *Game) materialiseControlLocked() []controlChange {
	if g.Battlefield == nil {
		return nil
	}
	var changed []controlChange
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil || c.effective.Controller == uuid.Nil {
			continue
		}
		if c.effective.Controller == c.Controller {
			continue
		}
		changed = append(changed, controlChange{
			card:   c.InstanceID,
			from:   c.Controller,
			to:     c.effective.Controller,
			source: c.effective.ControlSource,
		})
		c.Controller = c.effective.Controller
		c.SummonedThisTurn = true
		g.removeFromCombatLocked(c)
	}
	return changed
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
// effective is empty. Returns the COLOURS only; the tri-state that
// says what an empty list means (CR 903.4f, #844) stays inside the
// package, where the one narrowing function reads it.
func CommanderIdentityForTest(g *Game, p *Player) []string {
	return commanderIdentityFor(g, p).Colors
}
