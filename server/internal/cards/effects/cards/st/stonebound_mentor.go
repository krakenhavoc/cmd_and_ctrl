package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonebound Mentor — Creature — Spirit Advisor {1}{R}{W}, 3/3 (Strixhaven:
// School of Mages, #5502):
//
//	"Whenever one or more cards leave your graveyard, scry 1."
//
// Spirit Mascot's trigger (left_graveyard.go): every route a card leaves
// your graveyard by counts, and cards leaving together are one batch
// (OncePerBatch, CR 603.2c). A token ceasing to exist is not a card leaving.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c4b4dd4-4a80-436a-8526-5c58c1751ccd",
		Name:         "Stonebound Mentor",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, effects.WhenCardsLeaveYourGraveyard,
				"Stonebound Mentor — scry 1",
				effects.Do(effects.Scry{N: 1}))),
		},
	})
}
