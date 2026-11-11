package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benthic Biomancer — Creature — Merfolk Wizard Mutant {U}, 1/1 (Lost
// Caverns of Ixalan Commander, #5368):
//
//	"{1}{U}: Adapt 1.
//	 Whenever one or more +1/+1 counters are put on this creature, draw a
//	 card, then discard a card."
//
// Passive "are put": whoever puts them, and the counters it enters with,
// count. One trigger per batch of counters (CR 603.2c). Adapt checks for
// counters as it resolves (CR 701.46a), so the Biomancer's own adapt
// triggers its loot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9000917-4f09-42f4-ae85-fc1ac7571d93",
		Name:         "Benthic Biomancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}{U}: Adapt 1",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    ManaCost("{1}{U}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return adaptN(g, item, 1) },
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source == nil || ev.Target != source.InstanceID || ev.Label != game.CounterPlusOne {
					return false
				}
				return b39CountersPlacedOnThisObject(ev, g) > 0
			}, "Benthic Biomancer — draw a card, then discard a card", b35LootOne)),
		},
	})
}
