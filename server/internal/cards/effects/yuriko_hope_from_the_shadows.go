package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yuriko, Hope from the Shadows — Legendary Creature — Human Ninja {U},
// 1/1:
//
//	"Flash
//	 When Yuriko enters, choose one —
//	 • Target creature gets -X/-0 until end of turn, where X is the
//	   number of cards in your graveyard.
//	 • Surveil 2."
//
// A modal enters trigger (Divining Duelist's shape): the mode and the
// first bullet's target are chosen as the trigger goes on the stack.
// X is read as the trigger resolves (CR 608.2h), counting the
// controller's own graveyard.
//
// No simplification.
func init() {
	yuriko := WhenThisEnters("Yuriko — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	yuriko.Modes = ChooseOne(
		ModeDoing("Target creature gets -X/-0 until end of turn, where X is the number of cards in your graveyard.",
			TargetCreature("target creature"),
			rfCreatureBModeTargetDoes(func(ctx *Context, id uuid.UUID) error {
				x := b31GraveyardSize(ctx.Game, ctx.Controller())
				return BoostUntilEOT{
					Target: id, Power: -x,
					Label: "Yuriko, Hope from the Shadows — -X/-0",
				}.Apply(ctx)
			})),
		ModeDoing("Surveil 2.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return Surveil{N: 2}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "983f4fde-ab04-44f0-a294-c3b472de0b19",
		Name:            "Yuriko, Hope from the Shadows",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{yuriko},
	})
}
