package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Follow the Lumarets — Sorcery {1}{G} (Secrets of Strixhaven, #4456):
//
//	"Infusion — Look at the top four cards of your library. You may reveal a
//	 creature or land card from among them and put it into your hand. If you
//	 gained life this turn, you may instead reveal two creature and/or land
//	 cards from among them and put them into your hand. Put the rest on the
//	 bottom of your library in a random order."
//
// Infusion is an ability word (CR 207.2c) and does nothing itself. Vivien's
// Talent's dig (LookAtTopThenMayTakeToHand: a private look, the taken cards
// revealed, the rest to the bottom in a random order) with the ceiling read as
// the spell resolves (CR 608.2c): two when you gained life this turn, one
// otherwise. "May instead" is still a may, so the player can take one, or
// none, either way. The two cards each need only be a creature card or a land
// card, not one of each.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5bb38d08-582e-43aa-9507-7c20759978ec",
		Name:         "Follow the Lumarets",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			take := 1
			if b15LifeGainedThisTurn(ctx.Game, ctx.Controller()) > 0 {
				take = 2
			}
			return LookAtTopThenMayTakeToHand(4, Or(Creature(), Land()), take,
				"Follow the Lumarets — reveal creature or land cards and put them into your hand")(ctx.Game, item)
		},
	})
}
