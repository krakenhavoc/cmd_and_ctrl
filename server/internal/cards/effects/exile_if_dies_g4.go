package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_if_dies_g4.go — shared reads for ADR 0108 PR 1's group 4 cards
// (#1886): Cry of the Carnarium's "creature cards in all graveyards that
// were put there from the battlefield this turn", and The War Doctor's
// "one or more other cards are put into exile from anywhere".

// putIntoGraveyardFromBattlefieldThisTurn reports whether `cardID`'s
// most recent arrival in a graveyard happened this turn and came from
// the battlefield. It reads g.EventsThisTurn(), as the readers in
// graveyard_provenance.go do. One arrival is announced by several
// events (the zone move and the leaves-the-battlefield event), and the
// arrival counts if any of them names the battlefield as the old zone.
// A later move out of the graveyard ends the arrival (CR 400.7: a card
// that left the graveyard and came back is a new object, and only its
// newest arrival counts).
//
// Caller holds g.mu.
func putIntoGraveyardFromBattlefieldThisTurn(g *game.Game, cardID uuid.UUID) bool {
	arrived, fromBattlefield := false, false
	for _, ev := range g.EventsThisTurn() {
		if ev.CardID != cardID {
			continue
		}
		leaving := ev.OldZone == game.ZoneGraveyard && ev.NewZone != "" && ev.NewZone != game.ZoneGraveyard
		if leaving {
			arrived, fromBattlefield = false, false
			continue
		}
		if ev.NewZone != game.ZoneGraveyard || ev.OldZone == game.ZoneGraveyard {
			continue
		}
		if !arrived {
			arrived, fromBattlefield = true, false
		}
		fromBattlefield = fromBattlefield || ev.OldZone == game.ZoneBattlefield
	}
	return arrived && fromBattlefield
}

// exileCreatureCardsPutIntoGraveyardsFromBattlefieldThisTurn is Cry of
// the Carnarium's "Exile all creature cards in all graveyards that were
// put there from the battlefield this turn." A token is not a card
// (CR 108.2). "Creature card" is the card as it is in the graveyard,
// its printed type line. The set is read before the first card is
// exiled.
func exileCreatureCardsPutIntoGraveyardsFromBattlefieldThisTurn(ctx *Context) error {
	var ids []uuid.UUID
	for _, p := range ctx.Game.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsCreature() && !c.IsToken() && putIntoGraveyardFromBattlefieldThisTurn(ctx.Game, c.InstanceID) {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	for _, id := range ids {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// otherCardPutIntoExile is The War Doctor's "one or more other cards are
// put into exile from anywhere", for one event. "From anywhere" is never
// a leaves-the-battlefield trigger (CR 603.6c), so it is judged on the
// board after the move. Three event shapes put a card into exile:
//
//   - EventZoneMove, from any zone, including the stack and the
//     battlefield;
//   - EventDiscardCard, when a replacement exiled the discarded card;
//   - EventCounterSpell, which carries no destination: a spell countered
//     into exile (Dissipate, or a flashed-back spell) is a card now in
//     exile. An ability being countered names its own ID as Target, not
//     the card's, so it is not counted.
//
// A token is not a card (CR 108.2), and the ability's own source is not
// "other".
func otherCardPutIntoExile(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
		return false
	}
	switch ev.Kind {
	case game.EventZoneMove, game.EventDiscardCard:
		if ev.NewZone != game.ZoneExile || ev.OldZone == game.ZoneExile {
			return false
		}
	case game.EventCounterSpell:
		if ev.Target != ev.CardID {
			return false
		}
		if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneExile {
			return false
		}
	default:
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !c.IsToken()
}
