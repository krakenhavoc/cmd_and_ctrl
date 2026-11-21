package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wake the Dead — Instant {X}{B}{B} (Commander Masters, #5948):
//
//	"Cast this spell only during combat on an opponent's turn.
//	 Return X target creature cards from your graveyard to the battlefield.
//	 Sacrifice those creatures at the beginning of the next end step."
//
// Exactly X targets (CountFromX), each re-checked as the spell resolves
// (CR 608.2b); the ones still legal return together. Each creature that
// actually entered is sacrificed by its own delayed trigger at the next end
// step (CR 603.7), which is this turn's, since the spell is cast in combat.
// A creature that has left by then is not sacrificed (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69bf79c3-87d1-4244-8d58-989307c91a12",
		Name:         "Wake the Dead",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, caster uuid.UUID, _ game.Card) bool {
			return game.PhaseOf(g.Turn.Step) == game.PhaseCombat &&
				g.Turn.ActiveSeat >= 0 && g.Turn.ActiveSeat < len(g.Seats) && g.Seats[g.Turn.ActiveSeat].ID != caster
		},
		CastConditionLabel: "Cast this spell only during combat on an opponent's turn.",
		Targets: targetsCountedByX(TargetCardInGraveyard("X target creature cards from your graveyard",
			YouOwn(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets:  legalTargetCardIDs(ctx),
				ThenTail: wakeTheDeadSacrificeTail,
			}.Apply(ctx)
		},
	})
}

// wakeTheDeadSacrificeTail schedules the sacrifice of every creature the
// return put onto the battlefield. An on-disk key: never renamed, never
// reused.
var wakeTheDeadSacrificeTail = TogetherTail("gap-b111/wake-the-dead-sacrifice",
	func(ctx *Context, entered []uuid.UUID, _ int) error {
		for _, id := range entered {
			c, ok := battlefieldCardForEffect(ctx.Game, id)
			if !ok {
				continue
			}
			if err := (ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Item.Controller,
				Label:      "Wake the Dead — sacrifice " + c.Name,
				Body:       cmdr2S12SoulstokeSacrifice,
				Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})
