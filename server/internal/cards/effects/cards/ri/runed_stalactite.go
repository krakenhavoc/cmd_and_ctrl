package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runed Stalactite — Artifact — Equipment {1} (Lorwyn, S82 OLD b557):
//
//	"Equipped creature gets +1/+1 and is every creature type.
//	 Equip {2}"
//
// Amorphous Axe's layer 4 every-creature-type fact with a +1/+1 layer 7c pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c169307-71f5-4679-8b5a-d60c1109b9db",
		Name:         "Runed Stalactite",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.EquippedCreatureIs(game.Layer4Type, "Equipped creature is every creature type.", func(c *game.Characteristic) {
				c.AllCreatureTypes = true
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
