package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lead the Stampede — Sorcery {2}{G} (Ikoria: Lair of Behemoths, #5512):
//
//	"Look at the top five cards of your library. You may reveal any number
//	 of creature cards from among them and put the revealed cards into your
//	 hand. Put the rest on the bottom of your library in any order."
//
// Commune with Nature's shape with "any number": a private look, an
// optional reveal of any number of creature cards (only creature cards are
// offered), and the rest bottomed in the player's order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea91be73-40f5-461a-bb39-d60216829a53",
		Name:         "Lead the Stampede",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    effects.Creature(),
				Optional: true,
				Label:    "Lead the Stampede — you may reveal any number of creature cards and put them into your hand",
				Key:      effects.TakeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
