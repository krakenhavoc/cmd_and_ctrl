package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Callaphe, Beloved of the Sea — Legendary Enchantment Creature — Demigod {1}{U}{U}, */3
// (Theros Beyond Death, #5514):
//
//	"Callaphe's power is equal to your devotion to blue.
//	 Creatures and enchantments you control have "Spells your opponents cast that
//	 target this permanent cost {1} more to cast.""
//
// The power is a characteristic-defining ability (CR 604.3, layer 7a) re-read every
// recompute, devotion to blue through Callaphe's own text. The second ability is
// Jubilant Skybonder's shape: a granted cost modifier (CR 601.2f) reading the
// spell's targets, given to each creature and enchantment you control, itself
// included.
//
// No simplification.
func init() {
	const grant = "callaphe-beloved-of-the-sea/target-tax"
	tax := effects.CostsMore(1, "Spells your opponents cast that target this permanent cost {1} more to cast.",
		effects.OpponentsSpell(), effects.ThatTargetThisPermanent())
	tax.ReadsTargets = true
	effects.Register(effects.Spec{
		OracleID:     "faf081dc-1fb2-47ac-a9c7-8001b4904120",
		Name:         "Callaphe, Beloved of the Sea",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:           grant,
			CostModifiers: []game.CostModifier{tax},
			Text:          "Spells your opponents cast that target this permanent cost {1} more to cast.",
		}},
		Static: []game.StaticAbility{
			{
				Label:    "Power is equal to your devotion to blue.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7A_CDA,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return target.InstanceID == source.InstanceID
				},
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					c.Power = effects.DevotionTo(g, source.Controller, source.TextColor("U"))
				},
			},
			{
				Label: "Creatures and enchantments you control have \"Spells your opponents cast that target this permanent cost {1} more to cast.\"",
				Layer: game.Layer6Ability,
				AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
					return (target.IsCreature() || target.HasCardType("enchantment")) && target.Controller == source.Controller
				},
				GrantAbilities: []string{grant},
			},
		},
	})
}
