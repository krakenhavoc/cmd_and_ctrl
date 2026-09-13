package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mind Stone — Artifact {2}:
//
//	"{T}: Add {C}."
//	"{1}, {T}, Sacrifice Mind Stone: Draw a card."
//
// The card that made two-mana rocks playable in slow decks: ramp
// early, cash it in for a card once you no longer need the mana.
//
// The second ability is the first catalog use of a THREE-component
// activated cost — mana, tap, and sacrifice-self together — which
// S21 sub-PR 2's AbilityCost made expressible. Cost order is handled
// engine-side (mana → tap → life → sacrifice), and every component
// is validated before any is paid, so an attempt with the stone
// already tapped fails without eating it.
//
// Note the two abilities both tap: the mana ability and the draw
// ability compete for the same tap, exactly as in paper.
func init() {
	Register(Spec{
		OracleID:     "c97361b5-af16-4a7b-af85-a429dbaf4ad2",
		Name:         "Mind Stone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{1}, {T}, Sacrifice Mind Stone: Draw a card.",
			Cost: game.AbilityCost{
				Tap:           true,
				SacrificeSelf: true,
				Mana:          "{1}",
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return g.DrawNForEffect(item.Controller, 1)
			},
		}},
	})
}
