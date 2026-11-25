package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightning Diadem — Enchantment — Aura {5}{R} (Jumpstart, #5510):
//
//	"Enchant creature
//	 When this Aura enters, it deals 2 damage to any target.
//	 Enchanted creature gets +2/+2."
//
// The pump is a layer 7c +2/+2. The enters trigger's target is chosen as it
// goes on the stack (CR 603.3d) and may be the enchanted creature; the Aura
// deals the damage as an object, so it still does if it is gone by then. A
// target no longer legal makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Lightning Diadem — it deals 2 damage to any target"
	effects.Register(effects.Spec{
		OracleID:     "1e4de77b-27cb-4afd-8d4b-1cc89f2a6986",
		Name:         "Lightning Diadem",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(2, 2)},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters(label, effects.DealDamageToTheTarget(2)),
				effects.TargetAny()),
				effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
