package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Escape from Orthanc — Instant {W} (The Lord of the Rings: Tales of
// Middle-earth, #5948):
//
//	"Target creature gets +1/+3 and gains flying until end of turn. Untap it."
//
// The pump and flying are one effect on the target as it resolves (CR 611.2c);
// the untap follows. An already untapped creature is simply not untapped
// (CR 701.26b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2025f2e-c931-48dd-af21-8f0a591befb4",
		Name:         "Escape from Orthanc",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 3, "flying")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Toughness: 3, Label: "Escape from Orthanc — +1/+3"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"flying"}, Label: "Escape from Orthanc — flying"}).Apply(ctx); err != nil {
				return err
			}
			return effects.UntapTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
