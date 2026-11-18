package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witch's Web — Instant {1}{G} (Innistrad: Crimson Vow):
//
//	"Target creature gets +3/+3 and gains reach until end of turn. Untap
//	 it."
//
// Acrobatic Leap's shape with reach: pump and keyword until the cleanup
// step, then the untap (CR 701.26b). A target that is no longer legal
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ec9a14a-0a62-4f8a-9147-3cab01411a65",
		Name:         "Witch's Web",
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "reach")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := pumpAndGrantFirstTarget(3, 3, "Witch's Web — +3/+3 and reach", "reach")(ctx.Game, item); err != nil {
				return err
			}
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
