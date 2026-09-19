package game

import "github.com/google/uuid"

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
// back to timestamp order (613.8c's escape).
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
	// started records that a live source's continuous effect has
	// already applied to an object — CR 613.6's "starts to apply".
	started map[sourceTarget]bool
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
			st.started = make(map[sourceTarget]bool, 8)
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

// recordApplied notes that `src`'s effect has applied to `target`.
func (st *layerPassState) recordApplied(src *Card, target *Card) {
	if st == nil || !st.track || src == nil {
		return
	}
	st.started[sourceTarget{source: src.InstanceID, target: target.InstanceID}] = true
}

// stillApplies answers CR 613.6 for an effect whose source has
// already had its abilities removed in this pass: may it apply to
// this object anyway?
//
// Yes in exactly one case — it is the later-layer half of a
// continuous effect that had already started applying to this same
// object before the removal landed. "If an effect starts to apply in
// one layer and/or sublayer, it will continue to be applied to the
// same set of objects in each other applicable layer and/or sublayer,
// even if the ability generating the effect is removed during this
// process." Humility taking its own abilities away in layer 6 and
// still making every creature 1/1 in layer 7b is that sentence.
//
// No in every other case, including within the removal's own bucket,
// where timestamp order has already decided (two Song of the Dryads
// enchanting each other) and CR 613.6 has nothing to say because
// nothing has "started" in a later layer yet.
func (st *layerPassState) stillApplies(src *Card, target *Card, bucketIndex int, continues bool) bool {
	if st == nil || !st.track || src == nil || !continues {
		return false
	}
	at, ok := st.removedAt[src.InstanceID]
	if !ok || bucketIndex <= at {
		return false
	}
	return st.started[sourceTarget{source: src.InstanceID, target: target.InstanceID}]
}

// applyBucketLocked applies one timestamp-ordered bucket, reordering
// it by CR 613.8 dependency first where that layer asks for it.
//
// The loop is CR 613.8b + 613.8c literally: take the first effect
// that depends on nothing else still waiting, apply it, and ask
// again. `firstIndependent` returning -1 is the loop case — every
// remaining effect depends on another — and CR 613.8c says to give up
// on dependency there and fall back to timestamp order, which is what
// index 0 is.
//
// The early `break` is the fast path and the reason this is
// affordable: once no dependency remains among the effects still to
// apply, the rest is plain timestamp order and no further trials are
// run. On a board with no dependency at all that is one relation
// computation per bucket and nothing else.
func (g *Game) applyBucketLocked(bucket []ContinuousEffect, l Layer, bucketIndex int, st *layerPassState) {
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
	for _, eff := range bucket {
		g.applyOneEffectLocked(eff, bucketIndex, st)
	}
}

// firstIndependent returns the index of the first effect that depends
// on no other effect in the set, or -1 when every one of them does —
// the CR 613.8c loop.
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
		c.AbilitiesRemoved = true
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
		a.Controller == b.Controller &&
		a.Restrictions == b.Restrictions &&
		sameStringSlice(a.Types, b.Types) &&
		sameStringSlice(a.Subtypes, b.Subtypes) &&
		sameStringSlice(a.Supertypes, b.Supertypes) &&
		sameStringSlice(a.Colors, b.Colors) &&
		sameStringSlice(a.Abilities, b.Abilities)
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
