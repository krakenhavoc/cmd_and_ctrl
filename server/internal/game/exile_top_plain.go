package game

import "github.com/google/uuid"

// ExileTopKeyedForEffect is "exile the top N cards of <a player>'s
// library" with no permission stamped on them (Triple Triad: the play
// permission is worked out afterwards, once every library has been exiled
// from and the mana values can be compared). It is
// ExileTopWithPermissionKeyedForEffect's routing without a grant: the same
// plan and the same CR 614 window, with `then` told the cards that landed in
// exile once they have all settled, or run at once on an empty library. A
// zero `then` is no continuation. `actor` is whoever the exile is done for.
//
// Caller must hold g.mu (write).
func (g *Game) ExileTopKeyedForEffect(fromPlayer, actor uuid.UUID, n int, then KeyedTail) error {
	grant := KeyedTail{Ref: impulseExileGranted, Player: actor}
	if !then.IsZero() {
		next := copyKeyedTail(then)
		grant.Next = &next
	}
	plan, err := g.impulseExilePlanLocked(fromPlayer, actor, n)
	if err != nil {
		return err
	}
	if len(plan) == 0 {
		return grant.Run(g, nil)
	}
	return g.routeAllKeyedLocked(zoneRoute{Dst: ZoneExile, Actor: actor}, plan, grant)
}
