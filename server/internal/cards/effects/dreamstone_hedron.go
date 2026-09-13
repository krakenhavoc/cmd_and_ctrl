package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dreamstone Hedron — Artifact {6} (EDHREC rank 1938):
//
//	"{T}: Add {C}{C}{C}.
//	 {3}, {T}, Sacrifice this artifact: Draw three cards."
//
// Hedron Archive's big brother: a six-mana rock that taps for three
// and cashes in for three cards. The same three-component activated
// cost (mana, tap, sacrifice self); the two abilities compete for
// the one tap, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e2575be-c596-4c8c-bf07-7941ca065721",
		Name:         "Dreamstone Hedron",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}",
			Label:    "Add {C}{C}{C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{3}, {T}, Sacrifice this artifact: Draw three cards.",
			Cost:  Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 3}.Apply(NewContext(g, item))
			},
		}},
	})
}
