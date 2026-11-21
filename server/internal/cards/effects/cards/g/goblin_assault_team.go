package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Assault Team — Creature — Goblin Warrior {3}{R}, 4/1 (War of the Spark, #5520):
//
//	"Haste
//	 When this creature dies, put a +1/+1 counter on target creature you control."
//
// Haste is printed data; the dies trigger is Steadfast Sentry's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d464b54c-3b02-4c4e-99ef-1ad75f336c65",
		Name:            "Goblin Assault Team",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Goblin Assault Team — put a +1/+1 counter on target creature you control",
				effects.RfCreatureFCountersOnTarget(1)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
