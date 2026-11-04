package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mikey & Leo, Chaos & Order — {G/W}{G/W} Legendary Creature — Mutant Ninja
// Turtle 2/2 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Whenever you put a counter on a creature you control, draw a card.
//	 This ability triggers only once each turn."
//
// Generous Patron's "whenever you put one or more counters on a creature"
// (WheneverYouPutOneOrMoreCountersOnACreature) narrowed to a creature you
// control: any kind of counter, placed by you (the event's recorded
// placer, so a counter an opponent puts on your creature does not count),
// Mikey & Leo included. "Triggers only once each turn" is Mirelurk
// Queen's TriggersOnlyOnceEachTurn, per object: the first placement of
// the turn triggers, a later one does not, and a trigger that was then
// countered still used up the turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a86117c1-b83d-4350-9a10-a23b7e67e1bd",
		Name:         "Mikey & Leo, Chaos & Order",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(WheneverYouPutOneOrMoreCountersOnACreature(creatureYouControl,
				"Mikey & Leo, Chaos & Order — draw a card", Do(DrawCards{N: 1}))),
		},
	})
}
