package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Creeping Corrosion — Sorcery {2}{G}{G} (Mirrodin Besieged,
// #5576):
//
//	"Destroy all artifacts."
//
// Pulverize's sweep without its alternative cost: every artifact on the
// battlefield as it resolves, the caster's included. A regeneration shield
// still saves one (CR 701.19a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33eefb22-51a6-4630-9bd2-0895997f6f33",
		Name:         "Creeping Corrosion",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Artifact()}.Apply(ctx)
		},
	})
}
