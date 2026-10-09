package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Murmuring Volume — Artifact — Book {3} (Reality Fracture):
//
//	"{T}: Add one mana of any color.
//	 {2}, {T}, Discard a card: Draw a card."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4baa7847-dfcc-4aa3-aa64-2deeaa3f013b",
		Name:         "Murmuring Volume",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Discard a card: Draw a card.",
			Cost:  Plus(ManaCost("{2}"), TapCost(), DiscardACard()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNForEffect(item.Controller, 1)
			},
		}},
	})
}
