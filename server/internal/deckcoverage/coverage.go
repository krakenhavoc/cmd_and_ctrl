// Package deckcoverage answers ADR 0095's question about a whole
// decklist: how much of it will the engine automate? It is the one
// place a card's coverage bucket is decided; the public checker
// (POST /deck-coverage), the Discord commands and the filed deck
// request all render the Report built here.
//
// It is its own package rather than a function in internal/deck
// because it reads the catalogue's completeness verdicts through
// internal/catalog, which imports internal/cards/effects — and the
// effects package's own tests import internal/deck, so deck cannot
// import them back.
//
// Every value here is derived from public data (the Scryfall dump and
// this repository). A Report carries card names, oracle IDs, buckets
// and caveat sentences, and never art or oracle text: the line ADR
// 0092 drew for the public roadmap.
package deckcoverage

import (
	"errors"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bucket is what the engine does with one card (ADR 0095 §1).
type Bucket string

const (
	// Manual: printed rules the engine will not run. Exactly
	// game.Unimplemented — the bit the stack overlay's `manual` chip,
	// the upload summary and the bot's improvisation already share.
	Manual Bucket = "manual"
	// Unreviewed: catalogued, but nobody has audited it against its
	// printed text (effects.CompletenessUnreviewed).
	Unreviewed Bucket = "unreviewed"
	// Caveats: catalogued, with declared simplifications.
	Caveats Bucket = "caveats"
	// Automated: catalogued and CompletenessFull.
	Automated Bucket = "automated"
	// NoEffect: nothing to automate — a vanilla creature, a card whose
	// text is keywords the engine enforces, a basic land
	// (!game.NeedsCatalogEffect).
	NoEffect Bucket = "no_effect"
)

// Buckets is every bucket, in report order: the most actionable first.
// Cards in a Report are sorted by this order, then by name.
var Buckets = []Bucket{Manual, Unreviewed, Caveats, Automated, NoEffect}

// Card is one distinct card in the deck.
type Card struct {
	Name     string `json:"name"`
	OracleID string `json:"oracle_id"`
	// Count is the copies in the list, summed across the command zone
	// and the main deck.
	Count  int    `json:"count"`
	Bucket Bucket `json:"bucket"`
	// Caveats is the card's player-facing caveat sentences, exactly as
	// the catalogue page shows them. Non-empty only in the Caveats
	// bucket.
	Caveats []string `json:"caveats,omitempty"`
}

// Report is one deck's coverage (ADR 0095 §1's CoverageReport).
type Report struct {
	DeckName string `json:"deck_name"`
	// Source is "moxfield", "archidekt" or "text".
	Source string `json:"source"`
	// SourceURL is the deck's canonical link; empty for pasted text.
	SourceURL string `json:"source_url,omitempty"`
	// DeckKey is "moxfield:<id>" or "archidekt:<id>" for a link, and
	// "list:<hash>" (ListKey) for a pasted list — whatever the caller
	// put on the Deck; Build does not derive it.
	DeckKey    string   `json:"deck_key,omitempty"`
	Commanders []string `json:"commanders"`
	// Counts is by distinct card, and always carries all five buckets.
	Counts map[Bucket]int `json:"counts"`
	// Cards is sorted by bucket (see Buckets), then by name.
	Cards []Card `json:"cards"`
	// Unknown is the names the card index could not resolve. They are
	// in no bucket.
	Unknown []string `json:"unknown"`
	// Violations is deck.Validate's answer, reported and never
	// blocking: the checker answers "how much of this is automated",
	// not "is this legal".
	Violations []deck.Violation `json:"violations"`
}

// Needed reports whether the deck has anything to request: a manual or
// an unreviewed card (ADR 0095 §3 step 2).
func (r *Report) Needed() bool {
	return r.Counts[Manual]+r.Counts[Unreviewed] > 0
}

// CardsIn returns the report's cards in one bucket, in report order.
func (r *Report) CardsIn(b Bucket) []Card {
	var out []Card
	for _, c := range r.Cards {
		if c.Bucket == b {
			out = append(out, c)
		}
	}
	return out
}

// Deck is a fetched or parsed decklist plus where it came from.
type Deck struct {
	Name      string
	Source    string // "moxfield" | "archidekt" | "text"
	SourceURL string
	DeckKey   string
	Entries   []deck.Entry
}

// ErrNoIndex is returned when no card index is loaded.
var ErrNoIndex = errors.New("deckcoverage: no card index loaded")

// Build resolves d against idx and sorts every card into its bucket.
//
// Sideboard rows are resolved and handed to deck.Validate (which warns
// about them) but are not bucketed: a Commander deck has no sideboard,
// and a builder's "considering" pile is not the deck.
func Build(idx *cards.Index, d Deck) (*Report, error) {
	if idx == nil {
		return nil, ErrNoIndex
	}
	verdicts := catalogVerdicts(idx)

	r := &Report{
		DeckName:   strings.TrimSpace(d.Name),
		Source:     d.Source,
		SourceURL:  d.SourceURL,
		DeckKey:    d.DeckKey,
		Commanders: []string{},
		Counts:     make(map[Bucket]int, len(Buckets)),
		Cards:      []Card{},
		Unknown:    []string{},
		Violations: []deck.Violation{},
	}
	for _, b := range Buckets {
		r.Counts[b] = 0
	}

	list := &deck.List{Name: r.DeckName}
	byKey := map[string]int{} // distinct-card key -> index in r.Cards
	unknownSeen := map[string]bool{}
	for _, e := range d.Entries {
		if e.Count <= 0 {
			continue
		}
		c, ok := idx.FindByName(e.Name)
		if !ok {
			n := strings.TrimSpace(e.Name)
			if k := strings.ToLower(n); !unknownSeen[k] {
				unknownSeen[k] = true
				r.Unknown = append(r.Unknown, n)
			}
			continue
		}
		for range e.Count {
			switch {
			case e.IsCommander:
				list.Commanders = append(list.Commanders, c)
			case e.IsSideboard:
				list.Sideboard = append(list.Sideboard, c)
			default:
				list.Mainboard = append(list.Mainboard, c)
			}
		}
		if e.IsSideboard {
			continue
		}
		if e.IsCommander {
			r.Commanders = append(r.Commanders, c.Name)
		}
		key := c.OracleID.String()
		if c.OracleID == uuid.Nil {
			key = "name:" + strings.ToLower(c.Name)
		}
		if i, ok := byKey[key]; ok {
			r.Cards[i].Count += e.Count
			continue
		}
		bucket, caveats := bucketOf(c, e.IsCommander, verdicts)
		byKey[key] = len(r.Cards)
		oracle := ""
		if c.OracleID != uuid.Nil {
			oracle = c.OracleID.String()
		}
		r.Cards = append(r.Cards, Card{
			Name:     c.Name,
			OracleID: oracle,
			Count:    e.Count,
			Bucket:   bucket,
			Caveats:  caveats,
		})
	}

	rank := make(map[Bucket]int, len(Buckets))
	for i, b := range Buckets {
		rank[b] = i
	}
	sort.SliceStable(r.Cards, func(a, b int) bool {
		ca, cb := r.Cards[a], r.Cards[b]
		if ca.Bucket != cb.Bucket {
			return rank[ca.Bucket] < rank[cb.Bucket]
		}
		return strings.ToLower(ca.Name) < strings.ToLower(cb.Name)
	})
	for _, c := range r.Cards {
		r.Counts[c.Bucket]++
	}

	if err := deck.Validate(list); err != nil {
		var ve *deck.ValidationError
		if errors.As(err, &ve) {
			r.Violations = append(r.Violations, ve.Violations...)
		}
	}
	return r, nil
}

// bucketOf decides one card's bucket. The order matters only for a
// catalogued card whose printed text is nothing the engine lacks —
// Serra Angel's keywords — which lands in NoEffect, as ADR 0095's
// table orders it: there is nothing on it for anybody to review.
func bucketOf(c cards.Card, isCommander bool, verdicts map[string]catalog.Entry) (Bucket, []string) {
	gc := deck.ToGameCard(c, isCommander)
	switch {
	case game.Unimplemented(gc):
		return Manual, nil
	case !gc.NeedsEffect:
		return NoEffect, nil
	}
	v, ok := verdicts[c.OracleID.String()]
	if !ok {
		// Not Unimplemented, so the engine has an entry for it, but
		// the catalogue page knows no verdict: say so honestly rather
		// than calling it automated.
		return Unreviewed, nil
	}
	switch v.Completeness {
	case effects.CompletenessFull.String():
		return Automated, nil
	case effects.CompletenessCaveats.String():
		return Caveats, append([]string(nil), v.Caveats...)
	default:
		return Unreviewed, nil
	}
}

// catalogVerdicts indexes catalog.Build by oracle ID. catalog.Build is
// the one place completeness is merged across a card's faces (one
// caveated half makes a caveat card; an MDFC whose front has no spec
// earns a caveat no card file declares), so the checker and the
// catalogue page can never disagree about a card.
func catalogVerdicts(idx *cards.Index) map[string]catalog.Entry {
	built := catalog.Build(idx)
	out := make(map[string]catalog.Entry, len(built.Cards))
	for _, e := range built.Cards {
		out[e.OracleID] = e
	}
	return out
}
