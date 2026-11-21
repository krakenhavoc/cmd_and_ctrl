package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spicy Oatmeal Pizza — {2}{R} Artifact — Food (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"When this artifact enters, it deals 4 damage to any target and 3
//	 damage to you.
//	 {2}, {T}, Sacrifice this artifact: You gain 3 life."
//
// The enters trigger is Daredevil's Billy Club's "damage to any target and
// damage to you" with 4 and 3 (tmtB6DamageAnyTargetAndYou): if the target is
// gone when it resolves, the trigger does nothing, the 3 to you included (CR
// 608.2b). The Food ability is the catalog's shared one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9665f940-c0a4-4d3e-b889-8049cf549a4f",
		Name:         "Spicy Oatmeal Pizza",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Spicy Oatmeal Pizza — 4 damage to any target and 3 damage to you",
				effects.TmtB6DamageAnyTargetAndYou(4, 3)), effects.TargetAny()),
		},
		Activated: []effects.ActivatedAbility{effects.SamiFoodLifeAbility()},
	})
}
