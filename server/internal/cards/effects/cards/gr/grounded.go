package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grounded — {1}{G} Aura: "Enchant creature. Enchanted creature loses
// flying."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cca3557c-c911-494e-95e7-cebec3442d63",
		Name:         "Grounded",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RemoveFromAttached("flying")},
	})
}
