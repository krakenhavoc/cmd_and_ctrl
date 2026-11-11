package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_b25_helpers.go — shared bodies for the -1/-1 counter placers of
// Commander wave 2, batch 25 (#5368). Append-only.

// cmdr2B25CountersOnTarget is "put N -1/-1 counters on target creature":
// the first target, if it is still legal as the ability resolves
// (CR 608.2b).
func cmdr2B25CountersOnTarget(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
			return nil
		}
		return AddCounter{Target: item.Targets[0].ID, Kind: game.CounterMinusOne, N: n}.Apply(ctx)
	}
}

// cmdr2B25CountersOnEach is "put a -1/-1 counter on each creature" that
// `match` admits (nil admits every creature), except the source when
// others is set ("each other creature"). The set is read before the first
// counter goes on, so a creature is counted once, and each is a member of
// a group (CR 608.2h).
func cmdr2B25CountersOnEach(others bool, match CardPredicate) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		var ids []uuid.UUID
		for _, c := range g.BattlefieldCardsForEffect() {
			if !c.IsCreature() || (others && c.InstanceID == item.SourceCardID) {
				continue
			}
			if match != nil && !match(g, item.Controller, c) {
				continue
			}
			ids = append(ids, c.InstanceID)
		}
		for _, id := range ids {
			if !onBattlefield(g, id) {
				continue
			}
			if err := (AddCounter{Target: id, Kind: game.CounterMinusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
				return err
			}
		}
		return nil
	}
}

// --- batch 26: -1/-1 counter payoffs (#5368) ------------------------

// cmdr2B25MinusCountersPlacedOnACreature is "one or more -1/-1 counters are
// put on a creature": a counter-placement event that raised a creature's
// -1/-1 total (b33CountersPlacedDelta), whoever put them. One event is one
// trigger however many counters it placed.
func cmdr2B25MinusCountersPlacedOnACreature(ev game.Event, g *game.Game) bool {
	if b33CountersPlacedDelta(ev, game.CounterMinusOne, g) <= 0 {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.Target)
	return ok && c.IsCreature() && onBattlefield(g, ev.Target)
}

// cmdr2B25CountersOnEachTarget is "put a -1/-1 counter on each of up to N
// target creatures": every target still legal as the ability resolves
// (CR 608.2b), one counter each.
func cmdr2B25CountersOnEachTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// cmdr2B25CounterTotal is the number of counters of every kind in a map,
// the "number of counters on that creature" of a last-known read.
func cmdr2B25CounterTotal(counters map[string]int) int {
	n := 0
	for _, c := range counters {
		if c > 0 {
			n += c
		}
	}
	return n
}

// cmdr2B25CreatureDiedWithACounter is "a creature with a counter on it
// dies": a creature went to the graveyard having had a counter of any kind
// (CR 603.10a), read from the departure record.
func cmdr2B25CreatureDiedWithACounter(ev game.Event, g *game.Game) bool {
	if _, ok := diedCreature(ev, g); !ok {
		return false
	}
	return cmdr2B25CounterTotal(g.LastKnownCountersForEffect(ev.CardID)) > 0
}

// cmdr2B25OpponentCreatureDiedWithACounter is cmdr2B25CreatureDiedWithACounter
// narrowed to a creature the source's controller's opponent controlled as it
// died (leftUnderControlOf, CR 603.10a).
func cmdr2B25OpponentCreatureDiedWithACounter(ev game.Event, source *game.Card, g *game.Game, minusOnly bool) bool {
	dead, ok := diedCreature(ev, g)
	if !ok {
		return false
	}
	if controller := leftUnderControlOf(ev, dead); controller == source.Controller || g.PlayerByIDForEffect(controller) == nil {
		return false
	}
	counters := g.LastKnownCountersForEffect(ev.CardID)
	if minusOnly {
		return counters[game.CounterMinusOne] > 0
	}
	return cmdr2B25CounterTotal(counters) > 0
}

// cmdr2B25PutLandFromHandOrGraveyardTapped is "you may put a land card from
// your hand or graveyard onto the battlefield tapped": the pick over the
// lands in both zones, the chosen card entering from whichever holds it
// when the answer arrives (Bonny Pall's shape, entering tapped).
func cmdr2B25PutLandFromHandOrGraveyardTapped(source string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		p := g.PlayerByIDForEffect(item.Controller)
		if p == nil {
			return nil
		}
		var ids []uuid.UUID
		for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
			if z == nil {
				continue
			}
			for _, c := range z.Cards {
				if c.IsLand() && !c.IsToken() {
					ids = append(ids, c.InstanceID)
				}
			}
		}
		if len(ids) == 0 {
			return nil
		}
		g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  item.Controller,
			Source:   item.SourceCardID,
			Question: source + " — you may put a land card from your hand or graveyard onto the battlefield tapped",
			Cards:    ids,
			Min:      0,
			Max:      1,
			ThenKey:  cmdr2B25LandPicked,
		})
		return nil
	}
}

// cmdr2B25LandPicked puts the picked land onto the battlefield tapped from
// the zone that holds it now. An on-disk key: never renamed, never reused.
var cmdr2B25LandPicked = game.RegisterChooseCardsThen("choose-cards/cmdr2-b25-land-from-hand-or-graveyard-tapped", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	id := r.Picked[0]
	z := g.FindCardZoneForEffect(id)
	if z == nil {
		return nil
	}
	switch z.Kind {
	case game.ZoneHand:
		return g.PutFromHandOntoBattlefieldKeyedForEffect(id, game.HandEntryOptions{Controller: r.Chooser, Tapped: true}, game.KeyedTail{})
	case game.ZoneGraveyard:
		return g.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, r.Chooser, true, game.KeyedTail{})
	}
	return nil
})

// itoaB25 renders a count for a label.
func itoaB25(n int) string { return strconv.Itoa(n) }
