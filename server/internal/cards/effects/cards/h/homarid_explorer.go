package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Homarid Explorer — Creature — Homarid Scout {3}{U}, 3/3 (Dominaria, #5530):
//
//	"When this creature enters, target player mills four cards."
//
// Wall of Lost Thoughts' enters trigger: target player mills four.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ac30745c-addf-433a-be73-d3d9f4cc4328",
		Name:         "Homarid Explorer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Homarid Explorer — target player mills four cards", effects.TargetPlayerMills(4)),
				effects.TargetPlayer("target player")),
		},
	})
}
