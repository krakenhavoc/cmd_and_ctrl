package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Infernal Contract — Sorcery {B}{B}{B} (Mirage, S82 OLD b800, #5633):
//
//	"Draw four cards. You lose half your life, rounded up."
//
// Death Wish's life loss: half of the life total as the spell resolves,
// rounded up (CR 107.1a), after the four draws, lost rather than dealt (CR
// 119.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4372205f-45e6-49c7-b1fd-13a41e20d4de",
		Name:         "Infernal Contract",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 4},
		OnResolve:    effects.OldB800InfernalContract,
	})
}
