package ws

import (
	"os"
	"path/filepath"
	"testing"
)

// TestEphemeralRoomLeavesNothingOnDisk — the tutorial's practice table
// (ADR 0076 §2.2) has no resume, so its room must not write the
// artifacts a restart would rebuild it from, even on a manager that
// persists every other room. The control room beside it proves the
// manager in this test does write them.
func TestEphemeralRoomLeavesNothingOnDisk(t *testing.T) {
	dir := t.TempDir()
	mgr := restoreTestManager(t, dir)

	practice := newPersistGame(t)
	room := mgr.CreateEphemeral(practice)
	if room == nil || mgr.Get(practice.ID) != room {
		t.Fatal("CreateEphemeral did not register the room")
	}
	if again := mgr.CreateEphemeral(practice); again != room {
		t.Error("a second CreateEphemeral for the same game replaced the room")
	}
	if room.ReplayPath() != "" {
		t.Errorf("an ephemeral room reports a replay path %q", room.ReplayPath())
	}

	control := newPersistGame(t)
	controlRoom := mgr.Create(control)

	for _, r := range []*Room{room, controlRoom} {
		g := r.Game
		if _, _, err := r.Apply(g.Seats[0].ID, func() error {
			g.WithWriteLock(func() { g.Seats[0].Life = 7 })
			return nil
		}); err != nil {
			t.Fatalf("Apply: %v", err)
		}
	}

	if _, err := os.Stat(restorePointPath(dir, control.ID)); err != nil {
		t.Fatalf("setup: the control room wrote no restore point: %v", err)
	}
	for _, path := range []string{
		restorePointPath(dir, practice.ID),
		snapshotPath(dir, practice.ID),
		replayPath(dir, practice.ID),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("ephemeral room wrote %s (err=%v)", filepath.Base(filepath.Dir(path))+"/"+filepath.Base(path), err)
		}
	}

	// And a restart rebuilds only the control room.
	outcomes := restoreTestManager(t, dir).RestoreRooms()
	for _, o := range outcomes {
		if o.GameID == practice.ID {
			t.Errorf("a restart rebuilt the ephemeral room")
		}
	}
}
