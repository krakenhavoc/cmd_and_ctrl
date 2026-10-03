package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mana_spent_entry.go — the catalog side of ADR 0109 §11 (#1552): a
// permanent that reads the mana spent to cast ANOTHER object, or reads
// its own as it enters. Append-only, like every mechanic-named helper
// file.
//
// The engine half is game.Game.EntrySpentForEffect, the spend record
// handed to a replacement inside the CR 614 entry window, and the same
// game.ManaSpent view every other reader of that record uses (CR 400.7d).
// So "mana from an artifact source", "mana from a Treasure" and "no
// mana was spent" are one question each, answered in game/mana_spent.go.
//
// Every reader here inherits ADR 0068 §3: a payment the engine waived
// (strict mana off) is UNKNOWN, so it counts no mana from any source and
// is never "no mana was spent". The cards say so in their caveats.

// CreaturesYouControlEnterWithCountersPerManaFrom is "each creature you
// control enters with an additional +1/+1 counter on it for each mana
// from <a source of kinds> spent to cast it" — Coin of Mastery's
// artifact sources, and with `others` Kalain, Reclusive Painter's "other
// creatures you control … from a Treasure". A CR 614.1c replacement on
// the source, applying to another object's entry, read off the entering
// spell's payment. A creature that was not cast entered with nothing
// spent, so it gets no counter.
func CreaturesYouControlEnterWithCountersPerManaFrom(label string, kinds game.ManaSourceKinds, others bool) game.ReplacementEffect {
	count := func(ev *game.ReplacementEvent, g *game.Game) int {
		return g.EntrySpentForEffect(ev).CountFrom(kinds)
	}
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if ev.Kind != game.RepEventMove || ev.NewZone != game.ZoneBattlefield || src == nil {
				return false
			}
			if others && ev.CardID == src.InstanceID {
				return false
			}
			entering, ok := g.LookupCardForEffect(ev.CardID)
			if !ok || entering.Controller != src.Controller || !entering.IsCreature() {
				return false
			}
			return count(ev, g) > 0
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(game.CounterPlusOne, count(ev, g))
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: label,
	}
}

// SelfEntersWithCountersIfNoManaSpent is "this creature enters with <n>
// <kind> counters on it if it wasn't cast or no mana was spent to cast
// it" (Freestrider Commando). One question, because an entry that was
// not a cast spent nothing that is known: EntrySpentForEffect's zero
// answers None. A plotted Commando cast for free gets them; a waived
// payment (strict mana off) is unknown and does not.
func SelfEntersWithCountersIfNoManaSpent(kind string, n int) game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventZoneMove},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			return ev.Kind == game.RepEventMove &&
				ev.NewZone == game.ZoneBattlefield &&
				src != nil && ev.CardID == src.InstanceID &&
				g.EntrySpentForEffect(ev).None()
		},
		Replace: func(ev *game.ReplacementEvent, _ *game.Game, _ *game.Card) error {
			ev.AddCounterAtETB(kind, n)
			return nil
		},
	}
}

// noneCastWithManaInBatch is Satoru, the Infiltrator's intervening "if"
// (CR 603.4) over the whole set: "if none of them were cast or no mana
// was spent to cast them", where "them" is every nontoken creature that
// entered under `controller`'s control in the event batch `batch`
// (CR 603.2c). It is false when any of them was cast with mana spent.
//
// A creature still on the battlefield answers from its provenance
// (CR 400.7d): not cast, or cast with a KNOWN nothing spent. A waived
// payment is unknown and counts as mana spent, the weaker answer. One
// that has left the battlefield since has lost that record, so it counts
// as mana spent if it entered from the stack — weaker again, never
// stronger.
func noneCastWithManaInBatch(g *game.Game, batch uint64, controller uuid.UUID) bool {
	events := g.EventsThisTurn()
	fromStack := map[uuid.UUID]bool{}
	for _, ev := range events {
		if ev.Batch == batch && ev.Kind == game.EventZoneMove && ev.OldZone == game.ZoneStack && ev.NewZone == game.ZoneBattlefield {
			fromStack[ev.CardID] = true
		}
	}
	for _, ev := range events {
		if ev.Batch != batch || ev.Kind != game.EventETB || ev.CardID == uuid.Nil {
			continue
		}
		c, ok := g.LookupCardForEffect(ev.CardID)
		if !ok || !c.IsCreature() || IsToken(c) {
			continue
		}
		if !onBattlefield(g, ev.CardID) {
			if c.Owner == controller && fromStack[ev.CardID] {
				return false
			}
			continue
		}
		if c.Controller != controller {
			continue
		}
		cast := c.Provenance.FromZone != "" || fromStack[ev.CardID]
		if cast && !c.Provenance.Spent().None() {
			return false
		}
	}
	return true
}
