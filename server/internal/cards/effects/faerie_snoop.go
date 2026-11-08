package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Faerie Snoop — Creature — Faerie Detective {1}{U}{B}, 1/4 (Murders at
// Karlov Manor, #4469):
//
//	"Flying
//	 Disguise {1}{U/B}{U/B} (You may cast this card face down for {3} as
//	 a 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)
//	 When this creature is turned face up, look at the top two cards of
//	 your library. Put one into your hand and the other into your
//	 graveyard."
//
// The face-up trigger (CR 708.8) is Marchesa, Dealer of Death's paid
// half without the payment: one choose_cards prompt over the top two,
// exactly one to hand and the rest to the graveyard (takeRestGraveyard).
// With one card left in the library that card goes to hand; with none,
// nothing happens.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "b759e8a6-8dd0-48ce-8816-e45b35a79895",
		Name:             "Faerie Snoop",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Disguise("{1}{U/B}{U/B}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Faerie Snoop — look at the top two cards of your library, one to hand and the other to your graveyard",
				faerieSnoopLook),
		},
	})
}

// faerieSnoopLook puts one of the top two cards into the controller's
// hand and the other into their graveyard.
func faerieSnoopLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player: player,
		Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
		Max:    1,
		Label:  "Faerie Snoop — put one of them into your hand and the other into your graveyard",
		Key:    takeRestGraveyard,
	}.Apply(ctx)
}
