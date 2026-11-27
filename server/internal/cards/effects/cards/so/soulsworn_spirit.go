package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soulsworn Spirit — Creature — Spirit {3}{U}, 2/1 (Return to Ravnica, #5566):
//
//	"This creature can't be blocked.
//	 When this creature enters, detain target creature an opponent controls."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e67e3cf1-6da9-4b52-a945-47bf54f69ffa",
		Name:         "Soulsworn Spirit",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Soulsworn Spirit — detain target creature an opponent controls", effects.OldB423DetainTargets("Soulsworn Spirit — detain")),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
