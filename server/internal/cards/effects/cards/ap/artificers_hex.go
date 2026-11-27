package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Artificer's Hex — Enchantment — Aura {B} (Magic 2014, S82 old sets):
//
//	"Enchant Equipment
//	 At the beginning of your upkeep, if enchanted Equipment is attached to a
//	 creature, destroy that creature."
//
// "Enchant Equipment" is the enchant-permanent spec narrowed to the Equipment
// subtype. The upkeep trigger is the Aura controller's own (CR 503.1), with an
// intervening "if" (CR 603.4) read at both ends: the enchanted Equipment must
// be attached to a creature as the upkeep begins and again as the ability
// resolves, and "that creature" is the one it is attached to then.
//
// No simplification.
func init() {
	const label = "Artificer's Hex — destroy the creature the enchanted Equipment is attached to"
	effects.Register(effects.Spec{
		OracleID:     "d14d9f44-9cff-4a11-be86-0cea5fe9b551",
		Name:         "Artificer's Hex",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.HasSubtype("Equipment")),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "the enchanted Equipment is attached to a creature", Test: effects.S82OldB399EquipmentIsOnACreature},
				effects.AtYourUpkeep(label, effects.S82OldB399DestroyEnchantedEquipmentsCreature)),
		},
	})
}
