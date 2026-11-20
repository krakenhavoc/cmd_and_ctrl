package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dimir Locket — Artifact {3} (Guilds of Ravnica, #5525):
//
//	"{T}: Add {U} or {B}.
//	 {U/B}{U/B}{U/B}{U/B}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's shape: a two-colour mana rock, and a draw-two whose cost
// is four hybrid symbols (each paid with {U} or {B}), the tap and the
// sacrifice, all paid at announcement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ae4b679-2e3c-4f60-bcf4-b279917cdc54",
		Name:         "Dimir Locket",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U|B}",
			Label:    "Add {U} or {B}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{U/B}{U/B}{U/B}{U/B}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{U/B}{U/B}{U/B}{U/B}"},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
