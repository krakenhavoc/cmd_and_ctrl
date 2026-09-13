// Package catalog publishes the card-effect catalog: the list of
// cards this engine actually automates, joined against Scryfall
// printing data so it can be rendered, and labelled with how
// faithfully each one implements its printed text.
//
// It exists as its own package rather than as another route on
// internal/lobby for two reasons. The join needs both
// internal/cards (printings) and internal/cards/effects (specs), and
// internal/cards/effects is otherwise blank-imported exactly once,
// by main, precisely so nothing depends on it. Putting the join here
// keeps that single import explicit and keeps a read-only public
// surface out of the lobby's session-shaped handler file.
//
// Everything here is read-only and derived from data that is already
// public: the Scryfall bulk dump and the contents of this
// repository. No game, seat, session or player state is reachable
// from this package, which is what makes the routes safe to serve
// unauthenticated.
package catalog

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Entry is one card in the published catalog: what the card is, and
// what the engine promises about it.
//
// JSON field names are the client's contract — client/src/lib/catalog.ts
// mirrors this struct, as protocol.ts mirrors the protocol package.
// Update both sides in lockstep.
type Entry struct {
	// OracleID is the catalog key and the card's printing-independent
	// identity. Face-suffixed registry keys ("<id>#1") collapse into
	// the one Entry for their card — a modal double-faced card is one
	// card, however many halves the registry declares.
	OracleID string `json:"oracle_id"`

	// ScryfallID is the representative printing, and the only thing
	// the image route accepts. Empty when the bulk dump has no
	// printing for this oracle ID — see Missing.
	ScryfallID string `json:"scryfall_id,omitempty"`

	Name       string `json:"name"`
	ManaCost   string `json:"mana_cost,omitempty"`
	TypeLine   string `json:"type_line,omitempty"`
	OracleText string `json:"oracle_text,omitempty"`

	// Types is the primary card types parsed off the type line
	// ("Legendary Artifact Creature — Golem" yields Artifact and
	// Creature), lowercased. Parsed server-side so the page's type
	// filter is one equality check and so the parsing has a Go test
	// rather than living in a template.
	Types []string `json:"types,omitempty"`

	// ColorIdentity is the Commander-relevant colour set (WUBRG
	// letters). It is what the page filters on: in Commander the
	// identity is the question players actually ask, and it gives
	// lands and mana rocks a sensible answer where Colors is empty.
	ColorIdentity []string `json:"color_identity,omitempty"`
	Colors        []string `json:"colors,omitempty"`

	// Completeness is "full", "caveats" or "unreviewed" — see
	// effects.Completeness. Never omitted: the whole point of the
	// page is that this value is always stated.
	Completeness string `json:"completeness"`

	// Caveats are the printed clauses the engine does not model, in
	// player-facing language. Non-empty exactly when Completeness is
	// "caveats".
	Caveats []string `json:"caveats,omitempty"`

	// Faces names the printed faces for a multi-faced card, front
	// first. Empty for ordinary cards.
	Faces []Face `json:"faces,omitempty"`

	// Missing reports that the catalog registers this oracle ID but
	// the loaded Scryfall dump has no printing for it. The entry
	// still appears — the engine really does automate it — but with
	// only the name the spec carries and no art. A non-zero count of
	// these means the dump is stale relative to the catalog, which is
	// worth seeing rather than hiding.
	Missing bool `json:"missing,omitempty"`
}

// Face is one printed half of a multi-faced card.
type Face struct {
	Index      int    `json:"index"`
	Name       string `json:"name"`
	TypeLine   string `json:"type_line,omitempty"`
	OracleText string `json:"oracle_text,omitempty"`
	ManaCost   string `json:"mana_cost,omitempty"`

	// Automated reports whether the registry has a spec for this
	// specific face. The distinction is real and currently visible:
	// the modal-double-faced lands register only their land back, so
	// their creature front is not automated at all.
	Automated bool `json:"automated"`
}

// Counts is the summary the page shows above the grid.
type Counts struct {
	Total      int `json:"total"`
	Full       int `json:"full"`
	Caveats    int `json:"caveats"`
	Unreviewed int `json:"unreviewed"`
	Missing    int `json:"missing"`
}

// Response is the body of GET /catalog.
type Response struct {
	Counts Counts  `json:"counts"`
	Cards  []Entry `json:"cards"`
}

