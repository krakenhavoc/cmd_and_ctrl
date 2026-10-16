package game

import "github.com/google/uuid"

// granted_statics.go lifts ADR 0093 Decision 10 for a static ability a
// layer-6 grant gives a permanent: The Irencrag gains "Equipped creature
// gets +3/+3" (#2562, amendment 2026-10-10, layer 7c), and Dancer's
// Chakrams gives the equipped creature "Other commanders you control get
// +2/+2 and have lifelink" (#3784, amendment 2026-10-15, layers 6 and
// 7c).
//
// Why Decision 10 refused it: the pass gathers every static before it
// walks layer 1, and a static that only exists once layer 6 has granted
// it was never gathered. The two halves of the fix:
//
//   - LAYER 6. A granted layer-6 static comes into existence part-way
//     through the layer-6 bucket, the moment the effect that grants its
//     bundle applies. CR 613.8a says that is a dependency: the grant
//     changes the static's existence, so the static waits for it. And
//     CR 613.7a gives the static its timestamp — the host's, or the
//     grant's if that is later — which is never earlier than the grant's.
//     So the bucket loop (applyGrantingBucketLocked) inserts each such
//     static into the part of the bucket still to run, at its timestamp,
//     straight after the grant applies. A removal still waiting in the
//     bucket that applies to the HOST would end the static's existence
//     (the second CR 613.8a dependency, Decision 3's), so the static goes
//     after it too. When its turn comes the static asks whether the host
//     still has the grant (grantedStaticEffect.AppliesTo), which is what
//     a removal that applied in between decides.
//   - LAYER 7. Gathered once the layer-6 bucket and the can't-have strip
//     have run, so each host's GrantedAbilities is final, at the same
//     CR 613.7a timestamp. 7b and 7e are allowed now that the timestamp
//     is right: a granted "set" or "switch" sorts against the host's own
//     effects by it.
//
// What a granted static may NOT be (GrantedStaticProblem):
//
//   - Layers 1-5. The object's own layers 1-5 have been applied by the
//     time layer 6 gives it the ability, and CR 613.6 does not reach
//     back. No card in the catalog asks for one.
//   - 7a. A characteristic-defining ability is never granted (CR 604.3a).
//   - A removal, a grant, or a borrow. Each one changes which grants
//     exist, which would put the bucket back into a loop whose
//     termination nothing guarantees ("this creature has 'creatures you
//     control have "…"'"). Lifelink, flying and every other keyword is
//     an Apply that appends a token, not a bundle, so no printed card
//     needs a granted grant.
//
// The rest is free:
//
//   - CR 613.6. A removal on the host sorted after the grant empties
//     GrantedAbilities in its own slot, so the static no longer exists;
//     one sorted before it leaves the grant, and the static applies. So
//     the effect is not `live`: the grant list says whether the host
//     still has the ability, and the host's own AbilitiesRemoved (which
//     The Irencrag's own "loses all other abilities" sets) is about its
//     OWN abilities, not this one.
//   - CR 707.2. A layer-6 grant is not copiable, so neither is the
//     static it carries.
//   - "Equipped creature" and "other" are the host's, read by the
//     static's own AppliesTo with the host as its source (CR 301.5f).

// GrantedStaticProblem says why a static ability cannot be carried by a
// bundle a LAYER-6 grant names, or "" when it can. A copy grant (CR
// 707.9a) may carry any static: the copy's own statics are gathered with
// its own key, before layer 1.
func GrantedStaticProblem(s StaticAbility) string {
	switch {
	case !grantableLayer(s.Layer, s.SubLayer):
		return "a layer-6 grant can give a static ability only in layer 6 or in layer 7b, 7c or 7e (ADR 0093 Decision 10, amended 2026-10-10 and 2026-10-15)"
	case s.AppliesTo == nil || s.Apply == nil:
		return "a granted static needs AppliesTo and Apply"
	case s.RemovesAbilities, s.ContinuesAfterRemoval, len(s.GrantAbilities) > 0, s.Borrows != nil:
		return "a granted static cannot remove, grant or borrow abilities"
	case s.ActiveWhen.IsGate():
		return "a granted static cannot declare ActiveWhen — a designation is the host's"
	case len(s.Zones) > 0 || s.AffectsSpells || s.GrantZone != "" || len(s.AffectsZones) > 0:
		return "a granted static functions on the battlefield over permanents only"
	case s.DependsOnHandSize || s.DependsOnLifeTotal || s.DependsOnAttackingStatus ||
		s.DependsOnSpellsCast || s.DependsOnExile || s.DependsOnManaPool || s.DependsOnLibraryTop ||
		len(s.DependsOnEvents) > 0:
		return "a granted static cannot declare an invalidation hint — the hints are read off the battlefield's own statics"
	}
	return ""
}

