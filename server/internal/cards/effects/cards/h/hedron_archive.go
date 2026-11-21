package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "32263baa-d3f0-463f-92b3-4e9938476add",
		Name:         "Hedron Archive",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Draw two cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 2},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
