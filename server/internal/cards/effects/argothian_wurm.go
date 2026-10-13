package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argothian Wurm — Creature — Wurm {3}{G}, 6/6:
//
//	"Trample
//	 When this creature enters, any player may sacrifice a land of their
//	 choice. If a player does, put this creature on top of its owner's
//	 library."
//
// ADR 0150 (#1903): the sacrifice is offered to every player in APNAP
// order from the active player, the Wurm's controller included, until
// one makes it (anyPlayerMaySacrificeALandToTuckThis). A player with no
// land can't pay (CR 118.3).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8aa1893-60e5-439e-9316-872a9f6b6c94",
		Name:            "Argothian Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			anyPlayerMaySacrificeALandToTuckThis("Argothian Wurm"),
		},
	})
}
