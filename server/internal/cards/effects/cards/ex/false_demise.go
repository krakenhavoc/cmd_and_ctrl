package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// False Demise — Enchantment — Aura {2}{U} (Mercadian Masques, S82 OLD batch 685, #5619):
//
//	"Enchant creature
//	 When enchanted creature dies, return that card to the battlefield under your control."
//
// Minion's Return without flash: when the enchanted creature dies it comes back under the Aura's controller, a new object (CR 400.7). A card that has left the graveyard by then stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50ffa9f0-fb3b-4de4-b463-6d7ee8b5e93a",
		Name:         "False Demise",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("False Demise — return that card to the battlefield under your control",
				effects.Cmdr2B09ReturnTheEventCardUnderYourControl),
		},
	})
}
