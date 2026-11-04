package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonshaker Cavalry — Creature — Spirit Knight {5}{W}{W}{W}, 6/6
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying
//	 When this creature enters, creatures you control gain flying and
//	 get +X/+X until end of turn, where X is the number of creatures you
//	 control."
//
// Craterhoof Behemoth's trigger with flying for trample
// (socB9TeamKeywordAndPump). X is counted as the trigger resolves, the
// Cavalry included, and neither X nor the affected set changes later in
// the turn (the rulings, CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b3f16f34-d78c-4b61-821b-05fc6bd0da37",
		Name:            "Moonshaker Cavalry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Moonshaker Cavalry — flying and +X/+X", socB9TeamKeywordAndPump("Moonshaker Cavalry", "flying")),
		},
	})
}
