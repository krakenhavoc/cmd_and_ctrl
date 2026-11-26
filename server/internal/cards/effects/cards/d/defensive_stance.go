package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defensive Stance — Enchantment — Aura {U}:
//
//	"Enchant creature
//	 Enchanted creature gets -1/+1."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "19959b6d-eb8f-4273-94c8-65210e1998ab",
		Name:         "Defensive Stance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(-1, 1)},
	})
}
