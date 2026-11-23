package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aspirant's Ascent — Instant {U} (Phyrexia: All Will Be One, #5479):
//
//	"Until end of turn, target creature gets +1/+3 and gains flying and
//	 toxic 1."
//
// Toxic is cumulative (CR 702.164b): a creature that already has toxic
// 2 has a total of 3 under it, because the grant appends through
// GrantKeywordUntilEOT.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed9c488a-9d44-4d2d-bcf8-81352193b8c2",
		Name:         "Aspirant's Ascent",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3, "flying", "toxic 1")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.OnTheTargetCard(func(ctx *effects.Context, id uuid.UUID) error {
				if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 3, Label: "Aspirant's Ascent — +1/+3"}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying", "toxic 1"},
					Label: "Aspirant's Ascent — flying and toxic 1"}.Apply(ctx)
			}),
		),
	})
}
