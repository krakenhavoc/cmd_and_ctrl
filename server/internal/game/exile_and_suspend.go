package game

import "github.com/google/uuid"

// exile_and_suspend.go — "exile that card with three time counters on it
// instead of putting it into your graveyard as it resolves. Then if the
// exiled card doesn't have suspend, it gains suspend" (Gandalf of the
// Secret Fire; CR 702.62a).
//
// Like exile_with_counter.go and exile_and_plot.go it is the spell's own
// resolution exit (resolution_exits.go): it never fires for a countered or
// fizzled spell, and it competes with the other exits under CR 616.1. The
// spell is marked while it is on the stack; the counters and the granted
// suspend arrive once the card is in exile (GainSuspendForEffect).

// MarkSpellExilesAndSuspendsForEffect marks the spell `spellID` so that it
// is exiled with n time counters, and gains suspend if it has none, instead
// of going to its owner's graveyard as it resolves. Reports false when the
// ID is not a spell still on the stack, or is a copy (a copy is not a card,
// CR 707.10a, and ceases to exist instead), or n is not positive.
//
// Caller must hold g.mu.
func (g *Game) MarkSpellExilesAndSuspendsForEffect(spellID uuid.UUID, n int) bool {
	_, item, ok := g.stolenSpellLocked(spellID)
	if !ok || item.IsCopy || n <= 0 {
		return false
	}
	item.ExileAndSuspend = n
	return true
}
