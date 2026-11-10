package game

import "github.com/google/uuid"

// cast_bottom_follow_up.go — "If that spell would be put into a graveyard,
// put it on the bottom of its owner's library instead" for a cast made at the
// card's PRINTED cost (#4452, Kylox's Voltstrider).
//
// Quintorius, Loremaster's clause rides the permission's own priced offer
// (CastPermission.BottomInsteadOfGraveyard, stamped through the synthesised
// alternative cost), which is the right shape when the permission prices the
// cast. Kylox's Voltstrider lets you cast the spell paying its mana cost, and
// a permission that names no price opens the zone without taking a claim
// away: the card may still be cast for any alternative cost it prints
// (permissionClaimOK, CR 118.9a). So the clause is the permission's cast
// follow-up instead: it runs once the spell is on the stack, and marks that
// spell. It applies to the spell cast through the permission only, never to
// the card's own cast from the zone.

// SetSpellBottomInsteadOfGraveyardForEffect marks the spell `spellID` on the
// stack so that, if it would be put into a graveyard as it resolves or is
// countered, it goes to the bottom of its owner's library instead (StackItem.
// BottomInsteadOfGraveyard). Nothing happens for a spell that is not on the
// stack, or a copy (a copy is not a card and ceases to exist, CR 707.10a).
//
// Caller must hold g.mu (write).
func (g *Game) SetSpellBottomInsteadOfGraveyardForEffect(spellID uuid.UUID) {
	item := g.StackMeta[spellID]
	if item == nil || item.IsCopy {
		return
	}
	item.BottomInsteadOfGraveyard = true
}
