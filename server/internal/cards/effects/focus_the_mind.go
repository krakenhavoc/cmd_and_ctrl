package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Focus the Mind — Instant {4}{U} (Tarkir: Dragonstorm, #4463):
//
//	"This spell costs {2} less to cast if you've cast another spell this
//	 turn.
//	 Draw three cards, then discard a card."
//
// The reduction reads the caster's cast tally as the cost is determined
// (CR 601.2f). The spell being cast is not counted yet, because a spell
// becomes cast only at CR 601.2i, so any spell in the tally is another
// one. The body is Ancestral Reminiscence's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "695a84b5-96b0-4e48-b615-15624efab3b7",
		Name:         "Focus the Mind",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Discards: 1},
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you've cast another spell this turn.",
				func(q game.CostQuery) bool {
					return q.Game != nil && q.Game.CastTallyFor(q.Controller).Total > 0
				}),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return drawThenDiscard(ctx.Game, ctx.Controller(), ctx.Source(), 3, 1, "")
		},
	})
}
