package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b212_helpers.go — shared pieces of Dominaria batch 212 (#5530).

// S82OldB212EnchantedHistoricPermanentLeft is "When enchanted permanent leaves
// the battlefield, if it was historic" (Curator's Ward): any departure of the
// permanent this Aura is attached to, where it was an artifact, a legendary or
// a Saga as it last existed on the battlefield (CR 700.6, CR 603.10a). The
// event is harvested while the Aura is still attached, so the attachment names
// the permanent.
func S82OldB212EnchantedHistoricPermanentLeft(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !EnchantedPermanentLeft(ev, source, lki, g) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return LeftAsType(ev, c, "artifact") || LeftAsSupertype(ev, c, "Legendary") || LeftAsSubtype(ev, c, "Saga")
}

// S82OldB212ReturnDiedCardToOwnersHand is "return that card to its owner's
// hand" for a trigger whose event names the card that died (Demonic Vigor): the
// card comes back from the graveyard it just reached. A token has no card to
// return, and a card that has since left the graveyard stays where it is.
func S82OldB212ReturnDiedCardToOwnersHand(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	id := item.Trigger.Event.CardID
	z := g.FindCardZoneForEffect(id)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}

// S82OldB212ExileTopCardLinked is "Exile the top card of your library" whose card
// another ability of the same source reads back as "exiled with this creature"
// (Rona, Disciple of Gix; CR 607.2a): the exile stamps the link to the source
// object. A source that has left exiles with no link.
func S82OldB212ExileTopCardLinked(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Library == nil || len(p.Library.Cards) == 0 {
		return nil
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	ref, ok := g.SourceObjectForEffect(item)
	if !ok {
		ref = game.ObjectRef{}
	}
	return g.ExileLinkedForEffect(ref, top)
}

// S82OldB212UrzasTomeDiscardUnlessExile is the tail of Urza's Tome: "Then
// discard a card unless you exile a historic card from your graveyard." With a
// historic card in the graveyard the player is asked which to exile (none is a
// choice, and then they discard); with none they discard at once. An empty hand
// discards nothing (CR 701.9a).
func S82OldB212UrzasTomeDiscardUnlessExile(ctx *Context) error {
	g, player := ctx.Game, ctx.Controller()
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	var historic []game.Card
	for _, c := range p.Graveyard.Cards {
		if B09IsHistoric(c) {
			historic = append(historic, c)
		}
	}
	if len(historic) == 0 {
		g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: ctx.Source(), N: 1, Question: "Urza's Tome — discard a card"})
		return nil
	}
	ids := make([]uuid.UUID, 0, len(historic))
	for _, c := range historic {
		ids = append(ids, c.InstanceID)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  player,
		Source:   ctx.Source(),
		Question: "Urza's Tome — exile a historic card from your graveyard? (If you don't, discard a card.)",
		Cards:    ids,
		Min:      0,
		Max:      1,
		Zone:     game.ZoneGraveyard,
		LiveZone: game.ZoneGraveyard,
		ThenKey:  s82OldB212TomeExiled,
	})
	return nil
}

// s82OldB212TomeExiled is the answer: a card picked is exiled, and none picked
// (or one that left the graveyard meanwhile) means the discard. An on-disk key:
// never renamed, never reused.
var s82OldB212TomeExiled = game.RegisterChooseCardsThen("choose-cards/s82-old-b212-urzas-tome-exile", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) > 0 {
		return g.ExileCardForEffect(r.Picked[0])
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: r.Chooser, Source: r.Source, N: 1, Question: "Urza's Tome — discard a card"})
	return nil
})

// S82OldB212ReturnThisFromGraveyardToHand is "return this card from your
// graveyard to your hand" for an ability that worked from the graveyard
// (Lingering Phantom): the card must still be the object the ability came
// from (CR 400.7), so one that left and came back is not returned.
func S82OldB212ReturnThisFromGraveyardToHand(g *game.Game, item *game.StackItem) error {
	z := g.FindCardZoneForEffect(item.SourceCardID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	if ref := item.SourceObject; ref.ID == item.SourceCardID && ref.ID != uuid.Nil {
		for i := range z.Cards {
			if z.Cards[i].InstanceID == ref.ID && z.Cards[i].ObjectEpoch != ref.Epoch {
				return nil
			}
		}
	}
	return ReturnFromGraveyard{Target: item.SourceCardID, Dest: game.ZoneHand}.Apply(NewContext(g, item))
}
