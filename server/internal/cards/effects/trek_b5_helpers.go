package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b5_helpers.go — Star Trek, batch 5 (#4973): the pieces the batch's
// cards share.

// trekB5DifferentlyNamedLandsYouControl is the cost-side count for "Strange
// new worlds — this spell costs {X} less to cast, where X is the number of
// differently named lands you control": each land the caster controls,
// counted once per English name (socB23DifferentlyNamedLands, read off a
// CostQuery instead of a resolving item). The live battlefield, so an
// animated or type-changed land counts by its effective types.
func trekB5DifferentlyNamedLandsYouControl() func(q game.CostQuery) int {
	return func(q game.CostQuery) int {
		if q.Game == nil || q.Game.Battlefield == nil {
			return 0
		}
		names := map[string]bool{}
		for _, c := range q.Game.Battlefield.Cards {
			if c.InstanceID != q.Card.InstanceID && c.Controller == q.Controller && c.IsLand() {
				names[c.Name] = true
			}
		}
		return len(names)
	}
}

// trekB5CountersOnThis is the number of `kind` counters on the source of a
// resolving ability: live while it is on the battlefield, its last-known
// counters once it has left (CR 608.2h).
func trekB5CountersOnThis(g *game.Game, source uuid.UUID, kind string) int {
	if z := g.FindCardZoneForEffect(source); z != nil && z.Kind == game.ZoneBattlefield {
		return b39CountersOn(g, source, kind)
	}
	return g.LastKnownCountersForEffect(source)[kind]
}
