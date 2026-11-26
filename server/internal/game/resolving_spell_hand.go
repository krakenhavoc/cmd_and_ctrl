package game

import "github.com/google/uuid"

// resolving_spell_hand.go — "return <this spell> to its owner's hand" from
// inside its own resolution (Molten Birth: "flip a coin. If you win the flip,
// return Molten Birth to its owner's hand", S82 OLD batch 401, #5563).
//
// The answer to a coin call comes while the spell is still resolving (its card
// is in the stack zone until the effect, and every prompt it queued, is done),
// so the card is moved by the spell's own resolution exit, not by a return
// from the graveyard: the item's HandInsteadOfGraveyard, the flag Soulfire
// Grand Master stamps on a copied-from spell, is set on the resolving item and
// read when the card leaves the stack (resolution_exits.go). The flag is
// already part of StackItem's persisted shape, so this adds nothing to a
// snapshot.

// ReturnResolvingSpellToHandForEffect marks the spell `spellID`, which is
// resolving now, so that its card goes to its owner's hand instead of its
// graveyard as it finishes resolving. False for an ID that is not the
// resolving spell and for a copy (a copy is not a card and ceases to exist,
// CR 707.10a); nothing is marked then.
//
// Caller must hold g.mu (write).
func (g *Game) ReturnResolvingSpellToHandForEffect(spellID uuid.UUID) bool {
	if spellID == uuid.Nil {
		return false
	}
	_, item, ok := g.resolvingSpellLocked(spellID)
	if !ok || item == nil || item.IsCopy {
		return false
	}
	item.HandInsteadOfGraveyard = true
	return true
}
