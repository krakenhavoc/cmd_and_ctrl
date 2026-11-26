package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Provoke — Instant {1}{G} (Conspiracy, S82 OLD batch 364, #5557):
//
//	"Untap target creature you don't control. That creature blocks this turn
//	 if able.
//	 Draw a card."
//
// Three instructions in printed order (CR 608.2c): the untap, Academic
// Dispute's block requirement for the turn (BlockRequirementUntilEOT), then
// the draw. A target that has gone by resolution skips the first two and the
// draw still happens; a spell whose only target is gone is countered whole.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "731821b1-89fd-4166-94b8-35a1364e4d9d",
		Name:         "Provoke",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("target creature you don't control", effects.Not(effects.YouControl())),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.UntapTarget{Target: id}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Provoke — blocks this turn if able",
				}.Apply(ctx)
			}),
			effects.DrawCards{N: 1},
		),
	})
}
