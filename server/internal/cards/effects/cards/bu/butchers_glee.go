package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Butcher's Glee — Instant {2}{B} (Dragons of Tarkir, #5551):
//
//	"Target creature gets +3/+0 and gains lifelink until end of turn. Regenerate it. (Damage dealt by a creature with lifelink also causes its controller to gain that much life.)"
//
// Rush of Adrenaline's pump-and-keyword body, then a regeneration shield on
// the same creature (CR 701.19a), which lasts the turn.
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd637245-52bc-4cc9-810d-253e69dc3394",
		Name:         "Butcher's Glee",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 0, "lifelink")),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.BoostUntilEOT{Target: id, Power: 3, Label: "Butcher's Glee — +3/+0"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Butcher's Glee — lifelink"}.Apply(ctx)
			}),
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				return effects.Regenerate{Target: id}.Apply(ctx)
			}),
		),
	})
}
