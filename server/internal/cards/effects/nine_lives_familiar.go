package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nine-Lives Familiar — Creature — Cat {1}{B}{B}, 1/1 (Foundations,
// #4465):
//
//	"This creature enters with eight revival counters on it if you cast
//	 it.
//	 When this creature dies, if it had a revival counter on it, return
//	 it to the battlefield with one fewer revival counter on it at the
//	 beginning of the next end step."
//
// The counters are a CR 614.1c clause read off the cast record
// (CountersIfCast), so a Familiar put onto the battlefield any other way
// has none and its death does nothing. The death trigger's intervening
// "if" (CR 603.4) reads the counters the creature had as it left, from
// its last-known information (CR 608.2h). It schedules a delayed trigger
// for the next end step (CR 603.7) naming the card as the object it is in
// the graveyard now and the number of counters to give it back; if the
// card has left the graveyard by then it is a new object and nothing
// returns (CR 400.7). The return is an entry with the counters riding the
// event, so a counter replacement sees them, and under its owner's
// control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                   "5049da27-f38b-4fe6-a387-51510bd1f44e",
		Name:                       "Nine-Lives Familiar",
		Completeness:               CompletenessFull,
		EntersWithCountersFromCast: []game.EntryCountersFromCast{CountersIfCast(revivalCounterName, 8)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, chars game.Characteristic, g *game.Game) bool {
				return ThisDied(ev, source, chars, g) && nineLivesRevivalCountersWhenItLeft(g, ev.CardID) > 0
			},
			Key:    "Nine-Lives Familiar — return it at the beginning of the next end step",
			Effect: nineLivesScheduleReturn,
		}},
	})
}

// revivalCounterName is the counter's name, as the card prints it.
const revivalCounterName = "revival"

// nineLivesRevivalCountersWhenItLeft is how many revival counters the
// creature had as it left the battlefield (last-known information).
func nineLivesRevivalCountersWhenItLeft(g *game.Game, id uuid.UUID) int {
	if info, ok := g.LastKnownPermanentForEffect(id); ok {
		return info.Counters[revivalCounterName]
	}
	return 0
}

// nineLivesReturnBody is the delayed trigger's body: give the card back
// to the battlefield with one fewer revival counter, if it is still the
// object that died, in its owner's graveyard. An on-disk key: never
// renamed, never reused.
var nineLivesReturnBody = game.DelayedBody("nine-lives-familiar/return", nineLivesReturn)

// nineLivesScheduleReturn schedules the next-end-step return, naming the
// card as the object it is in the graveyard and one fewer counter than it
// had.
func nineLivesScheduleReturn(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	return ScheduleDelayedTrigger{
		Label: "Nine-Lives Familiar — return it to the battlefield with one fewer revival counter",
		Cards: []uuid.UUID{id},
		Body:  nineLivesReturnBody,
		Params: game.EffectParams{
			Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch},
			Amount: nineLivesRevivalCountersWhenItLeft(g, id) - 1,
		},
	}.Apply(NewContext(g, item))
}

func nineLivesReturn(g *game.Game, _ *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	var counters map[string]int
	if p.Amount > 0 {
		counters = map[string]int{revivalCounterName: p.Amount}
	}
	_, err := g.ReturnFromGraveyardWithCountersForEffect(id, uuid.Nil, false, counters)
	return nothingIfGone(err)
}
