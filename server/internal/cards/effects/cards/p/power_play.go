package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Power Play — Conspiracy (Conspiracy, #5557):
//
//	"(Start the game with this conspiracy face up in the command zone.)
//	 You are the starting player. If multiple players would be the starting
//	 player, one of those players is chosen at random."
//
// A conspiracy without hidden agenda starts face up (CR 315.2). The engine
// reads it as the deal begins, once every conspiracy is in the command zone
// (game.conspiracyStartingSeatLocked): its owner takes the first turn whoever
// the opening roll chose, and when several players have a Power Play one of
// them is picked at random.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "15710c57-369a-442b-b7ea-175f3d9d164d",
		Name:           "Power Play",
		Completeness:   effects.CompletenessFull,
		StartingPlayer: true,
	})
}
