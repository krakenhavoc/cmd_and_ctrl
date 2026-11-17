package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Saiba Cryptomancer — Creature — Moonfolk Ninja {1}{U}, 0/1 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Hexproof"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// hexproof until end of turn; the card has it itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "15f89b06-5025-4c9f-bf51-e25bc5c2dd6b",
		Name:            "Saiba Cryptomancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "hexproof"},
		Triggered: []game.TriggeredAbility{
			Backup("Saiba Cryptomancer", 1, "", "hexproof"),
		},
	})
}
