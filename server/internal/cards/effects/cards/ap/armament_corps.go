package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armament Corps — Creature — Human Soldier {2}{W}{B}{G}, 4/4 (Khans of
// Tarkir):
//
//	"When this creature enters, distribute two +1/+1 counters among one or two
//	 target creatures you control."
//
// Splendid Agony's distribution with +1/+1 counters, on a trigger whose one or
// two targets (yours, the Corps itself a legal one) are chosen as it goes on the
// stack with at least one counter each (CR 601.2d, 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b93427c-5b40-4119-96ae-7dce3ca2bbb9",
		Name:         "Armament Corps",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.B06SelfETB,
			Targets:   effects.TargetCreature("one or two target creatures you control", effects.YouControl()).WithCount(1, 2).Dividing(effects.Divide(2)),
			Key:       "Armament Corps — distribute two +1/+1 counters among one or two target creatures you control",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutDividedCounters(effects.NewContext(g, item), game.CounterPlusOne)
			},
		}},
	})
}
