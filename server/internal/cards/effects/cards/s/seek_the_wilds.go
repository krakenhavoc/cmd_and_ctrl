package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Seek the Wilds — Sorcery {1}{G} (Game Night):
//
//	"Look at the top four cards of your library. You may reveal a creature
//	 or land card from among them and put it into your hand. Put the rest on
//	 the bottom of your library in any order."
//
// Ancient Stirrings' dig over creature and land cards: optional, at most
// one, revealed, the rest to the bottom in an order the player chooses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3865d219-44b7-49ce-ab26-b833acb39765",
		Name:         "Seek the Wilds",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Match:    effects.Or(effects.Creature(), effects.Land()),
				Max:      1,
				Optional: true,
				Label:    "Seek the Wilds — reveal a creature or land card and put it into your hand",
				Key:      effects.TakeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
