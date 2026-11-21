package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func QgCounters(g *game.Game, id uuid.UUID) map[string]int {
	c := FindBattlefieldCardForTest(g, id)
	if c == nil {
		return nil
	}
	return c.Counters
}

type QgMove struct {
	FromZone         string   `json:"from_zone"`
	AlternativeCost  string   `json:"alternative_cost"`
	CounterSourceIDs []string `json:"counter_source_ids"`
	CounterCounts    []int    `json:"counter_counts"`
	CounterKind      string   `json:"counter_kind"`
	CounterKinds     []string `json:"counter_kinds"`
}
