package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Planar Genesis — Instant {G}{U} (Modern Horizons 3, #5948):
//
//	"Look at the top four cards of your library. You may put a land card
//	 from among them onto the battlefield tapped. If you don't, put a card
//	 from among them into your hand. Put the rest on the bottom of your
//	 library in a random order."
//
// Elvish Rejuvenator's private look and optional tapped land, with the
// "if you don't" branch written into the land pick's keyed tail: when no
// land arrived (declined, none among them, or a replacement kept it off)
// the player must take a card of any kind into their hand (the take is
// mandatory while any card is left, unlike a "may"), the rest going to
// the bottom in a random order either way. The cards are not revealed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b368121b-9984-4a01-8d35-d1e4a8a87539",
		Name:         "Planar Genesis",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return PutFromLibraryOntoBattlefield{
				Player:   item.Controller,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(item.Controller, 4),
				Match:    Land(),
				Max:      1,
				Optional: true,
				Tapped:   true,
				Label:    "Planar Genesis — you may put a land card from among them onto the battlefield tapped",
				ThenTail: planarGenesisAfterLand,
			}.Apply(ctx)
		},
	})
}

// planarGenesisAfterLand is the rest of the clause: a land that arrived
// ends the take, and the rest goes to the bottom in a random order;
// otherwise one card from among them goes to the hand first. An on-disk key.
var planarGenesisAfterLand = PutFromLibraryTail("planar-genesis/after-land",
	func(ctx *Context, res PutFromLibraryResult) error {
		if len(res.Entered) > 0 {
			return PutRestOnBottomInRandomOrder(ctx.Game, res)
		}
		return TakeFromLibraryToHand{
			Player: res.Player,
			Cards:  res.Rest,
			Max:    1,
			Label:  "Planar Genesis — put a card from among them into your hand",
			Key:    takeRestBottomRandom,
		}.Apply(ctx)
	})
