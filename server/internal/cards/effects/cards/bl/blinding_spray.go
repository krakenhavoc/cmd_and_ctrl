package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blinding Spray — Instant {4}{U} (Khans of Tarkir):
//
//	"Creatures your opponents control get -4/-0 until end of turn.
//	 Draw a card."
//
// The creatures are the ones opponents control as it resolves (CR 611.2c);
// a creature that arrives later is not affected.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88ce2017-83a6-4b55-afba-371875bf2172",
		Name:         "Blinding Spray",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: effects.ResolveSteps(
			effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.OpponentControls()), Power: -4,
				Label: "Blinding Spray — creatures your opponents control get -4/-0",
			},
			effects.DrawCards{N: 1},
		),
	})
}
