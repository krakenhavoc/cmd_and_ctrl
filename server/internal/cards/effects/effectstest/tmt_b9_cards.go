package effectstest

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- shared helpers ------------------------------------------------------------

// tmtB9Counters reads one counter kind on a battlefield permanent.
func TmtB9Counters(g *game.Game, id uuid.UUID, kind string) int {
	n := 0
	g.ReadSnapshot(func() { n = CountersOn(g, id, kind) })
	return n
}

// tmtB9Colors is a permanent's effective colours, upper-cased and joined.
func TmtB9Colors(t *testing.T, g *game.Game, id uuid.UUID) string {
	t.Helper()
	c := LayeredCard(t, g, id)
	out := ""
	for _, col := range c.EffectiveColors() {
		out += strings.ToUpper(col)
	}
	return out
}
