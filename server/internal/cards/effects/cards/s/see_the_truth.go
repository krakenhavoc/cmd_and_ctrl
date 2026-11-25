package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// See the Truth — Sorcery {1}{U} (Core Set 2021, #5511):
//
//	"Look at the top three cards of your library. Put one of those cards into
//	 your hand and the rest on the bottom of your library in any order. If this
//	 spell was cast from anywhere other than your hand, put each of those cards
//	 into your hand instead."
//
// Look at the top three, then either a mandatory pick of one with the rest on
// the bottom in the order you choose (TakeRestBottomAny), or, for a spell cast
// from a graveyard, exile or the command zone, all of them into your hand.
// The origin is read off the spell's own cast (CR 601.2a records the zone it
// was cast from); a copy was never cast and so was not cast from elsewhere.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b23c1f40-291d-43b3-971d-97ded8b367cb",
		Name:         "See the Truth",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 3),
				Max:    1,
				All:    effects.S82OldB110CastFromElsewhere(ctx.Game, item.SourceCardID),
				Label:  "See the Truth — put one of those cards into your hand and the rest on the bottom of your library in any order",
				Key:    effects.TakeRestBottomAny,
			}.Apply(ctx)
		},
	})
}
