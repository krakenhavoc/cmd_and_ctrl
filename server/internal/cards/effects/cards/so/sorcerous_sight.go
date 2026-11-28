package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sorcerous Sight — Sorcery {U} (Portal, S82 OLD batch 783, #5630):
//
//	"Look at target opponent's hand.
//	 Draw a card."
//
// The look is private to the caster (CR 701.20e); the draw is the next
// instruction (CR 608.2c). With the target illegal on resolution the spell is
// countered by the rules and nothing happens, the draw included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20370c3b-231f-4d9d-8b6e-f1eb25fa4b5d",
		Name:         "Sorcerous Sight",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.LookAtTargetPlayersHandThenDraw,
	})
}
