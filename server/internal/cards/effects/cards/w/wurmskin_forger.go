package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wurmskin Forger — Creature — Elf Warrior {5}{G}{G}, 2/2 (Mirrodin, S82
// OLD, #5606):
//
//	"When this creature enters, distribute three +1/+1 counters among one,
//	 two, or three target creatures."
//
// Biogenic Upgrade's distribution, on an enters trigger (Broodlord's
// shape): one to three targets, any creatures (the Forger itself and
// opponents' included), chosen with the division as the trigger goes on the
// stack (CR 603.3d, 601.2d), each target getting at least one. Each target
// still legal as it resolves gets its share (CR 608.2b).
//
// No simplification.
func init() {
	const key = "Wurmskin Forger — distribute three +1/+1 counters among one, two, or three target creatures"
	effects.Register(effects.Spec{
		OracleID:     "c94332ac-982a-4c0f-85d4-43e950ac3069",
		Name:         "Wurmskin Forger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Targets:   effects.TargetCreature("one, two, or three target creatures").WithCount(1, 3).Dividing(effects.Divide(3)),
			Key:       key,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PutDividedCounters(effects.NewContext(g, item), game.CounterPlusOne)
			},
		}},
	})
}
