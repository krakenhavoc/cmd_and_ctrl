package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// the_ozolith_test.go — #1218: the card proof for the counters-at-
// departure half of "Post-departure LKI / 'exiled with this' record"
// (docs/engine-seams.md). Every collection test destroys the creature
// with no other event in between the departure and the read, which is
// the property that was missing: MoveCard zeroes Card.Counters as
// part of the same mutation that emits the departure event, so a
// watcher reading the card AFTER the move sees none regardless of
// what it actually had.

const OzolithOracle = "1946ded1-5f53-409f-b0a6-5433bb0357d2"

// seedCreatureWithCounters puts a creature on the battlefield with
// the given counters, under owner's control.
func SeedCreatureWithCounters(g *game.Game, owner uuid.UUID, counters map[string]int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Counter Bearer",
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
		Counters:   counters,
	})
	return id
}

// countersOn reads the live Counters map off a battlefield card.
func AllCountersOn(t *testing.T, g *game.Game, id uuid.UUID) map[string]int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	return c.Counters
}
