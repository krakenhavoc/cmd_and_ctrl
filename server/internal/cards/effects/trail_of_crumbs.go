package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trail of Crumbs — {1}{G} Enchantment: "When this enchantment enters,
// create a Food token. Whenever you sacrifice a Food, you may pay {1}.
// If you do, look at the top two cards of your library. You may reveal
// a permanent card from among them and put it into your hand. Put the
// rest on the bottom of your library in any order."
func init() {
	Register(Spec{
		OracleID:     "185d6356-8f8d-4b32-8259-b4eb0a95866c",
		Name:         "Trail of Crumbs",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Trail of Crumbs — create a Food token",
				Do(CreateToken{Template: FoodToken(), N: 1})),
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b11SacrificedAFood(ev, source, g)
			}, "Trail of Crumbs — you may pay {1} to look at the top two cards", trailOfCrumbsMayPay),
		},
	})
}

func trailOfCrumbsMayPay(g *game.Game, item *game.StackItem) error {
	return MayPay{
		Chooser:  item.Controller,
		Cost:     "{1}",
		Question: "Trail of Crumbs — pay {1} to look at the top two cards of your library?",
		Then:     trailOfCrumbsPaid,
	}.Apply(NewContext(g, item))
}

var trailOfCrumbsPaid = PayUnlessBranches("pay-unless/trail-of-crumbs-look-at-top-two",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player:   player,
			Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 2),
			Match:    Permanent(),
			Max:      1,
			Optional: true,
			Label:    "Trail of Crumbs — reveal a permanent card and put it into your hand",
			Key:      takeRestBottomAny,
		}.Apply(ctx)
	}, nil)
