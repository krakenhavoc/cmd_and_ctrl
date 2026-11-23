package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suppression Bonds — Enchantment — Aura {3}{W} (Magic Origins, #5549):
//
//	"Enchant nonland permanent
//	 Enchanted permanent can't attack or block, and its activated abilities can't be activated."
//
// Petrify's restriction on any nonland permanent: it can't attack or
// block, and none of its activated abilities, mana abilities included,
// can be activated.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e844ded0-821a-4378-8f58-bf1e1c0269d5",
		Name:         "Suppression Bonds",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("enchant nonland permanent", effects.Nonland()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
	})
}
