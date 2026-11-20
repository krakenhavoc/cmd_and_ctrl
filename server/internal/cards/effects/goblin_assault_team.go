package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Assault Team — Creature — Goblin Warrior {3}{R}, 4/1 (War of the Spark, #5520):
//
//	"Haste
//	 When this creature dies, put a +1/+1 counter on target creature you control."
//
// Haste is printed data; the dies trigger is Steadfast Sentry's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d464b54c-3b02-4c4e-99ef-1ad75f336c65",
		Name:            "Goblin Assault Team",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Goblin Assault Team — put a +1/+1 counter on target creature you control",
				rfCreatureFCountersOnTarget(1)),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
