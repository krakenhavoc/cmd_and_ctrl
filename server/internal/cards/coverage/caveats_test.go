package coverage

import (
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// TestNoCaveatContradictsItsOwnCard is the hard guard.
//
// A caveat here names a mechanic that the very same Spec declares.
// The card does the thing; the catalog page tells the player it does
// not. That is not a stale note, it is a false statement shipped to a
// reader, and there is no allowlist for it.
//
// This is the shape that bit four cards in one September session —
// faithless_looting still saying "Flashback isn't implemented" after
// #411 added Flashback("{2}{R}") to its own Spec, weftstalker_ardent
// saying the same about warp after #413. In every case the mechanic
// and the caveat were in the same file, four lines apart.
func TestNoCaveatContradictsItsOwnCard(t *testing.T) {
	for _, f := range Contradictions() {
		fix := `Delete the caveat. If nothing else on the card is
  simplified, the whole declaration becomes:

      Completeness: CompletenessFull,

  and Register will reject leftover Caveats text, so the compiler and
  the boot check finish the job for you.`
		if f.Mechanic.Confidence == Heuristic {
			fix = `Either delete the caveat (the card now does the thing), or
  TIGHTEN THE PROBE in coverage/caveats.go — a heuristic probe reads a
  proxy, and this may be the false positive its comment warns about.
  Do not add an allowlist; a probe that cannot tell the difference
  should be made able to, or removed.`
		}
		t.Errorf(`a caveat contradicts the card it is printed on.

%s
The card declares the mechanic and the caveat denies it. The catalog
page publishes Spec.Caveats verbatim, so this text is on screen for
a player right now.

  %s`, f.Describe(), fix)
	}
}

// adoptableGaps pins the caveats that name a mechanic the ENGINE has
// and this card has not taken up. Each one is honest about its card —
// the clause really does not happen — and expired in its reasoning:
// the machinery exists, another card is already using it, and closing
// the gap is a small edit.
//
// It is a pinned table rather than a required-empty list for the same
// reason snapshot_drift_test classifies fields instead of banning
// them: the PR that lands a mechanic cannot be made responsible for
// adopting it on every card that ever mentioned it, and a check that
// demanded so would be turned off. What the pin buys is that the set
// cannot change in either direction without somebody saying why.
//
// Key is Finding.Key() — "<card name> / <mechanic>".
var adoptableGaps = map[string]string{
	"Deadly Rollick / free cast": `#428 landed the conditional free cast and used it on
		Fierce Guardianship only. Deadly Rollick and the rest of the
		Commander Legends cycle are two lines each
		(FreeIfYouControlCommander), but each needs its own free-cast
		test, so they are their own piece of work rather than a
		rider on the drift guard.`,
}

// TestAdoptableGapsArePinned fails on any movement in that set.
func TestAdoptableGapsArePinned(t *testing.T) {
	live := map[string]Finding{}
	for _, f := range AdoptableGaps() {
		live[f.Key()] = f
	}

	for key, f := range live {
		if _, ok := adoptableGaps[key]; ok {
			continue
		}
		t.Errorf(`a caveat names a mechanic the engine now has.

%s  already implemented by: %s

The caveat is true of this card and out of date about the engine: the
machinery shipped, another catalog card is using it, and this one is
a small edit away.

  adopt it:  %s

If adopting it belongs to a different piece of work, pin it in
adoptableGaps in coverage/caveats_test.go with a one-line reason —
the pin is what keeps the gap visible instead of silent.`,
			f.Describe(), f.ExampleFor, f.Mechanic.Adopt)
	}

	var stale []string
	for key := range adoptableGaps {
		if _, ok := live[key]; !ok {
			stale = append(stale, key)
		}
	}
	sort.Strings(stale)
	for _, key := range stale {
		t.Errorf(`adoptableGaps pins %q, which no longer applies.

Either the card adopted the mechanic (delete the pin and, while you
are there, the caveat), or the card or mechanic was renamed. A pin
that matches nothing is a note nobody will ever read again.`, key)
	}
}

// TestEveryExactMechanicHasAnImplementor guards the guards.
//
// Every exact probe asks game.AlternativeCostByKey for a keyword that
// some catalog card really offers today. If the key string is
// mistyped, renamed by a later sprint, or the CatalogAlternativeCosts
// hook stops being wired at init, the probe does not fail — it
// quietly answers "no" for every card forever, and both caveat guards
// go silent while still reporting PASS.
//
// A silent guard is the failure mode this package was written to stop
// happening to documentation, so it is not allowed to happen to the
// package itself.
func TestEveryExactMechanicHasAnImplementor(t *testing.T) {
	for _, m := range Mechanics() {
		if m.Confidence != Exact {
			continue
		}
		if example, ok := EngineImplements(m); ok {
			t.Logf("%s: implemented by %s", m.Name, example)
			continue
		}
		t.Errorf(`no registered card satisfies the %q probe.

  probe: %s

Either the probe is broken (a renamed key, an unwired
CatalogAlternativeCosts hook) — in which case both caveat guards are
currently passing on everything and checking nothing — or the last
card using this mechanic left the catalog, in which case delete the
row from mechanics in coverage/caveats.go.`, m.Name, m.Evidence)
	}
}

// TestMechanicPhrasesAreDistinct keeps the table from growing a
// phrase that pulls two mechanics onto the same caveat. A caveat
// matched by two rows produces two failures for one problem and, more
// importantly, means one of the two attributions is wrong.
func TestMechanicPhrasesAreDistinct(t *testing.T) {
	for _, s := range effects.All() {
		for _, cv := range s.Caveats {
			var hit []string
			for _, m := range Mechanics() {
				if m.names(cv) {
					hit = append(hit, m.Name)
				}
			}
			if len(hit) > 1 {
				t.Errorf(`%s: one caveat names %s.

  caveat: %q

Two rows of the mechanics table claim the same sentence, so at least
one of them is about to report the wrong mechanic with full
confidence. Narrow the phrase list in coverage/caveats.go.`,
					s.Name, strings.Join(hit, " and "), cv)
			}
		}
	}
}
