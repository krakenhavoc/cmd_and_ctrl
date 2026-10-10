package decks

import (
	"sort"
	"strings"
	"testing"
)

// minorCaveats is the owner's short list of cards a curated deck may
// carry although the catalog grades them `caveats` rather than `full`
// (#2436). Each one's missing clause is cosmetic or weaker than
// printed on something its deck does not lean on. The list is closed:
// a new caveat card in a deck is a choice for the owner, not for
// whoever is adding the card, so it fails the test below until it is
// named here in a PR that says why.
var minorCaveats = map[string]string{
	// With strict mana off the scry 1 never happens; it still taps
	// for the commander's colours, which is what it is here for.
	"Path of Ancestry": "a mana land whose scry rider is the missing clause",
	// With strict mana off a spell cast with its mana can still be
	// countered; it is in the Simic deck as a one-drop mana creature.
	"Delighted Halfling": "a mana creature whose \"can't be countered\" rider is the missing clause",
	// Landfall always makes the Treasure, which is what a ramp deck
	// would pick.
	"Tireless Provisioner": "the Food option is never offered",
	// Dash is unavailable; the one-drop's combat trigger is what the
	// Izzet deck plays it for.
	"Ragavan, Nimble Pilferer": "dash is not available",
	// Simultaneous triggers deal their damage as separate hits, which
	// matters only beside a per-hit damage booster the deck lacks.
	"Magmakin Artillerist":  "several discards deal separate 1s",
	"Ingenious Artillerist": "several artifacts entering deal separate hits",
}

// TestEveryCuratedCardIsFull is the owner's card-quality bar for the
// curated decks (#2436): every non-basic card the engine carries out
// in full, except the named minor caveats above. It is stricter than
// TestEveryCardResolvesToARegisteredSpec, which only asks that a card
// is registered: a registered card with a caveat on the clause its
// deck plays it for is a card the bot pays for and does not get.
//
// The tutorial pair is not checked here; its decks are built for the
// walkthrough, not for balance, and tutorial_test.go holds them to
// their own bar.
func TestEveryCuratedCardIsFull(t *testing.T) {
	used := map[string]bool{}
	covs := Coverages(nil)
	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			var bad []string
			for _, c := range covs[d.ID].Imperfect {
				if _, ok := minorCaveats[c.Name]; ok && !c.Unreviewed {
					used[c.Name] = true
					continue
				}
				why := strings.Join(c.Caveats, " / ")
				if c.Unreviewed && why == "" {
					why = "not yet reviewed"
				}
				bad = append(bad, c.Name+": "+why)
			}
			if len(bad) > 0 {
				sort.Strings(bad)
				t.Errorf("%d card(s) are not graded full and are not on the minor-caveats list:\n\t%s\n\n"+
					"Replace the card with one the catalog grades full, or ask the owner to add it to "+
					"minorCaveats with the reason its caveat does not matter to this deck.",
					len(bad), strings.Join(bad, "\n\t"))
			}
		})
	}
	// A name on the list that no deck carries any more is a stale
	// exemption waiting to excuse a card nobody re-read.
	for name := range minorCaveats {
		if !used[name] {
			t.Errorf("minorCaveats names %q, which no curated deck carries as a caveat card; delete the entry", name)
		}
	}
}
