package game

import (
	"math/rand/v2"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// mana_runs_test.go is #3041: the pool stores runs of identical mana,
// not one token per mana, and every helper must answer exactly what the
// one-token-per-mana pool answered (expansion equivalence).

func TestAddManaMergesOnlyTheSameMana(t *testing.T) {
	src, other := uuid.New(), uuid.New()
	eldrazi := []string{ManaRestrictSubtype("Eldrazi")}
	var p ManaPool
	p.AddMana(
		ManaToken{Color: "G", Source: src},
		ManaToken{Color: "G", Source: src},
		ManaToken{Color: "G", Source: src},
	)
	if len(p) != 1 || p[0].Count != 3 || p.Total() != 3 {
		t.Fatalf("three {G} from one source = %+v, want one run of three", p)
	}
	// Each of these differs from the run in one respect a reader can
	// tell apart, so each starts a new run.
	p.AddMana(
		ManaToken{Color: "R", Source: src},
		ManaToken{Color: "G", Source: other},
		ManaToken{Color: "G", Source: other, SourceKinds: ManaSourceTreasure},
		ManaToken{Color: "G", Source: other, SourceKinds: ManaSourceTreasure, Restrictions: eldrazi},
		ManaToken{Color: "G", Source: other, SourceKinds: ManaSourceTreasure, Restrictions: append([]string(nil), eldrazi...)},
	)
	if len(p) != 5 || p.Total() != 8 {
		t.Fatalf("pool = %+v, want five runs holding eight mana", p)
	}
	if p[4].N() != 2 {
		t.Errorf("two restricted {G} with equal restrictions = %+v, want one run of two", p[4])
	}
	// Insertion order: a {G} from the first source after the {R} does
	// not rejoin the first run.
	p.AddMana(ManaToken{Color: "G", Source: src})
	if len(p) != 6 || p[0].N() != 3 {
		t.Errorf("a later {G} joined an earlier run: %+v", p)
	}
	wantColors := "GGGRGGGGG"
	var got strings.Builder
	for _, u := range p.Units() {
		got.WriteString(u.Color)
	}
	if got.String() != wantColors {
		t.Errorf("Units() = %s, want %s", got.String(), wantColors)
	}
}

func TestManaWithSpendRidersNeverMerges(t *testing.T) {
	src := uuid.New()
	rider := []ManaSpendRider{{Kind: ManaRiderCantBeCountered, Production: uuid.New()}}
	var p ManaPool
	p.AddMana(
		ManaToken{Color: "G", Source: src, Riders: copyManaRiders(rider)},
		ManaToken{Color: "G", Source: src, Riders: copyManaRiders(rider)},
		ManaToken{Color: "G", Source: src},
		ManaToken{Color: "G", Source: src},
	)
	if len(p) != 3 || p[0].N() != 1 || p[1].N() != 1 || p[2].N() != 2 {
		t.Fatalf("pool = %+v, want the two rider mana apart and the plain pair as one run", p)
	}
}

// A run of one is written as Count 0, so a single token and its JSON
// are what they were before runs, and an older restore point (no count
// key) reads as one mana per token.
func TestARunOfOneHasNoCount(t *testing.T) {
	var p ManaPool
	p.AddMana(ManaToken{Color: "U"})
	if p[0].Count != 0 {
		t.Errorf("a lone mana has Count %d, want 0", p[0].Count)
	}
	cost, _ := ParseCost("{U}")
	p.AddMana(ManaToken{Color: "U"})
	if !p.SpendMana(cost, 0) {
		t.Fatal("{U}{U} could not pay {U}")
	}
	if len(p) != 1 || p[0].Count != 0 {
		t.Errorf("after spending one of two = %+v, want a run of one written as Count 0", p)
	}
}

func TestSpendingTakesPartOfARunAndRecordsIt(t *testing.T) {
	marwyn := uuid.New()
	var p ManaPool
	for range 10 {
		p.AddMana(ManaToken{Color: "G", Source: marwyn, SourceKinds: ManaSourceCreature})
	}
	cost, _ := ParseCost("{X}{G}")
	spent, ok := p.SpendManaFor(cost, 3, ManaSpendContext{Purpose: SpendPurposeCast})
	if !ok {
		t.Fatal("ten {G} could not pay {3}{G}")
	}
	if len(p) != 1 || p.Total() != 6 {
		t.Errorf("pool = %+v, want one run of six", p)
	}
	if len(spent) != 1 || spent[0].N() != 4 || spent[0].Source != marwyn {
		t.Errorf("spent = %+v, want one run of four from the Marwyn", spent)
	}
	paid := PaidCost{Mana: spent}
	if paid.ManaSpentCount() != 4 || paid.SpentOfColor("G") != 4 || paid.Spent().CountFrom(ManaSourceCreature) != 4 {
		t.Errorf("record reads total %d, green %d, from creatures %d; want 4 each",
			paid.ManaSpentCount(), paid.SpentOfColor("G"), paid.Spent().CountFrom(ManaSourceCreature))
	}
	if ev := manaSpentEvent(uuid.Nil, uuid.Nil, spent); ev.Amount != 4 {
		t.Errorf("EventManaSpent.Amount = %d, want 4", ev.Amount)
	}
}

// Restricted mana stays its own run: a payment it may not make leaves
// it untouched, and one it may make burns it first, restriction-first,
// exactly as single tokens did.
func TestARestrictedRunStaysDistinct(t *testing.T) {
	temple, forest := uuid.New(), uuid.New()
	var p ManaPool
	for range 3 {
		p.AddMana(ManaToken{Color: "C", Source: temple, Restrictions: []string{ManaRestrictSubtype("Eldrazi")}})
	}
	for range 3 {
		p.AddMana(ManaToken{Color: "C", Source: forest})
	}
	if len(p) != 2 {
		t.Fatalf("pool = %+v, want the restricted and the plain {C} as two runs", p)
	}
	cost, _ := ParseCost("{4}")
	bolt := ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Instant"}}
	if p.CanPayFor(cost, 0, bolt) {
		t.Error("an instant paid {4} with mana only an Eldrazi may spend")
	}
	eldrazi := ManaSpendContext{Purpose: SpendPurposeCast, Types: []string{"Creature"}, Subtypes: []string{"Eldrazi"}}
	spent, ok := p.SpendManaFor(cost, 0, eldrazi)
	if !ok {
		t.Fatal("an Eldrazi could not pay {4} out of six {C}")
	}
	if len(spent) != 2 || spent[0].Source != temple || spent[0].N() != 3 || spent[1].Source != forest || spent[1].N() != 1 {
		t.Errorf("spent = %+v, want the whole restricted run first, then one plain {C}", spent)
	}
	if len(p) != 1 || p[0].Source != forest || p[0].N() != 2 {
		t.Errorf("pool = %+v, want two plain {C} left", p)
	}
}

// Converge and sunburst count colours, so SpendDistinctColors spends a
// fresh colour one mana at a time and only then takes the rest in bulk.
func TestDistinctColorsOverRuns(t *testing.T) {
	var p ManaPool
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		src := uuid.New()
		for range 100 {
			p.AddMana(ManaToken{Color: c, Source: src})
		}
	}
	cost, _ := ParseCost("{X}")
	spent, ok := p.SpendManaForWith(cost, 7, ManaSpendContext{}, SpendDistinctColors)
	if !ok {
		t.Fatal("a 500-mana pool could not pay {7}")
	}
	if got := (PaidCost{Mana: spent}).ColorsSpentCount(); got != 5 {
		t.Errorf("converge counted %d colours, want 5", got)
	}
	if p.Total() != 493 {
		t.Errorf("pool holds %d, want 493", p.Total())
	}
}

