package catalog

import (
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// SetSummary is one set in the catalogue's set picker: how many of its
// cards the engine plays, by how faithfully, and how many it does not
// play yet (#4790).
//
// Counts are of UNIQUE CARDS (oracle IDs), not printings: a card
// reprinted three times in a set is one card. Only cards a player could
// put in a deck count (see countsInSet), so Total is "the set's cards" as
// a player would count them, not Scryfall's record count.
//
// JSON field names are the client's contract; client/src/lib/catalog.ts
// mirrors this struct.
type SetSummary struct {
	// Code is Scryfall's lowercase set code ("spm"), the key
	// Entry.Sets refers to.
	Code string `json:"code"`
	Name string `json:"name"`
	// Released is the set's release date, "YYYY-MM-DD". Empty when the
	// dump carries none for the set.
	Released string `json:"released,omitempty"`

	// Total is the set's unique cards. Full, Caveats and Unreviewed split
	// the ones the catalogue holds by declared completeness, and
	// Unsupported is the rest: Total minus the three. The set's catalogued
	// cards are exactly Full+Caveats+Unreviewed.
	Total       int `json:"total"`
	Full        int `json:"full"`
	Caveats     int `json:"caveats"`
	Unreviewed  int `json:"unreviewed"`
	Unsupported int `json:"unsupported"`
}

// BuildWithSets is Build plus the set picker's data: Response.Sets, and
// Entry.Sets on every card. It is what GET /catalog serves.
//
// Build stays set-free on purpose. The roadmap, the deck checker and the
// deck library all call Build for its verdicts and none of them reads a
// set, so they should not each pay a pass over every printing in the
// dump.
func BuildWithSets(idx *cards.Index) Response {
	return buildWithSets(effects.All(), idx)
}

// buildWithSets is BuildWithSets with the spec list injected, as build is.
func buildWithSets(specs []effects.Spec, idx *cards.Index) Response {
	resp := build(specs, idx)
	attachSets(&resp, idx)
	return resp
}

// countsInSet reports whether a printing stands for a card that belongs
// in a set's tally: a real, playable, paper card.
//
// Excluded, because the catalogue could never "support" them and listing
// them as not-yet-supported would be noise: tokens, double-faced tokens,
// art-series cards, memorabilia and other placeholder records
// (cards.Card.Playable), emblems, planes, schemes and vanguards (none is
// a card you put in a Commander deck), digital-only printings (Alchemy
// and the like, not legal in paper Commander), and basic lands (nothing
// to automate).
func countsInSet(c cards.Card) bool {
	if c.SetCode == "" || c.OracleID == uuid.Nil || c.Digital || !c.Playable() {
		return false
	}
	switch c.Layout {
	case "emblem", "planar", "scheme", "vanguard":
		return false
	}
	// "Basic Land — Forest", "Basic Snow Land — Snow-Covered Forest".
	return !strings.HasPrefix(c.TypeLine, "Basic ")
}

// setAcc accumulates one set while the dump is walked.
type setAcc struct {
	name   string
	dates  map[string]int
	oracle map[uuid.UUID]struct{}
}

// attachSets walks the index's printings once, fills resp.Sets, and gives
// each catalogued entry the codes of the sets it was printed in.
//
// A nil idx leaves the response as build made it: no dump, no sets.
func attachSets(resp *Response, idx *cards.Index) {
	if idx == nil {
		return
	}

	sets := map[string]*setAcc{}
	idx.EachPrinting(func(c cards.Card) {
		if !countsInSet(c) {
			return
		}
		a := sets[c.SetCode]
		if a == nil {
			a = &setAcc{dates: map[string]int{}, oracle: map[uuid.UUID]struct{}{}}
			sets[c.SetCode] = a
		}
		if a.name == "" {
			a.name = c.SetName
		}
		if c.ReleasedAt != "" {
			a.dates[c.ReleasedAt]++
		}
		a.oracle[c.OracleID] = struct{}{}
	})

	byOracle := make(map[uuid.UUID]*Entry, len(resp.Cards))
	for i := range resp.Cards {
		if id, err := uuid.Parse(resp.Cards[i].OracleID); err == nil {
			byOracle[id] = &resp.Cards[i]
		}
	}

	out := make([]SetSummary, 0, len(sets))
	members := make(map[string][]uuid.UUID, len(sets))
	for code, a := range sets {
		s := SetSummary{
			Code:     code,
			Name:     a.name,
			Released: commonDate(a.dates),
			Total:    len(a.oracle),
		}
		if s.Name == "" {
			s.Name = strings.ToUpper(code)
		}
		for id := range a.oracle {
			e := byOracle[id]
			if e == nil {
				continue
			}
			switch e.Completeness {
			case effects.CompletenessFull.String():
				s.Full++
			case effects.CompletenessCaveats.String():
				s.Caveats++
			default:
				s.Unreviewed++
			}
			members[code] = append(members[code], id)
		}
		s.Unsupported = s.Total - s.Full - s.Caveats - s.Unreviewed
		out = append(out, s)
	}

	// Newest first; the code settles ties so the order is stable.
	sort.Slice(out, func(a, b int) bool {
		if out[a].Released != out[b].Released {
			return out[a].Released > out[b].Released
		}
		return out[a].Code < out[b].Code
	})

	// Entries list their sets in the same order, so the first code is
	// the newest set a card is in.
	for _, s := range out {
		for _, id := range members[s.Code] {
			byOracle[id].Sets = append(byOracle[id].Sets, s.Code)
		}
	}
	resp.Sets = out
}

// commonDate is the most frequent release date among a set's printings,
// the earliest on a tie. The most common wins over the earliest because
// a set's early promos and late reprints carry stray dates, while its
// bulk of cards all carry the set's own.
func commonDate(dates map[string]int) string {
	best, bestN := "", 0
	for d, n := range dates {
		if n > bestN || (n == bestN && d < best) {
			best, bestN = d, n
		}
	}
	return best
}
