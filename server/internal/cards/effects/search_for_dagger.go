package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Search for Dagger — Enchantment {1}{W} (the Squall deck request,
// #3918):
//
//	"Whenever your commander enters or attacks, look at the top six
//	 cards of your library. You may reveal a legendary creature card
//	 from among them and put it into your hand. Put the rest on the
//	 bottom of your library in a random order."
//
// "Your commander" is a commander you own (CR 903.3), read the way Tome
// of Legends reads it, so both partners count and a commander an
// opponent stole still triggers its owner's enchantment when it enters
// or attacks under them. The look is Horn of the Mark's: only you see
// the six, the card you take is revealed, and the rest go to the bottom
// in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9bdd6ece-b78c-4814-ac40-b48f91b405b5",
		Name:         "Search for Dagger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := g.LookupCardForEffect(ev.CardID)
					return ok && c.IsCommander && c.Owner == source.Controller
				},
				"Search for Dagger — look at the top six cards of your library",
				LookAtTopThenMayTakeToHand(6, And(Legendary(), Creature()), 1,
					"Search for Dagger — you may reveal a legendary creature card from among them and put it into your hand")),
		},
	})
}
