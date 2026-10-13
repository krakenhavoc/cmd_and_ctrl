package game

import "github.com/google/uuid"

// exile_with_counter.go — "exile that card with a <kind> counter on it
// instead of putting it into your graveyard as it resolves" (Goliath
// Daydreamer's dream counters).
//
// The replacement is the spell's own resolution exit, the same slot
// buyback, rebound and the Adventure exile occupy
// (resolution_exits.go), so it competes with them under CR 616.1 and
// never fires for a countered or fizzled spell. The card is marked
// while it is still on the stack; the counter goes on once the card is
// in exile.

// MarkSpellExilesWithCounterForEffect marks the spell `spellID` so that
// it is exiled with one `kind` counter on its card instead of going to
// its owner's graveyard as it resolves. Reports false when the ID is
// not a spell still on the stack, or is a copy (a copy is not a card,
// CR 707.10a, and ceases to exist instead).
//
// Caller must hold g.mu.
func (g *Game) MarkSpellExilesWithCounterForEffect(spellID uuid.UUID, kind string) bool {
	_, item, ok := g.stolenSpellLocked(spellID)
	if !ok || item.IsCopy || kind == "" {
		return false
	}
	item.ExileWithCounter = kind
	return true
}
