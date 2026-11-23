package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rakshasa's Bargain — Instant {2/B}{2/G}{2/U} (Tarkir: Dragonstorm,
// #4463):
//
//	"Look at the top four cards of your library. Put two of them into your
//	 hand and the rest into your graveyard."
//
// Polygraph Orb's look as a spell: the controller picks two of the top
// four (every card, with fewer than two in the library), and the rest are
// put into the graveyard, not milled. Each monocolored hybrid symbol is
// paid with its colour or with two generic mana (CR 107.4e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d87be8c6-7a1f-4db8-ada1-b1a030f03877",
		Name:         "Rakshasa's Bargain",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Max:    2,
				Label:  "Rakshasa's Bargain — put two of them into your hand and the rest into your graveyard",
				Key:    effects.TakeRestGraveyard,
			}.Apply(ctx)
		},
	})
}
