package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "a0218a14-8301-4484-aed5-90334349620e",
		Name:          "Warhost's Frenzy",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerPump},
		OptionalCosts: []game.AdditionalCost{Kicker("{B}")},
		OnResolve: ResolveSteps(
			BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 2, Label: "Warhost's Frenzy — +2/+0"},
			Instruction(func(ctx *Context) error {
				if !ctx.WasKicked() {
					return nil
				}
				return DelayedOnEvent{
					Label:     "Warhost's Frenzy — draw a card",
					On:        []game.EventKind{game.EventLTB},
					Condition: b24ACreatureYouControlDiedCondition,
					Body:      dmuB07DrawACardBody,
					Repeats:   true,
				}.Apply(ctx)
			}),
		),
	})
}
