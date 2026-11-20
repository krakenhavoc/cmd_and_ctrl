package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boros Cluestone — Artifact {3}:
//
//	"{T}: Add {R} or {W}.
//	 {R}{W}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {R}{W}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6c3b420-9aca-4fab-b7bb-1814bd2d93f6",
		Name:         "Boros Cluestone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|W}",
			Label:    "Add {R} or {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}{W}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{R}{W}"},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
