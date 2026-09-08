package game

import "github.com/google/uuid"

// sacrifice.go — S21 sub-PR 1: sacrifice as a first-class engine
// operation (CR 701.17). A permanent is sacrificed by its
// controller, as a cost (Goblin Bombardment, Treasure's mana
// ability) or as an effect's instruction. It is NOT destroyed:
// indestructible and regeneration don't apply, and a replacement
// keyed on destruction never sees it.
//
// The mechanics are deliberately thin — the permanent takes the
// ordinary route to its owner's graveyard, so dies-triggers, the
// CR 903.9 commander-zone replacement and LKI all keep working
// untouched. The only new thing is EventSacrifice, emitted while
// the card is still on the battlefield so "whenever you sacrifice"
// payoffs can read its characteristics before it moves.

// sacrificePermanentLocked emits EventSacrifice for the permanent's
// controller and routes it to its owner's graveyard. Caller must
// hold g.mu.
func (g *Game) sacrificePermanentLocked(cardID uuid.UUID) error {
	controller := g.controllerOfBattlefieldCardLocked(cardID)
	if controller == uuid.Nil {
		return ErrCardNotFound
	}
	g.EmitEvent(Event{
		Kind:   EventSacrifice,
		Actor:  controller,
		CardID: cardID,
	})
	return g.routeBattlefieldCardToOwnerGraveyardLocked(cardID)
}

// SacrificePermanent is the locking entry point: a player sacrifices
// a permanent they control. Rejects a permanent controlled by
// someone else (CR 701.17b — only a permanent's controller may
// sacrifice it) so the action layer can expose it directly.
//
// Caller must NOT hold g.mu.
func (g *Game) SacrificePermanent(playerID, cardID uuid.UUID) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.State != StateActive {
		return ErrGameNotActive
	}
	controller := g.controllerOfBattlefieldCardLocked(cardID)
	if controller == uuid.Nil {
		return ErrCardNotFound
	}
	if controller != playerID {
		return ErrCardCallerMismatch
	}
	if err := g.sacrificePermanentLocked(cardID); err != nil {
		return err
	}
	g.runStateChecksLocked()
	return nil
}
