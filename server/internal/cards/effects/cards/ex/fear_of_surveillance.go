package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fear of Surveillance — Enchantment Creature — Nightmare {1}{W}, 2/2
// (Duskmourn: House of Horror, #4466):
//
//	"Vigilance
//	 Whenever this creature attacks, surveil 1. (Look at the top card of
//	 your library. You may put it into your graveyard.)"
//
// Appendage Amalgam's attack trigger (CR 701.25a) on a white body with
// vigilance in place of flash.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "52478437-bb6c-48bb-855d-54a8a4d42d18",
		Name:            "Fear of Surveillance",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Fear of Surveillance — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
