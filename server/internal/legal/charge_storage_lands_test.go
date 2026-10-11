package legal_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// charge_storage_lands_test.go — #2951: the enumerator's half of Black
// Mana Battery, Bottomless Vault and Blast Zone, from the real catalog
// rows. dispatchAll is the assertion that matters: every enumerated
// move goes through the real dispatcher.

const (
	oracleBlackManaBattery = "dc3b2bab-755d-4b2f-97c0-70a74d721a79"
	oracleBottomlessVault  = "e43413e4-be17-49af-978a-26210d05f52a"
	oracleBlastZone        = "393a254f-be31-431a-9341-a51286f8cbce"
)

func countersOf(t *testing.T, m legal.Move) counterRestParams {
	t.Helper()
	return counterRestParamsOf(t, m)
}

// Black Mana Battery holding counters offers its payout as every
// counter it holds, and as banking when two mana are spare.
func TestBlackManaBatteryOffersThePayoutAndBanking(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bat := battlefieldCard(g, active, game.Card{
		Name: "Black Mana Battery", TypeLine: "Artifact", OracleID: oracleBlackManaBattery,
		Counters: map[string]int{"charge": 3},
	})
	mana(g, active, 2)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	payouts := movesFrom(moves, bat, legal.KindMana)
	if len(payouts) != 1 {
		t.Fatalf("want one payout, got %v", labels(payouts))
	}
	if p := countersOf(t, payouts[0]); len(p.CounterCounts) != 1 || p.CounterCounts[0] != 3 {
		t.Errorf("counter_counts = %v, want [3]", p.CounterCounts)
	}
	if len(activationsOf(moves, bat)) != 1 {
		t.Errorf("want the {2},{T} banking offered, got %v", labels(activationsOf(moves, bat)))
	}
	dispatchAll(t, g, active.ID, moves)
}

// With no counters Black Mana Battery still taps for its {B}: "any
// number" includes none, so the payout is offered with an empty
// payment, and the engine accepts it.
func TestBlackManaBatteryWithNoCountersStillTapsForBlack(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	bat := battlefieldCard(g, active, game.Card{
		Name: "Black Mana Battery", TypeLine: "Artifact", OracleID: oracleBlackManaBattery,
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	payouts := movesFrom(moves, bat, legal.KindMana)
	if len(payouts) != 1 {
		t.Fatalf("want the plain {B} offered, got %v", labels(payouts))
	}
	if p := countersOf(t, payouts[0]); len(p.CounterCounts) != 0 || len(p.CounterSourceIDs) != 0 {
		t.Errorf("an empty payment should carry no counter fields, got %+v", p)
	}
	dispatchAll(t, g, active.ID, payouts)
}

// A Bottomless Vault with no counters taps for nothing, so nothing is
// offered; with counters it is offered as every counter it holds.
func TestBottomlessVaultOffersOnlyAPayoutThatMakesMana(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	empty := battlefieldCard(g, active, game.Card{
		Name: "Bottomless Vault", TypeLine: "Land", OracleID: oracleBottomlessVault,
	})
	full := battlefieldCard(g, active, game.Card{
		Name: "Bottomless Vault", TypeLine: "Land", OracleID: oracleBottomlessVault,
		Counters: map[string]int{"storage": 2},
	})
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	if got := movesFrom(moves, empty, legal.KindMana); len(got) != 0 {
		t.Errorf("a Vault with no counters adds nothing and should not be offered, got %v", labels(got))
	}
	got := movesFrom(moves, full, legal.KindMana)
	if len(got) != 1 {
		t.Fatalf("want one payout, got %v", labels(got))
	}
	if p := countersOf(t, got[0]); len(p.CounterCounts) != 1 || p.CounterCounts[0] != 2 {
		t.Errorf("counter_counts = %v, want [2]", p.CounterCounts)
	}
	dispatchAll(t, g, active.ID, moves)
}

// Blast Zone offers its {X}{X} banking at the largest affordable X
// and its three-mana sweep, and the dispatcher accepts both.
func TestBlastZoneOffersBankingAndTheSweep(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	zone := battlefieldCard(g, active, game.Card{
		Name: "Blast Zone", TypeLine: "Land", OracleID: oracleBlastZone,
		Counters: map[string]int{"charge": 1},
	})
	mana(g, active, 6)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, zone)
	if len(acts) < 2 {
		t.Fatalf("want the banking and the sweep offered, got %v", labels(acts))
	}
	sawX := false
	for _, m := range acts {
		if xValueOf(t, m) == 3 {
			sawX = true
		}
	}
	if !sawX {
		t.Errorf("want the banking at X=3 on six spare mana, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}
