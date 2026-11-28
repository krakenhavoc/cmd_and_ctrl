package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jokulhaups — Sorcery {4}{R}{R} (Ice Age, S82 OLD, #5636):
//
//	"Destroy all artifacts, creatures, and lands. They can't be regenerated."
//
// Obliterate without the uncounterability: one shared body (S82OldB842DestroyArtifactsCreaturesAndLands).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7bdfe368-353c-4f8d-851f-05b7ea9f905d",
		Name:         "Jokulhaups",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve:    effects.S82OldB842DestroyArtifactsCreaturesAndLands,
	})
}
