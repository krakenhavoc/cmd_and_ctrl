package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unhinge — Sorcery {2}{B} (Torment, #5612):
//
//	"Target player discards a card.
//	 Draw a card."
//
// The discarding player picks the card (CR 701.9a): Skull Fracture's
// targetPlayerDiscards body, which does nothing when the player is no
// longer a legal target. The draw is the caster's and is a second step,
// so it waits for the discard's answer (CR 608.2c): a caster who
// targets themselves discards before drawing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb1ae4f2-4fec-4cac-8992-616af82e0ade",
		Name:         "Unhinge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1})},
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.TargetPlayerDiscards(1)(ctx.Game, ctx.Item)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
