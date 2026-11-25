package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Revoke Privileges — Enchantment — Aura {2}{W} (Kaladesh, #5539):
//
//	"Enchant creature
//	 Enchanted creature can't attack, block, or crew Vehicles."
//
// Bound in Gold's lock without the activated-ability clause: a layer 6
// restriction on the enchanted creature (CantAttackOrBlock and CantCrew, #5673).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60b2f991-47ef-4da3-ba61-0692695ea098",
		Name:         "Revoke Privileges",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static: []game.StaticAbility{
			effects.RestrictAttachedExtra(game.CantAttackOrBlock, game.CantCrew),
		},
	})
}
