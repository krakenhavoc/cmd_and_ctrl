package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "0cb7f468-070a-463c-b00c-16db3421d16a",
		Name:         "Culling Mark",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BlockRequirementUntilEOT{
					Target: id,
					Kind:   game.BlockRequirementBlocks,
					Label:  "Culling Mark — blocks this turn if able",
				}.Apply(ctx)
			}),
		),
	})
}
