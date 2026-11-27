package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grave Consequences — Instant {1}{B} (Judgment, #5611):
//
//	"Each player may exile any number of cards from their graveyard. Then
//	 each player loses 1 life for each card in their graveyard.
//	 Draw a card."
//
// One choose-cards run: each player is asked about their own graveyard in
// APNAP order and the picks are exiled together; then each player loses 1
// life per card left and the controller draws.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "435c8d98-9fa3-40d8-9b7c-32474099e08e",
		Name:         "Grave Consequences",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.OldB621GraveConsequences),
	})
}
