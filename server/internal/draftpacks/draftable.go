package draftpacks

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/catalog"
)

// IsFull answers whether the catalog holds a card, by oracle ID, with
// CompletenessFull.
type IsFull func(oracleID uuid.UUID) bool

// CatalogIsFull is IsFull over the real catalog, with the verdict the
// public catalogue shows: catalog.Build merges a multi-faced card's
// per-face entries, so a card with one face unautomated is not Full.
//
// It is computed from the registry on every call to CatalogIsFull;
// nothing is hand-kept.
func CatalogIsFull(idx *cards.Index) IsFull {
	full := map[uuid.UUID]bool{}
	want := effects.CompletenessFull.String()
	for _, e := range catalog.Build(idx).Cards {
		if e.Completeness != want {
			continue
		}
		if id, err := uuid.Parse(e.OracleID); err == nil {
			full[id] = true
		}
	}
	return func(id uuid.UUID) bool { return full[id] }
}

// Blockers are the cards of a pool that keep its set from being drafted:
// every card a pack can contain that the catalog does not hold as Full
// (ADR 0222 §4). Sorted by name.
//
// Basic lands are exempt. The engine plays a basic land without a
// catalog entry (there is nothing on one to automate), so the catalog
// registers none, and counting them would block every set.
func (p Pool) Blockers(isFull IsFull) []Card {
	var out []Card
	for _, c := range p.Contents() {
		if c.Basic || isFull(c.OracleID) {
			continue
		}
		out = append(out, c)
	}
	return out
}

// Draftable reports whether a draft may hand out this pool's packs: it
// can fill a pack, and no card of it is a blocker.
func (p Pool) Draftable(isFull IsFull) bool {
	return p.CanFill() && len(p.Blockers(isFull)) == 0
}

// SetStatus is one set in the draftable list. JSON field names are the
// client's contract; client/src/lib/draftSets.ts mirrors this struct.
type SetStatus struct {
	// Code is Scryfall's lowercase set code.
	Code string `json:"code"`
	Name string `json:"name"`
	// Released is "YYYY-MM-DD", empty when the data has none.
	Released string `json:"released,omitempty"`
	// Cards is how many different cards the set's packs can contain,
	// basic lands included.
	Cards int `json:"cards"`
	// Draftable is true only when Blockers is empty.
	Draftable bool `json:"draftable"`
	// Blockers are the names of the cards that are not Full in the
	// catalog, sorted. Empty for a draftable set.
	Blockers []string `json:"blockers"`
}

// Response is the body of GET /draft/sets.
type Response struct {
	// Sets is every set packs can be made for, newest first.
	Sets []SetStatus `json:"sets"`
	// Draftable counts the sets with Draftable true.
	Draftable int `json:"draftable"`
}

// listedSetTypes are the Scryfall set types sold in draftable boosters.
// A Commander precon, a promo set or a Secret Lair has no pack to draft.
var listedSetTypes = map[string]bool{
	"core":             true,
	"expansion":        true,
	"masters":          true,
	"draft_innovation": true,
}

// Sets is the draftable list, computed from the index and the catalog.
func Sets(idx *cards.Index) Response {
	if idx == nil {
		return Response{Sets: []SetStatus{}}
	}
	return sets(idx, CatalogIsFull(idx), overrides)
}

// sets is Sets with the catalog verdict and the override table injected.
func sets(idx *cards.Index, isFull IsFull, table map[string]Override) Response {
	found := collect(idx, nil)
	out := make([]SetStatus, 0, 64)
	for code, main := range found {
		var name, setType string
		dates := map[string]int{}
		for _, c := range main {
			if name == "" {
				name = c.SetName
			}
			if setType == "" {
				setType = c.SetType
			}
			if c.ReleasedAt != "" {
				dates[c.ReleasedAt]++
			}
		}
		released := commonDate(dates)
		if info, ok := idx.Set(code); ok {
			if info.Name != "" {
				name = info.Name
			}
			if info.SetType != "" {
				setType = info.SetType
			}
			if info.ReleasedAt != "" {
				released = info.ReleasedAt
			}
		}
		if !listedSetTypes[setType] {
			continue
		}
		r := recipeFor(table, code)
		var bonus map[uuid.UUID]cards.Card
		if r.BonusSheet != "" && r.BonusSheet != code {
			bonus = found[r.BonusSheet]
		}
		pool := buildPool(code, r, main, bonus)
		if !pool.CanFill() {
			continue
		}
		if name == "" {
			name = strings.ToUpper(code)
		}
		s := SetStatus{
			Code:     code,
			Name:     name,
			Released: released,
			Cards:    len(pool.Contents()),
			Blockers: []string{},
		}
		for _, b := range pool.Blockers(isFull) {
			s.Blockers = append(s.Blockers, b.Name)
		}
		s.Draftable = len(s.Blockers) == 0
		out = append(out, s)
	}
	sort.Slice(out, func(a, b int) bool {
		if out[a].Released != out[b].Released {
			return out[a].Released > out[b].Released
		}
		return out[a].Code < out[b].Code
	})
	resp := Response{Sets: out}
	for _, s := range out {
		if s.Draftable {
			resp.Draftable++
		}
	}
	return resp
}

// commonDate is the most frequent release date among a set's printings,
// the earliest on a tie (as the catalogue's set picker dates a set).
func commonDate(dates map[string]int) string {
	best, bestN := "", 0
	for d, n := range dates {
		if n > bestN || (n == bestN && d < best) {
			best, bestN = d, n
		}
	}
	return best
}

// Lister serves Sets for one index and remembers the answer until the
// index is reloaded. The catalog is fixed for the life of the binary, so
// the dump's load time is the only thing the list depends on.
type Lister struct {
	idx *cards.Index

	mu     sync.Mutex
	at     time.Time
	count  int
	cached Response
	have   bool
}

// NewLister returns a Lister over idx. A nil idx lists no sets.
func NewLister(idx *cards.Index) *Lister {
	return &Lister{idx: idx}
}

// Sets is Sets(idx), computed at most once per load of the index.
func (l *Lister) Sets() Response {
	if l == nil || l.idx == nil {
		return Response{Sets: []SetStatus{}}
	}
	at, count := l.idx.LoadedAt(), l.idx.Count()
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.have || !at.Equal(l.at) || count != l.count {
		l.cached = Sets(l.idx)
		l.at, l.count, l.have = at, count, true
	}
	return l.cached
}
