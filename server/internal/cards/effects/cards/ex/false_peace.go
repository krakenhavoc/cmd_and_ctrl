package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// False Peace — {W} Sorcery (Starter 1999, S82 OLD batch 701, #5620):
//
//	"Target player skips all combat phases of their next turn."
//
// Empty City Ruse's primitive without the opponent restriction
// (TargetPlayerSkipsCombatPhasesOfNextTurn; CR 500.11, 614.10).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7962db58-dbd9-4b94-8a21-a1625da4c384",
		Name:         "False Peace",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerSkipsCombatPhasesOfNextTurn{}.Apply(ctx)
		},
	})
}
