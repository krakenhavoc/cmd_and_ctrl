package game

import "github.com/google/uuid"

// library_shuffle_keep_top.go — "Then shuffle and put the last cards you
// looked at this way on top in any order" (Lim-Dûl's Vault, #6101).
//
// The cards named stay out of the shuffle: they are taken off the library,
// the rest is shuffled (ShuffleLibraryForEffect's shuffle stream, knowledge
// wipe and event), and they go back on top in the order given, the first
// ID on top. The player still knows them, as a tutor's searcher still
// knows the card it put on top after its shuffle (finishSearchLocked): the
// shuffle un-knows the cards it shuffled, not the ones it left alone.

// ShuffleLibraryKeepingOnTopForEffect shuffles playerID's library except
// `top`, then puts those cards on top in the order given, top card first.
// An ID not in that library is skipped. A missing player is a no-op, as for
// ShuffleLibraryForEffect.
//
// Caller must hold g.mu.
func (g *Game) ShuffleLibraryKeepingOnTopForEffect(playerID uuid.UUID, top []uuid.UUID) error {
	p := g.playerByIDLocked(playerID)
	if p == nil || p.Library == nil {
		return nil
	}
	kept := make([]Card, 0, len(top))
	for _, id := range top {
		c, err := p.Library.Remove(id)
		if err != nil {
			continue
		}
		kept = append(kept, c)
	}
	if err := g.ShuffleLibraryForEffect(playerID); err != nil {
		return err
	}
	for i := len(kept) - 1; i >= 0; i-- {
		c := kept[i]
		c.AddKnower(playerID)
		p.Library.InsertFromTop(c, 0)
	}
	return nil
}
