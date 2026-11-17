package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Talas Lookout — Creature — Human Pirate {2}{U}{U}, 3/2 (Dominaria
// United):
//
//	"Flying
//	 When this creature dies, look at the top two cards of your library. Put
//	 one of them into your hand and the other into your graveyard."
//
// Akal Pakal's look-two-take-one: the pick goes to hand and the rest to the
// graveyard, whichever is chosen. A library of one card gives that card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "68bde9ff-3206-4a9c-9233-4b2387dcd306",
		Name:            "Talas Lookout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Talas Lookout — look at the top two cards; one into your hand, the other into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
						Max:    1,
						Label:  "Talas Lookout — put one of them into your hand and the other into your graveyard",
						Key:    takeRestGraveyard,
					}.Apply(ctx)
				}),
		},
	})
}
