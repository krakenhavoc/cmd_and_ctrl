package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aerial Modification — Enchantment — Aura {4}{W} (Jumpstart 2022, #5948):
//
//	"Enchant creature or Vehicle
//	 As long as enchanted permanent is a Vehicle, it's a creature in addition
//	 to its other types.
//	 Enchanted creature gets +2/+2 and has flying."
//
// Aether Meltdown's enchant clause (an uncrewed Vehicle is a legal host and
// stays one), a layer-4 type addition for a Vehicle host, and the +2/+2 and
// flying for the enchanted creature (layers 7c and 6) once it is one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "290cd0b1-55d6-4b30-962d-075a6f75d562",
		Name:         "Aerial Modification",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.GapB169VehicleIsACreature(),
			effects.PumpEnchantedCreature(2, 2),
			effects.GrantToAttached("flying"),
		},
	})
}
