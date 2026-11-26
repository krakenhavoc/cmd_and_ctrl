package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viper's Kiss — Enchantment — Aura {B} (Theros, S82 OLD b390, #5561):
//
//	"Enchant creature
//	 Enchanted creature gets -1/-1, and its activated abilities can't be
//	 activated."
//
// Crippling Blight's shrink at -1/-1 with Arrest's restriction on the
// host: mana abilities are activated abilities too, so both bits are set
// (CantActivate, CantActivateMana).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35dead9a-20ab-45e9-be3e-36857c928ab9",
		Name:         "Viper's Kiss",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.PumpAttached(-1, -1),
			effects.RestrictAttached(game.CantActivate | game.CantActivateMana),
		},
	})
}
