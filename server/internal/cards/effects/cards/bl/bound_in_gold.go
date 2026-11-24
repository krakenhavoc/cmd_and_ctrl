package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bound in Gold — Enchantment — Aura {2}{W} (Kaldheim, #5505):
//
//	"Enchant permanent
//	 Enchanted permanent can't attack, block, or crew Vehicles, and its
//	 activated abilities can't be activated unless they're mana
//	 abilities."
//
// Intercessor's Arrest's lock, word for word (CantCrew, #5673).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ca597a0d-b510-4b25-9a83-4d4e613546f5",
		Name:         "Bound in Gold",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttachedExtra(game.CantAttackOrBlock|game.CantActivate, game.CantCrew),
		},
	})
}
