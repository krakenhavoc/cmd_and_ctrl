package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Obliterate — Sorcery {6}{R}{R} (Eighth Edition, #5607):
//
//	"This spell can't be countered.
//	 Destroy all artifacts, creatures, and lands. They can't be
//	 regenerated."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47496941-8007-45a0-a981-76c683cbddc1",
		Name:            "Obliterate",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve:       effects.S82OldB842DestroyArtifactsCreaturesAndLands,
	})
}
