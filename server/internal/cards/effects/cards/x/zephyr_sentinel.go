package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Zephyr Sentinel — Creature — Human Soldier {1}{U}, 2/1 (The Brothers'
// War, #5482):
//
//	"Flash
//	 Flying
//	 When this creature enters, return up to one other target creature you
//	 control to its owner's hand. If it was a Soldier, put a +1/+1 counter
//	 on this creature."
//
// Flock Impostor's bounce, with the Soldier test read off the creature as it
// stood on the battlefield, before it left. The counter lands on the
// Sentinel only while it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dcddd143-a7d9-4a38-a2da-dd6bd2e48205",
		Name:            "Zephyr Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Zephyr Sentinel — return up to one other target creature you control to its owner's hand; a +1/+1 counter on this creature if it was a Soldier", effects.BroB05ZephyrSentinelEnters),
				effects.Another(effects.UpToOneTargetCreature("up to one other target creature you control", effects.YouControl())),
			),
		},
	})
}
