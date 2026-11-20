package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Selesnya Locket — Artifact {3} (Guilds of Ravnica, #5525):
//
//	"{T}: Add {G} or {W}.
//	 {G/W}{G/W}{G/W}{G/W}, {T}, Sacrifice this artifact: Draw two cards."
//
// Azorius Locket's shape: a two-colour mana rock, and a draw-two whose cost
// is four hybrid symbols (each paid with {G} or {W}), the tap and the
// sacrifice, all paid at announcement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9ecd98d5-8d66-4cb0-bada-23dfa0faacc0",
		Name:         "Selesnya Locket",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G/W}{G/W}{G/W}{G/W}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{G/W}{G/W}{G/W}{G/W}"},
			Effect:  Do(DrawCards{N: 2}),
		}},
	})
}