// Build joins every registered spec against idx and returns the
// catalog sorted by name. A nil idx yields name-only entries — the
// server can still publish what it automates before the dump lands.
//
// Deliberately a pure function of (registry, index) with no caching:
// the registry is fixed at init and the index changes only on an
// explicit reload, so the cost is one pass over a few hundred specs
// per request, and a stale cache would be a worse bug than that pass
// is a cost.
func Build(idx *cards.Index) Response {
	return build(effects.All(), idx)
}

// build is Build with the spec list injected, so tests can exercise
// the grouping and completeness rules against a handful of synthetic
// specs instead of against whatever several hundred cards the real
// registry happens to hold this week.
func build(specs []effects.Spec, idx *cards.Index) Response {
	// Group registry keys by the card they belong to. A key is
	// either a bare oracle ID (face 0) or "<oracle_id>#N" — see
	// game.CatalogKey, which is where that encoding is defined.
	type group struct {
		specs map[int]effects.Spec // face index -> spec
	}
	groups := map[string]*group{}
	for _, spec := range specs {
		id, face := splitCatalogKey(spec.OracleID)
		g := groups[id]
		if g == nil {
			g = &group{specs: map[int]effects.Spec{}}
			groups[id] = g
		}
		g.specs[face] = spec
	}

	out := make([]Entry, 0, len(groups))
	for id, g := range groups {
		out = append(out, buildEntry(idx, id, g.specs))
	}

	sort.Slice(out, func(a, b int) bool {
		if la, lb := strings.ToLower(out[a].Name), strings.ToLower(out[b].Name); la != lb {
			return la < lb
		}
		return out[a].OracleID < out[b].OracleID
	})

	counts := Counts{Total: len(out)}
	for _, e := range out {
		switch e.Completeness {
		case effects.CompletenessFull.String():
			counts.Full++
		case effects.CompletenessCaveats.String():
			counts.Caveats++
		default:
			counts.Unreviewed++
		}
		if e.Missing {
			counts.Missing++
		}
	}
	return Response{Counts: counts, Cards: out}
}

// buildEntry renders one card from its specs plus, if the dump has
// one, a printing.
func buildEntry(idx *cards.Index, oracleID string, specs map[int]effects.Spec) Entry {
	e := Entry{OracleID: oracleID}

	// Name: the printing's wins below. This is the fallback for a
	// card the dump has no printing for, and it takes the lowest
	// registered face — face 0's name is the card's name, while a
	// back-face-only spec names just the back ("Akoum Teeth"), which
	// is the best available answer in that case and no worse than
	// showing a bare UUID.
	for _, face := range sortedFaces(specs) {
		if specs[face].Name != "" {
			e.Name = specs[face].Name
			break
		}
	}

	var print cards.Card
	var found bool
	if idx != nil {
		if parsed, err := uuid.Parse(oracleID); err == nil {
			print, found = idx.FindByOracleID(parsed)
		}
	}
	if !found {
		e.Missing = true
	} else {
		e.ScryfallID = print.ID.String()
		e.Name = print.Name
		e.ManaCost = print.ManaCost
		e.TypeLine = print.TypeLine
		e.OracleText = print.OracleText
		e.Types = primaryTypes(print.TypeLine)
		e.ColorIdentity = upperAll(print.ColorIdentity)
		e.Colors = upperAll(print.Colors)
		for i, f := range print.CardFaces {
			_, automated := specs[i]
			e.Faces = append(e.Faces, Face{
				Index:      i,
				Name:       f.Name,
				TypeLine:   f.TypeLine,
				OracleText: f.OracleText,
				ManaCost:   f.ManaCost,
				Automated:  automated,
			})
			if e.Types == nil {
				e.Types = primaryTypes(f.TypeLine)
			}
		}
	}

	completeness, caveats := mergeCompleteness(specs, e.Faces)
	e.Completeness = completeness.String()
	e.Caveats = caveats
	return e
}

