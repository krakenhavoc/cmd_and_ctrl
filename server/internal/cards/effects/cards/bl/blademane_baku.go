package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blademane Baku — Creature — Spirit {1}{R}, 1/1 (Betrayers of Kamigawa,
// #5602):
//
//	"Whenever you cast a Spirit or Arcane spell, you may put a ki counter on
//	 this creature.
//	 {1}, Remove X ki counters from this creature: For each counter removed,
//	 this creature gets +2/+0 until end of turn."
//
// Quillmane Baku's ki counter trigger and cost, with no tap. X is the number
// of counters the cost removes, announced with the activation, and the Baku
// gets +2/+0 for each; it does nothing if the Baku has left the battlefield
// or is a new object by then (CR 400.7). X may be 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9ca64b8e-e9d2-41b3-b229-6fcdc48a8371",
		Name:         "Blademane Baku",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WheneverYouCast(effects.Or(effects.Subtype("Spirit"), effects.Subtype("Arcane")),
				"Blademane Baku — put a ki counter on this creature", effects.PutACounterOnThis("ki")),
				"Blademane Baku — put a ki counter on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Remove X ki counters from this creature: For each counter removed, this creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.RemoveCountersXFromThis("ki", 0)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := ctx.CountersRemoved()
				if x <= 0 || !effects.OnBattlefield(g, item.SourceCardID) || effects.SourceIsNewObject(g, item) {
					return nil
				}
				return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2 * x, Label: "Blademane Baku — +2/+0 for each counter removed"}.Apply(ctx)
			},
		}},
	})
}
