package decks

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
)

// coverage.go answers the question a player actually asks of a
// pre-built deck: "are all its cards implemented?"
//
// decks_test.go answers a narrower one — every non-basic card resolves
// to a REGISTERED effects.Spec — and that is the floor this catalog is
// built on. It is not the same promise. effects.Completeness grades a
// registered card as `full`, `caveats` or `unreviewed`, and a deck of
// registered-but-caveated cards is not what "all their cards are
// implemented" means to the person picking it. The catalog page has
// drawn that distinction since it shipped; the deck picker draws the
// same one from the same numbers.
//
// The computation is catalog.Build's, not a second one. That matters
// more than it looks: completeness is merged across faces (one
// caveated half makes the card a caveat card), and an MDFC whose front
// face has no spec earns a caveat no card file declares. Recomputing
// that here would eventually disagree with /catalog about the same
// card, and the two numbers would be shown on the same site.

// CaveatCard is one card in a deck whose printed text the engine does
// not fully carry out, with the clauses it skips in player-facing
// language — exactly the strings catalog.Entry.Caveats carries.
type CaveatCard struct {
	Name string `json:"name"`
	// Caveats is non-empty for a `caveats` card and empty for an
	// `unreviewed` one, which is a card nobody has graded rather than
	// a card with a known gap.
	Caveats []string `json:"caveats,omitempty"`
	// Unreviewed distinguishes the two: "we know what this one skips"
	// from "nobody has checked this one". Both belong in the same
	// list — a player wants one honest count of cards that might not
	// do what they print — but they are not the same claim.
	Unreviewed bool `json:"unreviewed,omitempty"`
}

// Coverage is a deck's engine-coverage profile: how many of its cards
// the engine carries out in full, and which ones it does not.
//
// Counts are per CARD (distinct row), not per copy: a deck's thirty
// Forests are one basic-land row and count once in Basics. Cards +
// Basics is the number of distinct names in the list; Size() is the
// hundred physical cards.
type Coverage struct {
	// Cards is the number of non-basic cards graded — the sum of
	// Full, Caveats and Unreviewed.
	Cards int `json:"cards"`
	// Full is the count whose printed text the engine carries out
	// with no declared simplification.
	Full int `json:"full"`
	// Caveats is the count with at least one declared simplification.
	Caveats int `json:"caveats"`
	// Unreviewed is the count registered but never graded. Nonzero
	// here is a gap in the CATALOG's bookkeeping, not necessarily in
	// the card — say so rather than rounding it into either column.
	Unreviewed int `json:"unreviewed"`
	// Basics is the number of basic-land rows, which are exempt: the
	// engine synthesises a basic's mana ability from its type line
	// (see the package comment), so they are fully played while
	// carrying no catalog entry to grade.
	Basics int `json:"basics"`
	// Unregistered is the count with no effects.Spec at all. Always
	// zero — decks_test.go fails the build otherwise — and published
	// anyway so that a deck which somehow shipped with one is visible
	// in the picker instead of silently miscounted as full.
	Unregistered int `json:"unregistered"`
	// Imperfect lists every card behind the Caveats and Unreviewed
	// counts, by name, so the picker can show WHICH cards rather than
	// only how many. Sorted by name.
	Imperfect []CaveatCard `json:"imperfect,omitempty"`
}

// Complete reports whether every non-basic card in the deck is graded
// `full`. This is the "all their cards are implemented" claim, and it
// is the one the picker is allowed to make.
func (c Coverage) Complete() bool {
	return c.Caveats == 0 && c.Unreviewed == 0 && c.Unregistered == 0
}

// CoverageOf computes d's profile against the live effects registry.
//
// idx is optional and only affects nothing a player sees: without a
// Scryfall dump the completeness grades still come out of the
// registry, which is why this is testable in CI and why a deployment
// with no dump can still answer honestly.
//
// Prefer Coverages when profiling more than one deck — it joins
// against one catalog build instead of one per deck.
func CoverageOf(idx *cards.Index, d Deck) Coverage {
	return coverage(catalogByOracle(idx), d)
}

// Coverages profiles every pre-built deck, keyed by deck ID, against
// a single catalog build. This is what the picker route serves.
func Coverages(idx *cards.Index) map[string]Coverage {
	byOracle := catalogByOracle(idx)
	out := make(map[string]Coverage, len(all))
	for _, d := range all {
		out[d.ID] = coverage(byOracle, d)
	}
	return out
}

func catalogByOracle(idx *cards.Index) map[string]catalog.Entry {
	entries := catalog.Build(idx)
	byOracle := make(map[string]catalog.Entry, len(entries.Cards))
	for _, e := range entries.Cards {
		byOracle[e.OracleID] = e
	}
	return byOracle
}

func coverage(byOracle map[string]catalog.Entry, d Deck) Coverage {
	var cov Coverage
	for _, c := range d.Cards() {
		if c.Basic {
			cov.Basics++
			continue
		}
		cov.Cards++
		e, ok := byOracle[c.OracleID]
		if !ok {
			cov.Unregistered++
			cov.Imperfect = append(cov.Imperfect, CaveatCard{
				Name:       c.Name,
				Unreviewed: true,
				Caveats:    []string{"This card has no implementation in this build; it behaves as a manual sandbox card."},
			})
			continue
		}
		switch e.Completeness {
		case effects.CompletenessFull.String():
			cov.Full++
		case effects.CompletenessCaveats.String():
			cov.Caveats++
			cov.Imperfect = append(cov.Imperfect, CaveatCard{Name: c.Name, Caveats: e.Caveats})
		default:
			cov.Unreviewed++
			cov.Imperfect = append(cov.Imperfect, CaveatCard{Name: c.Name, Unreviewed: true})
		}
	}
	sort.Slice(cov.Imperfect, func(a, b int) bool { return cov.Imperfect[a].Name < cov.Imperfect[b].Name })
	return cov
}
