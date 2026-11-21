package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bruenor Battlehammer — Legendary Creature — Dwarf Warrior {2}{R}{W}, 5/3
// (Adventures in the Forgotten Realms Commander):
//
//	"Each creature you control gets +2/+0 for each Equipment attached to it.
//	 You may pay {0} rather than pay the equip cost of the first equip
//	 ability you activate each turn."
//
// The pump is a layer 7c static on each creature you control, counting every
// Equipment attached to it, whoever controls it, read on every recompute. The
// {0} is Kíli the Resourceful's CR 118.9 alternative cost on an equip
// ability, with no condition: each of your equip abilities has a twin row
// costing {0} until you have activated an equip ability this turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea7de96f-db69-421d-a951-ffaeac140379",
		Name:         "Bruenor Battlehammer",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Each creature you control gets +2/+0 for each Equipment attached to it.",
			Layer: game.Layer7PT, SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCreature() && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				c.Power += 2 * effects.S82AfrB08EquipmentAttachedTo(g, target.InstanceID)
			},
		}},
		EquipCostAlternatives: []game.EquipCostAlternative{{
			Label:         "You may pay {0} rather than pay the equip cost of the first equip ability you activate each turn.",
			Cost:          "{0}",
			FirstEachTurn: true,
		}},
	})
}
