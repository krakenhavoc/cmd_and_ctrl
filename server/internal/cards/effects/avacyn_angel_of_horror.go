package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Avacyn, Angel of Horror — Legendary Creature — Angel {5}{B}{B}{B}, 6/6:
//
//	"Flying, deathtouch
//	 Whenever Avacyn or another nontoken creature you control dies,
//	 return that card to the battlefield under your control at the
//	 beginning of the next end step."
//
// One trigger with two conditions: Avacyn herself dying (ThisDied), or
// Liesa's "another nontoken creature you control" dying. Avacyn's own
// clause names no token restriction, but she is never a token unless
// copied, and a token that died has ceased to exist (CR 111.7) and has
// nothing to return.
//
// The trigger schedules a delayed trigger for the next end step (CR
// 603.7) naming the card in the graveyard as the object it is NOW,
// with its epoch. If that card leaves the graveyard before the end
// step, even to come back, it is a new object (CR 400.7) and nothing
// returns. It comes back under the control of whoever controlled
// Avacyn's trigger, which is "your".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4af61d45-7114-4eb4-a905-9e03caf87454",
		Name:            "Avacyn, Angel of Horror",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ThisDied(ev, source, ch, g) || b19AnotherNontokenCreatureYouControlDied(ev, source, g)
			},
			Key:    "Avacyn, Angel of Horror — return that card to the battlefield at the beginning of the next end step",
			Effect: avacynScheduleReturn,
		}},
	})
}

// avacynScheduleReturn schedules the end-step return for the card that
// died, if it is still in a graveyard as the trigger resolves.
func avacynScheduleReturn(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label:  "Avacyn, Angel of Horror — return the card to the battlefield under your control",
		Cards:  []uuid.UUID{id},
		Body:   returnTheObjectFromGraveyardUnderYourControlBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(NewContext(g, item))
}

// returnTheObjectFromGraveyardUnderYourControl is the delayed body: the
// card <Object> names goes from its graveyard to the battlefield under
// the delayed trigger's controller, if it is still that object in a
// graveyard.
func returnTheObjectFromGraveyardUnderYourControl(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}.Apply(NewContext(g, item))
}
