package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b227_helpers.go — shared shapes of the GAP batch 227 cards
// (#5948): Backgrounds that grant an end-step or upkeep trigger gated on
// something that happened this turn. A file of its own, so concurrent
// batches never collide.

// InterveningIfYou narrows a trigger to an intervening "if" (CR 603.4) about
// its controller: it triggers only while `test` holds for the controller and
// does nothing if it has stopped holding as the ability resolves. `name`
// words the condition for the footprint.
func InterveningIfYou(name string, test func(g *game.Game, you uuid.UUID) bool, t game.TriggeredAbility) game.TriggeredAbility {
	when := t.AppliesTo
	t.AppliesTo = func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return source != nil && test(g, source.Controller) && (when == nil || when(ev, source, lki, g))
	}
	return InterveningIf(CondBoard{Name: name, Test: func(ctx *Context) bool {
		return test(ctx.Game, ctx.Controller())
	}}, t)
}

// CreatureCardWasPutIntoYourGraveyardThisTurn is "a creature card was put
// into your graveyard from anywhere this turn" (Cloakwood Hermit): a zone
// move into a graveyard of a creature card its owner is `you`, read off the
// turn's event log. A token is not a card.
func CreatureCardWasPutIntoYourGraveyardThisTurn(g *game.Game, you uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventZoneMove || ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard || ev.CardID == uuid.Nil {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && !c.IsToken() && c.Owner == you && c.HasCardType("creature") {
			return true
		}
	}
	return false
}

// damageSourceIsControlledBy reads who controlled the source of a damage
// event as it dealt it: the event's actor when it names one, else the
// controller the source card carries (a creature that has since died keeps
// it).
func damageSourceIsControlledBy(g *game.Game, ev game.Event, you uuid.UUID) bool {
	if ev.Actor != uuid.Nil {
		return ev.Actor == you
	}
	src, ok := g.LookupCardForEffect(ev.Source)
	return ok && src.Controller == you
}

// ASourceYouControlledDealtAtLeastDamageThisTurn is "a source you controlled
// dealt N or more damage this turn" (Dragon Cultist): one source's damage
// this turn, to anything, added up from the turn's damage events; the
// damage's actor is the player who controlled the source as it dealt it.
func ASourceYouControlledDealtAtLeastDamageThisTurn(g *game.Game, you uuid.UUID, n int) bool {
	dealt := map[uuid.UUID]int{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventDealDamage || ev.Source == uuid.Nil || ev.Amount <= 0 {
			continue
		}
		if !damageSourceIsControlledBy(g, ev, you) {
			continue
		}
		dealt[ev.Source] += ev.Amount
		if dealt[ev.Source] >= n {
			return true
		}
	}
	return false
}
