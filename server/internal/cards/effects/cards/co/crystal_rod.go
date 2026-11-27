package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crystal Rod — Artifact {1} (Eighth Edition, S82 OLD batch 609, #5607):
//
//	"Whenever a player casts a blue spell, you may pay {1}. If you do, you gain
//	 1 life."
//
// Any player's blue spell triggers it, its controller's included
// (OldB609PlayerCastsMayPayGainLife); the {1} is asked as the trigger
// resolves and the life is gained only if it is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e68bc048-1009-46a5-97d6-ec77a18067da",
		Name:         "Crystal Rod",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB609PlayerCastsMayPayGainLife("Crystal Rod", "U", crystalRodPaid),
		},
	})
}

var crystalRodPaid = effects.PayUnlessBranches("pay-unless/s82-old-b609-crystal-rod-life",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	}, nil)
