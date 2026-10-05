package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// entry_counters.go — the catalog's declarations for "this permanent
// enters with N counters on it", where N is read from the spell that
// became it (CR 614.1c, #1002).
//
// Three families of "enters with counters" live in this catalog, and
// they are three different slots because they read three different
// things:
//
//	b10EntersWithCounters      a PRINTED number      Replacements
//	b19EntersWithCountersCounted  a count off the BOARD  Replacements
//	this file                  the ANNOUNCEMENT      EntersWithCountersFromCast
//
// The first two are ordinary CR 614 self-replacements: everything they
// need is reachable from (g, src) while the entry window is open. The
// third is not — "X" and "each colour of mana spent to cast it" are
// facts about the cast, and a replacement is handed the game, the
// source and the event, none of which carries the resolving stack
// item. That is why sunburst and thirteen X-creatures used to put
// their counters on in OnResolve instead, a beat before the permanent
// existed, and each declared a caveat saying so.
//
// The engine now seeds these clauses onto the entry event from the
// StackItem that is right there (game/entry_counters.go), so a card
// file declares the arithmetic and nothing else. The constructors
// below are the whole vocabulary; a card never builds a
// game.EntryCountersFromCast by hand, exactly as mana_spent.go's rule
// is that a card never reaches into game.PaidCost by hand.

// XCounters is CR 614.1c's commonest spelling: "this permanent enters
// with X <kind> counters on it", the announced X and nothing else.
//
// Cast for X=0 it puts no counters on, so a printed 0/0 enters as the
// 0/0 it is and CR 704.5f puts it into its owner's graveyard at the
// next state-based check (#691, game.Card.ToughnessIsKnown). That is
// the printed behaviour, and it is why the clause must NOT floor at
// one.
func XCounters(kind string) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return cast.X },
	}
}

// CountersPerKick is "this permanent enters with `per` <kind> counters
// on it for each time it was kicked" (CR 702.33d) — Everflowing
// Chalice's charge counter per multikick. Kicker and multikicker
// count together, because no rules text tells them apart.
//
// A single kicker is the same clause paid at most once: Vodalian
// Serpent's "If this creature was kicked, it enters with four +1/+1
// counters on it" is CountersPerKick(game.CounterPlusOne, 4).
func CountersPerKick(kind string, per int) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return cast.Kicked * per },
	}
}

// CountersIfKickedWith is "if this creature was kicked with its
// <cost> kicker, it enters with <n> <kind> counters on it" (CR
// 702.33f, #2360) — the Volvers. `cost` is spelled as the card's
// Kickers declaration spells it. A clause per kicker, so a Volver
// kicked with both gets both.
func CountersIfKickedWith(kind, cost string, n int) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: kind,
		Count: func(cast game.CastCounts) int {
			if cast.KickedWith(cost) {
				return n
			}
			return 0
		},
	}
}

// CountersPerDelved is "this permanent enters with a <kind> counter on
// it for each <kind of> card exiled with it", where "exiled with it"
// is CR 607.2q's link to the cards delve exiled to pay for the spell
// (ADR 0100 sub-PR 2) — Murktide Regent's "a +1/+1 counter on it for
// each instant and sorcery card exiled with it".
//
// `match` is asked of each linked card as it sits in exile (its
// printed characteristics: a card off the battlefield has no layered
// ones), and nil counts every one. A card that left exile before the
// permanent entered is not asked at all: it is a new object (CR 400.7)
// and CastCounts.Delved no longer holds it.
func CountersPerDelved(kind string, match func(game.Card) bool) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind: kind,
		Count: func(cast game.CastCounts) int {
			n := 0
			for _, c := range cast.Delved {
				if match == nil || match(c) {
					n++
				}
			}
			return n
		},
	}
}
