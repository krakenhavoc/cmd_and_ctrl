package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Appendage Amalgam — Enchantment Creature — Horror {2}{B}, 3/2
// (Duskmourn: House of Horror, #4466):
//
//	"Flash
//	 Whenever this creature attacks, surveil 1. (Look at the top card of
//	 your library. You may put it into your graveyard.)"
//
// Flash is the printed keyword; the attack trigger is the shared Surveil
// primitive for its controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9fb95813-05e5-4aa0-bab9-53d8a071fe6a",
		Name:            "Appendage Amalgam",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Appendage Amalgam — surveil 1", effects.Do(effects.Surveil{N: 1})),
		},
	})
}
