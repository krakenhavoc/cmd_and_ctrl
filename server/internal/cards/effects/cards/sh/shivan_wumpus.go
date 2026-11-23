package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "4b78222f-6973-42a3-987e-eb8cc8fd6c76",
		Name:            "Shivan Wumpus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.AnyPlayerMaySacrificeALandToTuckThis("Shivan Wumpus"),
		},
	})
}
