package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hulking Ogre — Creature — Ogre {2}{R}, 3/3 (Urza's Destiny, #5621):
//
//	"This creature can't block."
//
// Hulking Cyclops's static restriction on itself (RestrictSelf with
// CantBlock), read by the block check (CR 509.1b). It can still attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a95aa9ba-fa61-4a3d-a493-1ea790bd20ad",
		Name:         "Hulking Ogre",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
