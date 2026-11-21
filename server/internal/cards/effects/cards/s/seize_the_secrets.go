package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seize the Secrets — Sorcery {2}{U} (Outlaws of Thunder Junction,
// #4468):
//
//	"This spell costs {1} less to cast if you've committed a crime this
//	 turn. (Targeting opponents, anything they control, and/or cards in
//	 their graveyards is a crime.)
//	 Draw two cards."
//
// The discount is Focus the Mind's self cost modifier, its condition
// this turn's event log holding a crime by the caster
// (YouCommittedACrimeThisTurn), read as the cost is determined (CR
// 601.2f). The body is Divination's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db070312-9a27-435f-8b23-e538441103ec",
		Name:         "Seize the Secrets",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLess(1, "This spell costs {1} less to cast if you've committed a crime this turn.",
				func(q game.CostQuery) bool {
					return q.Game != nil && effects.YouCommittedACrimeThisTurn(q.Game, q.Controller)
				}),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
