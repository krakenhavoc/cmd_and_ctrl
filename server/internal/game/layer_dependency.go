package game

import (
	"slices"

	"github.com/google/uuid"
)

// layer_dependency.go is CR 613.8 — dependency ordering WITHIN one
// layer — and CR 613.6 — what happens to an effect whose source
// loses its abilities part-way through the pass. ADR 0067 has the
// decision; this file is the whole of the implementation, and
// layers.go calls into it from exactly two places
// (applyLayerLocked → applyBucketLocked, applyOneEffectLocked → the
// layerPassState methods).
//
// CR 613.7 orders a layer by timestamp. CR 613.8a overrides that for
// a PAIR of effects in the same layer: effect A depends on effect B
// when applying B would change A's text or existence, WHAT A APPLIES
// TO, or what A does to the things it applies to. A dependent effect
// waits until what it depends on has applied (613.8b), the question
// is asked again after each application (613.8c), and a loop falls
// back to timestamp order (613.8b's last sentence).
//
// Detection here is by TRIAL APPLICATION and not by a declaration on
// every StaticAbility, which is the design ADR 0043 §5 priced and
// declined: a required "what I read / what I write" field on ~1700
// catalog entries, wrong invisibly until two specific cards meet.
// The trial asks the rules' own question of the effects themselves.
//
// The one thing the trial must not do is diff A's OUTPUT, which is
// why applyOnFixedBase exists. Urborg, Tomb of Yawgmoth on a Command
// Tower that a Song of the Dryads has already turned into a Forest
// produces a different characteristic than it would have on the
// untouched Tower — and Urborg does not depend on the Song there,
// because the Tower was a land either way and "append Swamp" is the
// same instruction. Running A against the SAME base in both worlds
// isolates what A does from what it was done to, which is the
// distinction CR 613.8a draws and an output diff cannot.

// dependencyOrderedLayers names the buckets that get CR 613.8
// treatment. Layer 4 is the only one the catalog can build a
// dependency in: a type-changing effect whose AppliesTo reads a card
// type or subtype (Urborg's "each land", Maskwood Nexus' "creatures
// you control") sits in the same bucket as the type-SETTING effects
// that write those (Song of the Dryads, Arixmethes' slumber, The
// Warring Triad's self type-strip, crew).
//
// The restriction is cost, not principle — the machinery below is
// layer-agnostic and adding a bucket is one line. Layer 7c is the
// bucket that makes the difference: a saturated board carries ten
// anthems there and none of them can change what another applies to,
// so paying O(N²) trial applications on every recompute would buy
// nothing. ADR 0067 §1 has the measured numbers and the rule for
// adding a layer: a catalogued pair, plus a benchmark.
//
// Layer 6 is not in the list, and is ordered all the same for the one
// kind of effect there that can depend on another: a static that
// declares StaticAbility.ReadsAbilities (Sephara, Sky's Blade's "other
// creatures you control with flying have indestructible"). The trial
// runs for those rows only — nextAbilityEffectLocked, below — so a
// layer-6 bucket with no such static pays one bool test per effect.
// ADR 0067's amendment of 2026-10-23 has the decision and the numbers.
var dependencyOrderedLayers = map[Layer]bool{
	Layer4Type: true,
}

// sourceTarget keys the CR 613.6 "has started to apply" record.
type sourceTarget struct {
	source uuid.UUID
	target uuid.UUID
}

