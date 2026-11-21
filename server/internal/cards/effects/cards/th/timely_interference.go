package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:      "d64feb23-8259-4445-9b40-efc8b4466679",
		Name:          "Timely Interference",
		Purpose:       game.Purpose{Draws: 1, Targets: game.ForTargets(effects.HarmUntilEOT(0, -1, 0))},
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{R}")},
		Targets:       effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: -1, Label: "Timely Interference — -1/-0 until end of turn"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if !ctx.WasKicked() {
					return nil
				}
				return effects.BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Timely Interference — blocks this turn if able",
				}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
