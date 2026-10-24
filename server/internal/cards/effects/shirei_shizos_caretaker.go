package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shirei, Shizo's Caretaker — Legendary Creature — Spirit {4}{B}, 2/2:
//
//	"Whenever a creature with power 1 or less is put into your
//	 graveyard from the battlefield, you may return that card to the
//	 battlefield at the beginning of the next end step if Shirei is
//	 still on the battlefield."
//
// Avacyn, Angel of Horror's shape with three conditions of its own:
//
//   - "Power 1 or less" is the creature's power as it last existed on
//     the battlefield (CR 603.10a), read off its departure record, so a
//     2/2 shrunk to 1/1 counts and a 1/1 pumped to 3/3 does not. A
//     negative power is 1 or less too.
//   - "Your graveyard" is the graveyard it went to: a card Shirei's
//     controller OWNS, whoever controlled it. A token is a creature put
//     into a graveyard too, but it has ceased to exist by the end step
//     and there is nothing to return.
//   - The return is a CR 603.7 delayed trigger naming the card as the
//     object it is now and Shirei as the object she is now. At the end
//     step the card comes back only if it is still that object in a
//     graveyard AND Shirei is still that object on the battlefield;
//     then its owner is asked the "may" (CR 603.5). A Shirei that left
//     and came back is a new object (CR 400.7) and returns nothing.
//
// Shirei is the ability's own source, so her own death (power 2) never
// triggers it, and a Shirei made 1/1 that dies is gone by the end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7b5c3d28-6d8b-488b-b088-567b24faadc7",
		Name:         "Shirei, Shizo's Caretaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, shireiSmallCreatureToYourGraveyard,
				"Shirei, Shizo's Caretaker — return that card at the beginning of the next end step",
				shireiScheduleReturn),
		},
	})
}

// shireiSmallCreatureToYourGraveyard is the trigger condition: a
// creature whose last-known power was 1 or less went from the
// battlefield to the graveyard of Shirei's controller.
func shireiSmallCreatureToYourGraveyard(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := diedCreature(ev, g)
	if !ok || dead.Owner != source.Controller {
		return false
	}
	power := dead.Power
	if info, ok := g.LastKnownPermanentForEffect(ev.CardID); ok {
		power = info.Power
	}
	return power <= 1
}

// shireiScheduleReturn schedules the end-step return, naming the card
// and Shirei as the objects they are now.
func shireiScheduleReturn(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	shirei, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Shirei, Shizo's Caretaker — you may return that card to the battlefield",
		Cards: []uuid.UUID{id},
		Body:  shireiReturnBody,
		Params: game.EffectParams{
			Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch},
			Amount: shirei.ObjectEpoch,
		},
	}.Apply(NewContext(g, item))
}

// shireiReturn is the delayed body (shireiReturnBody, delayed_bodies.go):
// if Shirei — the item's source, at the epoch in p.Amount — is still on
// the battlefield and the card is still that object in a graveyard, its
// owner may return it.
func shireiReturn(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	shirei, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || shirei.ObjectEpoch != p.Amount || !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return MayChoice{
		Player:   item.Controller,
		Question: "Shirei, Shizo's Caretaker — return " + c.Name + " to the battlefield?",
		YesLabel: "Return it",
		NoLabel:  "Leave it",
		Then:     shireiAnswer,
		Carry:    []uuid.UUID{id},
	}.Apply(NewContext(g, item))
}

// shireiAnswer is the "yes": return the creature (the carried ID) to the
// battlefield. Keyed (#4291).
var shireiAnswer = MayYes("confirm/shirei-return", func(ctx *Context, r game.ConfirmAnswer) error {
	return ReturnFromGraveyard{Target: carryAt(r, 0), Dest: game.ZoneBattlefield}.Apply(ctx)
})
