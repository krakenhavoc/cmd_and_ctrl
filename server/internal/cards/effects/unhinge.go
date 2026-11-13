package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cb1ae4f2-4fec-4cac-8992-616af82e0ade",
		Name:         "Unhinge",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Discards: 1})},
		Targets:      TargetPlayer("target player"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return targetPlayerDiscards(1)(ctx.Game, ctx.Item)
			}),
			DrawCards{N: 1},
		),
	})
}
