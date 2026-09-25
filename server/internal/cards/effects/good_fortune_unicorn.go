package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Good-Fortune Unicorn — Creature — Unicorn {1}{G}{W}, 2/2 (EDHREC
// rank 3510):
//
//	"Whenever another creature you control enters, put a +1/+1 counter
//	 on that creature."
//
// The counters deck's welcome mat. The condition is
// b13AnotherCreatureYouControlEntered — a token counts, the Unicorn
// itself does not — and the counter goes on the creature that
// entered, if it is still on the battlefield when the trigger
// resolves. The counter rides the CR 614 pipeline, so Hardened
// Scales and Doubling Season see it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e0cdff3-3ec5-41fc-8053-f072bae156b3",
		Name:         "Good-Fortune Unicorn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			},
			Key: "Good-Fortune Unicorn — put a +1/+1 counter on the creature that entered",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b33PutCounterOnEnteredCreature(item.Trigger.Event.CardID)(g, item)
			},
		}},
	})
}