// layerPassState is the CR 613.6 bookkeeping for one layer pass: who
// was silenced, in which layer, and which (source, object) pairs an
// effect had already started applying to by then.
//
// Every field is inert — and every map nil — on a board with no
// ability-removing effect on it, which is nearly every board. That is
// what `track` is for: the pass pays for this only when something can
// actually take an ability away.
type layerPassState struct {
	track bool
	// removedAt maps a permanent to the layerOrder index of the
	// bucket in which a CR 613.1f removal applied to it. Layer 6 for
	// "loses all abilities"; layer 4 for CR 305.7's "is a Forest".
	removedAt map[uuid.UUID]int
	// removedTS maps a permanent to the latest timestamp of a LAYER-6
	// removal that applied to it in this pass (ADR 0155 Decision 4.4).
	// The borrow step runs after every other layer-6 effect, so it reads
	// this to keep CR 613.6's timestamp order against a removal on a
	// recipient: a removal later than the borrow takes the lent rows.
	removedTS map[uuid.UUID]int64
	// started records that a live source's continuous effect has
	// already applied to an object — CR 613.6's "starts to apply" —
	// with the layerOrder index of the first bucket it applied in.
	started map[sourceTarget]int
	// sourceStarted is the same record per source: the first bucket in
	// which any of its effects applied to anything. The layer-6
	// pre-sort reads it (orderRemovalsBeforeTheirVictimsLocked).
	sourceStarted map[uuid.UUID]int
	// granted records each grant of a bundle that carries a static,
	// with the timestamp of the effect that wrote it (ADR 0093
	// amendment 2026-10-15, granted_statics.go). Unlike the fields
	// above it is kept on every board, since a grant needs no removal
	// to be in play; it stays nil on a board with no such bundle.
	granted []grantStamp
}

// newLayerPassState decides up front whether this pass has to track
// anything: one walk of the gathered effects, no allocation unless
// something on the board removes abilities.
func newLayerPassState(effects []ContinuousEffect) *layerPassState {
	st := &layerPassState{}
	for _, eff := range effects {
		if eff.RemovesAbilities() {
			st.track = true
			st.removedAt = make(map[uuid.UUID]int, 2)
			st.removedTS = make(map[uuid.UUID]int64, 2)
			st.started = make(map[sourceTarget]int, 8)
			st.sourceStarted = make(map[uuid.UUID]int, 4)
			break
		}
	}
	return st
}

// recordRemoval stamps the bucket a CR 613.1f removal applied to this
// object in.
func (st *layerPassState) recordRemoval(target *Card, bucketIndex int) {
	if st == nil || !st.track {
		return
	}
	if _, seen := st.removedAt[target.InstanceID]; !seen {
		st.removedAt[target.InstanceID] = bucketIndex
	}
}

// recordRemovalTimestamp keeps the latest timestamp of a layer-6
// removal that applied to this object.
func (st *layerPassState) recordRemovalTimestamp(target *Card, ts int64) {
	if st == nil || !st.track {
		return
	}
	if prev, seen := st.removedTS[target.InstanceID]; !seen || ts > prev {
		st.removedTS[target.InstanceID] = ts
	}
}

// removedAfter reports whether a layer-6 removal with a timestamp later
// than `ts` applied to this object in this pass. Free on a board with no
// removal.
func (st *layerPassState) removedAfter(target *Card, ts int64) bool {
	if st == nil || !st.track {
		return false
	}
	at, ok := st.removedTS[target.InstanceID]
	return ok && at > ts
}

// recordApplied notes that `src`'s effect has applied to `target` in
// the bucket at `bucketIndex`. The first bucket is the one kept.
func (st *layerPassState) recordApplied(src *Card, target *Card, bucketIndex int) {
	if st == nil || !st.track || src == nil {
		return
	}
	key := sourceTarget{source: src.InstanceID, target: target.InstanceID}
	if _, seen := st.started[key]; !seen {
		st.started[key] = bucketIndex
	}
	if _, seen := st.sourceStarted[src.InstanceID]; !seen {
		st.sourceStarted[src.InstanceID] = bucketIndex
	}
}

// startedBefore reports whether an effect of `src` applied to anything
// in a bucket earlier than `bucketIndex` in this pass.
func (st *layerPassState) startedBefore(src *Card, bucketIndex int) bool {
	if st == nil || !st.track || src == nil {
		return false
	}
	at, ok := st.sourceStarted[src.InstanceID]
	return ok && at < bucketIndex
}

