package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Minion's Return — Enchantment — Aura {2}{B} (Theros Beyond Death, #5514):
//
//	"Flash
//	 Enchant creature
//	 When enchanted creature dies, return that card to the
//	 battlefield under your control."
//
// Flash, and Kaya's Ghostform's return: when the enchanted creature
// dies, that card comes back under the Aura's controller, a new object
// (CR 400.7), wherever it came from. A card that has left the graveyard
// by then stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6cbe76c6-2b64-4389-b5f9-0b221cd92312",
		Name:            "Minion's Return",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         effects.EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Minion's Return — return that card to the battlefield under your control",
				effects.Cmdr2B09ReturnTheEventCardUnderYourControl),
		},
	})
}
