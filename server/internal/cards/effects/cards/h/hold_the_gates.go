package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hold the Gates — Enchantment {2}{W} (Gatecrash, S82 OLD #5565):
//
//	"Creatures you control get +0/+1 for each Gate you control and have
//	 vigilance."
//
// Two statics: a layer 7c bonus to toughness counted on every pass over the
// Gates its controller controls (Kemba's Banner's live count), and a layer 6
// vigilance grant (Brave the Sands). Both cover only creatures its controller
// controls; the enchantment is not a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e6fcbc0-691a-4bd7-b063-debc86f4506a",
		Name:         "Hold the Gates",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:    "Creatures you control get +0/+1 for each Gate you control.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.IsCreature() && target.Controller == source.Controller
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					c.Toughness += effects.CountPermanentsYouControl(g, source, func(p *game.Card) bool { return p.HasSubtypeWord("Gate") })
				},
			},
			effects.B16GrantKeywords(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			}, "vigilance"),
		},
	})
}
