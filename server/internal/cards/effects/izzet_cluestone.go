package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Izzet Cluestone — Artifact {3}:
//
//	"{T}: Add {U} or {R}.
//	 {U}{R}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {U}{R}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c76bb90-92fb-428b-8e94-def4cf4b6f2d",
		Name:         "Izzet Cluestone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U}{R}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{U}{R}"},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
