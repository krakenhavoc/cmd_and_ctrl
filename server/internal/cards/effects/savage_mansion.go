package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Savage Mansion — Land (Marvel's Spider-Man, #4460):
//
//	"This land enters tapped.
//	 {T}: Add {R} or {G}.
//	 {4}, {T}: Surveil 1."
//
// The enters-tapped replacement, a two-colour mana ability, and an
// instant-speed surveil that costs {4} and the tap. The surveil is the
// real look-at-one prompt (Surveil).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "1af15c1d-a41c-44cc-9614-d72694dd26e8",
		Name:          "Savage Mansion",
		Completeness:  CompletenessFull,
		Replacements:  []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{dualManaAbility("R", "G")},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Surveil 1.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Surveil{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}
