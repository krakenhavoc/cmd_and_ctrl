package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Armed Response — Instant {2}{W} (Fifth Dawn, #5604):
//
//	"Armed Response deals damage to target attacking creature equal to
//	 the number of Equipment you control."
//
// Slash of Light's count without the creatures: the Equipment its
// controller controls are counted as the spell resolves, attached or
// not. With none, it deals no damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0258a606-1625-402e-b0c4-ffca06779b5e",
		Name:         "Armed Response",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			n := effects.CountControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.HasSubtype("Equipment") })
			if n <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: n}.Apply(ctx)
		},
	})
}
