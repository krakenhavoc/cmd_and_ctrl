package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apothecary Stomper — Creature — Elephant {4}{G}{G}, 4/4 (Foundations,
// #4465):
//
//	"Vigilance
//	 When this creature enters, choose one —
//	 • Put two +1/+1 counters on target creature you control.
//	 • You gain 4 life."
//
// Charming Prince's modal enters trigger: the mode is chosen as the
// trigger goes on the stack, and the counters' target is chosen with it
// and re-checked at resolution (CR 608.2b). The Stomper itself is a
// legal target.
//
// No simplification.
func init() {
	trigger := effects.WhenThisEnters("Apothecary Stomper — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	trigger.Modes = effects.ChooseOne(
		effects.ModeDoing("Put two +1/+1 counters on target creature you control.",
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TwoPlusOneCountersOnModeTarget),
		effects.ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *effects.Context, _ int) error {
			return effects.GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
		}),
	)
	effects.Register(effects.Spec{
		OracleID:        "6a381121-9134-489c-8c7e-44be1936b0e2",
		Name:            "Apothecary Stomper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{trigger},
	})
}
