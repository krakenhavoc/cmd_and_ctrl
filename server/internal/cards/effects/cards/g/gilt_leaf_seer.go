package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gilt-Leaf Seer — Creature — Elf Shaman {2}{G}, 2/2 (Lorwyn, S82 OLD b547,
// #5593):
//
//	"{G}, {T}: Look at the top two cards of your library, then put them back
//	 in any order."
//
// Aven Fateshaper's look-and-reorder at two cards: the player sees the top
// two and puts them back in the order they choose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eab27405-f590-40b7-ad7c-0e6c52410c0a",
		Name:         "Gilt-Leaf Seer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}, {T}: Look at the top two cards of your library, then put them back in any order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{G}"), effects.TapCost()),
			Effect:  effects.Do(effects.LookAtTop{N: 2}),
		}},
	})
}
