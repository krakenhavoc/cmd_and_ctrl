package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Izzet Locket — Artifact {3} (Guilds of Ravnica, #5525):
//
//	"{T}: Add {U} or {R}.
//	 {U/R}{U/R}{U/R}{U/R}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's shape: a two-colour mana rock, and a draw-two whose cost
// is four hybrid symbols (each paid with {U} or {R}), the tap and the
// sacrifice, all paid at announcement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7ee07546-2c2e-4679-b671-7ea509b1d3ea",
		Name:         "Izzet Locket",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|R}",
			Label:    "Add {U} or {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U/R}{U/R}{U/R}{U/R}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{U/R}{U/R}{U/R}{U/R}"},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
