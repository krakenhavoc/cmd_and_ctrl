package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scavenging Scarab — Creature — Insect {3}{B} (Darksteel, #5605):
//
//	"This creature can't block."
//
// Spineless Thug's self restriction (CR 509.1b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ce2f3ad-23b3-40fe-bb31-25dbe9d1150d",
		Name:         "Scavenging Scarab",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
