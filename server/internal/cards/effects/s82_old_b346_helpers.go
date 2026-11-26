package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB346GemCounter is the gem counter (Briber's Purse): a generic resource
// counter with no rule of its own (CR 122.1).
const OldB346GemCounter = "gem"

// OldB346CardExiledFromYourGraveyard is one zone-move event that took a card
// the source's controller owns out of a graveyard and into exile (Rakshasa
// Vizier). A token is not a card (CR 108.2), and a card cast from the
// graveyard goes to the stack, not exile.
func OldB346CardExiledFromYourGraveyard(ev game.Event, source *game.Card, g *game.Game) bool {
	if source == nil || ev.Kind != game.EventZoneMove || ev.CardID == uuid.Nil ||
		ev.OldZone != game.ZoneGraveyard || ev.NewZone != game.ZoneExile {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !IsToken(c) && c.Owner == source.Controller
}

// OldB346CardsExiledFromYourGraveyard counts the cards of the trigger's batch
// that were put into exile from the controller's graveyard: "that many".
func OldB346CardsExiledFromYourGraveyard(g *game.Game, item *game.StackItem) int {
	if item.Trigger == nil {
		return 0
	}
	src, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok {
		return 0
	}
	batch := item.Trigger.Event.Batch
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Batch == batch && OldB346CardExiledFromYourGraveyard(ev, &src, g) {
			n++
		}
	}
	return n
}

// OldB346TwoCountersAndUntapTheTarget is "Put two +1/+1 counters on target
// creature. Untap it." as a spell's whole resolution (Soldier On, Dragonscale
// Boon): the counters, then the untap, on the one target still legal as the
// spell resolves (CR 608.2b).
func OldB346TwoCountersAndUntapTheTarget(_ *game.StackItem, ctx *Context) error {
	id, ok := B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return UntapTarget{Target: id}.Apply(ctx)
}
