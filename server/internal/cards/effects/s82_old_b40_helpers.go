package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b40_helpers.go — shared bodies for the S82 older-sets batch 40
// (Strixhaven: School of Mages, #5502).

// S82OldB40CardsPutIntoExileOnYourTurn is "one or more cards are put into
// exile during your turn" (Stonebinder's Familiar): a card, not a token,
// that moved into exile from any zone while it is the source controller's
// turn. A permanent leaves through EventLTB and a card from any other
// zone through EventZoneMove, so the two never double count.
func S82OldB40CardsPutIntoExileOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	switch {
	case ev.Kind == game.EventZoneMove && ev.NewZone == game.ZoneExile && ev.OldZone != game.ZoneBattlefield:
	case ev.Kind == game.EventLTB && ev.NewZone == game.ZoneExile:
	default:
		return false
	}
	if !IsYourTurn(g, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}

// S82OldB40DiedWithAPlusOneCounter is "when this creature dies, if it had a
// +1/+1 counter on it" (Promising Duskmage): last-known information
// (CR 603.10a, 603.4), read from the departed object's record.
func S82OldB40DiedWithAPlusOneCounter(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return CardDied(ev, source) && g.LastKnownCountersForEffect(ev.CardID)[game.CounterPlusOne] > 0
}

// S82OldB40CopyPerOtherInstantOrSorcery is Show of Confidence's "When you
// cast this spell, copy it for each other instant and sorcery spell you've
// cast this turn. You may choose new targets for the copies." It is
// storm's shape with its own count, read as the trigger resolves: the
// caster's instant and sorcery spells cast this turn, less this one.
func S82OldB40CopyPerOtherInstantOrSorcery() game.TriggeredAbility {
	t := Storm()
	t.Keyword = ""
	t.Key = "show-of-confidence/copy-per-other-instant-or-sorcery"
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		return stormItem(source.Name, ev.Actor, source.InstanceID)
	}
	t.Effect = func(g *game.Game, item *game.StackItem) error {
		n := g.CastTallyFor(item.Controller).InstantSorcery - 1
		if n <= 0 {
			return nil
		}
		return CopySpell{
			StackID:          item.SourceCardID,
			Controller:       item.Controller,
			Count:            n,
			ChooseNewTargets: true,
			FromLastKnown:    true,
		}.Apply(NewContext(g, item))
	}
	return t
}

// S82OldB40CeilHalfManaValue is "half that card's mana value, rounded up".
func S82OldB40CeilHalfManaValue(mv int) int { return (mv + 1) / 2 }

// S82OldB40PickInstantOrSorcery is Torrent Sculptor's "exile an instant or
// sorcery card from your graveyard", then the counters on the source by half
// its mana value rounded up. The pick is a keyed prompt (a restore point
// while it waits), the counters go on only while the Sculptor is the same
// permanent on the battlefield.
func S82OldB40PickInstantOrSorcery(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil || p.Graveyard == nil {
			return nil
		}
		var ids []uuid.UUID
		for _, c := range p.Graveyard.Cards {
			if c.IsInstant() || c.IsSorcery() {
				ids = append(ids, c.InstanceID)
			}
		}
		if len(ids) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  item.Controller,
			Source:   item.SourceCardID,
			Question: label + " — exile an instant or sorcery card from your graveyard",
			Cards:    ids,
			Min:      1,
			Max:      1,
			ThenKey:  s82OldB40InstantOrSorceryPicked,
		})
		return nil
	}
}

// s82OldB40InstantOrSorceryPicked exiles the picked card and puts half its
// mana value (rounded up) in +1/+1 counters on the source. An on-disk key:
// never renamed, never reused.
var s82OldB40InstantOrSorceryPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b40-exile-instant-or-sorcery-for-counters", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	card, ok := g.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	mv, _ := g.ManaValueForEffect(card)
	ctx := KeyedContext(g, r.Chooser, r.Source)
	if err := (ExileTarget{Target: card.InstanceID}).Apply(ctx); err != nil {
		return err
	}
	n := S82OldB40CeilHalfManaValue(mv)
	if n <= 0 || g.FindCardZoneForEffect(r.Source) == nil {
		return nil
	}
	if z := g.FindCardZoneForEffect(r.Source); z.Kind != game.ZoneBattlefield {
		return nil
	}
	return AddCounter{Target: r.Source, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
})

// S82OldB40LoremageCounter is the keyed tail of Oriq Loremage's search:
// "if it's an instant or sorcery card, put a +1/+1 counter on this
// creature" (the card went to the graveyard; the creature may be gone).
var S82OldB40LoremageCounter = CardTail("oriq-loremage/instant-or-sorcery-counter", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	found, ok := ctx.Game.LookupCardForEffect(o.IDs[0])
	if !ok || !(found.IsInstant() || found.IsSorcery()) {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(o.Tail.Source); z == nil || z.Kind != game.ZoneBattlefield {
		return nil
	}
	return AddCounter{Target: o.Tail.Source, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
})
