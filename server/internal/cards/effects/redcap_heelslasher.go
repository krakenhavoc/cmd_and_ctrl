package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Redcap Heelslasher — Creature — Goblin Rogue {3}{R}, 2/3 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 First strike"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// first strike until end of turn; the card has it itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d2134b68-d73f-4090-aaec-b2ee5c9f9f04",
		Name:            "Redcap Heelslasher",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			Backup("Redcap Heelslasher", 1, "", "first strike"),
		},
	})
}
