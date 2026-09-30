package effects

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// TestCorpusPrePlanFixturesRestoreWithTheirOrdinals restores every v7
// corpus fixture written before the turn plan (ADR 0059 sub-PR 2b,
// #753: no "turnPlan" and no "nextPhaseId" in the file) and checks its
// turn ordinals and step and phase counts against a fresh build of the
// same board — the live game at the same step. The fixtures are real
// pre-plan restore points, read as they are on disk.
func TestCorpusPrePlanFixturesRestoreWithTheirOrdinals(t *testing.T) {
	dir := filepath.Join(corpusRoot, "v7")
	checked := 0
	for _, b := range corpusBoards() {
		raw, err := os.ReadFile(filepath.Join(dir, b.name+".json"))
		if err != nil {
			continue
		}
		var probe struct {
			Snapshot map[string]json.RawMessage `json:"snapshot"`
		}
		if err := json.Unmarshal(raw, &probe); err != nil {
			t.Fatalf("%s: %v", b.name, err)
		}
		if _, ok := probe.Snapshot["nextPhaseId"]; ok {
			continue // written with the plan
		}
		checked++
		t.Run(b.name, func(t *testing.T) {
			var file corpusFile
			if err := json.Unmarshal(raw, &file); err != nil {
				t.Fatal(err)
			}
			restored, err := file.Snapshot.RestoreStrict()
			if err != nil {
				t.Fatalf("RestoreStrict: %v", err)
			}
			live := b.build(t)
			rt, lt := restored.Turn, live.Turn
			if rt.Step != lt.Step || rt.PhaseID != lt.PhaseID || rt.PhaseOrdinal != lt.PhaseOrdinal || rt.StepOrdinal != lt.StepOrdinal {
				t.Errorf("restored %s phase %d ordinals %d/%d; the board built live is at %s phase %d ordinals %d/%d",
					rt.Step, rt.PhaseID, rt.PhaseOrdinal, rt.StepOrdinal, lt.Step, lt.PhaseID, lt.PhaseOrdinal, lt.StepOrdinal)
			}
			rc, lc := restored.TurnTally, live.TurnTally
			if !reflect.DeepEqual(rc.StepsBegun, lc.StepsBegun) || !reflect.DeepEqual(rc.PhasesBegun, lc.PhasesBegun) || rc.PhaseStarted != lc.PhaseStarted {
				t.Errorf("restored counts steps %v phases %v started %d; live steps %v phases %v started %d",
					rc.StepsBegun, rc.PhasesBegun, rc.PhaseStarted, lc.StepsBegun, lc.PhasesBegun, lc.PhaseStarted)
			}
		})
	}
	if checked == 0 {
		t.Fatal("no pre-plan fixture found in v7; the corpus should hold the ones written before #753")
	}
}
