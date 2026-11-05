package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b1_helpers.go — shared helpers for Star Trek batch 1 (#4973).

// trekB1CountersNowOrLastKnown is how many `kind` counters a permanent
// has: now while it is on the battlefield, else as it last existed
// there (CR 608.2h).
func trekB1CountersNowOrLastKnown(g *game.Game, id uuid.UUID, kind string) int {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == id {
			return c.Counters[kind]
		}
	}
	if info, ok := g.LastKnownPermanentForEffect(id); ok {
		return info.Counters[kind]
	}
	return 0
}
