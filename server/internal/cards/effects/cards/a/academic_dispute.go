package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "b22de128-6dad-4e15-a547-1c9ca08df5ff",
		Name:         "Academic Dispute",
		Purpose:      game.Purpose{Learn: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Academic Dispute — blocks this turn if able",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.MayChoice{
					Question: "Academic Dispute — have the creature gain reach until end of turn?",
					Then:     academicDisputeReach,
					Carry:    []uuid.UUID{id},
				}.Apply(ctx)
			}),
			effects.Learn{},
		),
	})
}

// academicDisputeReach is the "yes": the creature gains reach, keyed (#4291).
var academicDisputeReach = effects.MayYes("confirm/academic-dispute-reach", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	return effects.GrantKeywordUntilEOT{
		Target:   effects.CarryAt(r, 0),
		Keywords: []string{"reach"},
		Label:    "Academic Dispute — reach until end of turn",
	}.Apply(ctx)
})
