package deck

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// collection.go — the card database as a player's collection (ADR 0198).
//
// A player whose deck seated no sideboard chooses "a card you own from
// outside the game" from every card there is, the casual reading of
// Scryfall's ruling ("a card you choose from outside the game comes from
// your personal collection"). This is game.OutsideCollectionSource over
// the loaded Scryfall index, installed by main.go. Every card it hands
// the game is built by toGameCard, the function a deck import uses, so a
// card taken from the collection is the card a deck list would have
// seated: same oracle ID (the catalog's key), same printing, same faces.
//
// It is also game.PlanarCardSource (ADR 0216 §2, amending ADR 0198):
// PlanarCards reads the planar rows the pool excludes, for a Planechase
// table's communal planar deck. The two readers never share a result.

// Collection implements game.OutsideCollectionSource over a card index.
type Collection struct {
	idx *cards.Index

	mu      sync.Mutex
	order   []uuid.UUID // oracle IDs in suggestion order
	builtAt time.Time   // the index load the order was built from
}

var _ game.OutsideCollectionSource = (*Collection)(nil)

// NewCollection returns the collection over idx.
func NewCollection(idx *cards.Index) *Collection {
	return &Collection{idx: idx}
}

// notInACollection are the type-line prefixes of cards that are never
// part of a deck or a game's pool: the Planechase, Archenemy and
// Conspiracy cards, vanguards and dungeons. A wish for "a card" must not
// hand a player a plane.
var notInACollection = []string{
	"Plane ", "Plane —", "Phenomenon", "Scheme", "Ongoing Scheme",
	"Conspiracy", "Vanguard", "Dungeon", "Emblem",
}

// collectible reports whether c is a card a player can own and play.
func collectible(c cards.Card) bool {
	if !c.Playable() || c.Name == "" {
		return false
	}
	for _, p := range notInACollection {
		if strings.HasPrefix(c.TypeLine, p) {
			return false
		}
	}
	return true
}

// Search returns up to limit collectible cards whose name matches q,
// best match first.
func (col *Collection) Search(q string, limit int) []game.Card {
	if col == nil || col.idx == nil || limit <= 0 {
		return nil
	}
	var out []game.Card
	// Over-fetch: the index's own search does not skip placeholders.
	for _, c := range col.idx.Search(q, limit*2) {
		if !collectible(c) {
			continue
		}
		out = append(out, toGameCard(c, false))
		if len(out) == limit {
			break
		}
	}
	return out
}

// Named returns the collectible card with this name, or false.
func (col *Collection) Named(name string) (game.Card, bool) {
	if col == nil || col.idx == nil {
		return game.Card{}, false
	}
	c, ok := col.idx.FindByName(name)
	if !ok || !collectible(c) {
		return game.Card{}, false
	}
	return toGameCard(c, false), true
}

// Suggest walks the collection in suggestion order: cards the engine
// automates (they have a catalog entry) before cards it does not, cards
// legal in Commander before cards that are not, and within each tier the
// most reprinted first, then by name. fn is called with each card until it
// returns false.
func (col *Collection) Suggest(fn func(game.Card) bool) {
	if col == nil || col.idx == nil {
		return
	}
	for _, id := range col.suggestionOrder() {
		c, ok := col.idx.FindByOracleID(id)
		if !ok || !collectible(c) {
			continue
		}
		if !fn(toGameCard(c, false)) {
			return
		}
	}
}

// suggestionOrder is the cached suggestion order, rebuilt when the index
// has been reloaded since it was built.
func (col *Collection) suggestionOrder() []uuid.UUID {
	col.mu.Lock()
	defer col.mu.Unlock()
	loaded := col.idx.LoadedAt()
	if col.order != nil && col.builtAt.Equal(loaded) {
		return col.order
	}
	type row struct {
		id        uuid.UUID
		name      string
		catalog   bool
		legal     bool
		printings int
	}
	var rows []row
	col.idx.EachOracle(func(c cards.Card, printings int) {
		if !collectible(c) {
			return
		}
		rows = append(rows, row{
			id:        c.OracleID,
			name:      c.Name,
			catalog:   game.IsCatalogCard != nil && game.IsCatalogCard(c.OracleID.String()),
			legal:     c.Legalities["commander"] == "legal",
			printings: printings,
		})
	})
	sort.Slice(rows, func(a, b int) bool {
		ra, rb := rows[a], rows[b]
		if ra.catalog != rb.catalog {
			return ra.catalog
		}
		if ra.legal != rb.legal {
			return ra.legal
		}
		if ra.printings != rb.printings {
			return ra.printings > rb.printings
		}
		return ra.name < rb.name
	})
	order := make([]uuid.UUID, len(rows))
	for i, r := range rows {
		order[i] = r.id
	}
	col.order = order
	col.builtAt = loaded
	return order
}

// PlanarCards is game.PlanarCardSource over the card database (ADR 0216
// §2): one game card per planar oracle ID of the asked kinds, built by
// toGameCard, in oracle-ID order so that minting, and so the game's
// object ordinals, are the same on every run of a seeded game.
//
// It is a second, separate reader of the index that selects exactly
// the planar rows collectible excludes (ADR 0198 §2, amended by ADR
// 0216): a printing is kept when it is Playable, its Scryfall layout is
// "planar", and its type line is a plane's ("Plane — …") or, when asked
// for, a phenomenon's. The outside-the-game pool never sees these
// cards, and this never sees any other.
func (col *Collection) PlanarCards(kinds game.PlanarKinds) []game.Card {
	if col == nil || col.idx == nil {
		return nil
	}
	var rows []cards.Card
	col.idx.EachOracle(func(c cards.Card, _ int) {
		if planarKindOf(c)&kinds != 0 {
			rows = append(rows, c)
		}
	})
	sort.Slice(rows, func(a, b int) bool { return rows[a].OracleID.String() < rows[b].OracleID.String() })
	out := make([]game.Card, len(rows))
	for i, c := range rows {
		out[i] = toGameCard(c, false)
	}
	return out
}

var _ game.PlanarCardSource = (*Collection)(nil)

// planarKindOf is the planar kind of a printing, or 0 for a card that
// is not a playable plane or phenomenon.
func planarKindOf(c cards.Card) game.PlanarKinds {
	if !c.Playable() || c.Name == "" || c.Layout != "planar" {
		return 0
	}
	switch {
	case strings.HasPrefix(c.TypeLine, "Plane —"), strings.HasPrefix(c.TypeLine, "Plane "):
		return game.PlanarPlanes
	case strings.HasPrefix(c.TypeLine, "Phenomenon"):
		return game.PlanarPhenomena
	}
	return 0
}
