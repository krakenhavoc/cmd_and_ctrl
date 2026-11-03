package game

import "github.com/google/uuid"

// spell_mana_value_lki.go — "that spell's mana value" for an ability
// that names a spell without targeting it (Deekah, Fractal Theorist's
// magecraft, #4834).
//
// The trigger goes on the stack above the spell, so it usually resolves
// while the spell is still there and the read is ManaValueForEffect's,
// with the announced X counted (CR 202.3e). When the spell has left the
// stack first (countered in response), CR 608.2h says the ability uses
// the spell's last-known information: the mana value it had on the
// stack, X included, read from the record every non-resolving exit from
// the stack files (stack_lki.go). A copy that ceased to exist as it was
// countered is found the same way.

// LastKnownSpellManaValueForEffect is the mana value of the spell
// `spellID` while it is on the stack, or as it last existed there when
// it has left without resolving. ok is false when it is neither, and
// the caller decides what an object in no such state is worth. Caller
// must hold g.mu.
func (g *Game) LastKnownSpellManaValueForEffect(spellID uuid.UUID) (int, bool) {
	if g.Stack != nil && g.Stack.Contains(spellID) {
		if c, ok := g.LookupCardForEffect(spellID); ok {
			return g.ManaValueForEffect(c)
		}
	}
	card, item, ok := g.lastKnownSpellLocked(spellID)
	if !ok {
		return 0, false
	}
	cost, err := ParseCost(manaCostForValue(card))
	if err != nil {
		return 0, false
	}
	x := 0
	if item != nil && item.XValue > 0 {
		x = item.XValue
	}
	return cost.ManaValueWithX(x), true
}
