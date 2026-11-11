package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Plasm Capture — Instant {G}{G}{U}{U}:
//
//	"Counter target spell. At the beginning of your next first main phase,
//	 add X mana in any combination of colors, where X is that spell's mana
//	 value."
//
// Mana Drain's shape with coloured mana and the precombat main only. The
// mana value is read while the spell is still on the stack (CR 202.3e)
// and the refund is a CR 603.7 delayed trigger gated to the caster's own
// turn: "your next first main phase" is the next precombat main on a turn
// of yours, which may be this turn's only when the Capture resolves
// before it begins (never, for an instant cast in a main phase).
func init() {
	Register(Spec{
		OracleID:     "3673219d-d2b5-43ad-b0c8-58b583491158",
		Name:         "Plasm Capture",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
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
			if err := (CounterTarget{StackID: stackID}).Apply(ctx); err != nil {
				return err
			}
			if mv <= 0 {
				return nil
			}
			return ScheduleDelayedTrigger{
				At:                 game.StepPrecombatMain,
				ControllerTurnOnly: true,
				Label:              "Plasm Capture — add " + strconv.Itoa(mv) + " mana in any combination of colors",
				Body:               cmdr2S42AnyColorRefundBody,
				Params:             game.EffectParams{Amount: mv},
			}.Apply(ctx)
		},
	})
}
