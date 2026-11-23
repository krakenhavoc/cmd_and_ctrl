package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stormblood Berserker — Creature — Human Berserker {1}{R}, 1/1 (Ravnica: Clue Edition, #5929): "Bloodthirst 2. Menace."
//
// Bloodthirst is the shared entry replacement (bloodthirst.go); menace is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "84e56125-3fbb-4229-8aa0-8f8acff27038",
		Name:            "Stormblood Berserker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Replacements:    []game.ReplacementEffect{effects.Bloodthirst(2, "Stormblood Berserker")},
	})
}
