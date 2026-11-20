package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimir Cluestone — Artifact {3}:
//
//	"{T}: Add {U} or {B}.
//	 {U}{B}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {U}{B}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "62f834c7-4765-4b6f-812c-7ce14e13ea0d",
		Name:         "Dimir Cluestone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U}{B}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{U}{B}"},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
