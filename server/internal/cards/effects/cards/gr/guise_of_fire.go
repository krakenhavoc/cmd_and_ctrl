package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guise of Fire — {R} Aura: "Enchant creature. Enchanted creature gets
// +1/-1 and attacks each combat if able."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69d4d459-d680-45d9-a3dc-468e4ecce12e",
		Name:         "Guise of Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, -1),
			effects.AttacksEachCombatWhere(effects.AttachedToSource),
		},
	})
}
