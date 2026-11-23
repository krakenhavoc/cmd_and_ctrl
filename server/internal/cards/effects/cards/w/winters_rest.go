package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Winter's Rest — Snow Enchantment — Aura {1}{U} (Jumpstart 2022, #5948):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 As long as you control another snow permanent, enchanted creature doesn't
//	 untap during its controller's untap step."
//
// Castaway's Despair with the untap hold gated on the Aura's controller
// controlling another snow permanent, read through the layers at each untap
// step.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9df37d14-89fa-4a28-98ee-426062b5740d",
		Name:         "Winter's Rest",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{{
			Label:     "enchanted creature doesn't untap while you control another snow permanent",
			Restricts: effects.GapB169EnchantedHeldWhileOtherSnow,
		}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Winter's Rest — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
	})
}
