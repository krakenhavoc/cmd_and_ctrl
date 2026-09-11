package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hedron Archive — Artifact {4} (EDHREC rank 545):
//
//	"{T}: Add {C}{C}.
//	 {2}, {T}, Sacrifice this artifact: Draw two cards."
//
// Mind Stone's big brother: a four-mana rock that taps for two and
// cashes in for two cards. The same three-component activated cost
// (mana, tap, sacrifice self); the two abilities compete for the
// one tap, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "32263baa-d3f0-463f-92b3-4e9938476add",
		Name:     "Hedron Archive",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{2}, {T}, Sacrifice this artifact: Draw two cards.",
			Cost:  Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
			},
		}},
	})
}
