package suite

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// harvest_test.go builds a three-record decision log by hand — Layer
// A, a Layer C window where the model agreed with the heuristic, and
// one where it did not — and checks that --disagree finds the third
// and only the third. That filter is what makes labelling affordable:
// a game writes thousands of windows and the handful where the two
// layers pulled apart are the ones a human can usefully arbitrate.

const harvestGame = "6f1b2a3c-4d5e-6f70-8192-a3b4c5d6e7f8"

func idx(i int) *int { return &i }

// record builds one log line.
func record(seat, seq int, layer string, heuristic int, parsed *int, moves []legal.Move, stack bool) decisionlog.Record {
	in := aiseat.Input{
		Seat:  uuid.MustParse("11111111-1111-1111-1111-111111111111"),
		Moves: moves,
		View: protocol.GameView{
			ID:   harvestGame,
			Turn: protocol.TurnView{Number: 4, Step: "precombat_main"},
		},
	}
	if stack {
		in.View.StackItems = []protocol.StackItemView{{}}
	}
	rec := decisionlog.Record{
		V: decisionlog.RecordVersion, Game: harvestGame, SeatIndex: seat,
		Seq: uint64(seq), Turn: 4, Step: "precombat_main",
		Policy: "assisted", Input: &in,
		Trace: aiseat.Trace{Layer: layer, HeuristicIndex: heuristic, ParsedIndex: parsed},
	}
	if layer == model.LayerC {
		rec.Trace.Model = "qwen3:14b"
	}
	return rec
}

func writeLog(t *testing.T, dir string, recs ...decisionlog.Record) string {
	t.Helper()
	path := filepath.Join(dir, harvestGame+".decisions.jsonl")
	f, err := os.Create(path) //nolint:gosec // t.TempDir
	if err != nil {
		t.Fatalf("create log: %v", err)
	}
	defer func() { _ = f.Close() }()
	enc := json.NewEncoder(f)
	for _, r := range recs {
		if err := enc.Encode(r); err != nil {
			t.Fatalf("encode: %v", err)
		}
	}
	return path
}

func harvestMoves() []legal.Move {
	return []legal.Move{
		{Kind: legal.KindPass, Type: "pass_priority", Label: "Pass priority", AlwaysLegal: true},
		{Kind: legal.KindLand, Type: "play_land", Label: "Play Mountain"},
		{Kind: legal.KindCast, Type: "cast_spell", Label: "Cast Bear"},
	}
}

func TestHarvestDisagreeFindsTheOneWindowWorthLabelling(t *testing.T) {
	logDir, outDir := t.TempDir(), t.TempDir()
	path := writeLog(t, logDir,
		// A: Layer A settled it. No model, no disagreement.
		record(0, 10, model.LayerA, aiseat.Decline, nil, harvestMoves(), false),
		// B: the model agreed with the heuristic.
		record(0, 20, model.LayerC, 1, idx(1), harvestMoves(), false),
		// C: the model wanted a different move, on a window with
		// something on the stack.
		record(1, 30, model.LayerC, 1, idx(2), harvestMoves(), true),
	)

	rep, err := Harvest([]string{path}, HarvestFilter{Disagree: true}, outDir)
	if err != nil {
		t.Fatalf("harvest: %v", err)
	}
	if rep.Records != 3 || rep.Full != 3 {
		t.Fatalf("scanned %d records (%d full), want 3/3", rep.Records, rep.Full)
	}
	if rep.Candidates != 1 || rep.Written != 1 {
		t.Fatalf("--disagree kept %d candidates and wrote %d, want 1/1", rep.Candidates, rep.Written)
	}

	positions, err := Load(outDir)
	if err != nil {
		t.Fatalf("load the inbox: %v", err)
	}
	p := positions[0]
	if want := "6f1b2a3c-s1-seq30"; p.ID != want {
		t.Errorf("id %q, want %q", p.ID, want)
	}
	if p.Labelled() {
		t.Errorf("a harvested position must arrive UNLABELLED; %s has %d accept matchers", p.ID, len(p.Expected.Accept))
	}
	// Auto-tags come off the move kinds, plus stack-response from the
	// view. Pass, mana and activate are deliberately not tags: they
	// are in almost every window.
	if got, want := p.Tags, []string{"cast", "land", "stack-response"}; !equalStrings(got, want) {
		t.Errorf("auto-tags %v, want %v", got, want)
	}
	if p.AtCapture.Heuristic == nil || p.AtCapture.Heuristic.Index != 1 || p.AtCapture.Heuristic.Label != "Play Mountain" {
		t.Errorf("at_capture.heuristic = %+v, want index 1 Play Mountain", p.AtCapture.Heuristic)
	}
	if p.AtCapture.Model == nil || p.AtCapture.Model.Index != 2 || p.AtCapture.Model.ID != "qwen3:14b" || p.AtCapture.Model.Layer != model.LayerC {
		t.Errorf("at_capture.model = %+v, want index 2 on qwen3:14b at layer C", p.AtCapture.Model)
	}
	if p.Source.Seat != 1 || p.Source.Seq != 30 || p.Source.Log != path {
		t.Errorf("source = %+v, want seat 1 seq 30 from %s", p.Source, path)
	}
}

