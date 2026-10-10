package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omnath, Locus of the Void — Legendary Creature — Elemental {7}, 6/6:
//
//	"Omnath gets +1/+1 for each unspent mana you have.
//	 If you would lose unspent mana, that mana becomes colorless
//	 instead.
//	 Landfall — Whenever a land you control enters, add {C}{C}."
//
// The conversion is Spec.ManaPool's ManaPoolBecomesColorless (Kruphix,
// God of Horizons does the same): the mana stays in the pool as {C}.
// The P/T clause is Omnath, Locus of Mana's layer 7c static over every
// colour, declaring DependsOnManaPool so the layer cache follows the
// pool. Landfall is an ordinary triggered ability; the {C}{C} it adds
// is kept as colorless mana by the clause above, so it carries over.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "579225b5-50e2-4891-8871-0bf6dbc07e33",
		Name:         "Omnath, Locus of the Void",
		Completeness: CompletenessFull,
		ManaPool: []game.ManaPoolStatic{
			{Kind: game.ManaPoolBecomesColorless},
		},
		Static: []game.StaticAbility{{
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7C_Modify,
			AppliesTo:         selfOnly,
			DependsOnManaPool: true,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				p := g.PlayerByIDForEffect(source.Controller)
				if p == nil {
					return
				}
				n := len(p.ManaPool)
				c.Power += n
				c.Toughness += n
			},
		}},
		Triggered: []game.TriggeredAbility{
			Landfall("Omnath, Locus of the Void — add {C}{C}", Do(AddMana{Produced: "{C}{C}"})),
		},
	})
}
