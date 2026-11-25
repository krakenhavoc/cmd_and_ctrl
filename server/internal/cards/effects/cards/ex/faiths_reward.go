package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faith's Reward — Instant {3}{W} (Conspiracy: Take the Crown, S82 older
// sets batch 271, #5540):
//
//	"Return to the battlefield all permanent cards in your graveyard that
//	 were put there from the battlefield this turn."
//
// Thrilling Encore's untargeted return over one graveyard: the set is read
// as the spell resolves, a card counts when its newest arrival in your
// graveyard this turn came from the battlefield (a card that was discarded,
// milled or countered is not one, and a token is not a card), and they
// all enter together under your control (CR 603.6a). An Aura chooses what
// it enchants as it enters and stays behind with nothing to enchant
// (CR 303.4f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2fbb6703-3517-4117-9ca4-45e90a6216ce",
		Name:         "Faith's Reward",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFromGraveyardTogether{
				Targets: effects.GraveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
					return c.IsPermanent() && effects.PutIntoGraveyardFromBattlefieldThisTurn(ctx.Game, c.InstanceID)
				}),
				Controller: ctx.Controller(),
			}.Apply(ctx)
		},
	})
}
