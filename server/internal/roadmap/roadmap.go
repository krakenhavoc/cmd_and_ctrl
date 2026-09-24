// Package roadmap is the single source for what the engine supports:
// one entry per keyword, mechanic and engine seam, each with a status a
// player can read (ADR 0092).
//
// # Why it exists
//
// Before this package the answer to "does the engine do X?" lived in
// four places, all written for engineers and none of them checked
// against each other: the open-seams table in docs/engine-seams.md
// (hand-typed, and by September 2026 describing several gaps that had
// already closed), the census in the card-coverage roadmap, the
// coverage package's curated mechanic table, and each card's Caveats.
//
// Here the registry (registry.go) is curated by hand, because a status
// is a judgement, but every judgement that CAN be checked is checked
// by registry_test.go against the live catalog:
//
//   - an implemented or partial item has to find at least one fully
//     automated card that uses it, or say why none exists;
//   - a card listed as waiting on a seam must not already be in the
//     catalog as complete — if it is, the seam has probably closed;
//   - every sentence a player reads passes the Caveats tone rule;
//   - docs/engine-seams.md's open table is generated from this
//     registry, so the two cannot drift.
//
// # What Build returns
//
// Build computes a plain-data Roadmap once per binary (the registry
// and the catalog are both fixed at init). It carries card NAMES and
// caveat sentences, and nothing else about a card: no art and no
// oracle text, which is what lets the page that serves it be public
// when the card catalog is not (ADR 0092 Decision 1).
package roadmap

import (
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/coverage"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
)

// Kind says what sort of thing an item is.
type Kind string

const (
	// KindKeyword is a keyword ability the engine enforces — one row
	// (or, for landwalk, one family) of game.CanonicalKeywordTable.
	KindKeyword Kind = "keyword"
	// KindMechanic is any other printed mechanic with a shape in the
	// catalog: an alternative cost, a special action, a designation.
	KindMechanic Kind = "mechanic"
	// KindSeam is an engine gap a card can be waiting on, from
	// docs/engine-seams.md. A seam that has fully closed stays in the
	// registry as implemented, so the page can say so.
	KindSeam Kind = "seam"
)

// Status is how much of an item works.
type Status string

const (
	StatusImplemented Status = "implemented"
	StatusPartial     Status = "partial"
	StatusMissing     Status = "missing"
)

// Item is one hand-curated registry entry. Only the fields a player
// reads are published; the rest steer Build and the tests.
type Item struct {
	// Slug is the stable identity: unique, lowercase, dashes. The page
	// links to it and the tests key on it.
	Slug string
	// Name is the display name.
	Name   string
	Kind   Kind
	Status Status

	// Summary is one or two player-facing sentences saying what the
	// item is. Held to the Caveats tone rule.
	Summary string
	// Missing says, for a partial or missing item, what does not work
	// yet — player-facing, same rule. Empty for an implemented item.
	Missing string

	// Rules are Comprehensive Rules citations, pinned to the edition
	// AGENTS.md §6 names. Only numbers the repository already cites
	// for the same rule appear here.
	Rules []string
	// Issue is the tracking issue; required for partial and missing.
	Issue int
	// ADR is the decision record's file name, or empty.
	ADR string

	// Keywords are the canonical keyword tokens this item covers (see
	// game.CanonicalKeywordTable). Only keyword items set it, and the
	// tests hold every token to exactly one item.
	Keywords []string

	// Probe reports whether a registered spec demonstrably uses the
	// item. It reads a declaration off the Spec, never a guess.
	Probe func(effects.Spec) bool
	// Printed is a regular expression over the card's printed oracle
	// text. A card matches when its text uses the mechanic AND its card
	// file declares it complete — the most honest probe for a
	// mechanic that is a closure rather than a Spec field (ward,
	// cascade, hideaway).
	Printed string
	// Mechanic names a row of coverage.Mechanics(). The item borrows
	// that row's phrases, and its probe when the probe is Exact; a
	// Heuristic probe is never used to pick examples.
	Mechanic string
	// Phrases are caveat phrases that name the item, for the partial
	// examples. Matched like coverage's (word-bounded, any case).
	Phrases []string

	// Examples are curated featured cards, shown first. Each must be
	// registered, complete, and satisfy the probe.
	Examples []string
	// NoCatalogExample, when set, excuses an implemented or partial
	// item from having an example, and says why (a vanilla keyword
	// creature needs no card file at all).
	NoCatalogExample string

	// Waiting are the cards (by catalog name) known to wait on this
	// item. A batch PR that skips a card for a seam appends it here.
	Waiting []string
	// Unblocks is the number of cards this item alone blocks: the
	// 2026-09-16 audit's only-blocker count, or 0 where that number
	// predates a half that has since shipped.
	Unblocks int
	// Pin puts the item in "up next" ahead of the ranking, in
	// ascending order. Zero means ranked.
	Pin int

	// EngineNotes is the engineering prose for engine-seams.md's
	// generated table. Never published on the page.
	EngineNotes string
	// Tracked overrides the table's Tracked cell (default "#Issue").
	Tracked string
}

