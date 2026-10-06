package suite

import (
	"context"
	"encoding/json"
	"flag"
	"math"
	"os"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
)

// baseline_test.go pins heuristic.BaselineConfig to the heuristic as
// it was before S66 (ADR 0126 §9).
//
// ADR 0126 changes the heuristic's prices one term at a time, and
// every new term must be zero in BaselineConfig, so that the arena's
// `heuristic-baseline` contestant keeps playing the old bot. This is
// the test that says it does: for every suite position it compares
// what BaselineConfig decides, and the full price list Rank puts on
// the moves, with the record taken before the first price changed.
//
// The record is testdata/baseline_rankings.json. It is written once,
// with
//
//	go test ./internal/aiseat/suite -run TestBaselineConfigRanksTheSuiteAsBefore -args -update-baseline
//
// and only rewritten when an ADR replaces the frozen baseline (ADR
// 0126 owner decision 5). A PR that adds a position adds its record
// the same way: the flag writes a record for a position that has
// none and never touches an existing one.

var updateBaseline = flag.Bool("update-baseline", false,
	"record BaselineConfig's ranking for suite positions that have no record yet (never rewrites one)")

const baselineRecord = "testdata/baseline_rankings.json"

// baselineEps is how far a recorded price may drift before it counts
// as a change. Prices are sums of float64 products; the tolerance is
// there for a compiler that fuses a multiply-add on another
// architecture, not for a tuning change, which moves a price by
// hundredths.
const baselineEps = 1e-9

type baselineCandidate struct {
	Index int     `json:"index"`
	Value float64 `json:"value"`
}

type baselineEntry struct {
	// Decision is Decide's index (aiseat.Decline when it declined).
	Decision int `json:"decision"`
	// Ranking is Rank's output, best first. Empty for the windows Rank
	// does not price (mulligans, combat, the opening roll).
	Ranking []baselineCandidate `json:"ranking,omitempty"`
}

func baselineOf(t *testing.T, p Position) baselineEntry {
	t.Helper()
	pol := heuristic.NewWithConfig(heuristic.BaselineConfig())
	d, err := pol.Decide(context.Background(), p.Input)
	if err != nil {
		t.Fatalf("%s: Decide: %v", p.ID, err)
	}
	e := baselineEntry{Decision: d.Index}
	for _, c := range heuristic.NewWithConfig(heuristic.BaselineConfig()).Rank(context.Background(), p.Input) {
		e.Ranking = append(e.Ranking, baselineCandidate{Index: c.Index, Value: c.Value})
	}
	return e
}

func TestBaselineConfigRanksTheSuiteAsBefore(t *testing.T) {
	positions := loadSuite(t)
	record := map[string]baselineEntry{}
	if blob, err := os.ReadFile(baselineRecord); err == nil {
		if err := json.Unmarshal(blob, &record); err != nil {
			t.Fatalf("%s: %v", baselineRecord, err)
		}
	} else if !os.IsNotExist(err) || !*updateBaseline {
		t.Fatalf("read %s: %v (record it with -args -update-baseline)", baselineRecord, err)
	}

	added := 0
	for _, p := range positions {
		got := baselineOf(t, p)
		want, ok := record[p.ID]
		if !ok {
			if *updateBaseline {
				record[p.ID] = got
				added++
				continue
			}
			t.Errorf("%s has no baseline record; run with -args -update-baseline to add one", p.ID)
			continue
		}
		if got.Decision != want.Decision {
			t.Errorf("%s: BaselineConfig decides move %d, the pre-S66 heuristic decided %d — a new pricing term is not zeroed in BaselineConfig",
				p.ID, got.Decision, want.Decision)
		}
		if len(got.Ranking) != len(want.Ranking) {
			t.Errorf("%s: BaselineConfig ranks %d moves, the record has %d", p.ID, len(got.Ranking), len(want.Ranking))
			continue
		}
		for i := range want.Ranking {
			g, w := got.Ranking[i], want.Ranking[i]
			if g.Index != w.Index || math.Abs(g.Value-w.Value) > baselineEps {
				t.Errorf("%s: rank %d is move %d at %.6f, the record has move %d at %.6f", p.ID, i, g.Index, g.Value, w.Index, w.Value)
			}
		}
	}
	for id := range record {
		found := false
		for _, p := range positions {
			if p.ID == id {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("%s: the baseline record names a position the suite no longer has", id)
		}
	}

	if *updateBaseline && added > 0 {
		blob, err := json.MarshalIndent(record, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(baselineRecord, append(blob, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
		t.Logf("recorded %d new position(s) in %s", added, baselineRecord)
	}
}

// TestBaselineConfigIsTheDefaultBeforeAnyS66Term pins the measurement
// PR's own claim: it adds no pricing term, so the baseline and the
// shipped tuning are one config. The first PR that adds a term
// deletes this test, in the same change that zeroes the term in
// BaselineConfig.
func TestBaselineConfigIsTheDefaultBeforeAnyS66Term(t *testing.T) {
	if heuristic.BaselineConfig() != heuristic.DefaultConfig() {
		t.Fatalf("BaselineConfig differs from DefaultConfig before any ADR 0126 term has landed")
	}
}
