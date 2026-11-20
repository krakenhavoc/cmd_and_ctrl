package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Cluestone — Artifact {3}:
//
//	"{T}: Add {W} or {U}.
//	 {W}{U}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {W}{U}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27e04c41-f42c-4d60-8a71-ec2d7c326f64",
		Name:         "Azorius Cluestone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U}",
			Label:    "Add {W} or {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{W}{U}"},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
