package game

import "github.com/google/uuid"

// attack_haste_waiver.go — "All creatures can attack your opponents and
// planeswalkers your opponents control as though those creatures had haste"
// (Frenzied Saddlebrute).
//
// The sibling of defender_waivers.go (#1936) for summoning sickness. The
// source's layer-6 static writes its controller onto every creature it
// applies to (Characteristic.AttackHasteFor); the attacker-side half,
// hasAttackHasteWaiver, lets a sick creature through AttackerEligible, and
// the per-target half, sicknessRefusalLocked, refuses it unless the target is
// one the waiver names. The creature does not gain haste: nothing but the
// declaration is lifted, so its {T} abilities stay sick (CR 302.6) and
// "has haste" readers see none.

// hasAttackHasteWaiver reports whether c can attack anything at all as though
// it had haste. The attacker-side half: which targets is
// sicknessRefusalLocked's question.
func hasAttackHasteWaiver(c *Card) bool {
	return c != nil && len(c.Effective().AttackHasteFor) > 0
}

// attackHasteWaivesLocked reports whether one of c's waivers lets it attack
// `target` as though it had haste: a player other than the waiver's holder, or
// a planeswalker the holder does not control. Players who have left the game
// and battles are not named by the printed text.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) attackHasteWaivesLocked(c *Card, target uuid.UUID) bool {
	for _, holder := range c.Effective().AttackHasteFor {
		switch g.classifyAttackTargetLocked(target) {
		case AttackTargetPlayer:
			if target != holder {
				return true
			}
		case AttackTargetPlaneswalker:
			if walker := findBattlefieldCard(g, target); walker != nil && walker.Controller != holder {
				return true
			}
		}
	}
	return false
}

// sicknessRefusalLocked is CR 302.6 asked of one target: nil for a creature
// that is not summoning sick or whose waiver covers `target`, ErrSummoningSick
// otherwise.
//
// Caller must hold g.mu with fresh layers.
func (g *Game) sicknessRefusalLocked(attacker *Card, target uuid.UUID) error {
	if !HasSummoningSickness(attacker) || g.attackHasteWaivesLocked(attacker, target) {
		return nil
	}
	return ErrSummoningSick
}
