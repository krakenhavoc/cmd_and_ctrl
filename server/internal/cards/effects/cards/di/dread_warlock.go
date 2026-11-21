package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Warlock — Creature — Human Wizard Warlock {1}{B}{B}, 2/2 (Magic 2010, #5585):
//
//	"This creature can't be blocked except by black creatures."
//
// Manta Ray's block restriction with black: only a black creature can
// block it (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5791710d-c718-4f64-9c14-3668154716f1",
		Name:         "Dread Warlock",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedExceptBy(effects.OnSelf(), effects.OfColor("B"), "black creatures"),
		},
	})
}
