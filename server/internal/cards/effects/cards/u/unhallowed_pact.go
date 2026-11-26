package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unhallowed Pact — Enchantment — Aura {2}{B} (Conspiracy, S82 OLD batch
// 362, #5557):
//
//	"Enchant creature
//	 When enchanted creature dies, return that card to the battlefield under
//	 your control."
//
// The dies trigger is harvested while the Aura is still attached and returns
// the card as a new object (CR 400.7) under the Aura controller's control
// (Kaya's Ghostform's return); a card that has left the graveyard by then
// stays where it is, and a token returns nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da1dd850-9118-42b9-8846-e6bd5042775c",
		Name:         "Unhallowed Pact",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Unhallowed Pact — return that card to the battlefield under your control",
				effects.Cmdr2B09ReturnTheEventCardUnderYourControl),
		},
	})
}
