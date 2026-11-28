package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Transmogrifying Licid — Artifact Creature — Licid {3}, 2/2 (Exodus, S82 OLD
// b740, #5626):
//
//	"{1}, {T}: This creature loses this ability and becomes an Aura
//	 enchantment with enchant creature. Attach it to target creature. You may
//	 pay {1} to end this effect.
//	 Enchanted creature gets +1/+1 and is an artifact in addition to its other
//	 types."
//
// The licid seam (#5781): LicidAbility, LicidForm and the enchant clause. In
// its Aura form it is an Aura enchantment and nothing else (CR 205.1a: it is no
// longer an artifact or a creature), and the Aura's own text is the ordinary
// attachment helpers: +1/+1 and the layer-4 added artifact type (Silverskin
// Armor's), both of which refer to nothing while the licid is a creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4f484927-0e8e-493d-93bc-ed190965dad9",
		Name:         "Transmogrifying Licid",
		Completeness: effects.CompletenessFull,
		ReturnedAura: effects.EnchantCreature(),
		Static: append(effects.LicidForm(0),
			effects.PumpAttached(1, 1),
			game.StaticAbility{
				Label:     "Enchanted creature is an artifact in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Artifact") {
						c.Types = append(c.Types, "Artifact")
					}
				},
			}),
		Activated: []effects.ActivatedAbility{effects.LicidAbility("{1}", "{1}")},
	})
}
