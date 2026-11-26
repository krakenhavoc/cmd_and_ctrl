package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Undead Leotau — {5}{B} Creature — Zombie Cat 3/4 (Shards of Alara, S82 OLD
// b516):
//
//	"{R}: This creature gets +1/-1 until end of turn.
//	 Unearth {2}{B} (...)"
//
// Breathstealer's pump in red, beside unearth.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd7df06c-9434-46ce-893c-80081fbc2df7",
		Name:         "Undead Leotau",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{R}: This creature gets +1/-1 until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: -1}},
				Cost:    effects.ManaCost("{R}"),
				Effect:  effects.ThisGetsUntilEndOfTurn(1, -1, "Undead Leotau — +1/-1 until end of turn"),
			},
			effects.Unearth("{2}{B}"),
		},
	})
}
