package game

import "github.com/google/uuid"

// search_cant.go — "Your opponents can't search libraries" (Stranglehold,
// S82 older sets, #5462).
//
// CR 701.23a: to search a zone is to look at its cards and find one. A
// player who can't search a library does not look and finds nothing, and the
// instructions around the search still happen (CR 608.2c): a "then shuffle"
// shuffles, and what a search's tail does with "the cards found" does it
// with none. No EventSearchLibrary is emitted, because no library was
// searched (the "libraryUntouched" reading of an and/or search).
//
// Like Aven Mindcensor's top-N narrowing (search_top_n.go) it lives in the
// one search primitive, SearchLibraryThenForEffect, so every search the
// catalog makes is covered by construction. The searcher is
// SearchLibrarySpec.Player, who does the looking; the permanent's controller
// must be their opponent and still in the game. The graveyard or hand half of
// an and/or search is not a library and is not blocked.

// CatalogOpponentsCantSearch reports whether a battlefield permanent has the
// printed "your opponents can't search libraries", keyed by
// CatalogAbilityKey. Populated from effects.Spec.OpponentsCantSearchLibraries.
// A nil hook finds none.
var CatalogOpponentsCantSearch func(key string) bool

// searchCantLocked reports whether `searcher` can't search libraries because
// an opponent's permanent says so. Judged through catalogAbilityKeyOf, so a
// permanent that lost its abilities blocks nothing (CR 613.1f).
//
// Caller must hold g.mu.
func (g *Game) searchCantLocked(searcher uuid.UUID) bool {
	if g.Battlefield == nil || CatalogOpponentsCantSearch == nil || searcher == uuid.Nil {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == uuid.Nil || c.Controller == searcher || !g.playerInGameLocked(c.Controller) {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && CatalogOpponentsCantSearch(key) {
			return true
		}
	}
	return false
}

// skipBlockedSearchLocked is a library search its searcher can't make: the
// shuffle the card prints still happens, and the card's tail runs with
// nothing found.
//
// Caller must hold g.mu.
func (g *Game) skipBlockedSearchLocked(spec SearchLibrarySpec, p *Player) error {
	if spec.Shuffle {
		g.shuffleLibraryLocked(p, ShuffleByEffect)
	}
	return g.runSearchThenLocked(spec, nil)
}