// grantableLayer is the layer half of GrantedStaticProblem.
func grantableLayer(l Layer, sub SubLayer) bool {
	switch l {
	case Layer6Ability:
		return true
	case Layer7PT:
		return sub == SubLayer7B_Set || sub == SubLayer7C_Modify || sub == SubLayer7E_Switch
	}
	return false
}

// grantStamp is one grant of a static-bearing bundle the layer-6 bucket
// applied in this pass: which host, which bundle, which grantor, and the
// timestamp of the effect that wrote it (CR 613.7a needs that, and
// GrantedAbility does not carry it — it is pass-local, so it lives on
// the pass state and never reaches a snapshot).
type grantStamp struct {
	host    *Card
	key     string
	grantor uuid.UUID
	ts      int64
}

// grantedStaticEffect is one static of a granted bundle, bound to its
// host. Not a staticContinuousEffect, on purpose: that type's `source`
// is what the CR 613.6 silencing and Decision 3's removal pre-sort read,
// and neither is right for a granted static — whether it exists is the
// host's grant list, asked in AppliesTo.
type grantedStaticEffect struct {
	ability   StaticAbility
	host      *Card
	key       string
	grantor   uuid.UUID
	timestamp int64
}

func (e grantedStaticEffect) Layer() (Layer, SubLayer) { return e.ability.Layer, e.ability.SubLayer }
func (e grantedStaticEffect) Timestamp() int64         { return e.timestamp }
func (e grantedStaticEffect) RemovesAbilities() bool   { return false }
func (e grantedStaticEffect) ContinuesAfterRemoval() bool {
	return false
}

// AppliesTo is the static's own predicate with the host as its source,
// behind the existence test: the host still has the grant that carries
// it. For a layer-7 static that is true by construction (it was gathered
// from the final grant list); for a layer-6 one it is what a removal on
// the host that applied after the grant, but before this static's turn
// in the bucket, has decided.
func (e grantedStaticEffect) AppliesTo(target *Card, g *Game) bool {
	if e.ability.AppliesTo == nil || !hostHasGrant(e.host, e.key, e.grantor) {
		return false
	}
	return e.ability.AppliesTo(target, g, e.host)
}

func (e grantedStaticEffect) Apply(c *Characteristic, target *Card, g *Game) {
	if e.ability.Apply != nil {
		e.ability.Apply(c, target, g, e.host)
	}
}

// hostHasGrant reports whether the host's layered grants, as the pass
// has written them so far, include this bundle from this grantor.
func hostHasGrant(host *Card, key string, grantor uuid.UUID) bool {
	for _, ga := range layeredGrants(host) {
		if ga.Key == key && ga.Source == grantor {
			return true
		}
	}
	return false
}

// grantedStaticTimestamp is CR 613.7a for a granted static: the host's
// timestamp, or the grant's if that is later.
func grantedStaticTimestamp(host *Card, grantTS int64) int64 {
	return max(host.layerTimestamp(), grantTS)
}

// bundleStatics returns the statics of a granted bundle, or nil for a
// bundle with none — nearly every bundle.
func bundleStatics(key string) []StaticAbility {
	d := catalogDef(key)
	if d == nil {
		return nil
	}
	return d.Static
}

// noteGrants records, for a granting effect that has just applied to
// `target`, each bundle it named that carries a static. Called from
// applyOneEffectLocked only for an effect that declares
// GrantAbilities, so every other effect pays one nil test.
func (st *layerPassState) noteGrants(target *Card, keys []string, grantor uuid.UUID, ts int64) {
	if st == nil {
		return
	}
	for _, k := range keys {
		k = GrantKey(k)
		if k == "" || len(bundleStatics(k)) == 0 {
			continue
		}
		st.granted = append(st.granted, grantStamp{host: target, key: k, grantor: grantor, ts: ts})
	}
}

// grantTimestamp is the timestamp of the effect that wrote this grant in
// this pass. The host's own timestamp when the pass has no record, which
// is the pre-#3784 answer and cannot happen for a grant the bucket wrote.
func (st *layerPassState) grantTimestamp(host *Card, ga GrantedAbility) int64 {
	if st != nil {
		for i := len(st.granted) - 1; i >= 0; i-- {
			s := st.granted[i]
			if s.host == host && s.key == ga.Key && s.grantor == ga.Source {
				return s.ts
			}
		}
	}
	return host.layerTimestamp()
}

