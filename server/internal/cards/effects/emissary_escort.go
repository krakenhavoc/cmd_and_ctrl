package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emissary Escort — Artifact Creature — Robot Soldier {1}{U}, 0/4 (Edge
// of Eternities, #4461):
//
//	"This creature gets +X/+0, where X is the greatest mana value among
//	 other artifacts you control."
//
// A layer 7c static (CR 613.4c) on the Escort alone, with Daretti,
// Rocketeer Engineer's greatest-mana-value scan narrowed to OTHER
// artifacts: the Escort is an artifact itself and never counts. Read
// live after the layers, so a permanent made an artifact by another
// effect counts, and with no other artifact X is 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d8b53f4-bdb5-4ba1-b6c0-f4846d8d3565",
		Name:         "Emissary Escort",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +X/+0, where X is the greatest mana value among other artifacts you control.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: selfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				x := 0
				for _, p := range g.BattlefieldCardsForEffect() {
					if p.InstanceID != source.InstanceID && p.Controller == source.Controller && p.IsArtifact() && p.ManaValue() > x {
						x = p.ManaValue()
					}
				}
				c.Power += x
			},
		}},
	})
}
