package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Demotion — Enchantment — Aura {W} (Guilds of Ravnica, #5525):
//
//	"Enchant creature
//	 Enchanted creature can't block, and its activated abilities can't be
//	 activated."
//
// Arrest's restriction without the attack bar: the creature can't block and no
// activated ability of it can be activated, mana abilities included (the card
// has no "unless they're mana abilities" clause, CR 602.5), loyalty abilities
// being activated abilities too (CR 606.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c333eb79-300e-4dc0-9126-8dca1bac6ef2",
		Name:         "Demotion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttached(game.CantBlock | game.CantActivate | game.CantActivateMana),
		},
	})
}
