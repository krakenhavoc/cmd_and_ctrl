package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rescue from the Underworld — Instant {4}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Choose target creature card in your graveyard. Return that card and
//	 the sacrificed card to the battlefield under your control at the
//	 beginning of your next upkeep. Exile Rescue from the Underworld."
//
// #2099, ADR 0026 amendment 2026-10-16. The spell schedules one delayed
// trigger (CR 603.7) for the beginning of its controller's next upkeep
// and names two objects in it (EffectParams.Objects):
//
//   - the target, as the object it is in the graveyard now;
//   - the sacrificed card, as the object the sacrifice made of it in the
//     public zone it went to (CR 400.7j: the spell's effects can find
//     an object its cost moved). That is the graveyard, or exile when a
//     replacement such as Rest in Peace sent it there (the 2013-09-15
//     ruling). It is the card one zone change after the permanent the
//     payment record names (PaidCost.SacrificedObjects), so a card that
//     has moved again since is not it.
//
// The trigger returns each of them that is still that object in that
// zone, together, as one event (CR 603.6a), under the trigger's
// controller (CR 603.7d). A card that left its zone before the trigger
// resolves, even to come back, is a new object and stays where it is
// (CR 603.7c, 400.7; the 2013-09-15 ruling). A sacrificed token has
// ceased to exist (CR 111.7), so only the target returns (the 2013-09-15
// ruling).
//
// A sacrificed commander its owner moves to the command zone does not
// return. The 2013-09-15 ruling says it would, but it predates CR
// 903.9a: a commander now goes to the graveyard first and moves to the
// command zone as a state-based action, which is a second zone change
// (CR 400.7). Left in the graveyard, it returns.
//
// If the target is illegal as the spell resolves, the spell does not
// resolve (CR 608.2b): no trigger, nothing returns, and the spell goes
// to the graveyard rather than exile. Otherwise it exiles itself as it
// resolves, not later (the 2013-09-15 ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "7217ea53-3685-48d6-9df1-58a551f8c979",
		Name:           "Rescue from the Underworld",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := rescueFromTheUnderworldSchedule(ctx); err != nil {
				return err
			}
			return ExileTarget{Target: ctx.Source()}.Apply(ctx)
		},
	})
}

// rescueFromTheUnderworldSchedule schedules the upkeep return of the
// target and of the card the cost sacrificed, each as the object it is
// now. The target is legal, or the spell would not be resolving.
func rescueFromTheUnderworldSchedule(ctx *Context) error {
	var objects []game.ObjectRef
	var cards []uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			continue
		}
		objects = append(objects, game.ObjectRef{ID: t.ID, Epoch: c.ObjectEpoch})
		cards = append(cards, t.ID)
		break
	}
	for _, ref := range ctx.Paid().SacrificedObjects {
		if obj, ok := cardTheCostMovedToAPublicZone(ctx.Game, ref); ok {
			objects = append(objects, obj)
			cards = append(cards, obj.ID)
		}
	}
	if len(objects) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:                 game.StepUpkeep,
		ControllerTurnOnly: true,
		Label:              "Rescue from the Underworld — return that card and the sacrificed card to the battlefield under your control",
		Cards:              cards,
		Body:               returnTheObjectsTogetherUnderYourControlBody,
		Params:             game.EffectParams{Objects: objects},
	}.Apply(ctx)
}
