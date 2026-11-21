package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravine Raider — Creature — Lizard Rogue {B}, 1/1 (Bloomburrow, #4467):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 {1}{B}: This creature gets +1/+1 until end of turn."
//
// Menace is a printed keyword. The pump is Burrog Banemaker's: no {T} in
// the cost, so it can be activated any number of times, at instant
// speed, while the Raider is summoning sick.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c2e84dc-b4f5-4eca-af07-1c53a1caae7c",
		Name:            "Ravine Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{B}: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    effects.ManaCost("{1}{B}"),
			Effect:  effects.ThisGetsUntilEndOfTurn(1, 1, "Ravine Raider — +1/+1 until end of turn"),
		}},
	})
}
