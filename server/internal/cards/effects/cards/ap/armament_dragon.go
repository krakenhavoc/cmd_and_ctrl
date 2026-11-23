package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armament Dragon — Creature — Dragon {3}{W}{B}{G}, 3/4 (#1658,
// unblocked by #1656):
//
//	"Flying
//	 When this creature enters, distribute three +1/+1 counters
//	 among one, two, or three target creatures you control."
//
// Verdurous Gearhulk's shape with a bounded count instead of "any
// number" — one to three targets you control, each getting at least 1
// of the 3 (CR 601.2d).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "23c75386-8011-4ee4-97e2-eaafeb20b788",
		Name:            "Armament Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets: effects.TargetCreature("one, two, or three target creatures you control", effects.YouControl()).
				WithCount(1, 3).Dividing(effects.Divide(3)),
			Key: "Armament Dragon — distribute three +1/+1 counters",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutDividedCounters(effects.NewContext(g, item), game.CounterPlusOne)
			},
		}},
	})
}
