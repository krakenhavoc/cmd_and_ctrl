package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stolen Goods — Sorcery {3}{U} (Outlaws of Thunder Junction Commander):
//
//	"Target opponent exiles cards from the top of their library until they
//	 exile a nonland card. Until end of turn, you may cast that card without
//	 paying its mana cost."
//
// The opponent's library is walked with the exile-until-nonland run
// (UntilNonland); every card passed over stays in exile (it is not put
// anywhere), and the nonland card gets a {0} cast permission for the turn at
// the card's own timing, lapsing with it still in exile. It is not cast
// during the resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9294d4d3-0df0-4070-9261-06f15b758a0d",
		Name:         "Stolen Goods",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			victim := cmdr2S43TargetOpponent(ctx)
			if victim == uuid.Nil {
				return nil
			}
			return ctx.Game.MillToZoneKeyedForEffect(victim, 0, game.ZoneExile, UntilNonland, ctx.ItemTail(cmdr2S43StolenGoodsGranted))
		},
	})
}
