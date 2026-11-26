package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jar of Eyeballs — Artifact {3} (Commander 2013, S82 older sets):
//
//	"Whenever a creature you control dies, put two eyeball counters on this
//	 artifact.
//	 {3}, {T}, Remove all eyeball counters from this artifact: Look at the
//	 top X cards of your library, where X is the number of eyeball counters
//	 removed this way. Put one of them into your hand and the rest on the
//	 bottom of your library in any order."
//
// Removing the counters is part of the cost, so X is the number the cost
// removed (CountersRemoved); with none, X is 0 and nothing is looked at. The
// look, the one taken (no "may") and the ordered bottom are Impulse's three
// pieces.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3075dadd-240f-4455-9286-9f1d48f53a3f",
		Name:         "Jar of Eyeballs",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.WheneverACreatureYouControlDies("Jar of Eyeballs — put two eyeball counters on this artifact",
				effects.Do(effects.CounterOnThis{Kind: "eyeball", N: 2})), game.Purpose{DeathPayoff: true}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Remove all eyeball counters from this artifact: Look at the top X cards of your library, where X is the number of eyeball counters removed this way. Put one of them into your hand and the rest on the bottom of your library in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.RemoveAllCountersFromThis("eyeball")),
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := effects.NewContext(g, item).CountersRemoved()
				if x <= 0 {
					return nil
				}
				return effects.OldB381LookAtTopXTakeOne(x, "Jar of Eyeballs — put one of them into your hand")(g, item)
			},
		}},
	})
}
