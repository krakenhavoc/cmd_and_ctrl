package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Reaper — Creature — Horror {3}{B}{B}{B}, 6/5 (Portal, #5630):
//
//	"Flying
//	When this creature enters, you lose 5 life."
//
// Foul Imp's body and enters trigger, with 5 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "bd73ab86-0ac9-4ce0-be41-f4ad257e74f6",
		Name:            "Dread Reaper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dread Reaper — you lose 5 life", effects.Do(effects.GainLife{Amount: -5})),
		},
	})
}
