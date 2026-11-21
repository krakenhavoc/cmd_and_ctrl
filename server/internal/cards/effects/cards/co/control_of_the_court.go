package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Control of the Court — Sorcery {1}{R} (Mystery Booster 2, #5948):
//
//	"Draw four cards, then discard three cards at random."
//
// Four draws, then three random discards from the hand as it stands once the
// draws have settled (CR 608.2c): the discard waits for any prompt a draw
// asks. The discards are the controller's random stream (CR 701.9b), so the
// player chooses nothing; a hand of fewer than three discards what it has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "755034b1-1859-4f18-9394-c0e1914a5b83",
		Name:         "Control of the Court",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4, Discards: 3},
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 4},
			effects.DiscardCards{N: 3},
		),
	})
}
