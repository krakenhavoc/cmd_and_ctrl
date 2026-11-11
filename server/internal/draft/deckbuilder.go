package draft

import (
	"sort"
	"strings"
)

// LimitedDeck is a deck built from one seat's card pool (ADR 0222 §5,
// #5440): the pool cards it plays, by their draft ID, and the basic
// lands added to them. Every other card of the pool is the sideboard
// (CR 100.4b).
type LimitedDeck struct {
	// Main is the IDs of the pool cards in the deck.
	Main []string `json:"main"`
	// Basics is how many of each basic land the deck adds from outside
	// the pool, by name (BasicLandNames). Any number is legal (CR
	// 100.2b).
	Basics map[string]int `json:"basics,omitempty"`
}

// Size is how many cards the deck holds.
func (d LimitedDeck) Size() int {
	n := len(d.Main)
	for _, k := range d.Basics {
		n += k
	}
	return n
}

func (d LimitedDeck) clone() LimitedDeck {
	out := LimitedDeck{Main: append([]string{}, d.Main...)}
	if len(d.Basics) > 0 {
		out.Basics = make(map[string]int, len(d.Basics))
		for k, v := range d.Basics {
			out.Basics[k] = v
		}
	}
	return out
}

// BasicLandNames is the basic lands a deck may add, in colour order
// (W, U, B, R, G).
var BasicLandNames = []string{"Plains", "Island", "Swamp", "Mountain", "Forest"}

// basicOfColor is the basic land that makes a colour of mana.
var basicOfColor = map[string]string{"W": "Plains", "U": "Island", "B": "Swamp", "R": "Mountain", "G": "Forest"}

// IsBasicLandName reports whether name is a basic land a deck may add.
func IsBasicLandName(name string) bool {
	for _, n := range BasicLandNames {
		if n == name {
			return true
		}
	}
	return false
}

// DeckBuilder builds a deck from a card pool for a seat no person is
// building for: a bot seat, and a person whose build timer ran out.
//
// A draft given no builder uses BotDeckBuilder, the draft bot's deck
// (#5439). A deck that fails the
// limited check is replaced with SimpleDeckBuilder's, so a draft never
// stalls on a builder. A DeckBuilder is called with the draft's lock
// held, so it must not block and must not call back into the draft.
type DeckBuilder interface {
	BuildLimitedDeck(pool []Card) LimitedDeck
}

// BotDeckBuilder is the DeckBuilder a draft uses when it is given none:
// the draft bot's BuildDeck (#5439), which takes the pool's two best
// colours by rating, the best cards under a curve check, and lands split
// by colour weight. The seed is the draft's, so a seeded draft builds the
// same decks every time.
type BotDeckBuilder struct{ Seed uint64 }

// BuildLimitedDeck builds the bot's deck.
func (b BotDeckBuilder) BuildLimitedDeck(pool []Card) LimitedDeck {
	main, lands := BuildDeck(pool, StandardBasics, b.Seed)
	return LimitedDeck{Main: main, Basics: lands}
}

// The shape of the default deck: 23 cards and 17 lands make 40.
const (
	simpleDeckSpells = 23
	simpleDeckLands  = 17
)

// SimpleDeckBuilder is the fallback DeckBuilder, and deliberately plain:
// the pool's two most common colours, its 23 cheapest playable cards,
// and 17 basic lands split by colour. A playable card is one that is
// not a land and has no colour outside the two.
type SimpleDeckBuilder struct{}

// BuildLimitedDeck builds the plain deck.
func (SimpleDeckBuilder) BuildLimitedDeck(pool []Card) LimitedDeck {
	isLand := func(c Card) bool { return c.Basic || strings.Contains(frontType(c.TypeLine), "Land") }

	// The two most common colours, a card counting once for each of its
	// colours. Ties go to the earlier colour in WUBRG order.
	count := map[string]int{}
	for _, c := range pool {
		if isLand(c) {
			continue
		}
		for _, col := range c.Colors {
			count[col]++
		}
	}
	order := []string{"W", "U", "B", "R", "G"}
	sort.SliceStable(order, func(i, j int) bool { return count[order[i]] > count[order[j]] })
	chosen := map[string]bool{order[0]: true, order[1]: true}

	var playable []Card
	for _, c := range pool {
		if isLand(c) {
			continue
		}
		ok := true
		for _, col := range c.Colors {
			if !chosen[col] {
				ok = false
			}
		}
		if ok {
			playable = append(playable, c)
		}
	}
	sort.SliceStable(playable, func(i, j int) bool { return playable[i].ManaValue < playable[j].ManaValue })
	if len(playable) > simpleDeckSpells {
		playable = playable[:simpleDeckSpells]
	}

	deck := LimitedDeck{Main: make([]string, 0, len(playable)), Basics: map[string]int{}}
	weight := map[string]int{}
	for _, c := range playable {
		deck.Main = append(deck.Main, c.ID)
		for _, col := range c.Colors {
			weight[col]++
		}
	}
	// Enough lands to make 40, which is 17 when the pool gave 23 cards
	// to play; split by how many of the deck's cards are each colour.
	lands := max(simpleDeckLands, simpleDeckSpells+simpleDeckLands-len(playable))
	a, b := order[0], order[1]
	total := weight[a] + weight[b]
	first := (lands + 1) / 2
	if total > 0 {
		first = (lands*weight[a] + total/2) / total
		// A colour the deck plays gets at least a few lands.
		if weight[a] > 0 && weight[b] > 0 {
			first = min(max(first, 3), lands-3)
		}
	}
	if first > 0 {
		deck.Basics[basicOfColor[a]] = first
	}
	if lands-first > 0 {
		deck.Basics[basicOfColor[b]] = lands - first
	}
	return deck
}

// frontType is the front face's type line of a two-faced card.
func frontType(typeLine string) string {
	front, _, _ := strings.Cut(typeLine, "//")
	return front
}
