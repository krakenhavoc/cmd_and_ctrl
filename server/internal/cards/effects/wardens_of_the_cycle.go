package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wardens of the Cycle — Creature — Elf Warlock {1}{B}{G}{G}, 3/4
// (Foundations, #4465):
//
//	"Morbid — At the beginning of your end step, if a creature died this
//	 turn, choose one —
//	 • You gain 2 life.
//	 • You draw a card and you lose 1 life."
//
// Cackling Prowler's morbid end-step trigger with its intervening "if"
// (condACreatureDiedThisTurn, CR 603.4), carrying Apothecary Stomper's
// "choose one" bullets. The mode is chosen as the trigger goes on the
// stack (CR 700.2b). A creature that died this turn stays dead for the
// rest of it, so the condition still holds as the trigger resolves.
//
// No simplification.
func init() {
	t := AtYourEndStep("Wardens of the Cycle — morbid: choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("You gain 2 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 2}.Apply(ctx)
		}),
		ModeDoing("You draw a card and you lose 1 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			if err := (DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: item.Controller, Amount: -1}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "fa7d09aa-2ebc-4571-81da-836cd80cf554",
		Name:         "Wardens of the Cycle",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{IfVentureCond(condACreatureDiedThisTurn, t)},
	})
}
