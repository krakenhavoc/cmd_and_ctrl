package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// rooms_b_helpers.go — shared bodies for the ADR 0103 Rooms whose doors
// need more than an existing primitive (#1756). Append-only.

// roomsBCardPutIntoYourGraveyardFromLibrary is "a card is put into your
// graveyard from your library" (Polluted Cistern): a mill or any other
// library-to-graveyard move of a card its controller owns. A card's
// graveyard is always its owner's, so ownership reads the destination.
func roomsBCardPutIntoYourGraveyardFromLibrary(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventMill && ev.Kind != game.EventZoneMove {
		return false
	}
	if ev.OldZone != game.ZoneLibrary || ev.NewZone != game.ZoneGraveyard || ev.CardID == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Owner == source.Controller
}

// roomsBEachOpponentLosesLifePerCardTypeMilled is Polluted Cistern's
// effect: each opponent loses 1 life for each card type among the cards
// that were put into the controller's graveyard from their library
// together with the card that triggered it (the same event batch,
// CR 603.2c). The types are each card's printed ones; a card that has
// left the graveyard since is still counted, as "those cards".
func roomsBEachOpponentLosesLifePerCardTypeMilled(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	batch := item.Trigger.Event.Batch
	types := map[string]bool{}
	for _, ev := range g.EventsThisTurn() {
		if ev.Batch != batch || !roomsBCardPutIntoYourGraveyardFromLibrary(ev, &game.Card{Controller: item.Controller}, game.Characteristic{}, g) {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok {
			continue
		}
		_, ts, _ := game.ParseTypeLine(c.TypeLine)
		for _, t := range ts {
			types[t] = true
		}
	}
	if len(types) == 0 {
		return nil
	}
	return eachOpponentLosesLife(g, item, len(types))
}

// roomsBMillThenReturnAChosenCreatureCard is Dim Oubliette: mill N
// cards, then the controller chooses a creature card from their
// graveyard (the milled ones included) and puts it onto the battlefield.
// The pick is mandatory; with one candidate there is nothing to choose.
func roomsBMillThenReturnAChosenCreatureCard(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return MillToZone{N: n, Then: func(ctx *Context, _ []uuid.UUID) error {
			return roomsBReturnAChosenCreatureCard(ctx)
		}}.Apply(NewContext(g, item))
	}
}

func roomsBReturnAChosenCreatureCard(ctx *Context) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil || p.Graveyard == nil {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range p.Graveyard.Cards {
		if c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return ReturnFromGraveyard{Target: candidates[0], Dest: game.ZoneBattlefield}.Apply(ctx)
	}
	item := ctx.Item
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Return a creature card from your graveyard to the battlefield",
		Cards:    candidates,
		Min:      1,
		Max:      1,
		Zone:     game.ZoneGraveyard,
		Then: func(g *game.Game, picked []uuid.UUID) error {
			if len(picked) == 0 {
				return nil
			}
			return ReturnFromGraveyard{Target: picked[0], Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
		},
	})
	return nil
}
