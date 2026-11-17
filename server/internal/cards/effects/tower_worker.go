package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tower Worker — Artifact Creature — Assembly-Worker {3}, 1/3 (The
// Brothers' War, #5482):
//
//	"Reach
//	 {T}: Add {C}. If you control creatures named Mine Worker and Power
//	 Plant Worker, add {C}{C}{C} instead."
//
// A mana ability (CR 605.1a), so the trio is read as the mana is produced
// (ProducedFunc): three {C} with a Mine Worker and a Power Plant Worker
// under your control, else one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2f995678-c315-4afe-845b-8c8ae863de6e",
		Name:            "Tower Worker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{Tap: true},
			ProducedFunc: func(g *game.Game, controller, _ uuid.UUID) string {
				if s82BroB11AssemblyTrio(g, controller, "Mine Worker", "Power Plant Worker") {
					return "{C}{C}{C}"
				}
				return "{C}"
			},
			Label: "{T}: Add {C}. If you control creatures named Mine Worker and Power Plant Worker, add {C}{C}{C} instead.",
		}},
	})
}
