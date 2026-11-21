package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spined Megalodon — Creature — Shark {5}{U}{U}, 5/7 (Core Set 2021, #5511):
//
//	"Hexproof
//	 Whenever this creature attacks, scry 1."
//
// Prognostic Sphinx's attack scry at 1, with hexproof printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3a124fe5-f525-43e5-9746-501c68a98330",
		Name:            "Spined Megalodon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Purpose:         game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Spined Megalodon — scry 1", effects.Do(effects.Scry{N: 1})),
		},
	})
}
