package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Freestrider Lookout — Creature — Human Rogue {2}{G}, 3/3 (Outlaws of Thunder Junction):
//
//	"Reach
//	 Whenever you commit a crime, look at the top five cards of your library.
//	 You may put a land card from among them onto the battlefield tapped. Put
//	 the rest on the bottom of your library in a random order. This ability
//	 triggers only once each turn. (Targeting opponents, anything they
//	 control, and/or cards in their graveyards is a crime.)"
//
// The crime trigger held to once each turn (WheneverYouCommitACrimeOnceEachTurn),
// then Armored Skyhunter's look: PutFromLibraryOntoBattlefield over the five
// cards looked at, an optional land entering tapped, the rest to the bottom
// in an order drawn from the game's seeded RNG (a keyed tail, so a table
// waiting on the pick is a restore point).
//
// No simplification.
var freestriderLookoutRest = effects.PutFromLibraryTail("freestrider-lookout/rest-on-bottom",
	func(ctx *effects.Context, res effects.PutFromLibraryResult) error {
		return effects.PutRestOnBottomInRandomOrder(ctx.Game, res)
	})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7d3e234-0954-4422-96c0-cda91101faa3",
		Name:            "Freestrider Lookout",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Purpose:         game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCommitACrimeOnceEachTurn("Freestrider Lookout — look at the top five cards; you may put a land onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					controller := item.Controller
					return effects.PutFromLibraryOntoBattlefield{
						Player:   controller,
						Cards:    g.LookAtTopOfLibraryForEffect(controller, 5),
						Match:    effects.Land(),
						Max:      1,
						Optional: true,
						Tapped:   true,
						Label:    "Freestrider Lookout — you may put a land card from among them onto the battlefield tapped",
						ThenTail: freestriderLookoutRest,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