// stillApplies answers CR 613.6 for an effect whose source has
// already had its abilities removed in this pass: may it apply to
// this object anyway?
//
// Yes in exactly one case — it is the later-layer half of a
// continuous effect that had already started applying to this same
// object in an EARLIER layer. "If an effect starts to apply in one
// layer and/or sublayer, it will continue to be applied to the same
// set of objects in each other applicable layer and/or sublayer,
// even if the ability generating the effect is removed during this
// process." Humility taking its own abilities away in layer 6 and
// still making every creature 1/1 in layer 7b is that sentence.
//
// That includes the removal's own bucket (#3862, ADR 0067's amendment
// of 2026-10-15). Idol of False Gods is a creature from layer 4 at
// eight counters, so its annihilator in layer 6 is the same effect
// carrying on: an older "loses all abilities" sorts before it and does
// not stop it, and a newer one sorts after it and takes the keyword
// away. Only an effect that has not started yet is stopped, which is
// what two Song of the Dryads enchanting each other are: each one's
// first layer is the removal's own, so the earlier one wins by
// timestamp and the other never starts.
func (st *layerPassState) stillApplies(src *Card, target *Card, bucketIndex int, continues bool) bool {
	if st == nil || !st.track || src == nil || !continues {
		return false
	}
	at, ok := st.started[sourceTarget{source: src.InstanceID, target: target.InstanceID}]
	return ok && at < bucketIndex
}

// applyBucketLocked applies one timestamp-ordered bucket, reordering
// it by CR 613.8 dependency first where that layer asks for it.
//
// The loop is CR 613.8b + 613.8c literally: take the first effect
// that depends on nothing else still waiting, apply it, and ask
// again. `firstIndependent` returning -1 is the loop case — every
// remaining effect depends on another — and CR 613.8b says to give up
// on dependency there and fall back to timestamp order, which is what
// index 0 is.
//
// The early `break` is the fast path and the reason this is
// affordable: once no dependency remains among the effects still to
// apply, the rest is plain timestamp order and no further trials are
// run. On a board with no dependency at all that is one relation
// computation per bucket and nothing else.
func (g *Game) applyBucketLocked(bucket []ContinuousEffect, l Layer, bucketIndex int, st *layerPassState) {
	// ADR 0155 Decision 4: a borrow ("has all activated abilities of …")
	// applies after every other layer-6 effect, as one step at the end of
	// the bucket. Taken out here; a bucket with none comes back as it was.
	var borrows []ContinuousEffect
	if l == Layer6Ability {
		bucket, borrows = splitBorrows(bucket)
	}
	// ADR 0093 Decision 3: in the layer-6 bucket a removal goes before
	// every effect whose SOURCE it silences, whatever the timestamps.
	// Only on a board where something removes abilities at all — which
	// `st.track` has already decided — so every other pass pays one
	// bool test.
	if l == Layer6Ability && len(bucket) > 1 && st != nil && st.track && g.Battlefield != nil {
		bucket = g.orderRemovalsBeforeTheirVictimsLocked(bucket, bucketIndex, st)
	}
	if len(bucket) > 1 && dependencyOrderedLayers[l] && g.Battlefield != nil {
		remaining := bucket
		for len(remaining) > 1 {
			edges := g.dependencyEdgesLocked(remaining)
			if edges == nil {
				break
			}
			k := firstIndependent(edges)
			if k < 0 {
				k = 0
			}
			g.applyOneEffectLocked(remaining[k], bucketIndex, st)
			remaining = append(remaining[:k:k], remaining[k+1:]...)
		}
		bucket = remaining
	}
	if l == Layer6Ability {
		// ADR 0093 amendment 2026-10-15: a static a grant brings in
		// joins the rest of this bucket (granted_statics.go).
		g.applyGrantingBucketLocked(bucket, bucketIndex, st)
	} else {
		for _, eff := range bucket {
			g.applyOneEffectLocked(eff, bucketIndex, st)
		}
	}
	if len(borrows) > 0 && g.Battlefield != nil {
		g.applyBorrowsLocked(borrows, st)
	}
}

