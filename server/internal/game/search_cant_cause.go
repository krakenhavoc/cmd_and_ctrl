package game

import "github.com/google/uuid"

// search_cant_cause.go — "Spells and abilities your opponents control can't
// cause their controller to search their library" (Ashiok, Dream Render,
// S82 old sets).
//
// It is a prohibition (CR 101.2) on the spell or ability's own controller
// searching their own library, not on a search a spell makes someone else
// do: Assassin's Trophy's victim, who is not the controller, still searches,
// and so does a controller searching an opponent's library (Bribery). Like
// Opposition Agent's statics it sits in the one search primitive,
// SearchLibraryThenForEffect, so every search the catalog makes is covered.
// The search does not happen: no look, no find, no shuffle and no
// "whenever a player searches" event.

// CatalogOpponentsCantCauseSearch reports whether a battlefield permanent
// with the given CatalogAbilityKey has the printed static above. Populated
// from effects.Spec.OpponentsCantCauseSearch. A nil hook finds none.
var CatalogOpponentsCantCauseSearch func(key string) bool

// searchCausedByForbiddenLocked reports whether the resolving spell or
// ability, controlled by `searcher`, may not make them search their own
// library because an opponent controls such a permanent. A permanent that
// lost its abilities forbids nothing (catalogAbilityKeyOf, CR 613.1f).
//
// Caller must hold g.mu.
func (g *Game) searchCausedByForbiddenLocked(searcher uuid.UUID) bool {
	if g.Battlefield == nil || CatalogOpponentsCantCauseSearch == nil || searcher == uuid.Nil {
		return false
	}
	if g.effectCauseControllerLocked() != searcher {
		return false
	}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller == uuid.Nil || c.Controller == searcher || !g.playerInGameLocked(c.Controller) {
			continue
		}
		if key := catalogAbilityKeyOf(c); key != "" && CatalogOpponentsCantCauseSearch(key) {
			return true
		}
	}
	return false
}
