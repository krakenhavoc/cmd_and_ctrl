package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viashino Cutthroat — Creature — Lizard {2}{R}{R}, 5/3 (Urza's Legacy,
// #5623):
//
//	"Haste
//	 At the beginning of the end step, return this creature to its
//	 owner's hand."
//
// Viashino Sandstalker's shape through the shared ulgB1ReturnAtEachEndStep
// trigger: "the end step" is any player's, so one put onto the battlefield
// on another turn goes back at that turn's end, and only while it is still
// the same object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6dc1305a-d6ca-4280-ba5d-5ed91f78ced0",
		Name:            "Viashino Cutthroat",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.UlgB1ReturnAtEachEndStep("Viashino Cutthroat"),
		},
	})
}