// expanded is the pool one entry per mana, built without AddMana's
// merge: the pool the solver saw before #3041.
func expanded(p ManaPool) ManaPool {
	return ManaPool(unitsOf(p))
}

// TestRunsSpendExactlyWhatSingleTokensSpent is the equivalence itself,
// over random pools and costs: the solver on runs and the solver on the
// same mana one token each agree on payable, on what is spent, on what
// is left and on what is missing, in order, for both strategies.
func TestRunsSpendExactlyWhatSingleTokensSpent(t *testing.T) {
	rng := rand.New(rand.NewPCG(3041, 1))
	sources := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	colors := []string{"W", "U", "B", "R", "G", "C"}
	restrictions := [][]string{nil, nil, nil, {ManaRestrictSubtype("Elf")}, {ManaRestrictNoGeneric}}
	symbols := []string{"{W}", "{U}", "{B}", "{R}", "{G}", "{C}", "{G/U}", "{W/B}", "{R/P}", "{1}", "{2}", "{X}"}
	contexts := []ManaSpendContext{
		{},
		{Purpose: SpendPurposeCast, Types: []string{"Creature"}, Subtypes: []string{"Elf"}},
		{Purpose: SpendPurposeCast, Types: []string{"Sorcery"}},
	}
	for trial := range 3000 {
		var p ManaPool
		for range rng.IntN(8) {
			tok := ManaToken{
				Color:        colors[rng.IntN(len(colors))],
				Source:       sources[rng.IntN(len(sources))],
				Restrictions: restrictions[rng.IntN(len(restrictions))],
			}
			for range 1 + rng.IntN(4) {
				p.AddMana(tok)
			}
		}
		var b strings.Builder
		for range rng.IntN(5) {
			b.WriteString(symbols[rng.IntN(len(symbols))])
		}
		cost, err := ParseCost(b.String())
		if err != nil {
			t.Fatalf("ParseCost(%q): %v", b.String(), err)
		}
		if len(cost.Required) > 0 && rng.IntN(4) == 0 {
			cost.Required[rng.IntN(len(cost.Required))].AnyMana = true
		}
		x := rng.IntN(6)
		ctx := contexts[rng.IntN(len(contexts))]
		flat := expanded(p)
		for _, strategy := range []ManaSpendStrategy{SpendPreserveColors, SpendDistinctColors} {
			gotLeft, gotSpent, gotOK := p.attemptSpend(cost, x, ctx, strategy)
			wantLeft, wantSpent, wantOK := flat.attemptSpend(cost, x, ctx, strategy)
			if gotOK != wantOK ||
				!reflect.DeepEqual(unitsOf(gotLeft), unitsOf(wantLeft)) ||
				!reflect.DeepEqual(unitsOf(gotSpent), unitsOf(wantSpent)) {
				t.Fatalf("trial %d, %q X=%d strategy %d, pool %+v:\nruns   ok=%v spent=%+v left=%+v\ntokens ok=%v spent=%+v left=%+v",
					trial, cost.String(), x, strategy, p, gotOK, gotSpent, gotLeft, wantOK, wantSpent, wantLeft)
			}
		}
		if got, want := p.MissingFor(cost, x, ctx), flat.MissingFor(cost, x, ctx); !reflect.DeepEqual(got, want) {
			t.Fatalf("trial %d, %q X=%d, pool %+v: MissingFor runs %v, tokens %v", trial, cost.String(), x, p, got, want)
		}
		if got, want := poolShortfalls(p, cost, x, ctx), poolShortfalls(flat, cost, x, ctx); !reflect.DeepEqual(got, want) {
			t.Fatalf("trial %d, %q X=%d, pool %+v: poolShortfalls runs %v, tokens %v", trial, cost.String(), x, p, got, want)
		}
	}
}

