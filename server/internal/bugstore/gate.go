package bugstore

// gate.go — the two small pieces of state the bug-report throttle
// keeps across a restart (ADR 0017 §10): the admin kill switch with
// today's filing count and the recent-filing index the duplicate check
// reads, in one file beside the report directories, and the reports
// that were queued instead of filed.
//
// One JSON file rather than a table: the data is a few hundred bytes,
// written on the report path only, and the artifact directory already
// is the feature's storage. Prune's scan ignores plain files, so
// gate.json is never reclaimed with the reports.

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

const gateFileName = "gate.json"

// QueuedReport is what a queued report remembers about the issue it
// would have opened. Title and Body are already redacted and rendered,
// exactly as they would have been filed.
type QueuedReport struct {
	// Reason is why it was kept: "paused", "account limit" or
	// "daily cap".
	Reason   string `json:"reason"`
	Kind     string `json:"kind"`
	Label    string `json:"label,omitempty"`
	Reporter string `json:"reporter"`
	Title    string `json:"title"`
	Body     string `json:"body"`
}

// Filing is one entry of the recent-filing index: an issue the server
// opened and the keys a later report is matched against.
type Filing struct {
	Keys   []string  `json:"keys"`
	Number int       `json:"number"`
	URL    string    `json:"url"`
	At     time.Time `json:"at"`
}

// GateState is the persisted throttle state.
type GateState struct {
	// Paused is the admin kill switch: while true nothing is sent to
	// GitHub and every report is queued.
	Paused   bool      `json:"paused"`
	PausedAt time.Time `json:"paused_at,omitempty"`
	// Day is the UTC date (2006-01-02) Filed counts, so the daily cap
	// resets when the date changes without a timer.
	Day    string   `json:"day,omitempty"`
	Filed  int      `json:"filed,omitempty"`
	Recent []Filing `json:"recent,omitempty"`
}

// LoadGate reads the persisted state. A missing file is the zero
// state, not an error; so is a disabled store.
func (s *Store) LoadGate() (GateState, error) {
	var st GateState
	if !s.Enabled() {
		return st, nil
	}
	buf, err := os.ReadFile(filepath.Join(s.dir, gateFileName)) //nolint:gosec // fixed name under the store dir
	if errors.Is(err, os.ErrNotExist) {
		return st, nil
	}
	if err != nil {
		return GateState{}, fmt.Errorf("bugstore: read gate: %w", err)
	}
	if err := json.Unmarshal(buf, &st); err != nil {
		return GateState{}, fmt.Errorf("bugstore: decode gate: %w", err)
	}
	return st, nil
}

// SaveGate writes the state atomically (temp file + rename), so a
// crash mid-write cannot leave a half-written file that loads as
// "not paused". A disabled store is a no-op.
func (s *Store) SaveGate(st GateState) error {
	if !s.Enabled() {
		return nil
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return fmt.Errorf("bugstore: create dir: %w", err)
	}
	buf, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("bugstore: encode gate: %w", err)
	}
	tmp, err := os.CreateTemp(s.dir, gateFileName+".*.tmp")
	if err != nil {
		return fmt.Errorf("bugstore: write gate: %w", err)
	}
	_, werr := tmp.Write(buf)
	cerr := tmp.Close()
	if werr != nil || cerr != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("bugstore: write gate: %w", errors.Join(werr, cerr))
	}
	if err := os.Rename(tmp.Name(), filepath.Join(s.dir, gateFileName)); err != nil {
		_ = os.Remove(tmp.Name())
		return fmt.Errorf("bugstore: write gate: %w", err)
	}
	return nil
}

// Queued lists the manifests of queued reports, newest first, at most
// limit (0 = all). Reports without a manifest, or filed ones, are
// skipped.
func (s *Store) Queued(limit int) ([]Manifest, error) {
	if !s.Enabled() {
		return nil, ErrDisabled
	}
	dirs, err := s.scan()
	if err != nil {
		return nil, err
	}
	var out []Manifest
	// scan is oldest-first; walk it backwards for newest-first.
	for i := len(dirs) - 1; i >= 0; i-- {
		buf, err := os.ReadFile(filepath.Join(s.dir, dirs[i].name, manifestFileName)) //nolint:gosec // name matched reportIDPattern
		if err != nil {
			continue
		}
		var m Manifest
		if json.Unmarshal(buf, &m) != nil || m.Queued == nil {
			continue
		}
		out = append(out, m)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
