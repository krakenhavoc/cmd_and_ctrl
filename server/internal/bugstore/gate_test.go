package bugstore

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGateStateRoundTripsAndMissingFileIsZero(t *testing.T) {
	s := New(t.TempDir(), "")
	st, err := s.LoadGate()
	if err != nil || st.Paused || st.Filed != 0 || len(st.Recent) != 0 {
		t.Fatalf("fresh store = %+v, %v; want the zero state", st, err)
	}
	want := GateState{
		Paused: true, PausedAt: time.Unix(1_700_000_000, 0).UTC(), Day: "2026-10-12", Filed: 3,
		Recent: []Filing{{Keys: []string{"t:bug|x"}, Number: 7, URL: "u", At: time.Unix(1_700_000_100, 0).UTC()}},
	}
	if err := s.SaveGate(want); err != nil {
		t.Fatal(err)
	}
	got, err := New(s.dir, "").LoadGate()
	if err != nil {
		t.Fatal(err)
	}
	if !got.Paused || got.Filed != 3 || got.Day != "2026-10-12" || len(got.Recent) != 1 || got.Recent[0].Number != 7 {
		t.Errorf("round trip = %+v", got)
	}
	// No temp files left behind, and Prune must not touch the state file.
	if _, err := s.Prune(time.Nanosecond, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(s.dir, gateFileName)); err != nil {
		t.Errorf("Prune removed the gate file: %v", err)
	}
	entries, _ := os.ReadDir(s.dir)
	if len(entries) != 1 {
		t.Errorf("dir holds %d entries, want only gate.json", len(entries))
	}
}

func TestGateOnADisabledStoreIsANoOp(t *testing.T) {
	var s *Store
	if st, err := s.LoadGate(); err != nil || st.Paused {
		t.Errorf("nil store load = %+v, %v", st, err)
	}
	if err := s.SaveGate(GateState{Paused: true}); err != nil {
		t.Errorf("nil store save = %v", err)
	}
	if _, err := s.Queued(0); err != ErrDisabled {
		t.Errorf("nil store Queued = %v, want ErrDisabled", err)
	}
}

func TestQueuedListsOnlyQueuedReportsNewestFirst(t *testing.T) {
	s := New(t.TempDir(), "")
	var queuedIDs []string
	for i, reason := range []string{"paused", "daily cap"} {
		r, err := s.NewReport()
		if err != nil {
			t.Fatal(err)
		}
		if err := r.CommitQueued("g", QueuedReport{Reason: reason, Title: "t", Body: "b"}); err != nil {
			t.Fatal(err)
		}
		queuedIDs = append(queuedIDs, r.ID)
		// Distinct creation times for a deterministic order.
		_ = i
		time.Sleep(2 * time.Millisecond)
	}
	filed, err := s.NewReport()
	if err != nil {
		t.Fatal(err)
	}
	if err := filed.Commit("g", "https://example.test/1"); err != nil {
		t.Fatal(err)
	}

	got, err := s.Queued(0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != queuedIDs[1] || got[1].ID != queuedIDs[0] {
		t.Fatalf("Queued = %v, want the two queued reports newest first", got)
	}
	if got[0].Queued.Reason != "daily cap" || got[0].IssueURL != "" {
		t.Errorf("newest = %+v", got[0])
	}
	if one, _ := s.Queued(1); len(one) != 1 {
		t.Errorf("limit 1 returned %d", len(one))
	}
}
