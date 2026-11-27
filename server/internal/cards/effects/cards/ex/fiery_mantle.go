package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fiery Mantle — Enchantment — Aura {1}{R} (Urza's Saga, #5624):
//
//	"Enchant creature
//	 {R}: Enchanted creature gets +1/+0 until end of turn.
//	 When this Aura is put into a graveyard from the battlefield, return it to
//	 its owner's hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b553ed4-12eb-4a53-8e1f-d47ec92b1d65",
		Name:         "Fiery Mantle",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Enchanted creature gets +1/+0 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{R}"),
			Effect:  effects.S82OldB399PumpEnchanted("Fiery Mantle", 1, 0),
		}},
		Triggered: []game.TriggeredAbility{effects.UlgB1AuraReturnsToHand("Fiery Mantle")},
	})
}
