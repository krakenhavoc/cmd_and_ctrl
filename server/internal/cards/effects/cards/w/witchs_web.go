package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6ec9a14a-0a62-4f8a-9147-3cab01411a65",
		Name:         "Witch's Web",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3, "reach")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := effects.PumpAndGrantFirstTarget(3, 3, "Witch's Web — +3/+3 and reach", "reach")(ctx.Game, item); err != nil {
				return err
			}
			return effects.UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
