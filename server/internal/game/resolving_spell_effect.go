package game

import "github.com/google/uuid"

// resolving_spell_effect.go — a read of the spell whose resolution is in
// flight, for a trigger that watches what a spell does as it resolves
// (Imodane, the Pyrohammer: "whenever an instant or sorcery spell you
// control that targets only a single creature deals damage to that
// creature").
//
// A resolving spell has left StackMeta (resolveTopOfStackLocked deletes the
// entry first), but the damage it deals is emitted while the resolution is
// still in flight, and the slot resolving_item.go keeps holds the item and
// the card exactly as the stack had them. This is the same read the copy
// legs use, handed to a catalog condition by value.

// ResolvingSpellForEffect returns the card and the stack item of the spell
// resolving now, when spellID names it. False for any other ID, for an
// ability, and once play has moved on (the slot is cleared at the next
// resolution or step). Both results are copies: the caller can't change the
// spell. Caller must hold g.mu (it is for already-locked card callbacks).
func (g *Game) ResolvingSpellForEffect(spellID uuid.UUID) (Card, StackItem, bool) {
	if spellID == uuid.Nil {
		return Card{}, StackItem{}, false
	}
	card, item, ok := g.resolvingSpellLocked(spellID)
	if !ok || item == nil {
		return Card{}, StackItem{}, false
	}
	out := *item
	out.Targets = append([]TargetRef(nil), item.Targets...)
	return card, out, true
}
