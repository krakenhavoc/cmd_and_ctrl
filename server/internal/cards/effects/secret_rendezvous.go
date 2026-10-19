package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secret Rendezvous — Sorcery {1}{W}{W} (EDHREC rank 1130):
//
//	"You and target opponent each draw three cards."
//
// White's group-hug draw three. The caster draws first, then the
// targeted opponent — the printed order — and a target that became
// illegal in response fizzles the whole spell, which the engine's
// CR 608.2b all-targets-illegal check does before OnResolve runs.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34b90d3e-f48c-41ff-b3e4-5cab7a9cd597",
		Name:         "Secret Rendezvous",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target opponent", Opponent()),
		// You draw three, and so does the opponent you aim it at.
		Purpose: game.Purpose{Draws: 3, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, Draws: 3})},
		OnResolve: ResolveSteps(
			// The target's draw waits for a prompt the caster's draw asks (CR 608.2c, #4155).
			DrawCards{N: 3},
			Instruction(func(ctx *Context) error {
				if len(ctx.Item.Targets) == 0 || ctx.Item.Targets[0].Kind != game.TargetPlayer {
					return nil
				}
				return DrawCards{Player: ctx.Item.Targets[0].ID, N: 3}.Apply(ctx)
			}),
		),
	})
}
