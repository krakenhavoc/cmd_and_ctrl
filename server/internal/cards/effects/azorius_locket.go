package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Azorius Locket — Artifact {3}:
//
//	"{T}: Add {W} or {U}.
//	 {W/U}{W/U}{W/U}{W/U}, {T}, Sacrifice this artifact: Draw two cards."
//
// A two-colour mana rock (Dimir Keyrune's mana ability) plus an activated
// ability whose cost is four hybrid symbols, each payable with {W} or {U},
// the tap and the sacrifice. All three are paid at announcement, so the
// Locket is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca13dda2-fbb0-4b6b-9126-3c8df04dee2b",
		Name:         "Azorius Locket",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U}",
			Label:    "Add {W} or {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{W/U}{W/U}{W/U}{W/U}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{W/U}{W/U}{W/U}{W/U}"},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
