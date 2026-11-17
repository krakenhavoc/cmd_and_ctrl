package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Accelerate — Instant {1}{R} (Torment, #5612):
//
//	"Target creature gains haste until end of turn.
//	 Draw a card."
//
// Expedite at instant speed: the haste ends at cleanup (CR 514.2) and is
// declared as a benefit, and the draw is the controller's. Written as
// steps (CR 608.2c), so a target that has gone by resolution only skips
// the haste and the card is still drawn (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79189464-9645-4411-8886-68fd40a588ed",
		Name:         "Accelerate",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "haste"))},
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Accelerate — haste until end of turn"}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
