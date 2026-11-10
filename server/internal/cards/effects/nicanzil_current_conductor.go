package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nicanzil, Current Conductor — Legendary Creature — Merfolk Scout {G}{U},
// 2/2 (The Lost Caverns of Ixalan, #4470):
//
//	"Whenever a creature you control explores a land card, you may put a
//	 land card from your hand onto the battlefield tapped.
//	 Whenever a creature you control explores a nonland card, put a +1/+1
//	 counter on Nicanzil."
//
// Both triggers watch game.EventExplored (CR 701.44b), which names the card
// the explore revealed (Event.ExploreRevealed), read wherever it is now,
// in the hand or on top of the library or in the graveyard. An explore that
// revealed nothing (an empty library) is neither. Nicanzil's own explore is
// one of them, and so is a creature that has since left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4cb86fb6-3e06-4c67-aee0-d5952178f664",
		Name:         "Nicanzil, Current Conductor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventExplored, nicanzilExplored(true),
				"Nicanzil, Current Conductor — you may put a land card from your hand onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return MayPutALandFromHandTapped("Nicanzil, Current Conductor").Apply(NewContext(g, item))
				}),
			On(game.EventExplored, nicanzilExplored(false),
				"Nicanzil, Current Conductor — put a +1/+1 counter on Nicanzil",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

// nicanzilExplored is "a creature you control explores a land card" (land
// true) or "a nonland card".
func nicanzilExplored(land bool) func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		if !aCreatureYouControlExplored(ev, source, lki, g) || ev.ExploreRevealed == uuid.Nil {
			return false
		}
		c, ok := g.LookupCardForEffect(ev.ExploreRevealed)
		return ok && c.IsLand() == land
	}
}
