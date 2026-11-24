package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Boarding Party — Creature — Human Pirate {5}{R}, 6/3 (Commander Legends):
//
//	"Haste
//	 Cascade (When you cast this spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)"
//
// Haste and the engine's cascade trigger, Bloodbraid Elf's shape.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef69ce86-48ae-4d90-8e29-051583671321",
		Name:            "Boarding Party",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.Cascade()},
	})
}
