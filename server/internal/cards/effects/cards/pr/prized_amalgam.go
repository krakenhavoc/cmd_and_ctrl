package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prized Amalgam — Creature — Zombie {1}{U}{B}, 3/3 (Innistrad: Midnight
// Hunt):
//
//	"Whenever a creature enters, if it entered from your graveyard or you
//	 cast it from your graveyard, return this card from your graveyard to
//	 the battlefield tapped at the beginning of the next end step."
//
// Bloodghast's graveyard trigger (InGraveyard: the ability works only from
// the graveyard, and "you" is the card's owner there, CR 108.4), watching
// every creature that enters, any player's: the entering creature came out
// of your graveyard (EnteredFromYourGraveyard), or you cast it from your
// graveyard (Archfiend's Vessel's cast read). The trigger has an
// intervening "if" (CR 603.4) read as the creature enters. The return is
// a delayed trigger (CR 603.7) naming the card as the object it is in the
// graveyard now (Nine-Lives Familiar's shape): a card that has left the
// graveyard by then is a new object and returns nothing (CR 400.7). A
// creature that enters at the same time the Amalgam does (the Amalgam
// itself returning from the graveyard) does not trigger it from the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8499b3ec-9a61-47a4-a880-e8230b41c81b",
		Name:         "Prized Amalgam",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventETB, gapB202EnteredFromYourGraveyardOrCast,
				"Prized Amalgam — return it to the battlefield tapped at the beginning of the next end step", prizedAmalgamSchedule)),
		},
	})
}

// gapB202EnteredFromYourGraveyardOrCast is "a creature enters, if it entered
// from your graveyard or you cast it from your graveyard", with "you" the
// source's controller (its owner while it sits in the graveyard).
func gapB202EnteredFromYourGraveyardOrCast(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
		return false
	}
	entered, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || !entered.IsCreature() {
		return false
	}
	if effects.EnteredFromYourGraveyard(ev, source, lki, g) {
		return true
	}
	cast, known := entered.CastByItsController()
	return known && cast && entered.Controller == source.Controller && entered.Provenance.FromZone == game.ZoneGraveyard
}

// prizedAmalgamReturnBody is the delayed trigger's body: put the Amalgam
// back tapped if it is still the object that was in the graveyard. An
// on-disk key: never renamed, never reused.
var prizedAmalgamReturnBody = game.DelayedBody("prized-amalgam/return", prizedAmalgamReturn)

func prizedAmalgamSchedule(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return effects.ScheduleDelayedTrigger{
		Label:  "Prized Amalgam — return it to the battlefield tapped",
		Cards:  []uuid.UUID{id},
		Body:   prizedAmalgamReturnBody,
		Params: game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
	}.Apply(effects.NewContext(g, item))
}

func prizedAmalgamReturn(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(id, uuid.Nil, true, nil)
	return effects.NothingIfGone(err)
}
