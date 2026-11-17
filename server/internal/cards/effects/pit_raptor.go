package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pit Raptor — Creature — Bird Mercenary {2}{B}{B}, 4/3 (Prophecy,
// #5617):
//
//	"Flying, first strike
//	 At the beginning of your upkeep, sacrifice this creature unless you
//	 pay {2}{B}{B}."
//
// Two printed keywords, and Whipstitched Zombie's upkeep
// (UpkeepSacrificeUnlessYouPay) at {2}{B}{B}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "92cdcafc-ffeb-436d-b2bd-d6ad895d0e08",
		Name:            "Pit Raptor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike"},
		Triggered: []game.TriggeredAbility{
			UpkeepSacrificeUnlessYouPay("Pit Raptor", "{2}{B}{B}"),
		},
	})
}
