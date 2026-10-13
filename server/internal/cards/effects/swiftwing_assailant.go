package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swiftwing Assailant — Creature — Bird Warrior {3}{W}, 3/3:
//
//	"Flying
//	 Start your engines!
//	 Max speed — This creature gets +0/+1 and has vigilance."
//
// ADR 0138 (#2711).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "46b36d25-25bd-48c5-9a53-75230bb27dee",
		Name:            "Swiftwing Assailant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", StartYourEngines},
		Static: append([]game.StaticAbility{MaxSpeedSelfPump(0, 1)},
			MaxSpeedSelfKeywords("vigilance")...),
	})
}
