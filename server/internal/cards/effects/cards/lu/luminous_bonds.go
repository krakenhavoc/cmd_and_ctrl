package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luminous Bonds — Enchantment — Aura {2}{W} (Rivals of Ixalan, #5531):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block."
//
// Pacifism's text: the restriction is the Aura's effect on its host, not
// an ability the creature has.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2c006a63-0550-4a87-a5fe-f4247857271c",
		Name:         "Luminous Bonds",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
	})
}
