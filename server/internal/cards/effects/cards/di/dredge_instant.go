package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dredge — Instant {B} (Invasion, S82 #5616):
//
//	"Sacrifice a creature or land.
//	 Draw a card."
//
// The controller chooses what to sacrifice as the spell resolves (a land
// creature counts as either); the draw follows the sacrifice, and still happens
// when they control nothing to sacrifice (CR 701.21a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efc9498a-b740-44de-aef3-01a1fa4ee49a",
		Name:         "Dredge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve:    effects.DredgeResolve,
	})
}
