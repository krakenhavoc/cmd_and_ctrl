package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gang of Devils — Creature — Devil {5}{R}, 3/3 (Conspiracy: Take the
// Crown, S82 older sets batch 270, #5540):
//
//	"When this creature dies, it deals 3 damage divided as you choose among
//	 one, two, or three targets."
//
// Orca, Siege Demon's death trigger with Mogg Mob's fixed division: the
// division is announced with the targets as the ability goes on the stack
// (CR 601.2d via CR 603.3d), each target gets at least 1, and the damage
// comes from the Devil as it last existed (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62110234-690b-4498-8076-26ba910f9f92",
		Name:         "Gang of Devils",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.CardDied(ev, source)
			},
			Targets: effects.TargetAny().WithCount(1, 3).Dividing(effects.Divide(3)),
			Key:     "Gang of Devils — 3 damage divided among one, two, or three targets",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DealDividedDamage(effects.NewContext(g, item))
			},
		}},
	})
}
