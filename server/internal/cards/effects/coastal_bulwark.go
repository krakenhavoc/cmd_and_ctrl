package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coastal Bulwark — Artifact Creature — Wall {2}, 1/3 (The Brothers' War, #5482):
//
//	"Defender
//	 This creature gets +2/+0 as long as you control an Island.
//	 {2}, {T}: Surveil 1."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "449ed4b4-9e1e-4181-9090-af91802eba49",
		Name:            "Coastal Bulwark",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+0 as long as you control an Island.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfWhileYouControlA(Subtype("Island")),
			Apply:     pumpBy(2, 0),
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Mana: "{2}", Tap: true},
			Effect:  Do(Surveil{N: 1}),
		}},
	})
}
