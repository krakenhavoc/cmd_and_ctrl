package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Omnath, Locus of Mana — Legendary Creature — Elemental {2}{G}, 1/1:
//
//	"You don't lose unspent green mana as steps and phases end.
//	 Omnath gets +1/+1 for each unspent green mana you have."
//
// The keep clause is Spec.ManaPool (#2166). The P/T clause is a layer 7c
// static that reads the controller's pool, so it declares
// DependsOnManaPool (#2363): a pool change is not a zone move or a counter,
// and without the hint the layer cache would serve the old size until
// something unrelated invalidated it. The hint is gated, so a table with
// no such permanent pays nothing on a land tap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9da896e1-2256-425b-b801-1ae6f0470559",
		Name:         "Omnath, Locus of Mana",
		Completeness: CompletenessFull,
		ManaPool: []game.ManaPoolStatic{
			{Kind: game.ManaPoolKeep, Colors: []string{"G"}},
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
				n := 0
				for _, tok := range p.ManaPool {
					if tok.Color == "G" {
						n++
					}
				}
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
