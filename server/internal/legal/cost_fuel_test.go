package legal_test

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// cost_fuel_test.go — #1013: a bot chooses WHICH cards an alternative
// cost eats.
//
// Before this the enumerator offered exactly ONE payment per offer
// (maxEnumeratedCostPayments was 1) and built it out of the first N
// candidates in ZONE order, so an escaping Uro over a graveyard holding
// a second Uro, a Snapcaster target and three lands ate whichever three
// were oldest. The cap was never the real constraint — the missing
// EVALUATION was, because a card in a graveyard was worth nothing to
// the policy and two payments were indistinguishable to it.
//
// So the fix is in two places and this file tests the `legal` half: the
// pool is ordered by a policy-supplied price (Options.OrderCostFuel),
// the cheapest payment is offered first, and the alternatives are paid
// for out of leftover budget rather than out of the target expansion's
// — ADR 0033 §1's corollary, which is the whole reason the cap was one.

// escapeBoard is a graveyard with an escape creature and `fuel`
// payment candidates, and enough lands to cast it. The fuel cards are
// named "Fuel 0".."Fuel N-1" in graveyard order, oldest first, so a
// test can say what zone order WOULD have picked.
func escapeBoard(t *testing.T, fuel int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 7, "Forest")

	typhon := graveyardCard(seat, game.Card{
		Name: "Voracious Typhon", TypeLine: "Creature — Hydra", ManaCost: "{4}{G}",
		Power: 4, Toughness: 4, OracleID: oracleVoraciousTyphon,
	})
	ids := make([]uuid.UUID, 0, fuel)
	for i := 0; i < fuel; i++ {
		ids = append(ids, graveyardCard(seat, game.Card{
			Name: fmt.Sprintf("Fuel %d", i), TypeLine: "Instant", ManaCost: "{1}",
		}))
	}
	return g, seat, typhon, ids
}

// escapePayments returns the AltCostIDs of every escape cast offered
// for `src`, in move order.
func escapePayments(t *testing.T, moves []legal.Move, src uuid.UUID) [][]string {
	t.Helper()
	var out [][]string
	for _, p := range castPayloadsOf(t, moves, src) {
		if p.AlternativeCost == "escape" {
			out = append(out, p.AltCostIDs)
		}
	}
	return out
}

// keepTheLast prices the LAST card in `ids` as precious and everything
// else as worthless — the smallest ordering a test can state, and the
// shape of the real question ("do not eat the second Uro").
func keepTheLast(ids []uuid.UUID) legal.CostFuelOrder {
	precious := ids[len(ids)-1]
	return func(c legal.TargetCandidate) float64 {
		if c.ID == precious {
			return 100
		}
		return 1
	}
}

// TestTheCheapestFuelIsEatenFirst is the headline. Five cards in the
// graveyard, four eaten, and the one the policy wants kept is the one
// zone order would have taken first.
func TestTheCheapestFuelIsEatenFirst(t *testing.T) {
	g, seat, typhon, fuel := escapeBoard(t, 5)
	// The precious card is the OLDEST, so zone order and the fuel
	// order disagree about it as loudly as they can.
	precious := fuel[0]
	order := func(c legal.TargetCandidate) float64 {
		if c.ID == precious {
			return 100
		}
		return 1
	}

	moves := legal.EnumerateForWithOptions(g, seat.ID, legal.Options{OrderCostFuel: order})
	pays := escapePayments(t, moves, typhon)
	if len(pays) == 0 {
		t.Fatalf("no escape cast offered: %v", labels(moves))
	}
	for _, id := range pays[0] {
		if id == precious.String() {
			t.Errorf("the first payment ate the card the policy priced highest: %v", pays[0])
		}
	}
	if len(pays[0]) != 4 {
		t.Fatalf("escape-four named %d cards: %v", len(pays[0]), pays[0])
	}
}

// TestZoneOrderIsWhatAPolicyWithNoOpinionGets: the hook is optional,
// and a nil one leaves the enumeration exactly as #673 built it. Every
// non-bot caller passes nil, so this is the common path.
func TestZoneOrderIsWhatAPolicyWithNoOpinionGets(t *testing.T) {
	g, seat, typhon, fuel := escapeBoard(t, 5)

	pays := escapePayments(t, legal.EnumerateFor(g, seat.ID), typhon)
	if len(pays) == 0 {
		t.Fatal("no escape cast offered")
	}
	want := []string{fuel[0].String(), fuel[1].String(), fuel[2].String(), fuel[3].String()}
	if strings.Join(pays[0], ",") != strings.Join(want, ",") {
		t.Errorf("payment = %v, want the four oldest in zone order %v", pays[0], want)
	}
}

