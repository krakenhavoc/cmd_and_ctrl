package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b12_helpers.go — the shared pieces of the Phyrexia: All Will Be
// One oil cards of S82 batch 12 (#5479). Oil counters are plain named
// counters ("oil"), as Vat of Rebirth and Mindsplice Apparatus use them.

const s82OneB12Oil = "oil"

// s82OneB12EntersWithOil is "This <permanent> enters with N oil counters on
// it" (CR 614.1c): a self-replacement on the card's own entry, so the
// counters go through the counter window.
func s82OneB12EntersWithOil(n int, label string) game.ReplacementEffect {
	r := b10EntersWithCounters(s82OneB12Oil, n, label)
	r.RowLabel = label
	return r
}

// s82OneB12OilUpkeep is "At the beginning of your upkeep, remove an oil
// counter from this creature. Then if it has no oil counters on it, <ifNone>"
// (Archfiend of the Dross, Evolved Spinoderm). The counter is removed as the
// trigger resolves, and "no oil counters" is read after that, so a creature
// whose counters were spent in response still triggers its penalty. A source
// that left, or left and came back as a new object, is untouched (CR 400.7).
func s82OneB12OilUpkeep(label string, ifNone func(g *game.Game, item *game.StackItem) error) game.TriggeredAbility {
	return AtYourUpkeep(label, func(g *game.Game, item *game.StackItem) error {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
			return nil
		}
		if c.Counters[s82OneB12Oil] > 0 {
			if err := (AddCounter{Target: item.SourceCardID, Kind: s82OneB12Oil, N: -1}).Apply(NewContext(g, item)); err != nil {
				return err
			}
		}
		now, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !onBattlefield(g, item.SourceCardID) || now.Counters[s82OneB12Oil] > 0 {
			return nil
		}
		return ifNone(g, item)
	})
}

// s82OneB12GrantsWhileOil is a layer-6 keyword grant to the source "as long
// as it has <cond> oil counters on it".
func s82OneB12GrantsWhileOil(keyword, label string, cond func(oil int) bool) game.StaticAbility {
	return game.StaticAbility{
		Label: label,
		Layer: game.Layer6Ability,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && cond(target.Counters[s82OneB12Oil])
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			appendKeywordsTo(c, []string{keyword})
		},
	}
}

// s82OneB12SourceOil is the oil on the ability's source as it last existed
// (CR 608.2h): on the battlefield now, or its last-known information when a
// cost sacrificed it.
func s82OneB12SourceOil(ctx *Context) int {
	if info, ok := ctx.SourcePermanent(); ok {
		return info.Counters[s82OneB12Oil]
	}
	return 0
}

// s82OneB12SacrificeAtEnd schedules "Sacrifice it at the beginning of the
// next end step" for the tokens just made.
func s82OneB12SacrificeAtEnd(ctx *Context, label string, made []uuid.UUID) error {
	if len(made) == 0 {
		return nil
	}
	return ScheduleDelayedTrigger{
		At:         game.StepEnd,
		Controller: ctx.Controller(),
		Label:      label,
		Cards:      made,
		Body:       sacrificeListedCardsBody,
	}.Apply(ctx)
}

// s82OneB12OilEventsThisTurn is the Churning Reservoir gate (it needs a
// per-turn fact the engine keeps only as the event log): "an oil counter was
// removed from a permanent you controlled this turn or a permanent with an
// oil counter on it was put into a graveyard this turn".
//
// A removal is an EventCounterPlaced for oil whose new total is below the
// total the same permanent had just before (CounterCountBeforeForEffect),
// and the permanent is judged by its controller then. A permanent put into
// a graveyard is an EventLTB to the graveyard; its oil is the last total the
// log recorded for it before it left (counters cease to exist as it leaves,
// CR 122.2, with no removal event).
func s82OneB12OilEventsThisTurn(g *game.Game, controller uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		switch ev.Kind {
		case game.EventCounterPlaced:
			if ev.Label != s82OneB12Oil || g.CounterCountBeforeForEffect(ev) <= ev.Amount {
				continue
			}
			if s82OneB12ControllerOf(g, ev.Target) == controller {
				return true
			}
		case game.EventLTB:
			if ev.NewZone != game.ZoneGraveyard {
				continue
			}
			probe := game.Event{Kind: game.EventCounterPlaced, Target: ev.CardID, Label: s82OneB12Oil, Seq: s82OneB12DepartureSeq(g, ev)}
			if g.CounterCountBeforeForEffect(probe) > 0 {
				return true
			}
		}
	}
	return false
}

// s82OneB12DepartureSeq is the sequence number at which the permanent an
// EventLTB names started to leave: its move to the graveyard is logged just
// before the LTB, and a count walk stops at that move (CR 400.7), so the oil
// it had is the last total recorded before it.
func s82OneB12DepartureSeq(g *game.Game, ltb game.Event) uint64 {
	evs := g.EventsThisTurn()
	for i := len(evs) - 1; i >= 0; i-- {
		if evs[i].Seq <= ltb.Seq && evs[i].Kind == game.EventZoneMove && evs[i].CardID == ltb.CardID &&
			evs[i].NewZone != game.ZoneBattlefield {
			return evs[i].Seq
		}
	}
	return ltb.Seq
}

// s82OneB12ControllerOf is who controls the permanent now, or controlled it
// as it left (a permanent whose oil was removed and that then died).
func s82OneB12ControllerOf(g *game.Game, id uuid.UUID) uuid.UUID {
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return uuid.Nil
	}
	if onBattlefield(g, id) {
		return c.Controller
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == game.EventLTB && ev.CardID == id {
			return leftUnderControlOf(ev, c)
		}
	}
	return c.Controller
}

// s82OneB12Token builds an X-sized token template.
func s82OneB12Token(name, typeLine string, power, toughness int, colors, keywords []string) game.Card {
	if power < 0 {
		power = 0
	}
	if toughness < 0 {
		toughness = 0
	}
	return game.Card{
		Name:      name,
		TypeLine:  typeLine,
		Power:     power,
		Toughness: toughness,
		Colors:    colors,
		Keywords:  keywords,
	}
}
