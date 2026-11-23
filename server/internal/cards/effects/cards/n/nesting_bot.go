package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nesting Bot — Artifact Creature — Robot {W}, 1/1:
//
//	"Start your engines!
//	 When this creature dies, create a 1/1 colorless Servo artifact
//	 creature token.
//	 Max speed — This creature gets +1/+0."
//
// ADR 0138 (#2122, #2711). The dies trigger is the shared Servo token;
// max speed is a layer 7c pump switched on by the controller's speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ffb8c2e-1737-4df3-b310-384a5000bfa5",
		Name:            "Nesting Bot",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Nesting Bot — create a 1/1 colorless Servo artifact creature token", effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
		Static: []game.StaticAbility{effects.MaxSpeedSelfPump(1, 0)},
	})
}
