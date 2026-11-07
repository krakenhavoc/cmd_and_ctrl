package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pillage the Bog — Sorcery {B}{G} (Outlaws of Thunder Junction,
// #4468):
//
//	"Look at the top X cards of your library, where X is twice the
//	 number of lands you control. Put one of them into your hand and the
//	 rest on the bottom of your library in a random order.
//	 Plot {1}{B}{G} (You may pay {1}{B}{G} and exile this card from your
//	 hand. Cast it as a sorcery on a later turn without paying its mana
//	 cost. Plot only as a sorcery.)"
//
// X is counted once, as the spell resolves. The look is Impulse's with
// Memory Deluge's random
// bottom (takeRestBottomRandom): one card into your hand, which is
// mandatory, and the rest under the library in a random order. With no
// lands there is nothing to look at. Plot is Beastbond Outcaster's
// special action (CR 702.170a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "91740570-1a97-4513-8af9-b32c363daf15",
		Name:           "Pillage the Bog",
		Completeness:   CompletenessFull,
		SpecialActions: []game.SpecialAction{Plot("{1}{B}{G}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			x := 2 * countControlled(ctx.Game, player, func(c game.Card) bool { return c.IsLand() })
			if x == 0 {
				return nil
			}
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, x),
				Max:    1,
				Label:  "Pillage the Bog — put one of them into your hand and the rest on the bottom of your library in a random order",
				Key:    takeRestBottomRandom,
			}.Apply(ctx)
		},
	})
}
