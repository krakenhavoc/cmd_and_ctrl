package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "83917116-d2d8-4701-a79f-af17486706ae",
		Name:         "Light the Way",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Put a +1/+1 counter on target creature or Vehicle. Untap it.",
				TargetPermanent("target creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (AddCounter{Target: t.ID, Kind: "+1/+1", N: 1}).Apply(ctx); err != nil {
						return err
					}
					return UntapTarget{Target: t.ID}.Apply(ctx)
				}), game.Purpose{Answers: game.AnswerPump}),
			ModeWithPurpose(ModeDoing("Return target permanent you control to its owner's hand.",
				TargetPermanent("target permanent you control", YouControl()),
				BounceTheModesTarget), game.Purpose{Answers: game.AnswerProtect}),
		),
	})
}
