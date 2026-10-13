package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shivan Wumpus — Creature — Beast {3}{R}, 6/6:
//
//	"Trample
//	 When this creature enters, any player may sacrifice a land of their
//	 choice. If a player does, put this creature on top of its owner's
//	 library."
//
// ADR 0150 (#1903): Argothian Wurm's trigger in red
// (anyPlayerMaySacrificeALandToTuckThis).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b78222f-6973-42a3-987e-eb8cc8fd6c76",
		Name:            "Shivan Wumpus",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			anyPlayerMaySacrificeALandToTuckThis("Shivan Wumpus"),
		},
	})
}
