package game

// granted_statics.go lifts ADR 0093 Decision 10 for one kind of granted
// static ability (#2562, ADR 0093 amendment 2026-10-10): a power and
// toughness MODIFY, CR 613.4c's layer 7c. The Irencrag gains "Equipped
// creature gets +3/+3"; Gemcutter Buccaneer's Treasures have "Equipped
// creature gets +2/+0".
//
// Why Decision 10 refused it: the pass gathers every static before it
// walks layer 1, and a static that only exists once layer 6 has granted
// it was never gathered. Decision 10 also said what doing it properly
// needs, and this is that: a second gather after the layer-6 bucket,
// restricted to layer 7. Nothing in layers 1-6 can be granted this way,
// because the object's own layers 1-6 have already been applied by the
// time it has the ability; a granted static that would change a type or
// grant a keyword is a CR 613 question the engine still does not open.
//
// Restricted further to 7c because 7c is the one sublayer where the
// effect's timestamp cannot change the answer: every 7c effect adds, and
// addition commutes. A granted 7a or 7b static would need CR 613.7a's
// timestamp ("the timestamp of the effect that created the ability, if
// that is later"), which GrantedAbility does not carry. No card has
// needed one.
//
// The rest is free:
//
//   - CR 613.6. A removal on the host sorted after the grant empties
//     GrantedAbilities in its own slot, so the static is never gathered;
//     one sorted before it leaves the grant, and the static applies. So
//     the gathered effect is not `live`: the grant list already says
//     whether the host still has the ability, and the host's own
//     AbilitiesRemoved (which The Irencrag's own "loses all other
//     abilities" sets) is about its OWN abilities, not this one.
//   - CR 707.2. A layer-6 grant is not copiable, so neither is the
//     static it carries.
//   - "Equipped creature" is the host's attachment (CR 301.5f), read by
//     the static's own AppliesTo with the host as its source.

// GrantedStaticProblem says why a static ability cannot be carried by a
// bundle a LAYER-6 grant names, or "" when it can. A copy grant (CR
// 707.9a) may carry any static: the copy's own statics are gathered with
// its own key, before layer 1.
func GrantedStaticProblem(s StaticAbility) string {
	switch {
	case s.Layer != Layer7PT || s.SubLayer != SubLayer7C_Modify:
		return "a layer-6 grant can give a static ability only in layer 7c (ADR 0093 Decision 10, amended 2026-10-10)"
	case s.AppliesTo == nil || s.Apply == nil:
		return "a granted static needs AppliesTo and Apply"
	case s.RemovesAbilities, s.ContinuesAfterRemoval, len(s.GrantAbilities) > 0:
		return "a granted static cannot remove or grant abilities"
	case s.ActiveWhen.IsGate():
		return "a granted static cannot declare ActiveWhen — a designation is the host's"
	case len(s.Zones) > 0 || s.AffectsSpells:
		return "a granted static functions on the battlefield over permanents only"
	case s.DependsOnHandSize || s.DependsOnLifeTotal || s.DependsOnAttackingStatus ||
		s.DependsOnSpellsCast || s.DependsOnExile || s.DependsOnManaPool:
		return "a granted static cannot declare an invalidation hint — the hints are read off the battlefield's own statics"
	}
	return ""
}

// grantedStaticEffectsLocked is the second gather: one effect per
// layer-7c static of each bundle a permanent was granted in this pass,
// with the host as its source, ordered at the host's timestamp. Called
// once the layer-6 bucket has been applied, so GrantedAbilities is final.
// Nil, and no allocation, on a board where nothing was granted a static.
//
// Caller must hold g.mu in write mode (the layer pass does).
func (g *Game) grantedStaticEffectsLocked() []ContinuousEffect {
	if g.Battlefield == nil {
		return nil
	}
	var out []ContinuousEffect
	for i := range g.Battlefield.Cards {
		host := &g.Battlefield.Cards[i]
		for _, ga := range layeredGrants(host) {
			d := catalogDef(ga.Key)
			if d == nil {
				continue
			}
			for _, s := range d.Static {
				// Registration refuses any other shape in a bundle a
				// layer-6 grant names; skipping it here as well keeps a
				// copy-grant bundle granted by mistake from applying in
				// the wrong layer.
				if GrantedStaticProblem(s) != "" {
					continue
				}
				out = append(out, staticContinuousEffect{
					ability:   s,
					source:    host,
					timestamp: host.layerTimestamp(),
				})
			}
		}
	}
	return out
}
