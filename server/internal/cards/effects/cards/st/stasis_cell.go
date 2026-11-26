package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stasis Cell — Enchantment — Aura {4}{U} (Conspiracy, S82 OLD batch 362,
// #5557):
//
//	"Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 {3}{U}: Attach this Aura to target creature."
//
// Bitter Chill's untap lock, and Bound by Moonsilver's Umbra-style move
// (AttachSourceToTarget) with no timing restriction: the ability may be
// activated at instant speed by the Aura's controller. The target must be a
// creature the Aura can enchant (CR 303.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "8b2ba6f8-1060-4167-9941-3ab7bd4e0ca8",
		Name:                  "Stasis Cell",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}: Attach this Aura to target creature.",
			Cost:    effects.ManaCost("{3}{U}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.AttachSourceToTarget,
		}},
	})
}
