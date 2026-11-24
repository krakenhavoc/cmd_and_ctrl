package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time Walk — Sorcery {1}{U}: "Take an extra turn after this one."
// (Alpha, #5948). One of the Power Nine Oracle of the Alpha conjures
// (ADR 0099 amendment 2026-11-24).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d0209d3f-3f7e-4fd5-bce5-10bce6f29c86",
		Name:         "Time Walk",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{ExtraTurns: 1},
		OnResolve:    effects.YouTakeAnExtraTurn,
	})
}
