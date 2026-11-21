package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wakandan Drone Flock — {3}{W} Artifact Creature — Robot 3/3 (Marvel
// Super Heroes, #4455):
//
//	"Flying
//	 When this creature enters, scry 2."
//
// Faerie Seer's shape: an enters trigger that scries 2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "282d6e61-3240-4f1e-9c2d-8ddb59b03af4",
		Name:            "Wakandan Drone Flock",
		Purpose:         game.Purpose{Scry: 2},
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wakandan Drone Flock — scry 2", effects.Do(effects.Scry{N: 2})),
		},
	})
}
