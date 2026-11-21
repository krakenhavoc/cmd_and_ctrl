package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	enters := effects.WhenThisEnters("Rogue Artificial Intelligence — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = effects.ChooseOne(
		effects.ModeDoing("You gain 3 life.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
		}),
		effects.ModeDoing("Surveil 2.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.Surveil{Player: item.Controller, N: 2}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:     "4849f2f8-0c0e-470d-a4f5-76547a9570a0",
		Name:         "Rogue Artificial Intelligence",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated:    []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
