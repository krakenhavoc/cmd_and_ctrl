package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_shield_grants_test.go — ADR 0106 PR 3 (#1806): the turn
// grants, the one-use promises and the mark on one spell, end to end
// through the real catalog. The engine contract (sources 3 and 5 of the
// counter gate, the CR 601.2i spend, copies, restore points) is pinned
// with hand-built spells in game/counter_shield_grants_test.go.

const (
	CgVeilOfSummerOracle    = "002965be-a36f-4a09-9ce0-c6535bca1703"
	CgOvermasterOracle      = "acc842a4-26e8-435a-b717-c1043b581ff8"
	CgVexingShusherOracle   = "a20a7cf8-2075-47ad-9229-36264b112e61"
	CgInsistOracle          = "518684d9-467a-4936-a3d9-17d23db4622c"
	CgDomriOracle           = "afc2269c-d3b5-487d-9445-800c7a8e526b"
	CgMistriseVillageOracle = "339f5334-b65a-445a-a016-20e997e0b4bb"
	CgBoundDeterminedOracle = "df39fbee-6ddc-4733-a32e-cdaf369ecccf"
)

// cgShields is the seat's live grants and promises.
func CgShields(g *game.Game, p *game.Player) []game.CounterShieldGrantSource {
	var out []game.CounterShieldGrantSource
	g.WithWriteLock(func() { out = g.CounterShieldGrantsForEffect(p) })
	return out
}

// cgResolveTop passes priority until the top item of the stack has
// resolved, leaving everything under it in place.
func CgResolveTop(t *testing.T, g *game.Game) {
	t.Helper()
	n := len(g.StackMeta)
	for i := 0; i < 8 && len(g.StackMeta) >= n; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if len(g.StackMeta) >= n {
		t.Fatal("the top of the stack never resolved")
	}
}
