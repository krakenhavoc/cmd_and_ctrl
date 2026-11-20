package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Culling Mark — Sorcery {2}{G} (Born of the Gods,
// #5559):
//
//	"Target creature blocks this turn if able."
//
// Academic Dispute's first sentence (BlockRequirementUntilEOT on the
// target still legal as it resolves).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0cb7f468-070a-463c-b00c-16db3421d16a",
		Name:         "Culling Mark",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			OnTheTargetCard(func(ctx *Context, id uuid.UUID) error {
				return BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Culling Mark — blocks this turn if able",
				}.Apply(ctx)
			}),
		),
	})
}
