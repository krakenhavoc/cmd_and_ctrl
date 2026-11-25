package game

// attack_needs_keyword.go — "Creatures without flying or islandwalk can't
// attack" (Stormtide Leviathan; the same shape as Moat's "Creatures without
// flying can't attack").
//
// The restricting permanent's layer-6 static writes the keyword tokens a
// creature needs one of (Characteristic.CantAttackWithoutAnyOf) onto every
// creature it applies to, and this fold turns that into the CantAttack bit
// once every layer has run. It is a fold for the reason power_floor.go is
// one: whether a creature has flying is a layer-6 result, and a layer-6
// bucket sorted by timestamp would ask the question before a later flying
// grant had applied, or before the creature's own ability removal had run.
// A restriction has no layer of its own (restrictions.go), so applying the
// bit after the last layer is the rules' own order (CR 613.11). From the bit
// on, nothing is new: AttackerEligible and the legal-move enumerator read
// Restrictions, so they agree on the declaration (CR 508.1c).

// foldAttackNeedsKeywordLocked adds CantAttack to every permanent that was
// told it needs one of some keywords to attack and whose finished ability
// list has none of them.
//
// Caller must hold g.mu in write mode (the layer pass).
func (g *Game) foldAttackNeedsKeywordLocked() {
	if g.Battlefield == nil {
		return
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.effective == nil || len(c.effective.CantAttackWithoutAnyOf) == 0 {
			continue
		}
		has := false
		for _, kw := range c.effective.CantAttackWithoutAnyOf {
			if containsKeyword(c.effective.Abilities, kw) {
				has = true
				break
			}
		}
		if !has {
			c.effective.Restrictions |= CantAttack
		}
	}
}
