package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defend the Rider — Instant {G}:
//
//	"Choose one —
//	 • Target permanent you control gains hexproof and indestructible
//	   until end of turn.
//	 • Create a 1/1 colorless Pilot creature token with "This token
//	   saddles Mounts and crews Vehicles as though its power were 2
//	   greater.""
//
// The second mode is PilotToken (ADR 0186 §5, #3552). The first is a
// single-target keyword grant, pinned to the chosen permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "561af57d-d43d-4d98-9ef6-18a0843f22ae",
		Name:         "Defend the Rider",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target permanent you control gains hexproof and indestructible until end of turn.",
				effects.TargetPermanent("target permanent you control", effects.YouControl()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetCard {
						return nil
					}
					return effects.GrantKeywordUntilEOT{
						Target:   t.ID,
						Keywords: []string{"hexproof", "indestructible"},
						Label:    "Defend the Rider — hexproof and indestructible",
					}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof", "indestructible"))),
			effects.ModeDoing("Create a 1/1 colorless Pilot creature token.", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.CreateToken{Template: effects.PilotToken(), N: 1}.Apply(ctx)
				}),
		),
	})
}
