package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spineless Thug — Creature — Phyrexian Zombie Mercenary {1}{B}, 2/2
// (Nemesis, #5618):
//
//	"This creature can't block."
//
// Sightless Ghoul's restriction without undying: RestrictSelf with
// game.CantBlock.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78bbd213-e8f4-40da-aef5-853244e95756",
		Name:         "Spineless Thug",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
	})
}
