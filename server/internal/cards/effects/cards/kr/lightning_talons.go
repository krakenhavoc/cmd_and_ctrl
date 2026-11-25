package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Talons — Enchantment — Aura {2}{R} (Battlebond):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+0 and has first strike."
//
// Angelic Destiny's layered pump and keyword on the enchanted creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a2a7eca-3477-4858-9dab-ba80d4d704cd",
		Name:         "Lightning Talons",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 0),
			effects.GrantToAttached("first strike"),
		},
	})
}
