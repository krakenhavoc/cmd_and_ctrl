package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "23c75386-8011-4ee4-97e2-eaafeb20b788",
		Name:            "Armament Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Targets: TargetCreature("one, two, or three target creatures you control", YouControl()).
				WithCount(1, 3).Dividing(Divide(3)),
			Key: "Armament Dragon — distribute three +1/+1 counters",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return PutDividedCounters(NewContext(g, item), game.CounterPlusOne)
			},
		}},
	})
}
