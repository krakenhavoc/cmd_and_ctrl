package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 — Crimson Vow, batch 16: cards exiled with a permanent. The four
// Cemetery creatures exile one card from any graveyard, linked to their own
// trigger (CR 607.2a), and read it back through the card types it has.

// vowB16GraveyardCards is every card in every graveyard but `except`.
func vowB16GraveyardCards(g *game.Game, except uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		for _, id := range GraveyardCardsOfPlayer(g, p.ID) {
			if id != except {
				out = append(out, id)
			}
		}
	}
	return out
}

// vowB16QueueExile asks the item's controller for one card from any
// graveyard other than `except` (uuid.Nil: any card), and hands the pick to
// `then`. Nothing is asked with every graveyard empty.
func VowB16QueueExile(g *game.Game, item *game.StackItem, label string, except uuid.UUID, then game.ChooseCardsThen) {
	candidates := vowB16GraveyardCards(g, except)
	if len(candidates) == 0 {
		return
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: label,
		Cards:    candidates,
		Min:      1,
		Max:      1,
		ThenKey:  then,
		Asked:    game.NewAskedItem(item),
	})
}

// vowB16LinkedExileThen registers the answer that exiles the pick, linked to
// the asking trigger. key is an on-disk identity: never renamed, never
// reused.
func VowB16LinkedExileThen(key string) game.ChooseCardsThen {
	return game.RegisterChooseCardsThen(key, func(g *game.Game, r game.CardsChosen) error {
		if len(r.Picked) == 0 || r.Asked == nil {
			return nil
		}
		return ExileTarget{Target: r.Picked[0], Linked: true}.Apply(NewContext(g, r.Asked))
	})
}

// vowB16ExiledWith is the cards exiled with the object `source` by the
// trigger row `label` (CR 607.2a).
func vowB16ExiledWith(g *game.Game, source game.PermissionCardRef, label string) []*game.Card {
	return g.ExiledWithRowForEffect(source, game.AbilitySlotTriggered, label)
}

// vowB16ExiledWithCard is vowB16ExiledWith for a source still in play.
func VowB16ExiledWithCard(g *game.Game, source *game.Card, label string) []*game.Card {
	return vowB16ExiledWith(g, game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}, label)
}

// vowB16SharesATypeWithAny reports whether `c` has a card type that any of
// `exiled` has (CR 205.2a).
func VowB16SharesATypeWithAny(c game.Card, exiled []*game.Card) bool {
	for _, t := range SemblanceCardTypes {
		if !c.HasCardType(t) {
			continue
		}
		for _, e := range exiled {
			if e.HasCardType(t) {
				return true
			}
		}
	}
	return false
}

// vowB16PlayedCardBy is the card a play-a-land-or-cast-a-spell event is
// about and the player who did it. False for any other event.
func vowB16PlayedCardBy(ev game.Event, g *game.Game) (game.Card, uuid.UUID, bool) {
	if ev.Kind != game.EventCast && !B20LandPlayed(ev, g) {
		return game.Card{}, uuid.Nil, false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || ev.Actor == uuid.Nil {
		return game.Card{}, uuid.Nil, false
	}
	return c, ev.Actor, true
}

// vowB16PlaysSharingAType is the "whenever <who> plays a land or casts a
// spell, if it shares a card type with the exiled card" test, as the ability
// triggers: `who` says whether the player counts (nil: any player), and the
// exiled card is the one `label`'s row exiled with the source.
func VowB16PlaysSharingAType(label string, who func(actor uuid.UUID, source *game.Card) bool) When {
	return func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		c, actor, ok := vowB16PlayedCardBy(ev, g)
		if !ok || (who != nil && !who(actor, source)) {
			return false
		}
		return VowB16SharesATypeWithAny(c, VowB16ExiledWithCard(g, source, label))
	}
}

// vowB16PlaySharesStillHolds is the same test as the ability resolves
// (CR 603.4), the source read as it last existed if it has gone.
func VowB16PlaySharesStillHolds(label string) Condition {
	return CondBoard{
		Name: "the land played or spell cast shares a card type with the exiled card",
		Test: func(ctx *Context) bool {
			ref, ok := ctx.SourceRef()
			if !ok {
				return false
			}
			c, ok := ctx.Game.LookupCardForEffect(ctx.Trigger().Event.CardID)
			if !ok {
				return false
			}
			return VowB16SharesATypeWithAny(c, vowB16ExiledWith(ctx.Game, game.PermissionCardRef(ref), label))
		},
	}
}

// vowB16YouPlay is "you" for vowB16PlaysSharingAType.
func VowB16YouPlay(actor uuid.UUID, source *game.Card) bool { return actor == source.Controller }

// vowB16PlayEvents are the events a play-a-land-or-cast-a-spell ability
// watches.
var VowB16PlayEvents = []game.EventKind{game.EventCast, game.EventZoneMove}
