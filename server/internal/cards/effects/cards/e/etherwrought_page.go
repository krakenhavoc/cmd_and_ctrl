package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etherwrought Page — Artifact {1}{W}{U}{B} (Alara Reborn, #5586):
//
//	"At the beginning of your upkeep, choose one —
//	 • You gain 2 life.
//	 • Surveil 1.
//	 • Each opponent loses 1 life."
//
// Mister Hyde's shape: a modal upkeep trigger whose mode is chosen as it goes on
// the stack (CR 700.2). Surveil is the engine's own look-and-sort.
//
// No simplification.
func init() {
	page := effects.AtYourUpkeep("Etherwrought Page — choose one", func(_ *game.Game, _ *game.StackItem) error { return nil })
	page.Modes = effects.ChooseOne(
		effects.ModeDoing("You gain 2 life.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.GainLife{Amount: 2}.Apply(ctx)
		}),
		effects.ModeDoing("Surveil 1.", nil, func(_ *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.Surveil{Player: ctx.Controller(), N: 1}.Apply(ctx)
		}),
		effects.ModeDoing("Each opponent loses 1 life.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.EachOpponentLosesLifeNow(ctx.Game, item, 1)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "f1c87913-9374-41d7-ae70-444d427290b4",
		Name:         "Etherwrought Page",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{page},
	})
}
