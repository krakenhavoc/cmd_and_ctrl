package testso

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
)

const outOfBoundsOracle = "431b2e52-1290-47b1-920d-2c1817553d96"

// TestOutOfBoundsCountersTheTargetedSpell: a spell an opponent casts is
// countered, and its controller takes no damage.
func TestOutOfBoundsCountersTheTargetedSpell(t *testing.T) {
	effectstest.AssistCardSpec(t, outOfBoundsOracle, "Out of Bounds")

	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolt := effectstest.Batch01OpponentCasts(t, g, opp, "Lightning Bolt", effectstest.LightningBoltOracle, "", effectstest.B16TargetPlayer(me.ID))

	effectstest.CastCatalogSpell(t, g, "Out of Bounds", "Instant", outOfBoundsOracle, effectstest.B16TargetCard(bolt))
	effectstest.PassPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
	if me.Life != 40 {
		t.Errorf("caster at %d life, want 40: the countered spell did not resolve", me.Life)
	}
}
