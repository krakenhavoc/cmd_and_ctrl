package cardss

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scattering Stroke — Instant {2}{U}{U} (Commander 2011, #5574):
//
//	"Counter target spell. Clash with an opponent. If you win, at the beginning
//	 of your next main phase, you may add an amount of {C} equal to that
//	 spell's mana value."
//
// Mana Drain's refund (the same "your next main phase" step pick and the same
// delayed body) behind a clash: the spell's mana value is read while it is
// still on the stack (CR 202.3e) and BEFORE it is countered, then rides the
// clash's continuation as its number. A spell that can't be countered is still
// clashed over and paid out for. The refund is mana, which a player always
// wants, so the "may" asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "314a9604-34ec-4529-bd14-7b63be28669f",
		Name:         "Scattering Stroke",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			stackID := item.Targets[0].ID
			mv := 0
			if ctx.Game.StackItemForEffect(stackID) != nil {
				if card, ok := ctx.Game.LookupCardForEffect(stackID); ok {
					mv, _ = ctx.Game.ManaValueForEffect(card)
				}
			}
			if err := (effects.CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			return effects.ClashWithAnOpponent{Then: ctx.TailN(scatteringStrokeClashed, mv)}.Apply(ctx)
		},
	})
}

// scatteringStrokeClashed is "If you win, at the beginning of your next main
// phase, add {C} equal to that spell's mana value" (the tail's N). An on-disk
// key.
var scatteringStrokeClashed = effects.CardTail("then/scattering-stroke-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	mv := o.Tail.N
	if !effects.WonTheClash(o) || mv <= 0 {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		At:                 effects.ManaDrainNextMainPhaseStep(ctx.Game, ctx.Controller()),
		ControllerTurnOnly: true,
		Label:              "Scattering Stroke — add {C} × " + strconv.Itoa(mv),
		Body:               effects.ManaDrainRefundBody,
		Params:             game.EffectParams{Amount: mv},
	}.Apply(ctx)
})
