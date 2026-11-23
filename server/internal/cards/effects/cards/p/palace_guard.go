package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Palace Guard — Creature — Human Soldier {2}{W}, 1/4:
//
//	"This creature can block any number of creatures."
//
// CanBlockAnyNumber on itself (#1706). Blocking several attackers, it
// takes all of their damage and divides its own 1 among them
// (CR 510.1d).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c92f375-ae6a-4be5-a499-ab87d6bdc49b",
		Name:         "Palace Guard",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.CanBlockAnyNumber(effects.SelfOnly)},
	})
}
