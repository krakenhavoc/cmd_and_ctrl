package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inescapable Brute — Creature — Giant Warrior {5}{R}, 3/3 (Shadowmoor,
// #5591):
//
//	"Wither
//	 This creature must be blocked if able."
//
// Wither rides PrintedKeywords; the requirement is the shared MustBeBlocked
// static (Riveteers Decoy's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cb69e07c-c831-4992-8089-3a3a1278c45b",
		Name:            "Inescapable Brute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"wither"},
		Static:          []game.StaticAbility{effects.MustBeBlocked()},
	})
}
