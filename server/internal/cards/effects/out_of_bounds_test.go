package effects

import "testing"

const outOfBoundsOracle = "431b2e52-1290-47b1-920d-2c1817553d96"

// TestOutOfBoundsCountersTheTargetedSpell: a spell an opponent casts is
// countered, and its controller takes no damage.
func TestOutOfBoundsCountersTheTargetedSpell(t *testing.T) {
	assistCardSpec(t, outOfBoundsOracle, "Out of Bounds")

	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bolt := batch01OpponentCasts(t, g, opp, "Lightning Bolt", lightningBoltOracle, "", b16TargetPlayer(me.ID))

	castCatalogSpell(t, g, "Out of Bounds", "Instant", outOfBoundsOracle, b16TargetCard(bolt))
	passPriorityAroundTable(t, g)

	if !opp.Graveyard.Contains(bolt) {
		t.Error("the countered spell is not in its owner's graveyard")
	}
	if me.Life != 40 {
		t.Errorf("caster at %d life, want 40: the countered spell did not resolve", me.Life)
	}
}
