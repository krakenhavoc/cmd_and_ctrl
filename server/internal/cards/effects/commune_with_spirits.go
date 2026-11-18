package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "68f8ed23-024e-414f-bbb2-ab73fbe21b69",
		Name:         "Commune with Spirits",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Match:    Or(Enchantment(), Land()),
				Max:      1,
				Optional: true,
				Label:    "Commune with Spirits — you may reveal an enchantment or land card and put it into your hand",
				Key:      takeRevealRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
