package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Defang — {1}{W} Aura: "Enchant creature. Prevent all damage that would be
// dealt by enchanted creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a6c81b8-71f0-468b-85ad-d87e9a712ecf",
		Name:         "Defang",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtByAttached("Defang: prevent all damage that would be dealt by enchanted creature."),
		},
	})
}