// applyGrantingBucketLocked applies the layer-6 bucket in order, and
// after each effect inserts the layer-6 statics of the bundles that
// effect just granted into the part of the bucket still to run (see the
// file comment). A bucket that grants no static-bearing bundle runs
// exactly as the plain loop did.
func (g *Game) applyGrantingBucketLocked(bucket []ContinuousEffect, bucketIndex int, st *layerPassState) {
	if st == nil {
		for _, eff := range bucket {
			g.applyOneEffectLocked(eff, bucketIndex, st)
		}
		return
	}
	queue := bucket
	seen := len(st.granted)
	for i := 0; i < len(queue); i++ {
		g.applyOneEffectLocked(queue[i], bucketIndex, st)
		if len(st.granted) == seen {
			continue
		}
		fresh := st.granted[seen:]
		seen = len(st.granted)
		for _, gs := range fresh {
			for _, s := range bundleStatics(gs.key) {
				if s.Layer != Layer6Ability || GrantedStaticProblem(s) != "" {
					continue
				}
				eff := grantedStaticEffect{
					ability:   s,
					host:      gs.host,
					key:       gs.key,
					grantor:   gs.grantor,
					timestamp: grantedStaticTimestamp(gs.host, gs.ts),
				}
				queue = g.insertGrantedStaticLocked(queue, i+1, eff)
			}
		}
	}
}

// insertGrantedStaticLocked puts a granted layer-6 static into the rest
// of the bucket (queue[from:]): after every effect with an earlier or
// equal timestamp (CR 613.7), and after every removal still waiting that
// applies to its host, since that removal changes whether the static
// exists (CR 613.8a, ADR 0093 Decision 3).
func (g *Game) insertGrantedStaticLocked(queue []ContinuousEffect, from int, eff grantedStaticEffect) []ContinuousEffect {
	pos := from
	for pos < len(queue) && queue[pos].Timestamp() <= eff.timestamp {
		pos++
	}
	for j := len(queue) - 1; j >= pos; j-- {
		if queue[j].RemovesAbilities() && queue[j].AppliesTo(eff.host, g) {
			pos = j + 1
			break
		}
	}
	// A fresh slice, so the bucket applyLayerLocked built is never
	// written through. Only a pass that grants a layer-6 static gets
	// here, once per static.
	out := make([]ContinuousEffect, 0, len(queue)+1)
	out = append(out, queue[:pos]...)
	out = append(out, eff)
	return append(out, queue[pos:]...)
}

// grantedStaticEffectsLocked is the second gather: one effect per
// layer-7 static of each bundle a permanent was granted in this pass,
// with the host as its source, at the CR 613.7a timestamp. Called once
// the layer-6 bucket has been applied, so GrantedAbilities is final.
// The layer-6 statics of the same bundles were applied inside the
// bucket (applyGrantingBucketLocked) and are not gathered again.
// Nil, and no allocation, on a board where nothing was granted a static.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) grantedStaticEffectsLocked(st *layerPassState) []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		host := &g.Battlefield.Cards[i]
		for _, ga := range layeredGrants(host) {
			statics := bundleStatics(ga.Key)
			if len(statics) == 0 {
				continue
			}
			ts := grantedStaticTimestamp(host, st.grantTimestamp(host, ga))
			for _, s := range statics {
				// Registration refuses any other shape in a bundle a
				// layer-6 grant names; skipping it here as well keeps a
				// copy-grant bundle granted by mistake from applying in
				// the wrong layer.
				if s.Layer != Layer7PT || GrantedStaticProblem(s) != "" {
					continue
				}
				out = append(out, grantedStaticEffect{
					ability:   s,
					host:      host,
					key:       ga.Key,
					grantor:   ga.Source,
					timestamp: ts,
				})
			}
		}
	}
	return out
}

// grantKeysOf returns the bundle keys a layer-6 effect grants, or nil —
// the one test applyOneEffectLocked makes of every effect.
func grantKeysOf(eff ContinuousEffect) ([]string, uuid.UUID) {
	sce, ok := eff.(staticContinuousEffect)
	if !ok || len(sce.ability.GrantAbilities) == 0 {
		return nil, uuid.Nil
	}
	if sce.source == nil {
		return sce.ability.GrantAbilities, uuid.Nil
	}
	return sce.ability.GrantAbilities, sce.source.InstanceID
}
