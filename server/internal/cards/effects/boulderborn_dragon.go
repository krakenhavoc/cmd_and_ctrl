package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boulderborn Dragon — Artifact Creature — Dragon {5}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying, vigilance
//	 Whenever this creature attacks, surveil 1. (Look at the top card of
//	 your library. You may put it into your graveyard.)"
//
// Flying and vigilance ride PrintedKeywords; the attack trigger is the
// footprinted Surveil primitive (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1654cab1-0697-4547-b1fe-2d0a61385199",
		Name:            "Boulderborn Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Boulderborn Dragon — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
