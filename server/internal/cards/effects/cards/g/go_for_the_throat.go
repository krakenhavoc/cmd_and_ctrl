package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Go for the Throat — Instant {1}{B} (EDHREC rank 375):
//
//	"Destroy target nonartifact creature."
//
// Doom Blade with the clause swapped from colour to type. Artifact()
// reads the post-layer type set, so a creature animated into an
// artifact is correctly not a legal target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f092562-9e17-43cd-aeb8-d0567f99363e",
		Name:         "Go for the Throat",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		Targets:      effects.TargetCreature("target nonartifact creature", effects.Not(effects.Artifact())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return effects.DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
