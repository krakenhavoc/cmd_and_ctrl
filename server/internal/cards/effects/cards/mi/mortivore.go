package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mortivore — Creature — Lhurgoyf {2}{B}{B}, */*:
//
//	"Mortivore's power and toughness are each equal to the number of
//	 creature cards in all graveyards.
//	 {B}: Regenerate this creature."
//
// Lord of Extinction's layer 7a characteristic-defining ability over
// creature cards only (Nighthowler's count), plus Albino Troll's
// regeneration ability. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f554218-2072-4d9b-b4d4-34f161f1f2a9",
		Name:         "Mortivore",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Power and toughness are each equal to the number of creature cards in all graveyards.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := effects.B41CreatureCardsInAllGraveyards(g)
				c.Power = n
				c.Toughness = n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}: Regenerate this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Regenerate{Target: item.SourceCardID}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
