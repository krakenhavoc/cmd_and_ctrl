package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wibbly-wobbly, Timey-wimey — Sorcery {1}{U} (Doctor Who):
//
//	"Time travel. (For each suspended card you own and each permanent you
//	 control with a time counter on it, you may add or remove a time counter.)
//	 Draw a card."
//
// Time travel (CR 701.56a) first, then the draw; the draw waits for any
// question the time travel asks.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5207621d-aecf-40ae-8108-2bd571e3a427",
		Name:         "Wibbly-wobbly, Timey-wimey",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve:    effects.ResolveSteps(effects.TimeTravel{}, effects.DrawCards{N: 1}),
	})
}
