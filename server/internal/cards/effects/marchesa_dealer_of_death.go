package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marchesa, Dealer of Death — Legendary Creature — Human Rogue
// {U}{B}{R}, 3/4 (Outlaws of Thunder Junction, #4468):
//
//	"Whenever you commit a crime, you may pay {1}. If you do, look at
//	 the top two cards of your library. Put one of them into your hand
//	 and the other into your graveyard. (Targeting opponents, anything
//	 they control, and/or cards in their graveyards is a crime.)"
//
// The crime is Overzealous Muscle's YouCommittedCrime (CR 700.13), once
// per announcement and with no once-each-turn limit. The {1} is Arahbo's
// MayPay, offered on resolution; declining does nothing. Paid, the look
// is Orazca Puzzle-Door's: one of the top two into your hand, which is
// mandatory, and the other into your graveyard. With one card left in
// the library that card goes to your hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "37449c93-4d32-4f1a-982f-8f3416214368",
		Name:         "Marchesa, Dealer of Death",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBecomesTarget, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return YouCommittedCrime(ev, source)
			}, "Marchesa, Dealer of Death — you may pay {1}; if you do, look at the top two cards of your library (crime)",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Marchesa, Dealer of Death — pay {1} to look at the top two cards of your library, put one into your hand and the other into your graveyard?",
						Then:     marchesaPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// marchesaPaid is the paid half: the top two, one to hand and the other
// to the graveyard.
var marchesaPaid = PayUnlessBranches("pay-unless/marchesa-dealer-of-death-look",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		player := ctx.Controller()
		return TakeFromLibraryToHand{
			Player: player,
			Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 2),
			Max:    1,
			Label:  "Marchesa, Dealer of Death — put one of them into your hand and the other into your graveyard",
			Key:    takeRestGraveyard,
		}.Apply(ctx)
	}, nil)
