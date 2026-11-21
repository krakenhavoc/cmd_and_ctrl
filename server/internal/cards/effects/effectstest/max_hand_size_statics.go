package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0113 §3 (#2074): statics that set or change a maximum hand size,
// with the printed cards. Each test is one interaction the ADR lists.
// The engine fold itself is tested with stubs in
// game/max_hand_size_test.go.

const (
	MhJinGitaxiasOracle      = "eb23aed0-c450-4e57-96f2-2866dceca004"
	MhNullProfusionOracle    = "7077ee9a-59c6-4b8e-bf33-9b094412ad38"
	MhPriceOfKnowledgeOracle = "1c586aa7-7a61-464f-abba-b33f9a525f0e"
)

// mhTickingClock makes every timestamp distinct and increasing, so the
// order the test puts permanents onto the battlefield is the CR 613.7
// order.
func MhTickingClock(t *testing.T) {
	t.Helper()
	now := int64(1_000_000)
	restore := game.SetClockForTest(func() int64 { now += 10; return now })
	t.Cleanup(restore)
}

func MhMax(g *game.Game, p *game.Player) int {
	var n int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		n = g.EffectiveMaxHandSizeLocked(p)
	})
	return n
}