func TestHarvestFiltersAndSampling(t *testing.T) {
	logDir := t.TempDir()
	path := writeLog(t, logDir,
		record(0, 10, model.LayerA, aiseat.Decline, nil, harvestMoves(), false),
		record(0, 20, model.LayerC, 1, idx(1), harvestMoves(), false),
		record(1, 30, model.LayerC, 1, idx(2), harvestMoves(), true),
	)

	t.Run("escalated drops the Layer A window", func(t *testing.T) {
		rep, err := Harvest([]string{path}, HarvestFilter{Escalated: true}, t.TempDir())
		if err != nil {
			t.Fatalf("harvest: %v", err)
		}
		if rep.Written != 2 {
			t.Fatalf("wrote %d, want 2", rep.Written)
		}
	})
	t.Run("seat", func(t *testing.T) {
		rep, err := Harvest([]string{path}, HarvestFilter{Seats: []int{1}}, t.TempDir())
		if err != nil {
			t.Fatalf("harvest: %v", err)
		}
		if rep.Written != 1 {
			t.Fatalf("wrote %d, want 1", rep.Written)
		}
	})
	t.Run("tag", func(t *testing.T) {
		rep, err := Harvest([]string{path}, HarvestFilter{Tags: []string{"stack-response"}}, t.TempDir())
		if err != nil {
			t.Fatalf("harvest: %v", err)
		}
		if rep.Written != 1 {
			t.Fatalf("wrote %d, want 1", rep.Written)
		}
	})
	// The same logs and the same seed have to produce the same inbox,
	// or a harvested corpus cannot be reviewed.
	t.Run("sampling is deterministic", func(t *testing.T) {
		var first []string
		for i := 0; i < 2; i++ {
			out := t.TempDir()
			rep, err := Harvest([]string{path}, HarvestFilter{Limit: 2, Seed: 7}, out)
			if err != nil {
				t.Fatalf("harvest: %v", err)
			}
			if rep.Written != 2 {
				t.Fatalf("wrote %d, want 2", rep.Written)
			}
			positions, err := Load(out)
			if err != nil {
				t.Fatalf("load: %v", err)
			}
			var ids []string
			for _, p := range positions {
				ids = append(ids, p.ID)
			}
			if i == 0 {
				first = ids
			} else if !equalStrings(first, ids) {
				t.Fatalf("two harvests with seed 7 wrote %v and %v", first, ids)
			}
		}
	})
}

func TestHarvestTimeoutFilterIncludesModelTimeouts(t *testing.T) {
	rec := record(0, 1, model.LayerB, 0, nil, harvestMoves(), false)
	rec.Trace.Fallback = model.FallbackError
	rec.Trace.TimedOut = true
	if !keep(rec, HarvestFilter{Fallbacks: []string{aiseat.FallbackTimeout}}) {
		t.Fatal("timeout filter discarded a model timeout carried by Trace.TimedOut")
	}
}

// TestHarvestSkipsCompactRecords: escalated mode writes most windows
// without a view, and a position without a view cannot be rendered,
// replayed or labelled.
func TestHarvestSkipsCompactRecords(t *testing.T) {
	logDir, outDir := t.TempDir(), t.TempDir()
	compact := decisionlog.Record{
		V: decisionlog.RecordVersion, Game: harvestGame, SeatIndex: 0, Seq: 5,
		Moves: harvestMoves(), Trace: aiseat.Trace{Layer: model.LayerA, HeuristicIndex: aiseat.Decline},
	}
	path := writeLog(t, logDir, compact)
	rep, err := Harvest([]string{path}, HarvestFilter{}, outDir)
	if err != nil {
		t.Fatalf("harvest: %v", err)
	}
	if rep.Records != 1 || rep.Full != 0 || rep.Written != 0 {
		t.Fatalf("report %+v, want 1 record, 0 full, 0 written", rep)
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
