package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Once and Future — Instant {3}{G} (Throne of Eldraine, #5516):
//
//	"Return target card from your graveyard to your hand. Put up to one other
//	 target card from your graveyard on top of your library. Exile Once and
//	 Future.
//	 Adamant — If at least three green mana was spent to cast this spell,
//	 instead return those cards to your hand and exile Once and Future."
//
// Two target clauses from your graveyard (the second up to one, and another
// card), each checked again on resolution (CR 608.2b). The first returns to
// hand; the second goes on top of the library, or to hand as well when
// adamant was met. Then the spell exiles itself instead of going to the
// graveyard (Retrieve's self-exile); a spell countered or fizzled is not
// exiled, because it never resolved.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "f666c38d-e601-4aa9-b140-4b6607bb12dd",
		Name:           "Once and Future",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G"},
		Purpose:        game.Purpose{Answers: game.AnswerValue},
		Targets: effects.Clauses(
			effects.TargetCardInGraveyard("target card in your graveyard", effects.YouOwn()),
			effects.Distinct(effects.TargetCardInGraveyard("up to one other target card in your graveyard", effects.YouOwn()).WithCount(0, 1)),
		),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				second := game.ZoneLibrary
				if effects.AdamantSpent(ctx, "G", 3) {
					second = game.ZoneHand
				}
				for slot, dest := range []game.ZoneKind{game.ZoneHand, second} {
					for _, t := range ctx.ClauseTargets(slot) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (effects.ReturnFromGraveyard{Target: t.ID, Dest: dest}).Apply(ctx); err != nil {
							return err
						}
					}
				}
				return nil
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				// "Exile Once and Future." The spell moves itself, so the
				// resolution frame does not put it into the graveyard (#489).
				return ctx.Game.ExileCardForEffect(ctx.Item.SourceCardID)
			}),
		),
	})
}
