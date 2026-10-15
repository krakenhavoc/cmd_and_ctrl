package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lion Umbra (#3654): enchants a modified creature, +3/+3, reach and
// vigilance, and umbra armor.

// modifiedBear is a 2/2 with a +1/+1 counter on it, which makes it
// modified (CR 700.9) and so a legal host.
func modifiedBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Bear", TypeLine: testCreatureTypeLine,
		Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		Counters: map[string]int{"+1/+1": 1},
	})
}

func TestLionUmbraPumpsAndGivesReachAndVigilance(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := modifiedBear(g, me.ID)
	auraCast(t, g, "Lion Umbra", lionUmbraOracleID, bear)
	// 2/2 with a +1/+1 counter, then +3/+3. The counter is layer 7d,
	// which only the post-layer characteristics (CurrentPower) include.
	if c := layeredCard(t, g, bear); c.CurrentPower() != 6 || c.CurrentToughness() != 6 {
		t.Fatalf("enchanted bear is %d/%d, want 6/6", c.CurrentPower(), c.CurrentToughness())
	}
	assertEffectiveKeyword(t, g, bear, "reach")
	assertEffectiveKeyword(t, g, bear, "vigilance")
}

func TestLionUmbraReplacesADestroy(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := modifiedBear(g, me.ID)
	umbra := auraCast(t, g, "Lion Umbra", lionUmbraOracleID, bear)
	assertUmbraSavesFromDestroy(t, g, me, bear, umbra)
}
