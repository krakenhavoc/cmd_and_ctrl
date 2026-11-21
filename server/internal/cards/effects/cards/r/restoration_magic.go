package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restoration Magic — Instant {W} (#3777):
//
//	"Tiered (Choose one additional cost.)
//	 • Cure — {0} — Target permanent gains hexproof and indestructible
//	   until end of turn.
//	 • Cura — {1} — Target permanent gains hexproof and indestructible
//	   until end of turn. You gain 3 life.
//	 • Curaga — {3}{W} — Permanents you control gain hexproof and
//	   indestructible until end of turn. You gain 6 life."
//
// Tiered proof card (CR 702.183a). Curaga's "permanents you control"
// is read once as it resolves (CR 611.2c), so a permanent that enters
// later in the turn gains nothing.
//
// No simplification.
func init() {
	protect := []string{"hexproof", "indestructible"}
	cure := func(life int, label string) func(*game.StackItem, *effects.Context, int) error {
		return func(_ *game.StackItem, ctx *effects.Context, occ int) error {
			if t, ok := effects.ModeTarget(ctx, occ); ok {
				if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: protect, Label: label}).Apply(ctx); err != nil {
					return err
				}
			}
			if life == 0 {
				return nil
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: life}.Apply(ctx)
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "ddbe0d17-5cac-417c-afc0-e4aebae83205",
		Name:         "Restoration Magic",
		Completeness: effects.CompletenessFull,
		Modes: effects.Tiered(
			effects.ModeWithPurpose(effects.SpreeModeDoing("Cure — Target permanent gains hexproof and indestructible until end of turn.", "{0}",
				effects.TargetPermanent("target permanent"), cure(0, "Restoration Magic — Cure")), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible"))),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Cura — Target permanent gains hexproof and indestructible until end of turn. You gain 3 life.", "{1}",
				effects.TargetPermanent("target permanent"), cure(3, "Restoration Magic — Cura")), game.Purpose{LifeGain: 3, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible"))}),
			effects.SpreeModeDoing("Curaga — Permanents you control gain hexproof and indestructible until end of turn. You gain 6 life.", "{3}{W}", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					if err := (effects.GrantKeywordUntilEOT{Match: effects.YouControl(), Keywords: protect, Label: "Restoration Magic — Curaga"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: ctx.Controller(), Amount: 6}.Apply(ctx)
				}),
		),
	})
}
