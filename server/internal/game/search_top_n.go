package game

import "github.com/google/uuid"

// search_top_n.go — "If an opponent would search a library, that player
// searches the top four cards of that library instead" (Aven Mindcensor,
// #6090).
//
// CR 701.23a: "To search for a card in a zone, look at all cards in that
// zone (even if it's a hidden zone) and find a card that matches the given
// description." The static is a replacement (CR 614.1) of that look: the
// searcher looks at the top N cards of the library instead of all of it.
// Everything else about the search is unchanged: it is still a search (so
// "whenever a player searches" fires and "then shuffle" shuffles the WHOLE
// library), and the graveyard or hand half of an and/or search is not a
// library and is not narrowed.
//
// Like Opposition Agent's statics (search_control.go) it lives in the one
// search primitive, SearchLibraryThenForEffect, so every search the catalog
// makes is covered by construction: whose library ("a library", so
// Bribery's search of an opponent's library too), every destination and
// every count. The searcher is SearchLibrarySpec.Player, who does the
// looking; the permanent's controller must be their opponent and still in
// the game.
//
// More than one such permanent: each would replace the search with a search
// of the top N, and once one applies the others' replacement changes
// nothing more for an equal N; for unequal ones the searcher, as the
// affected player, would order them (CR 616.1), and the smallest is what
// any order ends on. So the smallest N wins.

// CatalogOpponentsSearchTopN reports the N of a battlefield permanent's
// "if an opponent would search a library, that player searches the top N
// cards of that library instead", keyed by CatalogAbilityKey. Zero is
// "no such ability". Populated from effects.Spec.OpponentsSearchTopN. A
// nil hook finds none.
var CatalogOpponentsSearchTopN func(key string) int

// searchTopNLocked is how many cards from the top of a library `searcher`
// looks at when they search one: 0 for the whole library. Judged through
// catalogAbilityKeyOf, so a permanent that lost its abilities narrows
// nothing (CR 613.1f).
//
// Caller must hold g.mu.
func (g *Game) searchTopNLocked(searcher uuid.UUID) int {
	if g.Battlefield == nil || CatalogOpponentsSearchTopN == nil || searcher == uuid.Nil {
		return 0
	}
	best := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == uuid.Nil || c.Controller == searcher || !g.playerInGameLocked(c.Controller) {
			continue
		}
		key := catalogAbilityKeyOf(c)
		if key == "" {
			continue
		}
		n := CatalogOpponentsSearchTopN(key)
		if n > 0 && (best == 0 || n < best) {
			best = n
		}
	}
	return best
}

// searchableLibraryCards is the part of lib a search by `searcher` looks
// at: the whole library, or its top N cards (the end of the slice is the
// top, Zone.Top) under an opponent's Aven Mindcensor.
//
// Caller must hold g.mu.
func (g *Game) searchableLibraryCards(searcher uuid.UUID, lib []Card) []Card {
	n := g.searchTopNLocked(searcher)
	if n <= 0 || n >= len(lib) {
		return lib
	}
	return lib[len(lib)-n:]
}