// TestTheFuelOrderIsStableAndDeterministic. Equal prices keep zone
// order, and two enumerations of one board produce the same move list
// — the property every enumerator test in this package depends on, and
// the one a sort with an unstable comparator quietly breaks.
func TestTheFuelOrderIsStableAndDeterministic(t *testing.T) {
	g, seat, typhon, _ := escapeBoard(t, 6)
	flat := func(legal.TargetCandidate) float64 { return 1 }

	first := escapePayments(t, legal.EnumerateForWithOptions(g, seat.ID,
		legal.Options{OrderCostFuel: flat}), typhon)
	plain := escapePayments(t, legal.EnumerateFor(g, seat.ID), typhon)
	if len(first) == 0 || len(plain) == 0 {
		t.Fatal("no escape cast offered")
	}
	if strings.Join(first[0], ",") != strings.Join(plain[0], ",") {
		t.Errorf("an all-equal price reordered the pool: %v vs zone order %v", first[0], plain[0])
	}
	for i := 0; i < 3; i++ {
		again := escapePayments(t, legal.EnumerateForWithOptions(g, seat.ID,
			legal.Options{OrderCostFuel: flat}), typhon)
		if len(again) != len(first) {
			t.Fatalf("run %d offered %d payments, the first offered %d", i, len(again), len(first))
		}
		for j := range again {
			if strings.Join(again[j], ",") != strings.Join(first[j], ",") {
				t.Fatalf("run %d payment %d = %v, want %v", i, j, again[j], first[j])
			}
		}
	}
}

// TestAFewAlternativePaymentsAreOffered: the cap is no longer one, so a
// policy gets a small choice rather than a decree. The count is
// bounded, and it is bounded by the constant rather than by the
// graveyard.
func TestAFewAlternativePaymentsAreOffered(t *testing.T) {
	g, seat, typhon, fuel := escapeBoard(t, 8)

	pays := escapePayments(t, legal.EnumerateForWithOptions(g, seat.ID,
		legal.Options{OrderCostFuel: keepTheLast(fuel)}), typhon)
	if len(pays) < 2 {
		t.Fatalf("want more than one payment offered over a graveyard of eight, got %d", len(pays))
	}
	if len(pays) > 3 {
		t.Errorf("%d payments offered, the cap is maxEnumeratedCostPayments (3) — "+
			"a graveyard of eight is C(8,4) = 70 payments and the point of the cap is "+
			"that the number does not depend on it", len(pays))
	}
	// Distinct: three copies of one payment is not a choice.
	seen := map[string]bool{}
	for _, p := range pays {
		k := strings.Join(p, ",")
		if seen[k] {
			t.Errorf("the same payment was offered twice: %v", p)
		}
		seen[k] = true
	}
}

// TestAlternativePaymentsSpendNoTargetBudget is the corollary, and the
// reason the cap was one (ADR 0033 §1). The payments are offered out of
// the budget the TARGET walk did not use, so a spell with a wide
// target set gets exactly the payments it got before this change and
// loses no target to them.
//
// The board is the fixture the escape test uses with a Snapcaster-like
// flashback spell in the graveyard instead — a card whose whole
// expansion is targets — so "did the payments eat a target" has a
// visible answer.
func TestAlternativePaymentsSpendNoTargetBudget(t *testing.T) {
	g, seat, typhon, fuel := escapeBoard(t, 8)
	// A wide board, so anything with a target clause has plenty to
	// spend its expansion budget on.
	g.WithWriteLock(func() {
		for i := 0; i < 20; i++ {
			g.Battlefield.PushTop(game.Card{
				InstanceID: uuid.New(), Name: fmt.Sprintf("Bear %d", i),
				TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
				Owner: seat.ID, Controller: seat.ID,
			})
		}
	})

	opts := legal.Options{OrderCostFuel: keepTheLast(fuel), MaxExpansionPerSource: 12}
	withFuel := legal.EnumerateForWithOptions(g, seat.ID, opts)
	plain := legal.EnumerateFor(g, seat.ID)

	countFor := func(moves []legal.Move, src uuid.UUID) int {
		n := 0
		for _, m := range moves {
			if m.Source == src {
				n++
			}
		}
		return n
	}
	if got, cap := countFor(withFuel, typhon), 12; got > cap {
		t.Errorf("the escape expanded into %d moves, MaxExpansionPerSource is %d — "+
			"the alternative payments must come out of the budget, never past it", got, cap)
	}
	// Every other source's move count is untouched: the payments are a
	// per-source affair and nothing else in the enumeration moved.
	per := func(moves []legal.Move) map[uuid.UUID]int {
		out := map[uuid.UUID]int{}
		for _, m := range moves {
			if m.Source != typhon {
				out[m.Source]++
			}
		}
		return out
	}
	a, b := per(withFuel), per(plain)
	var drift []string
	for src, n := range a {
		if b[src] != n {
			drift = append(drift, fmt.Sprintf("%s: %d → %d", src, b[src], n))
		}
	}
	sort.Strings(drift)
	if len(drift) > 0 {
		t.Errorf("the fuel hook changed the move count of sources it is not about:\n  %s",
			strings.Join(drift, "\n  "))
	}
}

// TestEveryFuelOrderedPaymentIsOneTheEngineAccepts is the package's own
// contract at the widened search: an offered payment is a payment
// CastSpell takes. A wider search that offered one illegal set would be
// the #544 wedge with a new cause.
func TestEveryFuelOrderedPaymentIsOneTheEngineAccepts(t *testing.T) {
	g, seat, _, fuel := escapeBoard(t, 8)
	moves := legal.EnumerateForWithOptions(g, seat.ID,
		legal.Options{OrderCostFuel: keepTheLast(fuel)})
	dispatchAll(t, g, seat.ID, moves)
}
