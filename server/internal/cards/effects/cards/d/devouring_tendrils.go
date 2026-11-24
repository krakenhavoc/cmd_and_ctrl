package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devouring Tendrils — Sorcery {1}{G} (Strixhaven: School of Mages, #5502):
//
//	"Target creature you control deals damage equal to its power to target
//	 creature or planeswalker you don't control. When the permanent you
//	 don't control dies this turn, you gain 2 life."
//
// Two target clauses (CR 601.2c). The damage is Rabid Bite's one-sided bite
// (the source is the creature you control, last-known power when it left,
// CR 608.2h). The second sentence creates a delayed trigger (CR 603.7a)
// about the opposing permanent as the object it is now, when that target is
// still legal: it fires if that permanent dies this turn, whether it was
// damaged by this spell or not, and a permanent that leaves and returns is a
// new object (CR 603.7c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0112586-3826-452e-9e44-7e6edd197f78",
		Name:         "Devouring Tendrils",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetPermanent("target creature or planeswalker you don't control",
				effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove}, // #4734: a one-sided fight
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.OneSidedBite(item, ctx); err != nil {
				return err
			}
			if theirs, ok := ctx.ClauseTarget(1); ok && theirs.Kind == game.TargetCard {
				effects.GainTwoLifeWhenItDiesThisTurn(ctx, theirs.ID, "Devouring Tendrils — you gain 2 life")
			}
			return nil
		},
	})
}
