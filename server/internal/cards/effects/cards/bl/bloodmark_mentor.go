package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodmark Mentor — Creature — Goblin Warrior {1}{R}, 1/1 (Shadowmoor, #5591):
//
//	"Red creatures you control have first strike."
//
// A layer 6 keyword grant over the red creatures its controller controls,
// itself included, read from current colours.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c22b218-cf40-47af-8c11-73c581a43390",
		Name:         "Bloodmark Mentor",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.OldB540ColorCreaturesYouControlHave("R", "first strike")},
	})
}
