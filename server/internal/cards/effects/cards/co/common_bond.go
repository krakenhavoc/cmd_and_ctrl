package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Common Bond — Instant {1}{G}{W} (Return to Ravnica, #5566):
//
//	"Put a +1/+1 counter on target creature.
//	 Put a +1/+1 counter on target creature."
//
// Two target clauses, so one creature may be both picks and get two counters. If
// one pick is illegal on resolution the other still gets its counter (CR 608.2b);
// the spell does nothing only when both are.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d52e9255-62e7-479f-8def-e92b96efdb89",
		Name:         "Common Bond",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetCreature("target creature"),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for slot := 0; slot < 2; slot++ {
				if t, ok := ctx.ClauseTarget(slot); ok {
					if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			return nil
		},
	})
}
