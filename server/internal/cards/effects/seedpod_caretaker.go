package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	trigger := WhenThisEnters("Seedpod Caretaker — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on target artifact or creature you control.",
			TargetPermanent("target artifact or creature you control", Or(Artifact(), Creature()), YouControl()),
			func(item *game.StackItem, ctx *Context, _ int) error {
				return PlusOneCounterOnTheTarget(ctx.Game, item)
			}),
		ModeDoing("Transform target Incubator token you control.",
			TargetIncubatorYouControl(),
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return transformTargetIncubator(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "257fbe05-b2cb-493c-a116-49f9000b2384",
		Name:         "Seedpod Caretaker",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{trigger},
	})
}
