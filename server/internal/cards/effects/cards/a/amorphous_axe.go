package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Amorphous Axe — Artifact — Equipment {2} (Commander Legends):
//
//	"Equipped creature gets +3/+0 and is every creature type.
//	 Equip {3}"
//
// "Every creature type" is the layer-4 all-types fact (CR 205.3m), as a
// changeling has it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe340efa-6b69-4ccc-a13c-3688f1c44c29",
		Name:         "Amorphous Axe",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.EquippedCreatureIs(game.Layer4Type, "Equipped creature is every creature type.", func(c *game.Characteristic) {
				c.AllCreatureTypes = true
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
