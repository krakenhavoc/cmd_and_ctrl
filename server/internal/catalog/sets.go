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

// pickerOwnSetTypes are the set types that keep their own picker entry
// even when Scryfall gives them a parent: a Commander precon or an
// eternal product is something people pick as itself (#4790).
var pickerOwnSetTypes = map[string]bool{
	"commander": true,
	"eternal":   true,
}

// pickerHiddenSetTypes are the set types a parentless set is left out of
// the picker for: promos, Secret Lair ("box"), joke sets, memorabilia,
// tokens, minigames, Alchemy and treasure chests are not sets a player
// looks for by name. Their cards stay in the catalogue; they just bring
// no picker entry (#4790).
var pickerHiddenSetTypes = map[string]bool{
	"promo":          true,
	"box":            true,
	"funny":          true,
	"memorabilia":    true,
	"token":          true,
	"minigame":       true,
	"alchemy":        true,
	"treasure_chest": true,
}

// maxSetChain bounds the walk up parent_set_code links, so a cycle in the
// sets file cannot hang the catalogue. Scryfall's chains are two deep.
const maxSetChain = 8

// pickerSet is the picker entry a printing in set code is counted under,
// and whether it has one at all (#4790). With no set table loaded it is
// the code itself, as before the table existed. Otherwise a subset folds
// into its root parent, stopping at a commander or eternal set, and a
// parentless set of a pickerHiddenSetTypes type has none. A code the
// table does not know is its own entry.
func pickerSet(idx *cards.Index, code string) (string, bool) {
	if !idx.SetsLoaded() {
		return code, true
	}
	info, ok := idx.Set(code)
	if !ok {
		return code, true
	}
	for range maxSetChain {
		if pickerOwnSetTypes[info.SetType] {
			return info.Code, true
		}
		if info.ParentSetCode == "" {
			return info.Code, !pickerHiddenSetTypes[info.SetType]
		}
		parent, ok := idx.Set(info.ParentSetCode)
		if !ok {
			// The parent is not in the table: the set stands for
			// itself rather than fold into a name we cannot show.
			return info.Code, true
		}
		info = parent
	}
	return info.Code, true
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
// When the index holds Scryfall's set table, subsets fold into their
// parent and promo-like sets drop out (pickerSet): Entry.Sets then
// carries the parent's code, once, and a set's counts are its unique
// cards across it and its subsets.
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
		code, ok := pickerSet(idx, c.SetCode)
		if !ok {
			return
		}
		a := sets[code]
		if a == nil {
			a = &setAcc{dates: map[string]int{}, oracle: map[uuid.UUID]struct{}{}}
			sets[code] = a
		}
		// A folded printing's set name is its subset's, not the entry's.
		if a.name == "" && code == c.SetCode {
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
		// The set table, when loaded, names and dates the entry: a
		// folded entry's printings carry their subsets' dates.
		if info, ok := idx.Set(code); ok {
			if info.Name != "" {
				s.Name = info.Name
			}
			if info.ReleasedAt != "" {
				s.Released = info.ReleasedAt
			}
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
