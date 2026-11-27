package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gate Hound — Creature — Dog {2}{W}, 1/1 (Ravnica: City of Guilds):
//
//	"Creatures you control have vigilance as long as this creature is enchanted."
//
// A layer-6 keyword grant to the controller's creatures (Ajani, the Greathearted's
// shape) that applies only while an Aura is attached to this creature, read live
// on every layer pass (CR 303.4), so it follows the Aura on and off and covers
// creatures that arrive later. The Hound itself is a creature you control, so it
// has vigilance too.
//
// No simplification.
func init() {
	grant := effects.B16GrantKeywords(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.IsCreature() && target.Controller == source.Controller && effects.IsEnchantedByAura(g, source.InstanceID)
	}, "vigilance")
	grant.Label = "Creatures you control have vigilance as long as this creature is enchanted."
	effects.Register(effects.Spec{
		OracleID:     "762e01d8-c16a-44b8-8ba0-052d1a2f312f",
		Name:         "Gate Hound",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{grant},
	})
}