// orderRemovalsBeforeTheirVictimsLocked is ADR 0093 Decision 3's
// structural pre-sort of the layer-6 bucket — CR 613.8a(b) for the one
// dependency this engine can have in layer 6:
//
//	A removal effect is applied before every effect whose SOURCE it
//	applies to. If two removals each apply to the other's source, that
//	is a dependency loop (CR 613.8b) and timestamp order decides.
//
// The EXISTENCE of a grantor's effect depends on a removal that would
// take the grantor's abilities away, so a Kenrith's Transformation on
// Gemhide Sliver strips the grant from every Sliver, and a Lord of
// Atlantis under a Darksteel Mutation stops granting islandwalk, however
// the timestamps fall. Before this, the bucket was timestamp order alone
// and a lord whose grant happened to sort first kept granting.
//
// Applying the removal first is all it takes: applyOneEffectLocked then
// finds the grantor silenced in this very bucket and skips its effects,
// as it always has for a removal that sorted earlier. Effects no removal
// waits on keep their timestamp order relative to each other, and the
// input is already timestamp-sorted, so a board with no such pair gets
// the input back. It is not ADR 0067's trial application: the relation
// is read off the removals' own AppliesTo, once per pass.
//
// The other layer-6 dependency the catalog reaches — a grant whose
// "applies to" reads an ability another effect adds or removes — is
// ordered later, inside the bucket, by nextAbilityEffectLocked (ADR 0067
// amendment of 2026-10-23).
//
// An effect that is the later-layer half of one that already started
// in an earlier layer is nobody's victim (CR 613.6, #3862): the removal
// can no longer change its existence, so the two are independent and
// keep their timestamp order.
func (g *Game) orderRemovalsBeforeTheirVictimsLocked(bucket []ContinuousEffect, bucketIndex int, st *layerPassState) []ContinuousEffect {
	var removals []int
	for i, e := range bucket {
		if e.RemovesAbilities() {
			removals = append(removals, i)
		}
	}
	if len(removals) == 0 {
		return bucket
	}
	n := len(bucket)
	var waits [][]int
	for i, e := range bucket {
		src := effectSourceLocked(e)
		if src == nil {
			continue
		}
		if e.ContinuesAfterRemoval() && st.startedBefore(src, bucketIndex) {
			continue
		}
		for _, r := range removals {
			if r == i || !bucket[r].AppliesTo(src, g) {
				continue
			}
			if waits == nil {
				waits = make([][]int, n)
			}
			waits[i] = append(waits[i], r)
		}
	}
	if waits == nil {
		return bucket
	}
	out := make([]ContinuousEffect, 0, n)
	done := make([]bool, n)
	for len(out) < n {
		pick, first := -1, -1
		for i := 0; i < n; i++ {
			if done[i] {
				continue
			}
			if first < 0 {
				first = i
			}
			ready := true
			for _, r := range waits[i] {
				if !done[r] {
					ready = false
					break
				}
			}
			if ready {
				pick = i
				break
			}
		}
		if pick < 0 {
			// CR 613.8b: every remaining effect waits on another — the
			// loop — so timestamp order, which is the earliest left.
			pick = first
		}
		done[pick] = true
		out = append(out, bucket[pick])
	}
	return out
}

// firstIndependent returns the index of the first effect that depends
// on no other effect in the set, or -1 when every one of them does —
// the CR 613.8b loop.
func firstIndependent(edges [][]bool) int {
	for i := range edges {
		depends := false
		for j := range edges[i] {
			if edges[i][j] {
				depends = true
				break
			}
		}
		if !depends {
			return i
		}
	}
	return -1
}

