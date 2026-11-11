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
// The pool is built to the set's own recipe, so a slot the recipe does
// not have cannot block: a draft booster from before mythic rares holds
// no mythic, and a bonus sheet counts only where the recipe has a bonus
// slot and names the sheet.
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
	// Booster is the kind of retail booster the set's packs approximate,
	// the draftable booster of the set's era: "play" or "draft".
	Booster Kind `json:"booster"`
	// Cards is how many different cards the set's packs can contain,
	// basic lands included when the booster has a land slot.
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
	resp, _ := setsAndPools(idx, isFull, table)
	return resp
}

// setsAndPools is sets, and the pool of every draftable set by code:
// what a draft draws its packs from (#5438).
func setsAndPools(idx *cards.Index, isFull IsFull, table map[string]Override) (Response, map[string]Pool) {
	pools := map[string]Pool{}
	found := collect(idx, nil)
	out := make([]SetStatus, 0, 64)
	for code, main := range found {
		meta := metaFor(idx, code, main)
		name, setType, released := meta.Name, meta.SetType, meta.Released
		if !listedSetTypes[setType] {
			continue
		}
		r := recipeFor(table, code, released)
		var bonus map[uuid.UUID]cards.Card
		if r.UsesBonusSheet() && r.BonusSheet != code {
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
			Booster:  r.Kind,
			Cards:    len(pool.Contents()),
			Blockers: []string{},
		}
		for _, b := range pool.Blockers(isFull) {
			s.Blockers = append(s.Blockers, b.Name)
		}
		s.Draftable = len(s.Blockers) == 0
		if s.Draftable {
			pools[code] = pool
		}
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
	return resp, pools
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
//
// It also remembers the pool of every draftable set, which is what a
// draft draws its packs from (DraftablePool), and which printings those
// pools hold, which is what the public image route serves
// (IsDraftPrinting).
type Lister struct {
	idx *cards.Index
	// isFull, when set, replaces the catalog's verdict. Tests only.
	isFull IsFull

	mu        sync.Mutex
	at        time.Time
	count     int
	cached    Response
	pools     map[string]Pool
	status    map[string]SetStatus
	printings map[uuid.UUID]bool
	// full is the catalog verdict the list was computed with.
	full IsFull
	have bool
}

// NewLister returns a Lister over idx. A nil idx lists no sets.
func NewLister(idx *cards.Index) *Lister {
	return &Lister{idx: idx}
}

// NewListerWith is NewLister with the catalog's verdict replaced, for a
// test that drafts a made-up set.
func NewListerWith(idx *cards.Index, isFull IsFull) *Lister {
	return &Lister{idx: idx, isFull: isFull}
}

// refreshLocked recomputes the list when the index has been reloaded.
func (l *Lister) refreshLocked() {
	at, count := l.idx.LoadedAt(), l.idx.Count()
	if l.have && at.Equal(l.at) && count == l.count {
		return
	}
	isFull := l.isFull
	if isFull == nil {
		isFull = CatalogIsFull(l.idx)
	}
	l.full = isFull
	l.cached, l.pools = setsAndPools(l.idx, isFull, overrides)
	l.status = make(map[string]SetStatus, len(l.cached.Sets))
	for _, s := range l.cached.Sets {
		l.status[s.Code] = s
	}
	l.printings = map[uuid.UUID]bool{}
	for _, p := range l.pools {
		for _, c := range p.Contents() {
			l.printings[c.ID] = true
		}
	}
	l.at, l.count, l.have = at, count, true
}

// Sets is Sets(idx), computed at most once per load of the index.
func (l *Lister) Sets() Response {
	if l == nil || l.idx == nil {
		return Response{Sets: []SetStatus{}}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refreshLocked()
	return l.cached
}

// DraftablePool is the pool of a set a draft may hand out, with its row
// of the list. ok is false for a set that is not in the list or is not
// draftable: the same verdict GET /draft/sets shows.
func (l *Lister) DraftablePool(set string) (Pool, SetStatus, bool) {
	if l == nil || l.idx == nil {
		return Pool{}, SetStatus{}, false
	}
	set = strings.ToLower(strings.TrimSpace(set))
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refreshLocked()
	p, ok := l.pools[set]
	return p, l.status[set], ok
}

// IsDraftPrinting reports whether id is a printing some draftable set's
// packs can contain: the scope of GET /draft/image/{id}.
func (l *Lister) IsDraftPrinting(id uuid.UUID) bool {
	if l == nil || l.idx == nil {
		return false
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refreshLocked()
	return l.printings[id]
}

// IsFull is the catalog verdict the list is computed with, by oracle
// ID: what a cube's cards are checked against (ADR 0222 §8), so a cube
// and a set are held to one answer. It is computed at most once per
// load of the index. A Lister with no index calls nothing Full.
func (l *Lister) IsFull() IsFull {
	if l == nil || l.idx == nil {
		return func(uuid.UUID) bool { return false }
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refreshLocked()
	return l.full
}

// IsCubePrinting reports whether id is a printing a cube draft can hand
// out: any printing of a card the catalog holds as Full, or of a basic
// land (ADR 0222 §8). A cube names its own printings, so they are not
// in a draftable set's pool; like that scope, this one is fixed by the
// dump and the catalog and holds no card the public catalogue does not
// already show.
func (l *Lister) IsCubePrinting(id uuid.UUID) bool {
	if l == nil || l.idx == nil {
		return false
	}
	c, ok := l.idx.Get(id)
	if !ok {
		return false
	}
	return isBasicLand(c) || l.IsFull()(c.OracleID)
}
