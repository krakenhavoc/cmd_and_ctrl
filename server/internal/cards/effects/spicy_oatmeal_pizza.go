package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "9665f940-c0a4-4d3e-b889-8049cf549a4f",
		Name:         "Spicy Oatmeal Pizza",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Spicy Oatmeal Pizza — 4 damage to any target and 3 damage to you",
				tmtB6DamageAnyTargetAndYou(4, 3)), TargetAny()),
		},
		Activated: []ActivatedAbility{samiFoodLifeAbility()},
	})
}
