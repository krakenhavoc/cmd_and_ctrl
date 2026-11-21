package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron Bully — Artifact Creature — Golem {3}, 1/1 (War of the Spark, #5520):
//
//	"Menace (This creature can't be blocked except by two or more creatures.)
//	 When this creature enters, put a +1/+1 counter on target creature."
//
// Menace is printed data; the enters trigger is Jeong Jeong's
// Deserters'. Any creature may be the target, itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d78155c-c63f-48a3-86e5-12b80400d8de",
		Name:            "Iron Bully",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Iron Bully — put a +1/+1 counter on target creature", effects.RfCreatureFCounterOnTargetCreature),
				effects.TargetCreature("target creature"),
			),
		},
	})
}
