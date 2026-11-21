package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tests for #2391: a clause printed after a draw ("draw two cards, then
// discard two") waits for a draw that paused on a prompt (CR 608.2c).
// Seat 1 draws, as in draw_instead_test.go.

func DiscardPromptFor(g *game.Game, p *game.Player) *game.PendingChoice {
	return LatestChoiceOfKindFor(g, game.PendingChoiceChooseCards, p.ID)
}

func StartLoot(g *game.Game, p *game.Player, draw, discard int) {
	g.WithWriteLock(func() {
		if err := effects.DrawThenDiscard(g, p.ID, uuid.Nil, draw, discard, ""); err != nil {
			panic(err)
		}
	})
}
