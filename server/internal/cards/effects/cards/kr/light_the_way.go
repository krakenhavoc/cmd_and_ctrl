package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Light the Way — Instant {W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Choose one —
//	 • Put a +1/+1 counter on target creature or Vehicle. Untap it.
//	 • Return target permanent you control to its owner's hand."
//
// A Vehicle that isn't a creature still gets the counter and the untap.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83917116-d2d8-4701-a79f-af17486706ae",
		Name:         "Light the Way",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Put a +1/+1 counter on target creature or Vehicle. Untap it.",
				effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (effects.AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
					return effects.UntapTarget{Target: t.ID}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
			effects.ModeWithPurpose(effects.ModeDoing("Return target permanent you control to its owner's hand.",
				effects.TargetPermanent("target permanent you control", effects.YouControl()),
				effects.BounceTheModesTarget), game.Purpose{Answers: game.AnswerProtect}),
		),
	})
}
