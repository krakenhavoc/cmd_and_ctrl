package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cant_have_counters_test.go — the #544 half of #4692. A Devoted Druid
// under Blossombind ("can't have counters put on it") can't pay its
// "Put a -1/-1 counter on this creature" cost (CR 118.3), so the
// enumerator does not offer the untap, and every move it does offer is
// accepted.

const (
	oracleBlossombindLegal  = "0ecb35b3-5bd1-4467-a4a3-ce468023ced2"
	oracleDevotedDruidLegal = "cb814e16-acf7-41d5-a357-1323dcc369f3"
)

func TestCantHaveCountersCostIsNeverOffered(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	clearHand(me)
	druid := battlefieldCard(g, me, game.Card{
		Name: "Devoted Druid", TypeLine: "Creature — Elf Druid", OracleID: oracleDevotedDruidLegal,
		Power: 0, Toughness: 2,
	})
	free := battlefieldCard(g, me, game.Card{
		Name: "Devoted Druid", TypeLine: "Creature — Elf Druid", OracleID: oracleDevotedDruidLegal,
		Power: 0, Toughness: 2,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	before := legal.EnumerateFor(g, me.ID)
	if len(movesOfKindFor(before, legal.KindActivate, druid)) == 0 {
		t.Fatalf("no Druid untap offered before Blossombind, so this test proves nothing: %v", labels(before))
	}

	aura := battlefieldCard(g, me, game.Card{
		Name: "Blossombind", TypeLine: "Enchantment — Aura", OracleID: oracleBlossombindLegal,
	})
	attachTo(t, g, aura, druid)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	if n := len(movesOfKindFor(moves, legal.KindActivate, druid)); n != 0 {
		t.Errorf("offered %d activations of a cost it can't pay: %v", n, labels(moves))
	}
	if len(movesOfKindFor(moves, legal.KindActivate, free)) == 0 {
		t.Errorf("the unenchanted Druid lost its untap: %v", labels(moves))
	}
	if err := g.Clone().ActivateCatalogAbility(me.ID, druid, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the engine accepted the withheld activation")
	}
}
