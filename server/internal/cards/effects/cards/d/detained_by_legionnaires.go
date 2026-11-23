package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Detained by Legionnaires — Enchantment — Aura {1}{W} (Assassin's
// Creed, #5466):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block."
//
// Pacifism under another name: the restriction is the Aura's, read off
// the attachment on every recompute (pacifism.go has the argument).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d59488b5-1118-4276-b862-11e860c7c32a",
		Name:         "Detained by Legionnaires",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock),
		},
	})
}
