package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furious Strength — Enchantment — Aura {2}{R} (Marvel Super Heroes
// Commander, #4455): "Enchant creature. Enchanted creature gets +2/+2 and
// has menace."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efefb339-bb7c-4b26-abed-8434f955122b",
		Name:         "Furious Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2), effects.GrantToAttached("menace")},
	})
}