// Roadmap is Build's result: plain data, ready to serialise.
type Roadmap struct {
	Counts Counts   `json:"counts"`
	Next   []string `json:"next"`
	Items  []Entry  `json:"items"`
}

// Counts is items by status.
type Counts struct {
	Implemented int `json:"implemented"`
	Partial     int `json:"partial"`
	Missing     int `json:"missing"`
}

// Entry is one published item.
type Entry struct {
	Slug            string           `json:"slug"`
	Name            string           `json:"name"`
	Kind            Kind             `json:"kind"`
	Status          Status           `json:"status"`
	Summary         string           `json:"summary"`
	Missing         string           `json:"missing,omitempty"`
	Rules           []string         `json:"rules,omitempty"`
	Issue           int              `json:"issue,omitempty"`
	ADR             string           `json:"adr,omitempty"`
	Pin             int              `json:"pin,omitempty"`
	Unblocks        int              `json:"unblocks,omitempty"`
	Examples        []Example        `json:"examples,omitempty"`
	PartialExamples []PartialExample `json:"partial_examples,omitempty"`
	Waiting         []string         `json:"waiting,omitempty"`
}

// Example is a fully automated card that uses the item.
type Example struct {
	Name     string `json:"name"`
	OracleID string `json:"oracle_id"`
}

// PartialExample is a card that works with a gap the item explains.
type PartialExample struct {
	Name   string `json:"name"`
	Caveat string `json:"caveat"`
}

// Limits on what Build publishes per item and in "up next".
const (
	MaxExamples        = 3
	MaxPartialExamples = 3
	MaxNext            = 6
)

// Items returns the registry. The slice is shared; do not modify it.
func Items() []Item { return items }

var (
	buildOnce sync.Once
	built     Roadmap
)

// Build returns the roadmap, computed once per binary. The result is
// shared between callers and must not be modified.
func Build() Roadmap {
	buildOnce.Do(func() { built = build(items, loadCards()) })
	return built
}

// card is one whole catalog card: every registry key for one oracle
// ID, with the catalog's merged verdict on it.
type card struct {
	id       string
	name     string
	specs    []effects.Spec
	hasFront bool
	verdict  string // effects.Completeness.String(), merged over faces
	caveats  []string
	text     string // every printed face's oracle text, joined
}

// full reports whether the card may be shown as fully automated. A
// card whose front face has no spec is never full: the public catalog
// shows it as partly manual, whatever its back face declares.
func (c *card) full() bool {
	return c.hasFront && c.verdict == effects.CompletenessFull.String()
}

// cardSet is the catalog, grouped by card and indexed by name.
type cardSet struct {
	cards  []*card // sorted by name
	byName map[string]*card
}

// loadCards groups the live registry by card. The completeness verdict
// is catalog.Build's — the same merge the public catalog page shows —
// so "full" means the same thing on both pages.
func loadCards() cardSet {
	byID := map[string]*card{}
	for _, s := range effects.All() {
		id, face := coverage.BaseOracleID(s.OracleID)
		c := byID[id]
		if c == nil {
			c = &card{id: id}
			byID[id] = c
		}
		c.specs = append(c.specs, s)
		if face == 0 {
			c.hasFront = true
			c.name = s.Name
		} else if c.name == "" {
			c.name = s.Name
		}
	}
	for _, e := range catalog.Build(nil).Cards {
		if c := byID[e.OracleID]; c != nil {
			c.verdict = e.Completeness
			c.caveats = e.Caveats
		}
	}
	printed := coverage.PrintedOracle()
	set := cardSet{byName: map[string]*card{}}
	for id, c := range byID {
		sort.Slice(c.specs, func(i, j int) bool { return c.specs[i].OracleID < c.specs[j].OracleID })
		if oc, ok := printed[id]; ok {
			var faces []string
			for _, f := range oc.AllFaces() {
				faces = append(faces, f.Text)
			}
			c.text = strings.Join(faces, "\n")
		}
		set.cards = append(set.cards, c)
	}
	sort.Slice(set.cards, func(i, j int) bool {
		if a, b := strings.ToLower(set.cards[i].name), strings.ToLower(set.cards[j].name); a != b {
			return a < b
		}
		return set.cards[i].id < set.cards[j].id
	})
	for _, c := range set.cards {
		if _, dup := set.byName[c.name]; !dup {
			set.byName[c.name] = c
		}
	}
	return set
}

// matcher is an item's compiled probe: does this card use the item?
type matcher struct {
	probes  []func(effects.Spec) bool
	printed *regexp.Regexp
	phrases func(string) bool
}

