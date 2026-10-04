package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2f554218-2072-4d9b-b4d4-34f161f1f2a9",
		Name:         "Mortivore",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				n := b41CreatureCardsInAllGraveyards(g)
				c.Power = n
				c.Toughness = n
			},
		}},
		Activated: []ActivatedAbility{{
			Label: "{B}: Regenerate this creature.",
			Cost:  ManaCost("{B}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Regenerate{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
