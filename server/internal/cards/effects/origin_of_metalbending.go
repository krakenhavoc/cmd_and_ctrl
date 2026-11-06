package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "177113f8-1188-4e5e-8ad7-34a7bb8ca7f1",
		Name:         "Origin of Metalbending",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				DestroyTheModesTarget),
			ModeDoing("Put a +1/+1 counter on target creature you control. It gains indestructible until end of turn.",
				TargetCreature("target creature you control", YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := PlusOneCounterOnTheModesTarget(nil, ctx, occ); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Origin of Metalbending — indestructible"}.Apply(ctx)
				}),
		),
	})
}
