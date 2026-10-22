package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Academic Dispute — Instant {R} (EDHREC rank 17933):
//
//	"Target creature blocks this turn if able. You may have it gain
//	 reach until end of turn.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// Three instructions in printed order, and each waits for the one
// before it (CR 608.2c): the block requirement for the turn
// (BlockRequirementUntilEOT, the Peema Aether-Seer shape), the "you
// may" about reach (asked of the spell's controller as it resolves, a
// MayChoice), and learn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b22de128-6dad-4e15-a547-1c9ca08df5ff",
		Name:         "Academic Dispute",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Academic Dispute — blocks this turn if able",
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return MayChoice{
					Question: "Academic Dispute — have the creature gain reach until end of turn?",
					OnYes: func(ctx *Context) error {
						return GrantKeywordUntilEOT{
							Target:   id,
							Keywords: []string{"reach"},
							Label:    "Academic Dispute — reach until end of turn",
						}.Apply(ctx)
					},
				}.Apply(ctx)
			}),
			Learn{},
		),
	})
}
