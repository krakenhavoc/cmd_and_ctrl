package game

import "github.com/google/uuid"

// exile_and_plot.go — "exile that spell instead of putting it into your
// graveyard as it resolves. If you do, it becomes plotted" (Lilah,
// Undefeated Slickshot; CR 702.170).
//
// Like exile_with_counter.go it is the spell's own resolution exit
// (resolution_exits.go): it never fires for a countered or fizzled spell,
// and it competes with the other exits under CR 616.1. The spell is marked
// while it is on the stack; it becomes plotted once the card is in exile.

// MarkSpellExilesAndPlotsForEffect marks the spell `spellID` so that it is
// exiled and plotted instead of going to its owner's graveyard as it
// resolves. Reports false when the ID is not a spell still on the stack, or
// is a copy (a copy is not a card, CR 707.10a, and ceases to exist instead).
//
// Caller must hold g.mu.
func (g *Game) MarkSpellExilesAndPlotsForEffect(spellID uuid.UUID) bool {
	_, item, ok := g.stolenSpellLocked(spellID)
	if !ok || item.IsCopy {
		return false
	}
	item.ExileAndPlot = true
	return true
}
