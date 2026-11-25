package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b042_helpers.go — shared pieces for the Strixhaven: School of Mages
// batch 042 (S82, #5502).

// OldB42OpponentHandLead is the largest number of cards by which a single
// opponent of `you` who is still in the game has more cards in hand than
// `you` (negative when every opponent has fewer). Teachings of the Archaics
// reads it as it resolves; the spell has already left the hand.
func OldB42OpponentHandLead(g *game.Game, you uuid.UUID) int {
	mine := 0
	if p := g.PlayerByIDForEffect(you); p != nil && p.Hand != nil {
		mine = len(p.Hand.Cards)
	}
	lead, any := 0, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == you || p.Hand == nil {
			continue
		}
		if d := len(p.Hand.Cards) - mine; !any || d > lead {
			lead, any = d, true
		}
	}
	return lead
}
