package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harrowing Journey — Sorcery {4}{B} (Dark Ascension, #5571):
//
//	"Target player draws three cards and loses 3 life."
//
// Blood Pact's body with three and 3: the target player draws, then loses
// the life (a loss, not damage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "70f3d545-d114-485d-8d8b-4a93b2393c41",
		Name:         "Harrowing Journey",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3, LifeLoss: 3}),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.B32TargetPlayerDrawsAndLosesLife(item, ctx, 3, 3)
		},
	})
}
