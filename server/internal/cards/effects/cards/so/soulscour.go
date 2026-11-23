package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulscour — Sorcery {7}{W}{W}{W} (Darksteel, #5605):
//
//	"Destroy all nonartifact permanents."
//
// Planar Cleansing's sweep with the artifact exception instead of the
// land one: lands are destroyed too. The sweep is declared as the
// nonland-permanent class with Partial set, since the class names no
// lands and the card spares artifacts (ADR 0126 §4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7644325-8a0f-4f40-bab6-518df7ac1d14",
		Name:         "Soulscour",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.Not(effects.Artifact())}.Apply(ctx)
		},
	})
}
