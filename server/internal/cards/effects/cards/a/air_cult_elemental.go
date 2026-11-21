package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Air-Cult Elemental — Creature — Elemental {4}{U}{U}, 2/5 (Adventures in
// the Forgotten Realms):
//
//	"Flying
//	 Whirlwind — When this creature enters, return up to one other target
//	 creature to its owner's hand."
//
// Rimekin Recluse's enter trigger: up to one creature other than the
// Elemental, any controller's; with no target it does nothing and a
// target that left is skipped (CR 608.2b). "Whirlwind" has no rules
// meaning (CR 207.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ec2d957-c6d9-469f-8190-8b621c490238",
		Name:            "Air-Cult Elemental",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Air-Cult Elemental — return up to one other target creature to its owner's hand", effects.BounceChosenTarget),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature")),
			),
		},
	})
}
