package game

import "github.com/google/uuid"

// dash.go — dash (CR 702.109, #3172): "You may cast this card for its
// dash cost. If you do, it gains haste, and it's returned from the
// battlefield to its owner's hand at the beginning of the next end step."
//
// The dash cost is an alternative cost (AlternativeCost.Dash), and like
// evoke and warp it attaches its clauses to the permanent's ENTRY
// (queueAltCostEntryTriggerLocked), the last point at which the stack
// item and the cost that was paid are both in hand:
//
//   - HASTE is a layer-6 scoped effect for as long as the creature
//     remains on the battlefield (CR 611.2b), pinned to the object, so a
//     creature that left and came back is not hasty by the old dash.
//   - The RETURN is a CR 603.7 delayed trigger at the next end step. It
//     returns the creature only if it is still the object that was dashed
//     (CR 400.7): the epoch is recorded when the trigger is scheduled, so
//     a creature that was flickered away and back in the meantime stays.
//     A creature that already left the battlefield is not returned from
//     anywhere else.
//
// The return is the delayed trigger's own resolution, so it uses the
// stack and every player may respond (Ragavan can be sacrificed, or
// exiled, in response and nothing then comes back).

// dashHasteLabel names the scoped effect on the wire.
const dashHasteLabel = "Dash — haste"

// dashReturnBody is the delayed trigger's body, a registered key (ADR
// 0041 phase 3). Assigned in init: a var initialiser would be an
// initialisation cycle through the exit primitives, the shape
// warpExileBody's comment explains.
var dashReturnBody BodyRef

func init() { dashReturnBody = DelayedBody("dash/return", dashReturn) }

// applyDashLocked gives a dashed creature its haste and schedules its
// return. Caller must hold g.mu.
func (g *Game) applyDashLocked(card Card, item *StackItem, alt *AlternativeCost) {
	live, ok := g.battlefieldCardLocked(card.InstanceID)
	if !ok {
		return
	}
	if d, ok := g.ForAsLongAsOnBattlefieldDuration(live.InstanceID); ok {
		g.RegisterScopedEffectForEffect(live.InstanceID,
			[]AffectedObject{PinObject(live.InstanceID, live.EnteredBattlefieldAt)},
			[]Mod{AddKeywordsMod("haste")}, d, dashHasteLabel)
	}
	g.ScheduleDelayedTriggerForEffect(DelayedTrigger{
		Controller:   item.Controller,
		SourceCardID: live.InstanceID,
		Label:        live.Name + " — " + alt.Label + ", return it to its owner's hand",
		At:           StepEnd,
		Cards:        []uuid.UUID{live.InstanceID},
		Body:         dashReturnBody,
		// The object that was dashed (CR 400.7), recorded now.
		Params: EffectParams{Amount: live.ObjectEpoch},
	})
}

// dashReturn is the delayed trigger's body: return the dashed creature to
// its owner's hand if it is still the object that was dashed.
func dashReturn(g *Game, it *StackItem, p EffectParams) error {
	for _, t := range it.Targets {
		if t.Kind != TargetCard {
			continue
		}
		c, ok := g.battlefieldCardLocked(t.ID)
		if !ok || c.ObjectEpoch != p.Amount {
			continue
		}
		if err := g.BounceToHandForEffect(t.ID); err != nil {
			return err
		}
	}
	return nil
}
