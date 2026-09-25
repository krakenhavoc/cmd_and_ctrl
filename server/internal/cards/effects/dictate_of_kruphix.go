package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dictate of Kruphix — Enchantment {1}{U}{U} (EDHREC rank 1867):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.)
//	 At the beginning of each player's draw step, that player draws
//	 an additional card."
//
// Howling Mine with flash and no "if untapped" clause — flashed in at
// the end of the turn before yours so you are the first to draw off
// it. The trigger is the Mine's: EACH player's draw step, the drawer
// is the event's Actor rather than the Dictate's controller, and it
// arrives after the turn-based draw (CR 504.1 does not use the
// stack), which is what "an additional card" means. Flash rides
// PrintedKeywords so the cast-timing gate reads it from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "be74a4c1-d569-4203-b28a-3e2ac6a82990",
		Name:            "Dictate of Kruphix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginDrawStep},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil
			},
			Key: "Dictate of Kruphix — draw an additional card",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Trigger.Event.Actor, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
