package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fearless Skald — Creature — Dwarf Berserker {4}{R}, 3/2 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Double strike"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// double strike until end of turn; the card has it itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f055a941-8b83-464e-a57f-8f188dc69aaa",
		Name:            "Fearless Skald",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Triggered: []game.TriggeredAbility{
			Backup("Fearless Skald", 1, "", "double strike"),
		},
	})
}
