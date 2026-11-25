package legal

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// graveyardGroupPayments is the one payment the enumerator offers for a
// PayActionGraveyardToLibraryBottom (#6121, Jötun Grunt's "two cards from
// a single graveyard" per age counter): `count` picks in whole groups of
// `group`, each group from one graveyard, so the engine's validator
// accepts it (#544). The opponents' graveyards are taken first, in seat
// order, then the seat's own; within a graveyard the pool's order. Nil
// when the graveyards can't supply it in whole groups.
func graveyardGroupPayments(g *game.Game, seat uuid.UUID, pool []uuid.UUID, count, group int) [][]uuid.UUID {
	if group < 1 {
		group = 1
	}
	if count < 1 || count%group != 0 {
		return nil
	}
	byOwner := map[uuid.UUID][]uuid.UUID{}
	var owners []uuid.UUID
	for _, id := range pool {
		c, ok := g.LookupCardForEffect(id)
		if !ok {
			continue
		}
		if _, seen := byOwner[c.Owner]; !seen {
			owners = append(owners, c.Owner)
		}
		byOwner[c.Owner] = append(byOwner[c.Owner], id)
	}
	ordered := make([]uuid.UUID, 0, len(owners))
	for _, o := range owners {
		if o != seat {
			ordered = append(ordered, o)
		}
	}
	if _, ok := byOwner[seat]; ok {
		ordered = append(ordered, seat)
	}
	var set []uuid.UUID
	for _, o := range ordered {
		cards := byOwner[o]
		for len(cards) >= group && len(set) < count {
			set = append(set, cards[:group]...)
			cards = cards[group:]
		}
	}
	if len(set) != count {
		return nil
	}
	return [][]uuid.UUID{set}
}
