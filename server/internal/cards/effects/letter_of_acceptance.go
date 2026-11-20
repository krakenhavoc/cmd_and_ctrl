package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Letter of Acceptance — Artifact {3} (Strixhaven: School of Mages,
// #5502):
//
//	"{T}: Add one mana of any color.
//	 {2}, {T}, Sacrifice this artifact: Draw a card."
//
// Alloy Myr's any-color mana ability and Mind Stone's cash-in row at {2}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0cb503e-9814-4280-9c29-4eff3379bc24",
		Name:         "Letter of Acceptance",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{2}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNForEffect(item.Controller, 1)
			},
		}},
	})
}
