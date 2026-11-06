package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mourner's Surprise — Sorcery {1}{B} (Outlaws of Thunder Junction,
// #4468):
//
//	"Return up to one target creature card from your graveyard to your
//	 hand. Create a 1/1 red Mercenary creature token with '{T}: Target
//	 creature you control gets +1/+0 until end of turn. Activate only as
//	 a sorcery.'"
//
// "Up to one" lets it be cast with no target at all, and then it only
// makes the Mercenary. With a target chosen that has left the graveyard
// by resolution, the spell's only target is illegal and it does nothing
// (CR 608.2b), the token included. The token is the set's Mercenary
// template (MercenaryToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "567fa6e0-fc7b-4502-9c51-f2538c411003",
		Name:         "Mourner's Surprise",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      TargetCardInGraveyard("up to one target creature card from your graveyard", Creature(), YouOwn()).WithCount(0, 1),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := returnTargetGraveyardCardToHand(ctx.Game, item); err != nil {
				return err
			}
			return createAMercenary()(ctx.Game, item)
		},
	})
}
