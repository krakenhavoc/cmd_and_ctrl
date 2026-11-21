package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warhost's Frenzy — Instant {2}{R} (Dominaria United):
//
//	"Kicker {B}
//	 Creatures you control get +2/+0 until end of turn. If this spell was
//	 kicked, whenever a creature you control dies this turn, draw a card."
//
// Gnawing Crescendo's pump followed by Waltz of Rage's "whenever a creature
// you control dies this turn" delayed trigger (CR 603.7b), which counts every
// creature, tokens included, and draws for each. The pump applies to the
// creatures you control as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "a0218a14-8301-4484-aed5-90334349620e",
		Name:          "Warhost's Frenzy",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerPump},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{B}")},
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 2, Label: "Warhost's Frenzy — +2/+0"},
			effects.Instruction(func(ctx *effects.Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.DelayedOnEvent{
					Label:     "Warhost's Frenzy — draw a card",
					On:        []game.EventKind{game.EventLTB},
					Condition: effects.B24ACreatureYouControlDiedCondition,
					Body:      effects.DmuB07DrawACardBody,
					Repeats:   true,
				}.Apply(ctx)
			}),
		),
	})
}
