package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wall of Glare — Creature — Wall {1}{W}, 0/5:
//
//	"Defender
//	 This creature can block any number of creatures."
//
// Defender is the engine's keyword; the second line is
// CanBlockAnyNumber on itself (#1706). With power 0 it deals no damage,
// so there is nothing to divide and no prompt.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a59897cb-7ead-494a-a1d0-c9baf79bf18b",
		Name:            "Wall of Glare",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static:          []game.StaticAbility{effects.CanBlockAnyNumber(effects.SelfOnly)},
	})
}