func compile(it Item) matcher {
	var m matcher
	if it.Probe != nil {
		m.probes = append(m.probes, it.Probe)
	}
	phrases := append([]string(nil), it.Phrases...)
	if it.Mechanic != "" {
		if mech, ok := mechanicByName(it.Mechanic); ok {
			if mech.Confidence == coverage.Exact {
				m.probes = append(m.probes, mech.Implements)
			}
			phrases = append(phrases, mech.Phrases...)
		}
	}
	if it.Printed != "" {
		m.printed = regexp.MustCompile(it.Printed)
	}
	if len(phrases) > 0 {
		m.phrases = coverage.PhraseMatcher(phrases)
	}
	return m
}

// hasProbe reports whether the item has any way to find a card at all.
func (m matcher) hasProbe() bool { return len(m.probes) > 0 || m.printed != nil }

// uses reports whether the card uses the item: a declaration on any of
// its faces, or its printed text.
func (m matcher) uses(c *card) bool {
	for _, p := range m.probes {
		for _, s := range c.specs {
			if p(s) {
				return true
			}
		}
	}
	return m.printed != nil && c.text != "" && m.printed.MatchString(c.text)
}

func mechanicByName(name string) (coverage.Mechanic, bool) {
	for _, m := range coverage.Mechanics() {
		if m.Name == name {
			return m, true
		}
	}
	return coverage.Mechanic{}, false
}

func build(reg []Item, cs cardSet) Roadmap {
	var out Roadmap
	for _, it := range reg {
		m := compile(it)
		e := Entry{
			Slug: it.Slug, Name: it.Name, Kind: it.Kind, Status: it.Status,
			Summary: it.Summary, Missing: it.Missing,
			Rules: it.Rules, Issue: it.Issue, ADR: it.ADR,
			Pin: it.Pin, Unblocks: it.Unblocks,
			Waiting:         it.Waiting,
			Examples:        examplesFor(it, m, cs),
			PartialExamples: partialExamplesFor(it, m, cs),
		}
		switch it.Status {
		case StatusImplemented:
			out.Counts.Implemented++
		case StatusPartial:
			out.Counts.Partial++
		default:
			out.Counts.Missing++
		}
		out.Items = append(out.Items, e)
	}
	out.Next = next(reg)
	return out
}

// examplesFor is the curated examples first, then every other complete
// card the probe finds, in name order, up to MaxExamples.
func examplesFor(it Item, m matcher, cs cardSet) []Example {
	var out []Example
	seen := map[string]bool{}
	add := func(c *card) {
		if len(out) < MaxExamples && !seen[c.id] {
			seen[c.id] = true
			out = append(out, Example{Name: c.name, OracleID: c.id})
		}
	}
	for _, name := range it.Examples {
		if c := cs.byName[name]; c != nil && c.full() {
			add(c)
		}
	}
	if m.hasProbe() {
		for _, c := range cs.cards {
			if len(out) >= MaxExamples {
				break
			}
			if c.full() && m.uses(c) {
				add(c)
			}
		}
	}
	return out
}

// partialExamplesFor lists cards that work with a gap this item
// explains: first the item's own waiting cards that are in the catalog
// with a caveat, then any caveat card whose caveat names the item.
func partialExamplesFor(it Item, m matcher, cs cardSet) []PartialExample {
	var out []PartialExample
	seen := map[string]bool{}
	add := func(c *card, caveat string) {
		if len(out) < MaxPartialExamples && !seen[c.id] {
			seen[c.id] = true
			out = append(out, PartialExample{Name: c.name, Caveat: caveat})
		}
	}
	for _, name := range it.Waiting {
		if c := cs.byName[name]; c != nil && len(c.caveats) > 0 {
			add(c, c.caveats[0])
		}
	}
	if m.phrases == nil {
		return out
	}
	for _, c := range cs.cards {
		for _, cv := range c.caveats {
			if m.phrases(cv) {
				add(c, cv)
				break
			}
		}
	}
	return out
}

// score is the "up next" ranking key: the cards this item alone
// blocks, plus the cards recorded as waiting on it.
func (it Item) score() int { return it.Unblocks + len(it.Waiting) }

// next is the "up next" list: pinned items in pin order, then every
// unfinished item with something waiting on it, most cards first.
func next(reg []Item) []string {
	var pinned, ranked []Item
	for _, it := range reg {
		switch {
		case it.Pin > 0:
			pinned = append(pinned, it)
		case it.Status != StatusImplemented && it.score() > 0:
			ranked = append(ranked, it)
		}
	}
	sort.SliceStable(pinned, func(i, j int) bool { return pinned[i].Pin < pinned[j].Pin })
	sort.SliceStable(ranked, func(i, j int) bool {
		if a, b := ranked[i].score(), ranked[j].score(); a != b {
			return a > b
		}
		return ranked[i].Name < ranked[j].Name
	})
	var out []string
	for _, it := range append(pinned, ranked...) {
		if len(out) == MaxNext {
			break
		}
		out = append(out, it.Slug)
	}
	return out
}
