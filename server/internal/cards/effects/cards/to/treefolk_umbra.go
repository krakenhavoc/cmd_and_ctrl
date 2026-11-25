package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treefolk Umbra — Enchantment — Aura {2}{G} (Modern Horizons):
//
//	"Enchant creature
//	 Enchanted creature gets +0/+2 and assigns combat damage equal to its
//	 toughness rather than its power.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// The pump is PumpAttached. The damage rule is Felothar's
// DamageByToughnessAlways flag (CR 510.1c, toughness whether or not it is the
// greater), set by a layer 6 static on the enchanted creature only. Umbra armor
// is the shared replacement (CR 702.89a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa321366-9bc1-48c7-873a-0b12a026ce49",
		Name:         "Treefolk Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(0, 2),
			{
				Layer:     game.Layer6Ability,
				Label:     "Enchanted creature assigns combat damage equal to its toughness rather than its power.",
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.DamageByToughnessAlways = true
				},
			},
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}