// The point of #3041: a pool of a million mana from one source is one
// run, so paying out of it, copying it and asking what is missing cost
// what a pool of one costs.
func TestAMillionManaIsOneRunAndCheapToPay(t *testing.T) {
	marwyn := uuid.New()
	var p ManaPool
	for range 1 << 20 {
		p.AddMana(ManaToken{Color: "G", Source: marwyn, SourceKinds: ManaSourceCreature})
	}
	if len(p) != 1 || p.Total() != 1<<20 {
		t.Fatalf("pool = %d runs holding %d, want one run of 2^20", len(p), p.Total())
	}
	cost, _ := ParseCost("{X}{G}{G}")
	allocs := testing.AllocsPerRun(20, func() {
		if !p.CanPayFor(cost, 1<<20-2, ManaSpendContext{}) {
			t.Fatal("2^20 {G} could not pay {X}{G}{G} with X = 2^20-2")
		}
	})
	if allocs > 20 {
		t.Errorf("CanPayFor on a one-run pool allocated %.0f times, want a handful", allocs)
	}
	if missing := p.MissingFor(cost, 1<<20, ManaSpendContext{}); len(missing) != 2 {
		t.Errorf("missing %v, want two {1}", missing)
	}
}

// CR 106.4's sweep loses a run whole and reports its mana, not its
// runs.
func TestTheSweepCountsLostManaNotRuns(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	src := uuid.New()
	for range 5 {
		me.ManaPool.AddMana(ManaToken{Color: "G", Source: src})
	}
	me.ManaPool.AddMana(ManaToken{Color: "R"})
	var lost int
	g.WithWriteLock(func() { lost = g.sweepManaPoolLocked(me) })
	if lost != 6 || len(me.ManaPool) != 0 {
		t.Errorf("sweep lost %d (pool %+v), want 6 and an empty pool", lost, me.ManaPool)
	}
}
