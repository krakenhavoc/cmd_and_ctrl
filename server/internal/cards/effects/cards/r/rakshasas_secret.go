package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rakshasa's Secret — Sorcery {2}{B} (Khans of Tarkir):
//
//	"Target opponent discards two cards. You mill two cards."
//
// Heartless Pillage's discard, chosen by the opponent (CR 701.9b), then the
// controller's mill (CR 701.17a), which waits for the discard prompt
// (CR 608.2c). An illegal target stops the whole spell (CR 608.2b), so the
// mill does not happen either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e753f50d-64ed-4822-a301-57659ee378b6",
		Name:         "Rakshasa's Secret",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, Discards: 2}),
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.OldB350TargetPlayerDiscards(2)),
			effects.Instruction(func(ctx *effects.Context) error {
				if effects.TargetedPlayer(ctx) == uuid.Nil {
					return nil
				}
				return effects.MillCards{N: 2}.Apply(ctx)
			}),
		),
	})
}
