package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Commune with Spirits — Sorcery {G} (Kamigawa: Neon Dynasty, S82 #5493):
//
//	"Look at the top four cards of your library. You may reveal an
//	 enchantment or land card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order."
//
// The look-reveal-and-take over four cards (Contagious Vorrac's shape,
// the rest on the bottom at random).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68f8ed23-024e-414f-bbb2-ab73fbe21b69",
		Name:         "Commune with Spirits",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			return effects.TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Match:    effects.Or(effects.Enchantment(), effects.Land()),
				Max:      1,
				Optional: true,
				Label:    "Commune with Spirits — you may reveal an enchantment or land card and put it into your hand",
				Key:      effects.TakeRevealRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
