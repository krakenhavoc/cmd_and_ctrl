package game

import "github.com/google/uuid"

// venture_limit.go — "Each opponent can't venture into the dungeon more
// than once each turn" (Keen-Eared Sentry, #4100), read at one gate:
// VentureForEffect.
//
// A RESTRICTION, like a draw limit (draw_limit.go): a venture the player
// can't make is not performed — the marker does not move, no room
// ability triggers, no dungeon is entered or completed — and the rest of
// the instruction that asked for it goes on. "Venture into Undercity"
// (the initiative) is a venture into the dungeon (CR 701.49d), so it is
// counted and limited the same.
//
// THE COUNT is the ventures the player has made this turn: their
// EventVentureRoom events since the turn began (every venture that moved
// a marker emits exactly one). Ventures made before the permanent
// arrived count; "each turn" is every player's turn.
//
// THE SOURCES. Battlefield statics only, read live through
// CatalogAbilityKey, so a permanent that has lost its abilities gives
// nothing. Nothing is stored, so nothing is snapshotted.

// VentureLimitStatic is one printed venture limit on a permanent:
// "Each opponent can't venture into the dungeon more than Max times each
// turn". Catalog data, never stored.
type VentureLimitStatic struct {
	// Label is the clause as printed.
	Label string
	// Max is how many times a covered opponent may venture each turn.
	Max int
}

// CatalogVentureLimits returns the venture limits a permanent with the
// given catalog key has. carddef.go sets it from CardDef.VentureLimits;
// a game-package test may stub it.
var CatalogVentureLimits func(key string) []VentureLimitStatic

// ventureLimitReachedLocked reports whether `p` may not venture again
// this turn: an opponent of theirs controls a permanent with a venture
// limit they have reached. Caller must hold g.mu. Reads only.
func (g *Game) ventureLimitReachedLocked(p *Player) bool {
	if p == nil || CatalogVentureLimits == nil || g.Battlefield == nil {
		return false
	}
	limit, limited := 0, false
	for i := range g.Battlefield.Cards {
		src := &g.Battlefield.Cards[i]
		if src.Controller == uuid.Nil || src.Controller == p.ID {
			continue
		}
		key := catalogAbilityKeyOf(src)
		if key == "" {
			continue
		}
		for _, s := range CatalogVentureLimits(key) {
			if !limited || s.Max < limit {
				limit, limited = s.Max, true
			}
		}
	}
	return limited && g.venturesThisTurnLocked(p.ID) >= limit
}

// venturesThisTurnLocked counts `player`'s ventures this turn. Caller
// must hold g.mu.
func (g *Game) venturesThisTurnLocked(player uuid.UUID) int {
	n := 0
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind == EventVentureRoom && ev.Actor == player {
			n++
		}
	}
	return n
}
