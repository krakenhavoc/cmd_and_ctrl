package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golden-Scale Aeronaut — Creature — Dwarf Pilot {4}{W}, 2/3 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Flying"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// flying until end of turn; the card has it itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d388080d-5ad4-4e27-b361-340f410423a7",
		Name:            "Golden-Scale Aeronaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Backup("Golden-Scale Aeronaut", 1, "", "flying"),
		},
	})
}
