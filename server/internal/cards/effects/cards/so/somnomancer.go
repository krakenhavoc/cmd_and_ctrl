package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Somnomancer — Creature — Kithkin Wizard {1}{W/U}, 2/1 (Shadowmoor, #5591):
//
//	"When this creature enters, you may tap target creature."
//
// The target is chosen as the trigger goes on the stack (CR 603.3d) and the
// "you may" is asked as it resolves (CR 603.5). A target gone by then is not
// tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c819ebc8-feb5-48fa-81d2-a4f5aaa4fdbb",
		Name:         "Somnomancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.Optional(effects.WhenThisEnters("Somnomancer — you may tap target creature", effects.TapTheFirstLegalTarget),
				"Somnomancer — tap the target creature?"), effects.TargetCreature("target creature")),
		},
	})
}
