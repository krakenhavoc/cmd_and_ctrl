package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func MustCost(t *testing.T, s string) game.ParsedCost {
	t.Helper()
	c, err := game.ParseCost(s)
	if err != nil {
		t.Fatalf("ParseCost(%q): %v", s, err)
	}
	return c
}

func PreviewFor(t *testing.T, g *game.Game, who uuid.UUID, cost string) ([]game.AutoTapPlanEntry, bool) {
	t.Helper()
	return g.AutoTapPlanPreferringExcluding(who, MustCost(t, cost), 0, nil, 0)
}

func PlanHas(plan []game.AutoTapPlanEntry, id uuid.UUID) bool {
	for _, e := range plan {
		if e.CardID == id {
			return true
		}
	}
	return false
}
