package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pop Quiz — Instant {2}{U} (EDHREC rank 25086):
//
//	"Draw a card.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// The draw can ask a question (a dredge offer), so learn is a step that
// waits for the answer (CR 608.2c): a player who dredges is not asked
// what to discard until the dredge has settled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dda9ac6c-1ab2-4f77-b499-f03d587163d8",
		Name:         "Pop Quiz",
		Purpose:      game.Purpose{Learn: 1},
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.DrawCards{N: 1}, effects.Learn{}),
	})
}
