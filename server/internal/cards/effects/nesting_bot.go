package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "7ffb8c2e-1737-4df3-b310-384a5000bfa5",
		Name:            "Nesting Bot",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Nesting Bot — create a 1/1 colorless Servo artifact creature token", Do(CreateToken{Template: TokenCard("1/1 colorless Servo artifact"), N: 1})),
		},
		Static: []game.StaticAbility{MaxSpeedSelfPump(1, 0)},
	})
}
