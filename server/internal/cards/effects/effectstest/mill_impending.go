package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// mill_impending_test.go — #1117 batch E4: Barrowgoyf, Six and
// Overlord of the Balemurk, plus the impending keyword the Overlord
// brought with it.

const (
	BarrowgoyfOracle = "74c0164c-130f-4572-9066-626c77f6e2ff"
	SixOracle        = "dbcbdf37-c40f-4068-b4a7-a849cab1056c"
	BalemurkOracle   = "652e87af-7cf0-407f-9fb7-1a630fb8dd47"
)

// miCounters reads a battlefield card's counter pile.
func MiCounters(t *testing.T, g *game.Game, id uuid.UUID, kind string) int {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s is not in the game", id)
	}
	return c.Counters[kind]
}

// miInHand reports whether the named player holds the card.
func MiInHand(p *game.Player, id uuid.UUID) bool {
	for _, c := range p.Hand.Cards {
		if c.InstanceID == id {
			return true
		}
	}
	return false
}
