package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b13_helpers.go — shared reads for the March of the Machine batch 13
// cards (S82, #5477). Everything here is private to that batch.

// s82OpponentHasGraveyardCards is "an opponent has N or more cards in their
// graveyard": any other player still in the game whose graveyard holds at
// least n cards, read live.
func s82OpponentHasGraveyardCards(g *game.Game, you uuid.UUID, n int) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == you || p.Eliminated || p.Graveyard == nil {
			continue
		}
		if len(p.Graveyard.Cards) >= n {
			return true
		}
	}
	return false
}

// s82AnOpponentCastBlueOrBlackSpellThisTurn is "an opponent cast a blue and/or
// black spell this turn", from the per-turn cast tally (a spell's colours as
// it was cast, whatever became of it since).
func s82AnOpponentCastBlueOrBlackSpellThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, p := range g.Seats {
		if p == nil || p.ID == you {
			continue
		}
		for _, col := range g.CastTallyFor(p.ID).SpellColorsCast() {
			if col == "U" || col == "B" {
				return true
			}
		}
	}
	return false
}
