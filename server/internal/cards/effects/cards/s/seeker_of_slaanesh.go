package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seeker of Slaanesh — Creature — Demon {3}{R}, 3/3 (Warhammer 40,000
// Commander):
//
//	"Haste
//	 Allure of Slaanesh — Each opponent must attack with at least one
//	 creature each combat if able."
//
// Allure of Slaanesh is a player-level CR 508.1d requirement (#6087): one
// requirement on each opponent, obeyed when any one of their creatures
// attacks anything, and never owed when none of them can attack for free.
// Two Seekers under different players are two requirements, which one
// attacker obeys together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47af21d3-b43f-4101-81cf-ca44db1a2706",
		Name:            "Seeker of Slaanesh",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Static:          []game.StaticAbility{effects.EachOpponentAttacksWithAtLeastOneCreature()},
	})
}
