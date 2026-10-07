package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lathnu Hellion — Creature — Hellion {2}{R}, 4/4:
//
//	"Haste
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 At the beginning of your end step, sacrifice this creature unless you pay {E}{E}."
//
// ADR 0129 §3 (#1995): "sacrifice it unless you pay {E}{E}" is CR
// 118.12a's pay-unless, with an energy payment; declining, or being short
// of the energy (CR 118.3), sacrifices the Hellion. The prompt holds the
// end step until it is answered.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6dd71453-63ec-4f4c-87d2-34d207398b9a",
		Name:            "Lathnu Hellion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Lathnu Hellion", 2),
			AtYourEndStep("Lathnu Hellion — sacrifice it unless you pay {E}{E}",
				sacrificeThisUnlessYouPayEnergy("Lathnu Hellion", "Lathnu Hellion", 2)),
		},
	})
}
