package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vision, Synthezoid Avenger — {4} Legendary Artifact Creature — Robot
// Hero, 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever a player casts a spell, if it isn't that player's turn,
//	 choose one —
//	 • Put a +1/+1 counter on Vision.
//	 • Vision phases out. (Treat him and anything attached to him as
//	   though they don't exist until your next turn.)"
//
// Flying is the printed keyword. The trigger is any player's cast event with
// an intervening "if it isn't that player's turn" (CR 603.4): the caster is
// not the active player, which is true as the trigger is put on the stack and
// cannot change while it waits, since the turn does not end with a spell on
// the stack. The bullet is chosen by Vision's controller as the trigger goes
// on the stack (CR 700.2b). The counter is put only if Vision is still on
// the battlefield. The phase-out is the engine's phasing (CR 702.26): Vision
// and what is attached to him phase in again during his controller's next
// untap step.
//
// No simplification.
func init() {
	cast := On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
		if ev.Actor == uuid.Nil || g.Turn.ActiveSeat < 0 || g.Turn.ActiveSeat >= len(g.Seats) {
			return false
		}
		active := g.Seats[g.Turn.ActiveSeat]
		return active != nil && active.ID != ev.Actor
	}, "Vision — choose one: a +1/+1 counter on Vision, or Vision phases out",
		func(*game.Game, *game.StackItem) error { return nil })
	cast.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on Vision.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
		ModeDoing("Vision phases out.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				if !onBattlefield(ctx.Game, ctx.Source()) {
					return nil
				}
				return PhaseOut{Targets: []uuid.UUID{ctx.Source()}}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:        "2f4360f6-9ac7-499f-b0a5-769550ea15ba",
		Name:            "Vision, Synthezoid Avenger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{cast},
	})
}
