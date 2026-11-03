package game

import "github.com/google/uuid"

// SpellXValueForEffect is the value of X announced for the spell
// `spellID` (CR 107.3a), for an effect that names the spell without
// targeting it: Brass Infiniscope's "when you next cast a spell with {X}
// in its mana cost this turn, you draw a card and gain half X life"
// (#4834). While the spell is on the stack it is read off its stack
// item; once the spell has left without resolving (countered in
// response) it is read off the turn's last-known record of it
// (stack_lki.go), because the effect uses the spell as it last existed
// (CR 608.2h). False when the turn has no record of the spell. A
// negative announcement reads as zero.
//
// Caller must hold g.mu.
func (g *Game) SpellXValueForEffect(spellID uuid.UUID) (int, bool) {
	if item := g.StackItemForEffect(spellID); item != nil && item.Kind == StackItemSpell {
		return max(item.XValue, 0), true
	}
	if _, item, ok := g.lastKnownSpellLocked(spellID); ok {
		return max(item.XValue, 0), true
	}
	return 0, false
}
