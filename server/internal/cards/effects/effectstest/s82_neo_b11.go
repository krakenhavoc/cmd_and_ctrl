package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attachForTest attaches an already-placed permanent to a host.
func AttachForTest(t *testing.T, g *game.Game, attachment, host uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.AttachForEffect(attachment, game.TargetRef{Kind: game.TargetCard, ID: host}); err != nil {
			t.Fatalf("AttachForEffect: %v", err)
		}
		g.RecomputeLayersIfStaleLocked()
	})
}

// setCounters sets a permanent's counters and refreshes the layers.
func S82SetCounters(g *game.Game, id uuid.UUID, counters map[string]int) {
	g.WithWriteLock(func() {
		FindBattlefieldCardForTest(g, id).Counters = counters
		g.BumpLayerVersionForTest()
		g.RecomputeLayersIfStaleLocked()
	})
}
