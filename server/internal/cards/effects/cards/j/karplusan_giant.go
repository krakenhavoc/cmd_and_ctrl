package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Karplusan Giant — Creature — Giant {6}{R}, 3/3 (Ice Age, S82 OLD, #5636):
//
//	"Tap an untapped snow land you control: This creature gets +1/+1 until end of
//	 turn."
//
// The cost taps a snow land (not a mana ability, no {T} symbol: Zombie
// Trailblazer's tap cost), so it can be paid repeatedly with different lands.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f0fac3bc-a566-448e-a9a2-10927a49f182",
		Name:         "Karplusan Giant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped snow land you control: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.TapAnUntapped("an untapped snow land you control", effects.Land(), effects.Snow(), effects.YouControl()),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Karplusan Giant — +1/+1 until end of turn"),
		}},
	})
}
