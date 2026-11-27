package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Star — Artifact {1} (Eighth Edition, S82 OLD batch 609, #5607):
//
//	"Whenever a player casts a red spell, you may pay {1}. If you do, you gain
//	 1 life."
//
// Any player's red spell triggers it, its controller's included
// (OldB609PlayerCastsMayPayGainLife); the {1} is asked as the trigger
// resolves and the life is gained only if it is paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e9ec67e1-7064-44d4-a1ed-04b7893ffb15",
		Name:         "Iron Star",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB609PlayerCastsMayPayGainLife("Iron Star", "R", ironStarPaid),
		},
	})
}

var ironStarPaid = effects.PayUnlessBranches("pay-unless/s82-old-b609-iron-star-life",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.GainLife{Player: ctx.Controller(), Amount: 1}.Apply(ctx)
	}, nil)
