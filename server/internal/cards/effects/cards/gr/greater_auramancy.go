package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Greater Auramancy — Enchantment {1}{W} (Enchanting Tales):
//
//	"Other enchantments you control have shroud.
//	 Enchanted creatures you control have shroud."
//
// Two layer-6 keyword grants, re-read on every layer pass. The first reaches
// each other enchantment its controller controls (never Greater Auramancy
// itself); the second each creature of its controller that has an Aura
// attached (Syr Armont's "enchanted creatures"), whoever controls the Aura.
// Shroud stops targeting by anyone, its controller included (CR 702.18).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e6e888ea-9935-4df2-a344-c54637e55f43",
		Name:         "Greater Auramancy",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsEnchantment() && target.Controller == source.Controller && target.InstanceID != source.InstanceID
			}, "shroud"),
			effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller && effects.IsEnchantedByAura(g, target.InstanceID)
			}, "shroud"),
		},
	})
}
