package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0e2575be-c596-4c8c-bf07-7941ca065721",
		Name:         "Dreamstone Hedron",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}",
			Label:    "Add {C}{C}{C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: Draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 3},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.DrawCards{Player: item.Controller, N: 3}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
