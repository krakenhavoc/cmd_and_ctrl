package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jilt — Instant {1}{U} (Apocalypse, #5614):
//
//	"Kicker {1}{R} (You may pay an additional {1}{R} as you cast this spell.)
//	 Return target creature to its owner's hand. If this spell was kicked, it
//	 deals 2 damage to another target creature."
//
// The second target exists only when the spell is kicked (CR 601.2c) and must
// differ from the first ("another"). A bounce target that is gone does not
// stop the damage to the other one, each target being checked on its own
// (CR 608.2b).
//
// No simplification.
func init() {
	first := func() *game.TargetSpec { return effects.TargetCreature("target creature") }
	effects.Register(effects.Spec{
		OracleID:     "382bc554-d170-4c19-82dd-829f254e855e",
		Name:         "Jilt",
		Completeness: effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Kicker("{1}{R}"),
			effects.Clauses(first(), effects.Distinct(effects.TargetCreature("another target creature"))))},
		Targets: first(),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			if !ctx.WasKicked() {
				return nil
			}
			other, found := ctx.ClauseTarget(1)
			if !found || other.Kind != game.TargetCard {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: other.ID, Amount: 2}.Apply(ctx)
		},
	})
}
