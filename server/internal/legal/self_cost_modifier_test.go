package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// self_cost_modifier_test.go — #746, ADR 0048 addendum §14: the
// enumerator prices a spell's own cost modifiers the way the cast
// does, and prices per target set only when something reads targets.
// Every move offered here is dispatched against a clone
// (dispatchAll), which is the parity check: an enumerated cast the
// engine refuses is the #544 failure.

const (
	oracleThoughtMonitor = "9deded8b-cec4-4ede-a50b-131404d456d4"
	oracleFireball       = "aa7714b0-2bfb-458a-8ebf-37ec2c53383e"
	oraclePriceOfFame    = "45d5cd4c-7285-4507-8cbf-eace7a734f41"
)

// castsOf returns the cast moves for one source card, with their
// decoded target counts and X.
func castsOf(t *testing.T, moves []legal.Move, src uuid.UUID) []struct{ targets, x int } {
	t.Helper()
	var out []struct{ targets, x int }
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p struct {
			Targets []json.RawMessage `json:"targets"`
			XValue  int               `json:"x_value"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		out = append(out, struct{ targets, x int }{len(p.Targets), p.XValue})
	}
	return out
}

// Affinity reaches the enumerator: Thought Monitor at {6}{U} is
// offered off four lands when its caster controls three artifacts, and
// not without them.
func TestEnumeratorPricesAffinity(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	monitor := handCard(active, game.Card{Name: "Thought Monitor", TypeLine: "Artifact Creature — Construct", ManaCost: "{6}{U}", OracleID: oracleThoughtMonitor})
	battlefieldCard(g, active, basic("Island", "Island"))
	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, basic("Plains", "Plains"))
	}
	advanceTo(t, g, game.StepPrecombatMain)

	if got := castsOf(t, legal.EnumerateFor(g, active.ID), monitor); len(got) != 0 {
		t.Fatalf("Thought Monitor offered off four lands and no artifacts: %v", got)
	}
	for i := 0; i < 3; i++ {
		battlefieldCard(g, active, game.Card{Name: "Ornithopter", TypeLine: "Artifact Creature — Thopter"})
	}
	moves := legal.EnumerateFor(g, active.ID)
	if got := castsOf(t, moves, monitor); len(got) != 1 {
		t.Fatalf("Thought Monitor with three artifacts and four lands: %d casts, want 1: %v", len(got), labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// Fireball off two Mountains: zero or one target at X=1, two targets
// at X=0, three targets never. The unaffordable three-target sets
// spend no expansion budget, and every offered cast is accepted.
func TestEnumeratorPricesFireballPerTargetSet(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	fireball := handCard(active, game.Card{Name: "Fireball", TypeLine: "Sorcery", ManaCost: "{X}{R}", OracleID: oracleFireball})
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	battlefieldCard(g, active, basic("Mountain", "Mountain"))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	casts := castsOf(t, moves, fireball)
	if len(casts) == 0 {
		t.Fatalf("Fireball not offered off two Mountains: %v", labels(moves))
	}
	byCount := map[int]int{}
	for _, c := range casts {
		byCount[c.targets]++
		want := map[int]int{0: 1, 1: 1, 2: 0}
		x, ok := want[c.targets]
		if !ok {
			t.Errorf("Fireball offered at %d targets off two Mountains", c.targets)
			continue
		}
		if c.x != x {
			t.Errorf("Fireball at %d targets offered X=%d, want %d", c.targets, c.x, x)
		}
	}
	if byCount[1] != 4 {
		t.Errorf("Fireball single-target casts = %d, want one per player (4)", byCount[1])
	}
	if byCount[2] == 0 {
		t.Errorf("Fireball two-target casts at X=0 are affordable and missing: %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}

// A target-reading REDUCTION: Price of Fame off two lands is castable
// only at a legendary creature. A nil-targets price ({3}{B}) would
// have hidden it altogether.
func TestEnumeratorPricesPriceOfFameByTarget(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	clearHand(active)
	fame := handCard(active, game.Card{Name: "Price of Fame", TypeLine: "Instant", ManaCost: "{3}{B}", OracleID: oraclePriceOfFame})
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	battlefieldCard(g, active, basic("Swamp", "Swamp"))
	legend := battlefieldCard(g, opp, game.Card{Name: "Legend", TypeLine: "Legendary Creature — Human", Power: 2, Toughness: 2})
	battlefieldCard(g, opp, creature("Bear", "{1}{G}", 2, 2))
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	n := 0
	for _, m := range moves {
		if m.Source != fame {
			continue
		}
		n++
		var p struct {
			Targets []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if len(p.Targets) != 1 || p.Targets[0].ID != legend.String() {
			t.Errorf("Price of Fame offered off two lands at a non-legend: %s", m.Params)
		}
	}
	if n != 1 {
		t.Errorf("Price of Fame casts = %d, want exactly the one at the legend: %v", n, labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
}
