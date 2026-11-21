package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// counter_unless_paid_halt_test.go — #951 at the cards.
//
// The bug the issue reports is not visible in ward_test.go and that is
// the point of this file: those tests answer the prompt one pass at a
// time, which happens to dodge the interleaving. The bug needs a pass
// to arrive WHILE the prompt is outstanding — the ordinary case at a
// table with more than two seats, since any other player passing is
// enough. All four seats of newCatalogGame are used on purpose.

// theThirdSeatPasses is the move that used to kill a warded permanent
// for free: a player with no stake in the exchange passing priority
// while the payer has not answered.
func TheThirdSeatPasses(t *testing.T, g *game.Game) error {
	t.Helper()
	return g.PassPriority()
}

// passUntilTaxed passes priority until `chooser` owes a pay-unless,
// stopping the moment one appears (or the gate refuses a pass, which
// is itself a prompt being open).
func PassUntilTaxed(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 8 && !HasPayUnlessFor(g, chooser); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority %d: %v", i, err)
		}
	}
	if !HasPayUnlessFor(g, chooser) {
		t.Fatalf("no pay-unless prompt ever reached %s", chooser)
	}
}
