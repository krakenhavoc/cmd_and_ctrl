package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zulaport Enforcer — {B} Creature — Human Warrior 1/1 (Rise of the Eldrazi, S82 OLD #5581):
//
//	Level up {4}
//	LEVEL 1-2: 3/3
//	LEVEL 3+: 5/5, This creature can't be blocked except by black creatures.
//
// A leveler: Level up is the keyword, each level symbol a band of LevelBands.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ab5fde8-e7db-4265-81c6-2a6488ba56c4",
		Name:         "Zulaport Enforcer",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.LevelUpKeyword("{4}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 2, Power: 3, Toughness: 3},
			effects.LevelBand{From: 3, Power: 5, Toughness: 5},
		),
		BlockRules: []game.BlockRule{
			effects.OldB478CantBeBlockedExceptByAtLevel(3, effects.And(effects.Creature(), effects.OfColor("B")), "black creatures"),
		},
	})
}
