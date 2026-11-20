package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golgari Locket — Artifact {3} (Guilds of Ravnica, #5525):
//
//	"{T}: Add {B} or {G}.
//	 {B/G}{B/G}{B/G}{B/G}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's shape: a two-colour mana rock, and a draw-two whose cost
// is four hybrid symbols (each paid with {B} or {G}), the tap and the
// sacrifice, all paid at announcement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dba6e24e-7875-46b9-b51d-4cd71568ecfa",
		Name:         "Golgari Locket",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B/G}{B/G}{B/G}{B/G}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{B/G}{B/G}{B/G}{B/G}"},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
