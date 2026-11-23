package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seedpod Caretaker — Creature — Phyrexian Cleric {2}{W}, 2/2 (March of
// the Machine):
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on target artifact or creature you control.
//	 • Transform target Incubator token you control."
//
// A modal trigger whose bullets each carry their own target (CR 700.2):
// the mode is chosen as the trigger goes on the stack, and the target is
// re-checked as it resolves.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Seedpod Caretaker — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on target artifact or creature you control.",
			effects.TargetPermanent("target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()),
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.PlusOneCounterOnTheTarget(ctx.Game, item)
			}),
		effects.ModeDoing("Transform target Incubator token you control.",
			effects.TargetIncubatorYouControl(),
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.TransformTargetIncubator(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "257fbe05-b2cb-493c-a116-49f9000b2384",
		Name:         "Seedpod Caretaker",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