// dependencyEdgesLocked computes the CR 613.8a relation over a
// bucket: edges[i][j] is "effs[i] depends on effs[j]". Returns nil —
// not an empty matrix — when nothing depends on anything, which is
// the answer on every board without one of the catalogued pairs and
// is what lets applyBucketLocked skip straight to timestamp order.
//
// Cost is O(N²) trial applications for a bucket of N effects, and the
// per-trial work is bounded by the number of objects the CANDIDATE
// DEPENDENCY applies to, not by the size of the battlefield: an
// effect can only change what another effect sees on the objects it
// applies to. Declared narrowing, ADR 0067 §1: an effect whose Apply
// reads a board-wide count rather than its target ("+1/+1 for each
// artifact you control") can be changed by an effect that never
// touches the same object. No catalogued layer-4 effect is in that
// class, and the ones that count a board live in layer 7c, which is
// not dependency-ordered.
func (g *Game) dependencyEdgesLocked(effs []ContinuousEffect) [][]bool {
	n := len(effs)
	var edges [][]bool
	for j := 0; j < n; j++ {
		idxs := g.effectTargetIndexesLocked(effs[j])
		if len(idxs) == 0 {
			// An effect that applies to nothing changes nothing, so
			// nothing can depend on it.
			continue
		}
		// The bases are the pre-B characteristics, cloned once and
		// reused for both halves of every probe, so that what is
		// compared is what A DOES and not what A was done to.
		bases := make([]Characteristic, len(idxs))
		for t, idx := range idxs {
			bases[t] = g.Battlefield.Cards[idx].effective.clone()
		}
		before := g.observeLocked(effs, j, idxs, bases)
		restore := g.installTrialLocked(effs[j], idxs)
		after := g.observeLocked(effs, j, idxs, bases)
		restore()
		for i := 0; i < n; i++ {
			if i == j {
				continue
			}
			for t := range idxs {
				if sameObservation(before[i][t], after[i][t]) {
					continue
				}
				if edges == nil {
					edges = make([][]bool, n)
					for k := range edges {
						edges[k] = make([]bool, n)
					}
				}
				edges[i][j] = true
				break
			}
		}
	}
	return edges
}

// observation is what one effect would do to one object: whether it
// applies at all (CR 613.8a's "what it applies to") and, if it does,
// the result of running it against a fixed base ("what it does to any
// of the things it applies to").
type observation struct {
	applies bool
	out     Characteristic
}

// observeLocked records every effect but effs[skip] against every
// probe object, in whatever world is currently installed.
func (g *Game) observeLocked(effs []ContinuousEffect, skip int, idxs []int, bases []Characteristic) [][]observation {
	out := make([][]observation, len(effs))
	for i := range effs {
		if i == skip {
			continue
		}
		out[i] = make([]observation, len(idxs))
		for t, idx := range idxs {
			target := &g.Battlefield.Cards[idx]
			if !effs[i].AppliesTo(target, g) {
				continue
			}
			out[i][t] = observation{
				applies: true,
				out:     g.applyOnFixedBase(effs[i], target, bases[t]),
			}
		}
	}
	return out
}

// applyOnFixedBase runs an effect against a copy of `base` rather
// than against the object's live characteristic. The effect still
// READS the live world through `target` and `g` — that is the point,
// it is what makes a dependency visible — but it WRITES into a
// characteristic that is identical in both halves of the probe.
func (g *Game) applyOnFixedBase(eff ContinuousEffect, target *Card, base Characteristic) Characteristic {
	c := base.clone()
	applyRaw(eff, &c, target, g)
	return c
}

// installTrialLocked applies one effect to clones of the objects it
// applies to and returns the closure that puts the real
// characteristics back. Nothing outside the probe ever sees the
// clones: the pass holds the write lock throughout.
//
// Silencing is deliberately NOT honoured here. CR 613.8a asks what
// would happen if the other effect were applied, which is a question
// about the effect and not about whether the pass is going to run it.
func (g *Game) installTrialLocked(eff ContinuousEffect, idxs []int) func() {
	saved := make([]*Characteristic, len(idxs))
	for t, idx := range idxs {
		c := &g.Battlefield.Cards[idx]
		saved[t] = c.effective
		clone := c.effective.clone()
		c.effective = &clone
	}
	for _, idx := range idxs {
		c := &g.Battlefield.Cards[idx]
		applyRaw(eff, c.effective, c, g)
	}
	return func() {
		for t, idx := range idxs {
			g.Battlefield.Cards[idx].effective = saved[t]
		}
	}
}

