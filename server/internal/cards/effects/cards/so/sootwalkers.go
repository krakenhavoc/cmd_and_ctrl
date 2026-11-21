package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sootwalkers — Creature — Elemental Rogue {2}{B/R}{B/R}, 3/3 (Shadowmoor,
// #5591):
//
//	"This creature can't be blocked by white creatures."
//
// A block rule, Shield Mare's: a blocker of that colour, as it is now, is
// refused (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1492e778-b85e-4675-b48c-bc0a2eb1d105",
		Name:         "Sootwalkers",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{
			effects.CantBeBlockedBy(effects.OnSelf(), effects.And(effects.Creature(), effects.OfColor("W")), "white creatures"),
		},
	})
}
