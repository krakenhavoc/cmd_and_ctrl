package legal_test

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// exile_total_mana_value_cost_test.go — #5401: "{2}{W}, {T}, Exile one
// or more other artifacts you control with total mana value X: Return
// target artifact card with mana value X or less from your graveyard to
// the battlefield." The exiled artifacts' total IS the announced X, so
// the enumerator offers one payment per total the artifacts reach, and
// pairs each target only with the smallest total that admits it: a move
// that exiles more than its target needs is never offered. Every move is
// one the engine accepts (#544).

const oracleFabricationFoundry = "0bb20c36-689f-4a7c-80b5-a12bda821847"

type exileTotalPayload struct {
	XValue            int      `json:"x_value"`
	ExilePermanentIDs []string `json:"exile_permanent_ids"`
	Targets           []struct {
		ID string `json:"id"`
	} `json:"targets"`
}

func artifactCard(name, cost string) game.Card {
	return game.Card{Name: name, TypeLine: "Artifact", ManaCost: cost}
}

func TestFabricationFoundryOffersTheSmallestTotalForEachTarget(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	clearHand(seat)
	basics(g, seat, "Plains", 3)
	foundry := battlefieldCard(g, seat, game.Card{
		Name: "Fabrication Foundry", TypeLine: "Artifact", ManaCost: "{1}{W}", OracleID: oracleFabricationFoundry,
	})
	mv := map[string]int{}
	for _, a := range []struct {
		name, cost string
		v          int
	}{{"Treasure", "", 0}, {"Bauble", "{1}", 1}, {"Signet", "{2}", 2}, {"Obelisk", "{4}", 4}} {
		mv[battlefieldCard(g, seat, artifactCard(a.name, a.cost)).String()] = a.v
	}
	// An opponent's artifact never pays.
	theirs := battlefieldCard(g, opp, artifactCard("Their Relic", "{3}"))
	targets := map[string]int{}
	for _, c := range []struct {
		name, cost string
		v          int
	}{{"One Relic", "{1}", 1}, {"Three Relic", "{3}", 3}, {"Six Relic", "{6}", 6}, {"Nine Relic", "{9}", 9}} {
		targets[graveyardCard(seat, artifactCard(c.name, c.cost)).String()] = c.v
	}
	advanceTo(t, g, game.StepPrecombatMain)

	acts := activationsOf(legal.EnumerateFor(g, seat.ID), foundry)
	var xs []int
	for _, m := range acts {
		var p exileTotalPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		total := 0
		for _, id := range p.ExilePermanentIDs {
			if id == foundry.String() || id == theirs.String() {
				t.Errorf("move %q exiles %s, which cannot pay", m.Label, id)
			}
			total += mv[id]
		}
		if len(p.ExilePermanentIDs) == 0 || total != p.XValue {
			t.Errorf("move %q exiles %v (total %d) for x_value %d", m.Label, p.ExilePermanentIDs, total, p.XValue)
		}
		if len(p.Targets) != 1 {
			t.Fatalf("move %q has %d targets", m.Label, len(p.Targets))
		}
		if tv, ok := targets[p.Targets[0].ID]; !ok || tv > p.XValue {
			t.Errorf("move %q targets mana value %d at X = %d", m.Label, tv, p.XValue)
		}
		xs = append(xs, p.XValue)
	}
	sort.Ints(xs)
	// Totals 0..7 are reachable; the targets of mana value 1, 3 and 6
	// each pair with the smallest total that admits them, and nothing
	// reaches 9.
	if want := []int{1, 3, 6}; len(xs) != len(want) || xs[0] != 1 || xs[1] != 3 || xs[2] != 6 {
		t.Errorf("x values %v, want %v (one per target, smallest total)", xs, want)
	}
	dispatchAll(t, g, seat.ID, acts)
}

func TestFabricationFoundryIsNotOfferedWithNoOtherArtifact(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	basics(g, seat, "Plains", 3)
	foundry := battlefieldCard(g, seat, game.Card{
		Name: "Fabrication Foundry", TypeLine: "Artifact", ManaCost: "{1}{W}", OracleID: oracleFabricationFoundry,
	})
	graveyardCard(seat, artifactCard("Zero Relic", ""))
	advanceTo(t, g, game.StepPrecombatMain)
	if acts := activationsOf(legal.EnumerateFor(g, seat.ID), foundry); len(acts) != 0 {
		t.Errorf("offered with nothing but the Foundry to exile: %v", labels(acts))
	}
}