// applyRaw is the effect's whole instruction — the engine's CR 613.1f
// removal half plus the card's Apply — with none of the pass's
// bookkeeping. Shared by the trial and by nothing else; the real pass
// goes through applyOneEffectLocked, which is this plus CR 613.6.
func applyRaw(eff ContinuousEffect, c *Characteristic, target *Card, g *Game) {
	if eff.RemovesAbilities() {
		c.Abilities = nil
		c.GrantedAbilities = nil
		c.BorrowedAbilities = nil
		c.AbilitiesRemoved = true
	}
	if l, _ := eff.Layer(); l == Layer4Type {
		// #2041: the CR 205.1a subtype loss is part of the
		// instruction, so a probe sees it (card_type_loss.go).
		applyTypeChange(eff, c, target, g)
		return
	}
	eff.Apply(c, target, g)
}

// effectTargetIndexesLocked lists the battlefield positions an effect
// currently applies to.
func (g *Game) effectTargetIndexesLocked(eff ContinuousEffect) []int {
	var out []int
	for i := range g.Battlefield.Cards {
		if eff.AppliesTo(&g.Battlefield.Cards[i], g) {
			out = append(out, i)
		}
	}
	return out
}

// sameObservation reports whether two trial observations are the same
// answer to CR 613.8a's question.
func sameObservation(a, b observation) bool {
	if a.applies != b.applies {
		return false
	}
	if !a.applies {
		return true
	}
	return sameCharacteristic(a.out, b.out)
}

// sameCharacteristic compares two characteristics field by field.
// Explicit rather than reflect.DeepEqual because nil and empty slices
// have to compare EQUAL here: an effect that appends nothing to a nil
// slice and one that appends nothing to an empty one are doing the
// same thing, and calling that a dependency would reorder a layer for
// no reason.
//
// Characteristic.PTDefined is deliberately absent: it is a fact about
// the PASS ("an effect defined this object's P/T"), not a
// characteristic of the object, and applyRaw — the probe's whole
// instruction — never writes it. Comparing it would answer CR 613.8a
// with bookkeeping.
func sameCharacteristic(a, b Characteristic) bool {
	return a.Power == b.Power &&
		a.Toughness == b.Toughness &&
		a.Name == b.Name &&
		a.AllCreatureTypes == b.AllCreatureTypes &&
		a.AbilitiesRemoved == b.AbilitiesRemoved &&
		slices.Equal(a.RemovedOwnAbilities, b.RemovedOwnAbilities) &&
		a.Controller == b.Controller &&
		a.Restrictions == b.Restrictions &&
		a.AdditionalBlocks == b.AdditionalBlocks &&
		a.BlocksAnyNumber == b.BlocksAnyNumber &&
		a.DamageByToughness == b.DamageByToughness &&
		a.DamageAsPositivePower == b.DamageAsPositivePower &&
		slices.Equal(a.CostPower, b.CostPower) &&
		a.CantBeAttacked == b.CantBeAttacked &&
		a.CantBeSacrificed == b.CantBeSacrificed &&
		a.CantTransform == b.CantTransform &&
		sameStringSlice(a.Types, b.Types) &&
		sameStringSlice(a.Subtypes, b.Subtypes) &&
		sameStringSlice(a.Supertypes, b.Supertypes) &&
		sameStringSlice(a.Colors, b.Colors) &&
		sameStringSlice(a.Abilities, b.Abilities) &&
		sameGrants(a.GrantedAbilities, b.GrantedAbilities) &&
		slices.Equal(a.BorrowedAbilities, b.BorrowedAbilities) &&
		slices.Equal(a.AttackRequirements, b.AttackRequirements) &&
		slices.EqualFunc(a.AttackTargetRestrictions, b.AttackTargetRestrictions, AttackTargetRestriction.Equal) &&
		slices.Equal(a.DefenderWaivers, b.DefenderWaivers) &&
		slices.Equal(a.BlockRequirements, b.BlockRequirements) &&
		sameStringSlice(a.CantHave, b.CantHave)
}

