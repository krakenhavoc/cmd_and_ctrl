package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trial of Zeal — {2}{R} Enchantment: "When this enchantment enters, it
// deals 3 damage to any target. When a Cartouche you control enters,
// return this enchantment to its owner's hand."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0798eb13-83f1-4b7c-b537-83a9da2915a6",
		Name:         "Trial of Zeal",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Trial of Zeal — 3 damage to any target", effects.DamageToFirstTargetEffect(3)),
				effects.TargetAny()), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.WhenACartoucheYouControlEntersReturnThis("Trial of Zeal"),
		},
	})
}
