package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hit It! — Enchantment {1}{R}{G} (Star Trek, #4973):
//
//	"Creatures you control have haste.
//	 When this enchantment enters, look at the top five cards of your
//	 library. Put a land card from among them onto the battlefield tapped
//	 and the rest into your graveyard."
//
// Fervor's haste grant (a layer-6 grant to every creature its controller
// controls, CR 613.1f) with Shigeki, Jukai Visionary's dig. The put is not
// a "may": with a land among the five, one of them enters tapped (its
// controller picks which); with none, all five go to the graveyard. A
// looked-at card that left the library before the put is ignored.
//
// "Look at", not "reveal" (CR 701.20e): the card printed "reveal" and its
// 2026 Oracle update reads "look at" (#5949), so only its controller sees
// the five cards before they are put; the other players see the land that
// enters and the cards that go to the graveyard. The put is still not a
// "may": only a search lets a player fail to find (CR 701.23b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62361438-3da4-4ce9-95b8-fc361b824f58",
		Name:         "Hit It!",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hit It! — look at the top five cards; put a land card from among them onto the battlefield tapped and the rest into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.PutFromLibraryOntoBattlefield{
						Player:   item.Controller,
						Cards:    g.LookAtTopOfLibraryForEffect(item.Controller, 5),
						Match:    effects.Land(),
						Max:      1,
						Tapped:   true,
						Label:    "Hit It! — put a land card onto the battlefield tapped",
						ThenTail: effects.PutRestIntoGraveyardTail,
					}.Apply(ctx)
				}),
		},
	})
}
