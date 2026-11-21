package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caught in the Brights — Enchantment — Aura {2}{W} (Jumpstart 2022, #5948):
//
//	"Enchant creature
//	 Enchanted creature can't attack or block.
//	 When a Vehicle you control attacks, exile enchanted creature."
//
// Pacifism's restriction, plus a trigger on each Vehicle the Aura's controller
// attacks with (Prodigy's Prototype's condition) that exiles whatever the Aura
// enchants as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fa131186-d730-460b-b9e5-b97d89857e24",
		Name:         "Caught in the Brights",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Static:       []game.StaticAbility{effects.RestrictAttached(game.CantAttackOrBlock)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.VehicleYouControlAttacked,
				"Caught in the Brights — exile enchanted creature", effects.R41ExileEnchanted),
		},
	})
}
