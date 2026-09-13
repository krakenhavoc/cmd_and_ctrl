package bugstore

// gamelog_test.go — S31 sub-PR 0: the pinned public game log.

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func TestPinGameLogWritesAndReadsBack(t *testing.T) {
	s := newTestStore(t)
	r, err := s.NewReport()
	if err != nil {
		t.Fatalf("NewReport: %v", err)
	}
	text := "# public game log\n  1  t2  cast     Aang cast Lightning Bolt\n  2  t2  resolve  Lightning Bolt resolved\n"
	pin, err := r.PinGameLog([]byte(text))
	if err != nil {
		t.Fatalf("PinGameLog: %v", err)
	}
	if pin.Bytes != int64(len(text)) {
		t.Errorf("Bytes = %d, want %d", pin.Bytes, len(text))
	}
	if pin.Lines != 3 {
		t.Errorf("Lines = %d, want 3", pin.Lines)
	}
	if r.GameLog() != pin {
		t.Error("GameLog() does not report the pin")
	}

	f, info, err := s.OpenGameLog(r.ID)
	if err != nil {
		t.Fatalf("OpenGameLog: %v", err)
	}
	defer func() { _ = f.Close() }()
	if info.Size() != pin.Bytes {
		t.Errorf("on-disk size = %d, want %d", info.Size(), pin.Bytes)
	}
	got, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != text {
		t.Errorf("round trip mismatch:\n%s", got)
	}

	// It lands in the manifest, which is what pruning and any later
	// forensics read.
	if err := r.Commit("game-1", "https://example.test/issues/1"); err != nil {
		t.Fatalf("Commit: %v", err)
	}
	raw, err := os.ReadFile(filepath.Join(s.dir, r.ID, manifestFileName))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	if m.GameLog == nil || m.GameLog.Lines != 3 {
		t.Errorf("manifest game log = %+v, want 3 lines", m.GameLog)
	}
}

func TestPinGameLogRefusesEmptyAndOversized(t *testing.T) {
	s := newTestStore(t)
	r, err := s.NewReport()
	if err != nil {
		t.Fatalf("NewReport: %v", err)
	}
	if _, err := r.PinGameLog(nil); !errors.Is(err, ErrNotFound) {
		t.Errorf("empty pin err = %v, want ErrNotFound", err)
	}
	if _, err := r.PinGameLog(bytes.Repeat([]byte("x"), MaxGameLogPinBytes+1)); !errors.Is(err, ErrTooLarge) {
		t.Errorf("oversized pin err = %v, want ErrTooLarge", err)
	}
	// Neither refusal may leave a file behind.
	if _, _, err := s.OpenGameLog(r.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("OpenGameLog after refused pins = %v, want ErrNotFound", err)
	}
}

func TestOpenGameLogRejectsBadIDs(t *testing.T) {
	s := newTestStore(t)
	for _, id := range []string{"", "not-a-uuid", "../etc", uuid.NewString()} {
		if _, _, err := s.OpenGameLog(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("OpenGameLog(%q) = %v, want ErrNotFound", id, err)
		}
	}
	if _, _, err := New("", "").OpenGameLog(uuid.NewString()); !errors.Is(err, ErrDisabled) {
		t.Errorf("disabled store OpenGameLog err = %v, want ErrDisabled", err)
	}
}
