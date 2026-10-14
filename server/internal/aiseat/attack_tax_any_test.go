package aiseat_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// TestTheBotSeesAnAnyDefenderTax — #3456: the bot's attack moves carry
// War Tax's price in their cost, which is what the heuristic's
// AttackTaxPenalty reads, and with no mana none is offered.
func TestTheBotSeesAnAnyDefenderTax(t *testing.T) {
	g, _, them := propagandaTable(t, 4)
	g.WithWriteLock(func() {
		g.AnyAttackTaxForEffect(uuid.Nil, "{1}", game.Duration{}, "War Tax test")
	})

	moves := legal.EnumerateFor(g, g.Seats[0].ID)
	found := false
	for _, m := range moves {
		if m.Kind != legal.KindAttack {
			continue
		}
		found = true
		if m.Cost == nil || m.Cost.Mana == "" {
			t.Errorf("attack %q carries no price", m.Label)
		}
	}
	if !found {
		t.Fatalf("no attack offered with four lands (target %s)", them)
	}
}
