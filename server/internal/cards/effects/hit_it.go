package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hit It! — Enchantment {1}{R}{G} (Star Trek, #4973):
//
//	"Creatures you control have haste.
//	 When this enchantment enters, look at the top five cards of your
//	 library. Put a land card from among them onto the battlefield tapped
//	 and the rest into your graveyard."
//
// Fervor's haste grant (a layer-6 grant to every creature its controller
// controls, CR 613.1f) with Elvish Rejuvenator's dig. The put is not a
// "may": with a land among the five, one of them enters tapped (its
// controller picks which); with none, all five go to the graveyard. A
// card that left the library before the put is ignored.
//
// "Look at", not "reveal" (CR 701.20): the Oracle text said "reveal"
// until an erratum (#5949). Only the controller sees the five while
// choosing; the other seats see each card as it reaches the battlefield
// or the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62361438-3da4-4ce9-95b8-fc361b824f58",
		Name:         "Hit It!",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Hit It! — look at the top five cards; put a land card from among them onto the battlefield tapped and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
						Match:    Land(),
						Max:      1,
						Tapped:   true,
						Label:    "Hit It! — put a land card onto the battlefield tapped",
						ThenTail: PutRestIntoGraveyardTail,
					}.Apply(ctx)
				}),
		},
	})
}
