package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Origin of Metalbending — Instant — Lesson {1}{G}:
//
//	"Choose one —
//	 • Destroy target artifact or enchantment.
//	 • Put a +1/+1 counter on target creature you control. It gains
//	   indestructible until end of turn."
//
// It's Clobberin' Time!'s destroy bullet and Spectacular Tactics's
// counter-then-grant bullet with indestructible. The chosen bullet's
// target is rechecked as the spell resolves (CR 608.2b). Lesson is a
// spell subtype on the type line and needs nothing here.
//
// No simplification. Avatar: The Last Airbender (#4459).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "177113f8-1188-4e5e-8ad7-34a7bb8ca7f1",
		Name:         "Origin of Metalbending",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				effects.DestroyTheModesTarget),
			effects.ModeDoing("Put a +1/+1 counter on target creature you control. It gains indestructible until end of turn.",
				effects.TargetCreature("target creature you control", effects.YouControl()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := effects.PlusOneCounterOnTheModesTarget(nil, ctx, occ); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Origin of Metalbending — indestructible"}.Apply(ctx)
				}),
		),
	})
}
