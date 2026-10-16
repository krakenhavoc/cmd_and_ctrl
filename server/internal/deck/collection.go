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
