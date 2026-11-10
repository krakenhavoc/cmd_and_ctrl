package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// search_all_zones_named.go — "search your graveyard, hand and/or library
// for a card named A and/or a card named B and put them onto the
// battlefield. If you search your library this way, shuffle." (Agency
// Outfitter, CR 701.23).
//
// Two prompts, because the library search is its own choice (CR 701.23a:
// the shuffle follows a library search, found or not, and a player may
// decline it): first "search your library too?", then ONE card pick over
// the candidates in the graveyard and hand and, if they said yes, the
// library, with at most one card per name. The cards named enter together
// (CR 603.6a), then the library is shuffled when it was searched. The
// pick is at least zero: a search may find nothing (CR 701.23b).

// searchAllZonesNamed starts the search for the controller. `names` are
// the card names printed on the card.
func searchAllZonesNamed(ctx *Context, question string, then game.ConfirmThen) error {
	return MayChoice{Question: question, Then: then}.Apply(ctx)
}

// searchAllZonesCandidates lists, in zone order, the cards named one of
// names in the player's graveyard, hand and (when library is set) library.
func searchAllZonesCandidates(g *game.Game, player uuid.UUID, names []string, library bool) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	zones := []*game.Zone{p.Graveyard, p.Hand}
	if library {
		zones = append(zones, p.Library)
	}
	var out []uuid.UUID
	for _, z := range zones {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			for _, n := range names {
				if c.Name == n {
					out = append(out, c.InstanceID)
					break
				}
			}
		}
	}
	return out
}

// oneCardPerName is the set rule "at most one card of each name".
func oneCardPerName(picked []game.Card) bool {
	seen := map[string]bool{}
	for _, c := range picked {
		if seen[c.Name] {
			return false
		}
		seen[c.Name] = true
	}
	return true
}

// searchAllZonesPut puts the picked cards onto the battlefield together
// from wherever they are, then shuffles when the library was searched.
func searchAllZonesPut(ctx *Context, picked []uuid.UUID, librarySearched bool) error {
	var entries []game.BatchEntry
	for _, id := range picked {
		z := ctx.Game.FindCardZoneForEffect(id)
		if z == nil {
			continue
		}
		entries = append(entries, game.BatchEntry{CardID: id, From: z.Kind})
	}
	if len(entries) > 0 {
		if err := ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(entries,
			game.ZoneEntryOptions{Controller: ctx.Controller()}, game.KeyedTail{}); err != nil {
			return err
		}
	}
	if librarySearched {
		return ctx.Game.ShuffleLibraryForEffect(ctx.Controller())
	}
	return nil
}
