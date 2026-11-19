package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trial of Zeal — {2}{R} Enchantment: "When this enchantment enters, it
// deals 3 damage to any target. When a Cartouche you control enters,
// return this enchantment to its owner's hand."
func init() {
	Register(Spec{
		OracleID:     "0798eb13-83f1-4b7c-b537-83a9da2915a6",
		Name:         "Trial of Zeal",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisEnters("Trial of Zeal — 3 damage to any target", damageToFirstTargetEffect(3)),
				TargetAny()), ForTargets(DamageToTarget(0, 3))),
			whenACartoucheYouControlEntersReturnThis("Trial of Zeal"),
		},
	})
}
