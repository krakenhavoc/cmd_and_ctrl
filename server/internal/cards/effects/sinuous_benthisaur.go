package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sinuous Benthisaur — Creature — Dinosaur {5}{U}, 4/4 (The Lost Caverns
// of Ixalan, #4470):
//
//	"When this creature enters, look at the top X cards of your library,
//	 where X is the number of Caves you control plus the number of Cave
//	 cards in your graveyard. Put two of those cards into your hand and
//	 the rest on the bottom of your library in a random order."
//
// X is counted as the trigger resolves (CR 608.2h): Caves are lands
// with the Cave subtype you control (b08LandsWithSubtypeControlled), plus
// Cave cards in your graveyard (b19ZombieCardsInGraveyard's subtype
// count). The take is mandatory: two cards, or every card looked at if
// there are fewer, and no reveal, since nothing says to reveal them. The
// rest go to the bottom at random (takeRestBottomRandom).
//
// No simplification.
func init() {
	const label = "Sinuous Benthisaur — put two of those cards into your hand"
	Register(Spec{
		OracleID:     "d3cad867-142c-4e9d-bcd8-0e7ed1fd2f77",
		Name:         "Sinuous Benthisaur",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(label, func(g *game.Game, item *game.StackItem) error {
				x := b08LandsWithSubtypeControlled(g, item.Controller, "Cave") +
					b19ZombieCardsInGraveyard(g, item.Controller, "Cave")
				if x <= 0 {
					return nil
				}
				return TakeFromLibraryToHand{
					Player: item.Controller,
					Cards:  g.LookAtTopOfLibraryForEffect(item.Controller, x),
					Max:    2,
					Label:  label,
					Key:    takeRestBottomRandom,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
