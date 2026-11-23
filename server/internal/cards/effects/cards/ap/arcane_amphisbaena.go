package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arcane Amphisbaena — Creature — Snake {1}{G}, 1/1 (Reality
// Fracture):
//
//	"Deathtouch
//	 When this creature enters, empower Jace 2."
//
// ADR 0139: an enters trigger on the keyword action.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5c151e3b-fcf2-4adb-a60b-9eff0c360705",
		Name:            "Arcane Amphisbaena",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Arcane Amphisbaena — empower Jace 2", effects.Do(effects.EmpowerJace{N: 2})),
		},
	})
}
