package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rogue Artificial Intelligence — Artifact Creature — Construct {4}{B}{B},
// 6/5 (Star Trek, #4973):
//
//	"When this creature enters, choose one —
//	 • You gain 3 life.
//	 • Surveil 2.
//	 Basic landcycling {2}"
//
// The mode is chosen as the trigger is put on the stack (CR 700.2b).
// Basic landcycling is the typecycling activation from hand (CR 702.29e).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Rogue Artificial Intelligence — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("You gain 3 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
		}),
		ModeDoing("Surveil 2.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return Surveil{Player: item.Controller, N: 2}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "4849f2f8-0c0e-470d-a4f5-76547a9570a0",
		Name:         "Rogue Artificial Intelligence",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated:    []ActivatedAbility{BasicLandcycling("{2}")},
	})
}
