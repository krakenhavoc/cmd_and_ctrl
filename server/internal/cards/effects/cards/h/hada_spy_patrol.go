package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hada Spy Patrol — Creature — Human Rogue {1}{U}, 1/1 (Commander 2013, S82
// older sets):
//
//	"Level up {2}{U} ({2}{U}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 1-2
//	 2/2
//	 This creature can't be blocked.
//	 LEVEL 3+
//	 3/3
//	 Shroud (This creature can't be the target of spells or abilities.)
//	 This creature can't be blocked."
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b): each band carries its
// own power/toughness box and statics, gated on the level counters. The
// unblockable restriction is the creature's own static (RestrictSelf); the
// shroud is the engine's keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8db3152a-11b1-4c59-a745-1c40a9b235b7",
		Name:         "Hada Spy Patrol",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{2}{U}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 2, Power: 2, Toughness: 2, Statics: []game.StaticAbility{
				effects.Labelled(effects.RestrictSelf(game.CantBeBlocked), "This creature can't be blocked."),
			}},
			effects.LevelBand{From: 3, Power: 3, Toughness: 3, Keywords: []string{"shroud"}, Statics: []game.StaticAbility{
				effects.Labelled(effects.RestrictSelf(game.CantBeBlocked), "This creature can't be blocked."),
			}},
		),
	})
}
