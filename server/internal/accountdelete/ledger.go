package accountdelete

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

// LedgerDir and LedgerFileName place the ledger under the data dir:
// <data dir>/deletions/ledger.jsonl.
const (
	LedgerDir      = "deletions"
	LedgerFileName = "ledger.jsonl"
)

// Ledger is the append-only record of deleted accounts (#3340).
//
// The tombstone (ADR 0149 section 2) lives in the user database only, so
// restoring an older database copy brings deleted accounts back. The
// ledger is a second place that remembers them: one JSON line per
// deletion, holding the account id and the deletion time and nothing
// else (no name, no Discord id: the id names no one once the account is
// gone). It is a file beside the database so the off-site backup carries
// it, and a restore takes it from the LATEST snapshot while the database
// comes from whichever snapshot was chosen. Re-applying it
// (ReadLedger then ReapplyLedger, run by cmd/restorecheck) deletes again
// whatever the older database still holds.
//
// Entries are only ever appended, so a newer ledger is a superset of an
// older one, and re-applying is idempotent.
type Ledger struct {
	path string
	mu   sync.Mutex
}

// NewLedger returns the ledger at <dataDir>/deletions/ledger.jsonl.
func NewLedger(dataDir string) *Ledger {
	return &Ledger{path: LedgerPath(dataDir)}
}

// LedgerPath is where the ledger lives under dataDir.
func LedgerPath(dataDir string) string {
	return filepath.Join(dataDir, LedgerDir, LedgerFileName)
}

// LedgerEntry is one line of the ledger.
type LedgerEntry struct {
	UserID    uuid.UUID `json:"user_id"`
	DeletedAt int64     `json:"deleted_at"` // Unix milliseconds
}

// Append adds one entry and syncs it to disk.
func (l *Ledger) Append(id uuid.UUID, at time.Time) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(l.path), 0o700); err != nil {
		return fmt.Errorf("accountdelete: ledger dir: %w", err)
	}
	line, err := json.Marshal(LedgerEntry{UserID: id, DeletedAt: at.UnixMilli()})
	if err != nil {
		return err
	}
	f, err := os.OpenFile(l.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("accountdelete: ledger open: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		_ = f.Close()
		return fmt.Errorf("accountdelete: ledger write: %w", err)
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return fmt.Errorf("accountdelete: ledger sync: %w", err)
	}
	return f.Close()
}

// ReadLedger returns the distinct accounts in the ledger at path, oldest
// first. A missing file is an empty ledger. A line that does not parse
// is an error naming its number: a half-written ledger must be looked at,
// not skipped, because skipping one line brings one person back.
func ReadLedger(path string) ([]LedgerEntry, error) {
	f, err := os.Open(path)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("accountdelete: ledger open: %w", err)
	}
	defer func() { _ = f.Close() }()

	var out []LedgerEntry
	seen := map[uuid.UUID]bool{}
	sc := bufio.NewScanner(f)
	for n := 1; sc.Scan(); n++ {
		if len(sc.Bytes()) == 0 {
			continue
		}
		var e LedgerEntry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil || e.UserID == uuid.Nil {
			return nil, fmt.Errorf("accountdelete: ledger %s line %d is not a deletion entry", path, n)
		}
		if seen[e.UserID] {
			continue
		}
		seen[e.UserID] = true
		out = append(out, e)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("accountdelete: ledger read: %w", err)
	}
	return out, nil
}

// ReapplyReport is what ReapplyLedger did.
type ReapplyReport struct {
	Applied        int // accounts deleted again (under dryRun: that would be)
	AlreadyDeleted int // tombstoned in the database already
	Unknown        int // not in this database (it predates the account)
	PlaymatIDs     []string
}

// ReapplyLedger deletes again every ledger account that the database
// holds as a live account. dryRun reports without writing.
func ReapplyLedger(ctx context.Context, s *SQLStore, entries []LedgerEntry, dryRun bool) (ReapplyReport, error) {
	var rep ReapplyReport
	for _, e := range entries {
		if dryRun {
			var deletedAt sql.NullInt64
			err := s.db.QueryRowContext(ctx, `SELECT deleted_at FROM users WHERE id = ?`, e.UserID.String()).Scan(&deletedAt)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				rep.Unknown++
			case err != nil:
				return rep, fmt.Errorf("accountdelete: read user: %w", err)
			case deletedAt.Valid:
				rep.AlreadyDeleted++
			default:
				rep.Applied++
			}
			continue
		}
		res, err := s.Reapply(ctx, e.UserID)
		switch {
		case errors.Is(err, ErrNotFound):
			rep.Unknown++
		case errors.Is(err, ErrAlreadyDeleted):
			rep.AlreadyDeleted++
		case err != nil:
			return rep, fmt.Errorf("re-applying deletion: %w", err)
		default:
			rep.Applied++
			rep.PlaymatIDs = append(rep.PlaymatIDs, res.PlaymatIDs...)
		}
	}
	return rep, nil
}
