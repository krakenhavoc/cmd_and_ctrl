package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "db070312-9a27-435f-8b23-e538441103ec",
		Name:         "Seize the Secrets",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(1, "This spell costs {1} less to cast if you've committed a crime this turn.",
				func(q game.CostQuery) bool {
					return q.Game != nil && YouCommittedACrimeThisTurn(q.Game, q.Controller)
				}),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