// mergeCompleteness reduces a card's per-face declarations to one
// verdict, and adds the caveat the registry shape itself implies.
//
// The reduction is pessimistic on purpose. A card is only "full"
// when every declaration says so; one unreviewed half makes the card
// unreviewed, and one caveat makes it a caveat card, because that is
// what a player would conclude if they read both halves.
func mergeCompleteness(specs map[int]effects.Spec, faces []Face) (effects.Completeness, []string) {
	var caveats []string
	sawFull, sawUnreviewed := false, false
	for _, face := range sortedFaces(specs) {
		s := specs[face]
		switch s.Completeness {
		case effects.CompletenessFull:
			sawFull = true
		case effects.CompletenessCaveats:
			caveats = append(caveats, s.Caveats...)
		default:
			sawUnreviewed = true
		}
	}

	// A printed face with no spec of its own is a gap the card files
	// cannot declare, because no card file owns that face. The modal
	// double-faced lands are the live case: the registry automates
	// the land back and nothing automates the creature front.
	if _, hasFront := specs[0]; !hasFront && len(faces) > 0 {
		var unautomated []string
		for _, f := range faces {
			if !f.Automated && f.Name != "" {
				unautomated = append(unautomated, f.Name)
			}
		}
		if len(unautomated) > 0 {
			caveats = append(caveats,
				"Only part of this card is automated — "+
					strings.Join(unautomated, " and ")+
					" falls back to manual play.")
		}
	}

	switch {
	case len(caveats) > 0:
		return effects.CompletenessCaveats, dedupe(caveats)
	case sawUnreviewed || !sawFull:
		return effects.CompletenessUnreviewed, nil
	default:
		return effects.CompletenessFull, nil
	}
}

// splitCatalogKey undoes game.CatalogKey: "<uuid>" is face 0 and
// "<uuid>#N" is face N. An unparseable suffix is treated as part of
// the ID rather than guessed at, so a malformed key shows up as a
// missing printing instead of silently merging into another card.
func splitCatalogKey(key string) (string, int) {
	hash := strings.LastIndexByte(key, '#')
	if hash < 0 {
		return key, 0
	}
	face := 0
	for _, r := range key[hash+1:] {
		if r < '0' || r > '9' {
			return key, 0
		}
		face = face*10 + int(r-'0')
	}
	if hash+1 >= len(key) {
		return key, 0
	}
	return key[:hash], face
}

// primaryTypes pulls the card types out of a type line, dropping
// supertypes (Legendary, Basic, Snow, World) and everything after
// the em dash. Lowercased so the client's filter is a plain compare.
//
// Split and modal type lines ("Instant // Land") are handled by
// scanning both halves, which is what a player filtering for "land"
// expects of Sea Gate Restoration.
func primaryTypes(typeLine string) []string {
	if typeLine == "" {
		return nil
	}
	known := []string{
		"artifact", "battle", "creature", "enchantment", "instant",
		"kindred", "land", "planeswalker", "sorcery", "tribal",
	}
	// Everything past an em dash is subtypes; they are not card types.
	head := typeLine
	if i := strings.Index(head, "—"); i >= 0 {
		// Keep the other half of a split line, whose dash is its own.
		parts := strings.Split(typeLine, "//")
		head = ""
		for _, p := range parts {
			if i := strings.Index(p, "—"); i >= 0 {
				p = p[:i]
			}
			head += " " + p
		}
	}
	lower := strings.ToLower(head)
	var out []string
	for _, t := range known {
		if containsWord(lower, t) {
			out = append(out, t)
		}
	}
	return out
}

// containsWord reports whether word appears in s delimited by
// non-letters, so "land" does not match "Landwalk" and "art" would
// not match "Artifact".
func containsWord(s, word string) bool {
	for i := 0; ; {
		j := strings.Index(s[i:], word)
		if j < 0 {
			return false
		}
		start := i + j
		end := start + len(word)
		beforeOK := start == 0 || !isLetter(s[start-1])
		afterOK := end == len(s) || !isLetter(s[end])
		if beforeOK && afterOK {
			return true
		}
		i = start + 1
		if i >= len(s) {
			return false
		}
	}
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func sortedFaces(specs map[int]effects.Spec) []int {
	out := make([]int, 0, len(specs))
	for f := range specs {
		out = append(out, f)
	}
	sort.Ints(out)
	return out
}

func upperAll(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, 0, len(in))
	for _, s := range in {
		out = append(out, strings.ToUpper(s))
	}
	sort.Strings(out)
	return out
}

func dedupe(in []string) []string {
	seen := make(map[string]bool, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

// IsCatalogPrinting reports whether scryfallID is the representative
// printing of a card the catalog registers. The public image route
// asks this before serving a byte: the page may publish art for the
// cards it publishes and nothing else, which keeps an unauthenticated
// route from being a general-purpose proxy for the whole Scryfall
// dump.
func IsCatalogPrinting(idx *cards.Index, scryfallID uuid.UUID) bool {
	if idx == nil {
		return false
	}
	c, ok := idx.Get(scryfallID)
	if !ok || c.OracleID == uuid.Nil {
		return false
	}
	rep, ok := idx.FindByOracleID(c.OracleID)
	if !ok || rep.ID != scryfallID {
		return false
	}
	return effects.Has(c.OracleID.String()) ||
		effects.Has(c.OracleID.String()+"#1")
}
