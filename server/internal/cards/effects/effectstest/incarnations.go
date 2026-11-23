package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func BearHasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	// The seeds below push straight onto a zone, which emits no
	// event, so nothing has bumped the layer version. Force the
	// recompute rather than relying on a staleness the test itself
	// never created — the INVALIDATION is pinned in
	// game/static_zones_test.go, where a real mill does the moving.
	g.BumpLayerVersionForTest()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("creature %s vanished", id)
	}
	return game.HasKeyword(&c, kw)
}
