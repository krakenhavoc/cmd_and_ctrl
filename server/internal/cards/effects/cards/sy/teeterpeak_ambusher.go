package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teeterpeak Ambusher — Creature — Goblin Warrior {1}{R}, 1/3 (Zendikar
// Rising, #5508):
//
//	"{2}{R}: This creature gets +2/+0 until end of turn."
//
// Rampaging Raptor's firebreathing-style self pump, as many times as it
// is paid for.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe769071-3f99-411b-878a-57a2200649a0",
		Name:         "Teeterpeak Ambusher",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{R}: This creature gets +2/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{R}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(2, 0, "Teeterpeak Ambusher — +2/+0"),
		}},
	})
}
