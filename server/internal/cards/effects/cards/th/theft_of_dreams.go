package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Theft of Dreams — Sorcery {2}{U} (Portal Second Age, S82 older sets batch 736, #5625):
//
//	"Draw a card for each tapped creature target opponent controls."
//
// Counted as the spell resolves, from the opponent's creatures through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "008011e2-7b82-4962-af6e-be627112f37f",
		Name:         "Theft of Dreams",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve:    effects.S82OldB736DrawPerTappedCreature,
	})
}
