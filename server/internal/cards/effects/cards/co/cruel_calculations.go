package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "69b96fe5-9733-4b43-bd66-075742e142d9",
		Name:         "Cruel Calculations",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 || ts[0].Kind != game.TargetPlayer {
				return nil
			}
			x := effects.RfCardsMilledThisTurn(ctx.Game, ts[0].ID)
			return effects.DrawCards{Player: ctx.Controller(), N: x}.Apply(ctx)
		},
	})
}
