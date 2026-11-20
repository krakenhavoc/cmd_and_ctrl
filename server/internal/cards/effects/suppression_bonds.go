package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e844ded0-821a-4378-8f58-bf1e1c0269d5",
		Name:         "Suppression Bonds",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("enchant nonland permanent", Nonland()),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			RestrictAttached(game.CantAttackOrBlock | game.CantActivate | game.CantActivateMana),
		},
	})
}
