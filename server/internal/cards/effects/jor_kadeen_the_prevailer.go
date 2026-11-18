package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jor Kadeen, the Prevailer — Legendary Creature — Human Warrior {3}{R}{W}, 5/4:
//
//	"First strike
//	 Metalcraft — Creatures you control get +3/+0 as long as you control three
//	 or more artifacts."
//
// A layer 7c modify over every creature its controller controls, Jor Kadeen
// included, gated on the artifact count read through the layered view on every
// recompute (an animated artifact counts, an artifact that leaves switches the
// bonus off).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7c4dbaa8-e24f-4126-81fe-3f218f520d8c",
		Name:            "Jor Kadeen, the Prevailer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Static: []game.StaticAbility{{
			Label:    "Metalcraft — Creatures you control get +3/+0 as long as you control three or more artifacts.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller &&
					jorKadeenPrevailerMetalcraft(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += 3
			},
		}},
	})
}

// jorKadeenPrevailerMetalcraft is metalcraft for the source's controller. It
// walks g.Battlefield.Cards directly: a static's applies-to runs under the
// layer pass's own read of the battlefield.
func jorKadeenPrevailerMetalcraft(g *game.Game, source *game.Card) bool {
	if g == nil || g.Battlefield == nil {
		return false
	}
	n := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == source.Controller && c.IsArtifact() {
			n++
			if n >= 3 {
				return true
			}
		}
	}
	return false
}