// sameGrants compares two layered-grant lists; nil equals empty, for
// sameStringSlice's reason.
func sameGrants(a, b []GrantedAbility) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// sameStringSlice compares element by element; nil equals empty.
func sameStringSlice(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- layer 6: a grant that reads an ability (ADR 0067 amendment 2026-10-23) ---
//
// Sephara, Sky's Blade's "Other creatures you control with flying have
// indestructible" applies to the creatures that have flying, and flying
// is itself added and removed in layer 6: by a lord, an Equipment, an
// until-end-of-turn grant, a flying counter, a printed keyword row that
// entered after her, or Gravity Sphere's "All creatures lose flying".
// Applying any of those changes what her effect applies to, so it
// depends on each of them (CR 613.8a) and waits until they have applied
// (CR 613.8b). In timestamp order alone a creature that gained flying
// later got no indestructible, and one that lost flying later kept it.
//
// The relation is ADR 0067's trial application, unchanged, asked for
// the rows that declare StaticAbility.ReadsAbilities and of nothing
// else. The narrowing is stated: a layer-6 effect that does not declare
// it is never treated as dependent. It can't be one unless it reads an
// ability, and the declaration is what keeps the bucket — which holds a
// printed keyword row for every catalogued permanent with a keyword —
// from paying O(N²) trials on every board.

// readsAbilities reports whether a layer-6 effect declares
// StaticAbility.ReadsAbilities. Only a battlefield (or emblem) static can:
// a granted bundle can't carry it (GrantedStaticProblem) and a resolving
// spell's effect locks its set as it resolves (CR 611.2c).
func readsAbilities(eff ContinuousEffect) bool {
	sce, ok := eff.(staticContinuousEffect)
	return ok && sce.ability.ReadsAbilities
}

// countAbilityReaders is the fast-path test: how many effects in the
// bucket declare ReadsAbilities. Zero on nearly every board, and then the
// bucket is applied exactly as before.
func countAbilityReaders(bucket []ContinuousEffect) int {
	n := 0
	for _, eff := range bucket {
		if readsAbilities(eff) {
			n++
		}
	}
	return n
}

// nextAbilityEffectLocked returns the index, in `rest` (the layer-6
// effects still to apply, in the order the bucket has them), of the one
// to apply next.
//
// CR 613.8b and 613.8c: the first effect that depends on nothing still
// waiting, asked again before every application. An effect that does not
// read an ability depends on nothing, so the first one of those ends the
// scan with no trial at all; a reader is applied when no remaining effect
// would change what it applies to or what it does. When every remaining
// effect is a reader waiting on another, the readers form a dependency
// loop, or wait on one, and firstInDependencyLoopLocked picks.
//
// `hint` is the index in rest of an effect rest[0] was found waiting on
// before the last application, or -1. It is tried first: a reader that
// waits on an effect near the end of the bucket is asked again at every
// step, and the answer is nearly always the same effect, so this keeps
// each step to one trial. The second result is that index for this step
// (an effect rest[0] waits on), or -1 when rest[0] does not wait.
func (g *Game) nextAbilityEffectLocked(rest []ContinuousEffect, hint int) (pick, headWaitsOn int) {
	headWaitsOn = -1
	if len(rest) < 2 || g.Battlefield == nil {
		return 0, -1
	}
	for k, eff := range rest {
		if !readsAbilities(eff) {
			return k, headWaitsOn
		}
		first := -1
		if k == 0 {
			first = hint
		}
		on := g.readerWaitsOnLocked(rest, k, first)
		if on < 0 {
			return k, headWaitsOn
		}
		if k == 0 {
			headWaitsOn = on
		}
	}
	return g.firstInDependencyLoopLocked(rest), headWaitsOn
}

// readerWaitsOnLocked returns the index in rest of an effect that rest[k],
// a reader, depends on, or -1 when it depends on none: an effect whose
// trial application to the objects it applies to changes whether rest[k]
// applies to one of them, or what it does to it (CR 613.8a, ADR 0067
// Decision 1, on the same fixed base). rest[first] is tried before the
// others when it is in range; the walk stops at the first dependency.
func (g *Game) readerWaitsOnLocked(rest []ContinuousEffect, k, first int) int {
	reader := rest[k]
	waitsOn := func(j int) bool {
		idxs := g.effectTargetIndexesLocked(rest[j])
		return len(idxs) > 0 && g.trialChangesLocked(reader, rest[j], idxs)
	}
	if first >= 0 && first < len(rest) && first != k && waitsOn(first) {
		return first
	}
	for j := range rest {
		if j == k || j == first {
			continue
		}
		if waitsOn(j) {
			return j
		}
	}
	return -1
}

// trialChangesLocked is one probe of dependencyEdgesLocked, for one pair:
// does applying `other` to the objects at idxs change what `reader` would
// do to any of them?
//
// The fixed base is cloned, and the reader run on it, only for an object
// the reader applies to before the trial: one it applies to in neither
// world is the same answer, and one it applies to in only one is a
// difference whatever the output. That is most of the objects a printed
// keyword row or a lord applies to, and it keeps the probe's allocations
// to the trial's own clones.
func (g *Game) trialChangesLocked(reader, other ContinuousEffect, idxs []int) bool {
	var bases []Characteristic
	var outs []Characteristic
	applied := make([]bool, len(idxs))
	for t, idx := range idxs {
		target := &g.Battlefield.Cards[idx]
		if !reader.AppliesTo(target, g) {
			continue
		}
		if bases == nil {
			bases = make([]Characteristic, len(idxs))
			outs = make([]Characteristic, len(idxs))
		}
		applied[t] = true
		bases[t] = target.effective.clone()
		outs[t] = g.applyOnFixedBase(reader, target, bases[t])
	}
	restore := g.installTrialLocked(other, idxs)
	defer restore()
	for t, idx := range idxs {
		target := &g.Battlefield.Cards[idx]
		applies := reader.AppliesTo(target, g)
		if applies != applied[t] {
			return true
		}
		if applies && !sameCharacteristic(outs[t], g.applyOnFixedBase(reader, target, bases[t])) {
			return true
		}
	}
	return false
}

// firstInDependencyLoopLocked is CR 613.8b's last sentence: "If several
// dependent effects form a dependency loop, then this rule is ignored and
// the effects in the dependency loop are applied in timestamp order." It
// returns the earliest effect in `rest` that is IN a loop — one that
// depends, through the relation, on itself. An effect that merely waits
// on a loop is not in it, and keeps waiting until the loop has applied.
// Index 0, the earliest, when the relation shows no loop at all, which
// nextAbilityEffectLocked never asks for.
func (g *Game) firstInDependencyLoopLocked(rest []ContinuousEffect) int {
	edges := g.dependencyEdgesLocked(rest)
	if edges == nil {
		return 0
	}
	for i := range rest {
		if dependsOnItself(edges, i) {
			return i
		}
	}
	return 0
}

// dependsOnItself reports whether effect i reaches itself along the
// "depends on" edges: whether it is in a dependency loop.
func dependsOnItself(edges [][]bool, i int) bool {
	seen := make([]bool, len(edges))
	stack := []int{i}
	for len(stack) > 0 {
		at := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		for j, dep := range edges[at] {
			if !dep {
				continue
			}
			if j == i {
				return true
			}
			if !seen[j] {
				seen[j] = true
				stack = append(stack, j)
			}
		}
	}
	return false
}

// moveToFront brings rest[k] to rest[0], keeping the others in order.
func moveToFront(rest []ContinuousEffect, k int) {
	eff := rest[k]
	copy(rest[1:k+1], rest[:k])
	rest[0] = eff
}
