package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Closing Statement — Instant {3}{W}{B} (Strixhaven: School of Mages,
// #5502):
//
//	"This spell costs {2} less to cast during your end step.
//	 Destroy target creature or planeswalker you don't control. Put a +1/+1
//	 counter on up to one target creature you control."
//
// Two target clauses, the second optional: neither half waits on the other
// (the counter is put on first only so the destruction is the last step). The discount is judged as the cost is determined
// (CR 601.2f): your own end step, and no other.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79ac7b6a-1c25-45f6-84b4-5c4a78d6fd6e",
		Name:         "Closing Statement",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(2, "Closing Statement — costs {2} less to cast during your end step", func(q game.CostQuery) bool {
				return q.Game != nil && q.Game.Turn.Step == game.StepEnd && effects.IsYourTurn(q.Game, q.Controller)
			}),
		},
		Targets: effects.Clauses(
			effects.TargetPermanent("target creature or planeswalker you don't control", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()),
			effects.TargetCreature("up to one target creature you control", effects.YouControl()).WithCount(0, 1),
		),
		Purpose: game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
