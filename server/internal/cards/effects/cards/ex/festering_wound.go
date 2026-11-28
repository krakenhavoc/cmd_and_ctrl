package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Festering Wound — {1}{B} Enchantment — Aura (Urza's Destiny, S82 OLD batch 703, #5621):
//
//	"Enchant creature
//	 At the beginning of your upkeep, you may put an infection counter on this
//	 Aura.
//	 At the beginning of the upkeep of enchanted creature's controller, this
//	 Aura deals X damage to that player, where X is the number of infection
//	 counters on this Aura."
//
// Two triggers: yours (the Aura's controller) grows it, and the enchanted
// creature's controller's upkeep (Stab Wound's trigger) is dealt X damage by
// the Aura (S82OldB703FesteringWoundDamage).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "32638d03-c101-42f0-8a4b-8c6290257a28",
		Name:         "Festering Wound",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			effects.S82OldB703MayPutCounterOnThisAtUpkeep("Festering Wound", "infection"),
			effects.On(game.EventBeginUpkeep, effects.UpkeepOfEnchantedCreaturesController,
				"Festering Wound — it deals X damage to that player, where X is the number of infection counters on it",
				effects.S82OldB703FesteringWoundDamage),
		},
	})
}
