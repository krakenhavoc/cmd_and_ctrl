package draftpacks

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/metrics"
)

// archetypes.go is a set's ten colour pairs for the drafter: the gold
// uncommons that name each pair, read from card data, and a short
// written label for what the pair does in this set (ADR 0222 §7,
// amendment 2026-11-11, #5703). It is what Arena shows when you choose
// a draft.
//
// The signposts are computed: every uncommon of the set's pool whose
// colours are exactly two. The labels are written by hand, one per pair
// per set, in archetypes.json, from the set's signpost cards and the
// mechanics printed on its cards in those colours (the file's own
// header says how). A pair with no label is shown with its cards and no
// label: the data supports a pair, not a name, in a set with no clear
// ten (a small or unusual one).

// Pairs are the ten two-colour pairs, allied first, each spelled as its
// two colours in WUBRG order.
var Pairs = []string{"WU", "UB", "BR", "RG", "WG", "WB", "UR", "BG", "WR", "UG"}

//go:embed archetypes.json
var archetypeLabelsJSON []byte

// archetypeLabels is the embedded labels file: set code, then pair.
var archetypeLabels = mustLabels(archetypeLabelsJSON)

func mustLabels(raw []byte) map[string]map[string]string {
	var f struct {
		Sets map[string]map[string]string `json:"sets"`
	}
	if err := json.Unmarshal(raw, &f); err != nil {
		panic(fmt.Sprintf("draftpacks: archetypes.json: %v", err))
	}
	return f.Sets
}

// Signpost is a gold uncommon of a set: a card that names a colour pair.
// JSON field names are the client's contract; client/src/lib/draftSets.ts
// mirrors this struct.
type Signpost struct {
	// ID is the printing's Scryfall ID: what GET /draft/image/{id} takes.
	ID       uuid.UUID `json:"id"`
	OracleID uuid.UUID `json:"oracle_id"`
	Name     string    `json:"name"`
	ManaCost string    `json:"mana_cost"`
}

// Pair is one colour pair of a set.
type Pair struct {
	// Colors is the pair as two letters in WUBRG order ("WU").
	Colors string `json:"colors"`
	// Label is the pair's short written label ("Birds and flyers"),
	// absent where none is written.
	Label string `json:"label,omitempty"`
	// Signposts are the set's exactly-two-colour uncommons of this
	// pair, sorted by name.
	Signposts []Signpost `json:"signposts"`
}

// Archetypes is a set's ten colour pairs: the body of
// GET /draft/sets/{code}/archetypes.
type Archetypes struct {
	Set  string `json:"set"`
	Name string `json:"name"`
	// Pairs is always the ten pairs, in Pairs order.
	Pairs []Pair `json:"pairs"`
}

// pairOf is the two-letter pair a card's colours make, false unless the
// card has exactly two colours.
func pairOf(colors []string) (string, bool) {
	has := map[string]bool{}
	for _, c := range colors {
		has[c] = true
	}
	if len(has) != 2 {
		return "", false
	}
	var out strings.Builder
	for _, c := range "WUBRG" {
		if has[string(c)] {
			out.WriteRune(c)
		}
	}
	return out.String(), out.Len() == 2
}

// cardColors is the colours of a printing: the card's, or its front
// face's for a card whose faces carry the colours.
func cardColors(c cards.Card) []string {
	if len(c.Colors) > 0 {
		return c.Colors
	}
	if len(c.CardFaces) > 0 {
		return c.CardFaces[0].Colors
	}
	return nil
}

// ArchetypesOf is the colour pairs of a pool, read from the index, with
// the written labels of the set. name is the set's display name.
func ArchetypesOf(idx *cards.Index, p Pool, name string) Archetypes {
	return archetypesWith(idx, p, name, archetypeLabels[p.Set])
}

func archetypesWith(idx *cards.Index, p Pool, name string, labels map[string]string) Archetypes {
	by := map[string][]Signpost{}
	for _, c := range p.Uncommons {
		printing, ok := idx.Get(c.ID)
		if !ok {
			continue
		}
		pair, ok := pairOf(cardColors(printing))
		if !ok {
			continue
		}
		by[pair] = append(by[pair], Signpost{ID: c.ID, OracleID: c.OracleID, Name: c.Name, ManaCost: printing.ManaCost})
	}
	out := Archetypes{Set: p.Set, Name: name, Pairs: make([]Pair, 0, len(Pairs))}
	for _, pair := range Pairs {
		sp := by[pair]
		if sp == nil {
			sp = []Signpost{}
		}
		sort.Slice(sp, func(i, j int) bool { return sp[i].Name < sp[j].Name })
		out.Pairs = append(out.Pairs, Pair{Colors: pair, Label: labels[pair], Signposts: sp})
	}
	return out
}

// Archetypes is the colour pairs of a set a draft may be made of, with
// its row of the list. ok is false for a set that is not in the list or
// is not draftable: the verdict GET /draft/sets shows.
func (l *Lister) Archetypes(set string) (Archetypes, bool) {
	pool, status, ok := l.DraftablePool(set)
	if !ok {
		return Archetypes{}, false
	}
	return ArchetypesOf(l.idx, pool, status.Name), true
}

// ArchetypesHandler serves a set's colour pairs:
//
//	GET /draft/sets/{code}/archetypes
//
// It is public like GET /draft/sets, for the same reason: the body is
// card names and printings of the public catalogue, joined to labels
// written for this repository, and card data does not sit behind a
// sign-in (PLAN.md §3, ADR 0222 §9). It reads nothing from the request
// but the set code and stores nothing. A set that is not draftable, or
// is unknown, is a 404.
func ArchetypesHandler(lister *Lister) http.Handler {
	mux := metrics.NewServeMux()
	mux.HandleFunc("GET /draft/sets/{code}/archetypes", func(w http.ResponseWriter, r *http.Request) {
		a, ok := lister.Archetypes(r.PathValue("code"))
		if !ok {
			writeErr(w, http.StatusNotFound, "not a draftable set")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "public, max-age=300")
		_ = json.NewEncoder(w).Encode(a)
	})
	return mux
}
