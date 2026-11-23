package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "46b36d25-25bd-48c5-9a53-75230bb27dee",
		Name:            "Swiftwing Assailant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", effects.StartYourEngines},
		Static: append([]game.StaticAbility{effects.MaxSpeedSelfPump(0, 1)},
			effects.MaxSpeedSelfKeywords("vigilance")...),
	})
}
