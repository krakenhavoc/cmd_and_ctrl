package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// One with the Stars — Enchantment — Aura {3}{U}:
//
//	"Enchant creature or enchantment
//	 Enchanted permanent is an enchantment and loses all other card types.
//	 (It still has its abilities, but it's no longer a creature.)"
//
// A layer 4 type change; the layer pass drops the subtypes the lost types
// carried (CR 205.1a), so a creature loses its creature types and an
// enchantment keeps Aura, Saga and the like.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0ab43fc3-53ec-4686-873d-8dbeee2634f4",
		Name:         "One with the Stars",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{{
			Label:                 "Enchanted permanent is an enchantment and loses all other card types.",
			Layer:                 game.Layer4Type,
			ContinuesAfterRemoval: true,
			AppliesTo:             effects.AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Types = []string{"Enchantment"}
			},
		}},
	})
}
