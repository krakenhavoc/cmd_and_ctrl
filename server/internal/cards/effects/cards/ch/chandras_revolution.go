package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chandra's Revolution — Sorcery {3}{R} (Aether Revolt, S82 older sets, #5536):
//
//	"Chandra's Revolution deals 4 damage to target creature. Tap target
//	 land. That land doesn't untap during its controller's next untap step."
//
// Two target clauses in printed order. Each part is done to its own target
// if that target is still legal (CR 608.2b), so a creature that has gone
// does not stop the land being tapped and frozen.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f72d2074-2f32-4b56-a172-947bf6104fae",
		Name:         "Chandra's Revolution",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPermanent("target land", effects.Land()),
		),
		Purpose: effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(0)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 4}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				t, ok := ctx.ClauseTarget(1)
				if !ok || t.Kind != game.TargetCard {
					return nil
				}
				return effects.TapAndFreeze{Targets: []uuid.UUID{t.ID}, Label: "Chandra's Revolution — doesn't untap during its controller's next untap step"}.Apply(ctx)
			}),
		),
	})
}
