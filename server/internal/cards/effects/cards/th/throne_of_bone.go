package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throne of Bone — Artifact {1} (Eighth Edition, S82 OLD batch 609, #5607):
//
//	"Whenever a player casts a black spell, you may pay {1}. If you do, you gain
//	 1 life."
//
// Any player's black spell triggers it, its controller's included
// (OldB609PlayerCastsMayPayGainLife); the {1} is asked as the trigger
// resolves and the life is gained only if it is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f73c7edf-ed2c-41e8-ac83-c83ddd543f14",
		Name:         "Throne of Bone",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB609PlayerCastsMayPayGainLife("Throne of Bone", "B", throneOfBonePaid),
		},
	})
}

var throneOfBonePaid = effects.PayUnlessBranches("pay-unless/s82-old-b609-throne-of-bone-life",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	}, nil)
