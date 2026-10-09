package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cruel Calculations — Sorcery {2}{U}:
//
//	"Draw X cards, where X is the number of cards that were put into
//	 target player's graveyard from their library this turn."
//
// X counts every card of the target player's that moved from their
// library to a graveyard this turn (a mill, a surveil, a self-mill
// cost), whether or not it is still there. A card that left the
// graveyard and was milled again counts again, as each is a card put
// there. It is read as the spell resolves; the caster draws.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69b96fe5-9733-4b43-bd66-075742e142d9",
		Name:         "Cruel Calculations",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 || ts[0].Kind != game.TargetPlayer {
				return nil
			}
			x := rfCardsMilledThisTurn(ctx.Game, ts[0].ID)
			return DrawCards{Player: ctx.Controller(), N: x}.Apply(ctx)
		},
	})
}
