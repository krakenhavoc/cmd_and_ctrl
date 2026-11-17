package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reality Anchor — Instant {1}{G} (Tempest, #5628):
//
//	"Target creature loses shadow until end of turn.
//	 Draw a card."
//
// Canopy Claws' "loses flying until end of turn" with shadow, then the
// draw (CR 608.2c). The loss is a layer-6 ability-removing effect with
// its own timestamp (CR 613.1f): a shadow grant with a later timestamp
// (a Soltari Emissary activated afterwards) puts it back, and the
// printed keyword or an earlier grant does not. It ends in the cleanup
// step (CR 514.2). A target gone in response means the spell doesn't
// resolve and nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a7b4a519-d85f-40ae-bb5f-08bc1d301418",
		Name:         "Reality Anchor",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(HarmUntilEOT(0, 0, 0))},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return untilEndOfTurn(ctx, id, nil, "Reality Anchor — loses shadow until end of turn",
					game.RemoveKeywordsMod("shadow"))
			}),
			DrawCards{N: 1},
		),
	})
}
