package game

import "github.com/google/uuid"

// ExileTopUntilForEffect is "<a player> exiles cards from the top of their
// library until they exile a card that <matches>" with nothing done to the
// cards afterwards (Umbris, Fear Manifest): every card walked over, and the
// hit, stay in exile. It is the walk cascade and discover use, with the hit
// test supplied by the caller. A library that runs out first is exiled
// whole. The hit test is called during the walk only and is never stored.
//
// Caller must hold g.mu (write).
func (g *Game) ExileTopUntilForEffect(library uuid.UUID, hit func(Card) bool) error {
	_, _, err := g.exileUntilLocked(library, hit)
	return err
}
