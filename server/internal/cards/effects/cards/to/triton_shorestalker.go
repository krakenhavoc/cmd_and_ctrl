package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triton Shorestalker — Creature — Merfolk Rogue {U}, 1/1:
//
//	"This creature can't be blocked."
//
// The S24 restriction bit on the creature itself, read at the block
// declaration (CR 509.1b). No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "71a60247-17e5-44bd-9fed-f7d550b09ef3",
		Name:         "Triton Shorestalker",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBeBlocked)},
	})
}
