package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortuary — Enchantment {3}{B} (Stronghold, S82 OLD b751, #5627):
//
//	"Whenever a creature is put into your graveyard from the battlefield, put
//	 that card on top of your library."
//
// A leaves-the-battlefield trigger on every creature that reached a graveyard
// it owner owns, the Mortuary's controller (CR 404.1: a card goes to its owner's
// graveyard; CR 603.10a looks back to the battlefield). The card goes to the top
// of its owner's library through the shared exit primitive when the trigger
// resolves, and a card that has since left the graveyard stays where it is
// (CR 608.2b). A token ceases to exist in the graveyard and has no card to put.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8bd2af55-5a48-4300-a893-ae6cd548db24",
		Name:         "Mortuary",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				dead, ok := effects.DiedCreature(ev, g)
				return ok && !dead.IsToken() && dead.Owner == source.Controller
			}, "Mortuary — put that card on top of your library", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				id := item.Trigger.Event.CardID
				z := g.FindCardZoneForEffect(id)
				if z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				return g.TuckToLibraryForEffect(id, false)
			}),
		},
	})
}
