package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightveil Sprite — Creature — Faerie Rogue {1}{U}, 1/2:
//
//	"Flying
//	 Whenever this creature attacks, surveil 1."
//
// Flying rides PrintedKeywords; the attack trigger is one surveil.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8ac23863-0aa7-41bc-b9b7-c5f0566172cf",
		Name:            "Nightveil Sprite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Nightveil Sprite — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
