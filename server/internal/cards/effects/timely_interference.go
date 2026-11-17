package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Timely Interference — Instant {U} (Dominaria United):
//
//	"Kicker {1}{R} (You may pay an additional {1}{R} as you cast this
//	 spell.)
//	 Target creature gets -1/-0 until end of turn. If this spell was
//	 kicked, that creature blocks this turn if able.
//	 Draw a card."
//
// Three instructions in printed order (CR 608.2c). The kicked block
// requirement is Academic Dispute's, a record pinned to the creature for
// the turn. If the target is gone the spell does nothing, the draw
// included (CR 608.2b).
func init() {
	Register(Spec{
		OracleID:      "d64feb23-8259-4445-9b40-efc8b4466679",
		Name:          "Timely Interference",
		Purpose:       game.Purpose{Draws: 1, Targets: game.ForTargets(HarmUntilEOT(0, -1, 0))},
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{1}{R}")},
		Targets:       TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BoostUntilEOT{Target: id, Power: -1, Label: "Timely Interference — -1/-0 until end of turn"}.Apply(ctx)
			}),
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				if !ctx.WasKicked() {
					return nil
				}
				return BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Timely Interference — blocks this turn if able",
				}.Apply(ctx)
			}),
			DrawCards{N: 1},
		),
	})
}
