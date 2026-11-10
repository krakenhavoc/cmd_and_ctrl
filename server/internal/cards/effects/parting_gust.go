package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Parting Gust — Instant {W} (Bloomburrow, #4467):
//
//	"Gift a tapped Fish (You may promise an opponent a gift as you cast
//	 this spell. If you do, they create a tapped 1/1 blue Fish creature
//	 token before its other effects.)
//	 Exile target nontoken creature. If the gift wasn't promised, return
//	 that card to the battlefield under its owner's control with a +1/+1
//	 counter on it at the beginning of the next end step."
//
// The exile is one event; a promise leaves the card there for good, and no
// promise schedules the CR 603.7 delayed return (a new object, with the
// counter put on as it enters, CR 614.12).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "367bb252-41e1-4f98-9d22-442dd332cccf",
		Name:         "Parting Gust",
		Completeness: CompletenessFull,
		Gift:         GiftATappedFish(),
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      TargetCreature("target nontoken creature", Not(IsTokenPredicate())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if ctx.GiftPromised() {
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						return ExileTarget{Target: t.ID}.Apply(ctx)
					}
				}
				return nil
			}
			return exileTargetsThenScheduleReturnBody(ctx,
				"Parting Gust — return the exiled card with a +1/+1 counter",
				returnExiledWithPlusOneCreatureBody)
		},
	})
}
