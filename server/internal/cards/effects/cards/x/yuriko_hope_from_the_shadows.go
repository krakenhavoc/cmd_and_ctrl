package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	yuriko := effects.WhenThisEnters("Yuriko — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	yuriko.Modes = effects.ChooseOne(
		effects.ModeDoing("Target creature gets -X/-0 until end of turn, where X is the number of cards in your graveyard.",
			effects.TargetCreature("target creature"),
			effects.RfCreatureBModeTargetDoes(func(ctx *effects.Context, id uuid.UUID) error {
				x := effects.B31GraveyardSize(ctx.Game, ctx.Controller())
				return effects.BoostUntilEOT{
					Target: id, Power: -x,
					Label: "Yuriko, Hope from the Shadows — -X/-0",
				}.Apply(ctx)
			})),
		effects.ModeDoing("Surveil 2.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.Surveil{N: 2}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "983f4fde-ab04-44f0-a294-c3b472de0b19",
		Name:            "Yuriko, Hope from the Shadows",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{yuriko},
	})
}
