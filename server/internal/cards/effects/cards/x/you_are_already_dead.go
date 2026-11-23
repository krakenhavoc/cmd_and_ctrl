package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// You Are Already Dead — Instant {B} (Kamigawa: Neon Dynasty, #5493):
//
//	"Destroy target creature that was dealt damage this turn.
//	 Draw a card."
//
// "Was dealt damage this turn" is the shared DealtDamageThisTurn
// predicate, read off the turn's event log at announce and resolution.
// The draw is the spell's second sentence and happens even if the creature
// can't be destroyed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1157d79-38ea-4a88-b346-7b12614ff78a",
		Name:         "You Are Already Dead",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := (effects.DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
