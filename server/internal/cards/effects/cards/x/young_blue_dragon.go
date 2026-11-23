package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Young Blue Dragon // Sand Augury — an Adventure creature (Commander Legends:
// Battle for Baldur's Gate, #5948):
//
//	Young Blue Dragon {4}{U}, Creature — Dragon, 3/3
//	  "Flying"
//	Sand Augury {1}{U}, Sorcery — Adventure
//	  "Scry 1, then draw a card."
//
// Joint Exploration's ordered steps: the scry prompt is answered before the
// draw, so the card left on top is the card drawn.
//
// No simplification.
func init() {
	const oracleID = "fff28b30-9903-4a0a-a723-31216abd457d"
	effects.Register(effects.Spec{
		OracleID:        oracleID,
		Name:            "Young Blue Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
	effects.Register(effects.Spec{
		OracleID:     oracleID + "#1",
		Name:         "Sand Augury",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1, Draws: 1},
		OnResolve:    effects.ResolveSteps(effects.Scry{N: 1}, effects.DrawCards{N: 1}),
	})
}
