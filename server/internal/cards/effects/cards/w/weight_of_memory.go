package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Weight of Memory — Sorcery {3}{U}{U} (Dominaria, S82 OLD, #5530):
//
//	"Draw three cards. Target player mills three cards."
//
// The caster draws first, then the target (a player, possibly the caster)
// mills three. A target that has become illegal fizzles the whole spell
// (CR 608.2b), so no cards are drawn then either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "826c5feb-9e82-49e0-bc6f-6c29004f8151",
		Name:         "Weight of Memory",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			effects.DrawCards{N: 3},
			effects.Instruction(func(ctx *effects.Context) error {
				player := effects.TargetedPlayer(ctx)
				if player == [16]byte{} {
					return nil
				}
				return effects.MillCards{Player: player, N: 3}.Apply(ctx)
			}),
		),
	})
}
