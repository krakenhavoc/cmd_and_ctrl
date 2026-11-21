package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inkfathom Infiltrator — Creature — Merfolk Rogue {U/B}{U/B}, 2/1
// (Shadowmoor, #5591):
//
//	"This creature can't block and can't be blocked."
//
// Changeling Outcast's restriction line exactly: both bits in one self
// static.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b43c7de-8c21-4137-b6b8-45965e4bb558",
		Name:         "Inkfathom Infiltrator",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock | game.CantBeBlocked)},
	})
}
