package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	etb := WhenThisEnters("Archive Arbiter — choose one", func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseOne(
		ModeDoing("Destroy target noncreature, nonland permanent.",
			TargetPermanent("target noncreature, nonland permanent", Noncreature(), Nonland()),
			DestroyTheModesTarget),
		ModeDoing("You gain 4 life.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return GainLife{Amount: 4}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "36cb3d18-019c-464d-b591-8c87e1b81f31",
		Name:            "Archive Arbiter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{etb},
	})
}
