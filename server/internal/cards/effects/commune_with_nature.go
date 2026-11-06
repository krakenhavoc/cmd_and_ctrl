package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commune with Nature — Sorcery {G} (Wilds of Eldraine, #4471):
//
//	"Look at the top five cards of your library. You may reveal a creature
//	 card from among them and put it into your hand. Put the rest on the
//	 bottom of your library in any order."
//
// Commune with Dinosaurs' dig (TakeFromLibraryToHand with
// takeRevealRestBottomAny) over creature cards: the pick is optional, at
// most one card, revealed, and the rest go to the bottom in an order the
// player chooses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d4ed4260-5f38-4b55-ba6e-9f54f04d2360",
		Name:         "Commune with Nature",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    Creature(),
				Max:      1,
				Optional: true,
				Label:    "Commune with Nature — reveal a creature card and put it into your hand",
				Key:      takeRevealRestBottomAny,
			}.Apply(ctx)
		},
	})
}
