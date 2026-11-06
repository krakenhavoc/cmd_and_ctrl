package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "52478437-bb6c-48bb-855d-54a8a4d42d18",
		Name:            "Fear of Surveillance",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Fear of Surveillance — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
