package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cagemail — Enchantment — Aura {1}{W} (Judgment):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+2 and can't attack."
//
// Weight of Conscience's attached can't-attack bit with a layer-7c pump.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fe1f40fe-011b-4686-b865-fed00c56d30a",
		Name:         "Cagemail",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.RestrictAttached(game.CantAttack),
		},
	})
}
