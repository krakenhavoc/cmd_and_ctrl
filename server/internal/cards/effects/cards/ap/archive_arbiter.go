package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archive Arbiter — Artifact Creature — Sphinx {6}, 4/4:
//
//	"Flying
//	 When this creature enters, choose one —
//	 • Destroy target noncreature, nonland permanent.
//	 • You gain 4 life."
//
// A modal enters trigger: the mode (and the target of the first mode)
// is chosen as the ability goes on the stack (CR 603.3c). With no
// legal target for the destroy mode only the life mode is offered.
//
// No simplification.
func init() {
	etb := effects.WhenThisEnters("Archive Arbiter — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = effects.ChooseOne(
		effects.ModeDoing("Destroy target noncreature, nonland permanent.",
			effects.TargetPermanent("target noncreature, nonland permanent", effects.Noncreature(), effects.Nonland()),
			effects.DestroyTheModesTarget),
		effects.ModeDoing("You gain 4 life.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GainLife{Amount: 4}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "36cb3d18-019c-464d-b591-8c87e1b81f31",
		Name:            "Archive Arbiter",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
