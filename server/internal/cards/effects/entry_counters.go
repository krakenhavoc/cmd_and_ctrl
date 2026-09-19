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
// Unused by the catalog today; it is here because it is the second
// half of what game.CastCounts exists to express, and a card that
// prints the clause should find the declaration already written
// rather than reach for the payment record itself.
func CountersPerKick(kind string, per int) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return cast.Kicked * per },
	}
}

// SunburstCounters is sunburst (CR 702.44a): "this permanent enters
// with a +1/+1 counter on it for each color of mana spent to cast it"
// — a +1/+1 counter for a creature, a charge counter for a
// non-creature artifact (CR 702.44b).
//
// DECLARED SIMPLIFICATION, and it is the only one left: a payment the
// engine did not take (permissive mode, a strict-mode override) claims
// no colours, so the permanent enters with no counters at all. That is
// ADR 0068 §3's rule for every reader of the paid-cost record —
// unknown answers weaker than printed — and a card that declares this
// says so in its Caveats.
//
// Pair it with Spec.WantsDistinctColors, which is what makes the cast
// gate spread the payment across colours rather than paying
// colourless-first; without that a sunburst creature cast off five
// lands into a wide pool would routinely enter smaller than it should.
func SunburstCounters(kind string) game.EntryCountersFromCast {
	return game.EntryCountersFromCast{
		Kind:  kind,
		Count: func(cast game.CastCounts) int { return cast.ColorsSpent },
	}
}
