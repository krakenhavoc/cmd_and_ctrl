package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "0258a606-1625-402e-b0c4-ffca06779b5e",
		Name:         "Armed Response",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			n := countControlled(ctx.Game, item.Controller, func(c game.Card) bool { return c.HasSubtype("Equipment") })
			if n <= 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: target, Amount: n}.Apply(ctx)
		},
	})
}
