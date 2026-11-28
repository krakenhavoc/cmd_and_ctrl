package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Anarchy — Sorcery {2}{R}{R} (Ice Age, S82 OLD, #5636):
//
//	"Destroy all white permanents."
//
// Virtue's Ruin's colour-gated sweep over every permanent type (CR 105: a white permanent is read
// through the layers; lands are colourless and stay).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b91ed762-9232-4b77-abf4-97f66637f49c",
		Name:         "Anarchy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve:    effects.S82OldB842DestroyAllWhitePermanents,
	})
}
