package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lieutenant Commander Data — Legendary Artifact Creature — Android
// Scientist {2}{U}, 2/2 (Star Trek, #4973):
//
//	"Whenever a player casts a noncreature spell, put a +1/+1 counter on
//	 Lieutenant Commander Data.
//	 {1}, Remove three counters from among creatures you control: Draw a
//	 card."
//
// The trigger watches every player's casts (Heartwood Storyteller's
// shape), reading the spell's types on the stack, so a creature spell
// with flash or an artifact creature does not count. The counter goes
// on only while Data is still the same object on the battlefield
// (CR 400.7).
//
// The cost is the any-kind "from among" removal (#943): three counters
// of whatever kinds, split across any creatures the activator controls,
// Data included. A pool that cannot pay {1}, or fewer than three
// counters among them, refuses the activation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebe988c1-96e0-4a99-86b5-d948837545de",
		Name:         "Lieutenant Commander Data",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !spell.IsCreature()
			},
			Key:    "Lieutenant Commander Data — put a +1/+1 counter on Lieutenant Commander Data",
			Effect: Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, Remove three counters from among creatures you control: Draw a card.",
			Cost: Plus(
				ManaCost("{1}"),
				RemoveCountersAmong("", 3, "creatures you control", Creature()),
			),
			Effect:  drawOneCardEffect,
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
		}},
	})
}
