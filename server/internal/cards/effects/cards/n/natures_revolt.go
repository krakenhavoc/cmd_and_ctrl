package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nature's Revolt — Enchantment {3}{G}{G} (Tempest, S82 OLD batch 765, #5628):
//
//	"All lands are 2/2 creatures that are still lands."
//
// Opalescence's two layers over every land, whoever controls it: the creature
// type added in layer 4 (CR 613.1d), so the land keeps its land type, and the
// base 2/2 in layer 7b (CR 613.4b). A land that came under a player's control
// this turn is summoning sick as any creature is (CR 302.6); the layer pass
// reads that off the permanent, so nothing more is needed here.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0df0503b-25e7-499b-a703-a867f2da790c",
		Name:         "Nature's Revolt",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			{
				Label:     "All lands are creatures that are still lands.",
				Layer:     game.Layer4Type,
				AppliesTo: naturesRevoltApplies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Creature") {
						c.Types = append(append([]string(nil), c.Types...), "Creature")
					}
				},
			},
			{
				Label:     "All lands have base power and toughness 2/2.",
				Layer:     game.Layer7PT,
				SubLayer:  game.SubLayer7B_Set,
				AppliesTo: naturesRevoltApplies,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power, c.Toughness = 2, 2
				},
			},
		},
	})
}

func naturesRevoltApplies(target *game.Card, _ *game.Game, _ *game.Card) bool {
	return target.IsLand()
}
