package cardsst

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sweep Away — Instant {2}{U} (Jumpstart, #5510):
//
//	"Return target creature to its owner's hand. If that creature is
//	 attacking, you may put it on top of its owner's library instead."
//
// Whether the creature is attacking is read as the spell resolves. Only then
// is the controller asked; a creature that is not attacking is simply
// returned. A token is returned or tucked and ceases to exist either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6901a89-ac7d-4797-a027-97bb1cf42e60",
		Name:         "Sweep Away",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if c.AttackingTarget == uuid.Nil {
				return effects.BounceToHand{Target: id}.Apply(ctx)
			}
			return effects.MayChoice{
				Question: "Sweep Away — put the attacking creature on top of its owner's library instead of returning it to hand?",
				YesLabel: "Put it on top of its library",
				NoLabel:  "Return it to hand",
				Then:     sweepAwayAnswer,
				Carry:    []uuid.UUID{id},
			}.Apply(ctx)
		},
	})
}

var sweepAwayAnswer = effects.MayAnswer("confirm/s82-old-b104-sweep-away", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if len(r.Carry) == 0 || !effects.OnBattlefield(ctx.Game, r.Carry[0]) {
		return nil
	}
	if r.Accepted {
		return ctx.Game.TuckToLibraryForEffect(r.Carry[0], false)
	}
	return effects.BounceToHand{Target: r.Carry[0]}.Apply(ctx)
})
