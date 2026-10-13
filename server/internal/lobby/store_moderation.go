package lobby

// store_moderation.go: the two tables migration 0016 adds (#2984,
// ADR 0075 section 2.6), on both Store implementations.
//
//   - table_blocks: who a table has refused to seat again.
//   - player_reports: "report player", read by an admin.

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
)

// BlockKind says what a table_blocks subject is.
type BlockKind string

const (
	// BlockUser: subject is a users(id). Holds against that account.
	BlockUser BlockKind = "user"
	// BlockDiscord: subject is a Discord snowflake. Holds against that
	// Discord account whether or not it has a users row yet.
	BlockDiscord BlockKind = "discord"
	// BlockName: subject is moderation.Key of a guest's name. Holds
	// against a guest (no account) who types the same name again, in any
	// case, spacing or leetspeak. It does NOT hold against a different
	// name; a guest has nothing else the server can recognise.
	BlockName BlockKind = "name"
)

// BlockRecord is one table_blocks row, minus its game.
type BlockRecord struct {
	Kind      BlockKind
	Subject   string
	CreatedAt time.Time
}

// ReportLine is one chat line in a report's excerpt, as the table saw it
// (already masked by the chat filter).
type ReportLine struct {
	Author string `json:"author"`
	Text   string `json:"text"`
	At     string `json:"at"`
}

// ReportRecord is one player_reports row.
type ReportRecord struct {
	ID               uuid.UUID
	CreatedAt        time.Time
	GameID           uuid.UUID
	GameName         string
	ReporterPlayerID uuid.UUID
	ReporterUserID   string // "" for a guest
	ReporterName     string
	ReportedPlayerID uuid.UUID
	ReportedUserID   string // "" for a guest
	ReportedName     string
	Reason           string
	Excerpt          []ReportLine
}

// --- SQL ----------------------------------------------------------------

func (s *SQLStore) AddBlock(ctx context.Context, gameID uuid.UUID, b BlockRecord) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO table_blocks (game_id, kind, subject, created_at) VALUES (?, ?, ?, ?)`,
		gameID.String(), string(b.Kind), b.Subject, toMillis(b.CreatedAt))
	return err
}

func (s *SQLStore) Blocks(ctx context.Context, gameID uuid.UUID) ([]BlockRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT kind, subject, created_at FROM table_blocks WHERE game_id = ? ORDER BY created_at, kind, subject`,
		gameID.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []BlockRecord
	for rows.Next() {
		var (
			kind, subject string
			at            int64
		)
		if err := rows.Scan(&kind, &subject, &at); err != nil {
			return nil, err
		}
		out = append(out, BlockRecord{Kind: BlockKind(kind), Subject: subject, CreatedAt: fromMillis(at)})
	}
	return out, rows.Err()
}

func (s *SQLStore) AddReport(ctx context.Context, r ReportRecord) error {
	excerpt, err := json.Marshal(r.Excerpt)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO player_reports (id, created_at, game_id, game_name, reporter_player_id, reporter_user_id, reporter_name,
		                             reported_player_id, reported_user_id, reported_name, reason, excerpt)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID.String(), toMillis(r.CreatedAt), r.GameID.String(), r.GameName,
		r.ReporterPlayerID.String(), nullString(r.ReporterUserID), r.ReporterName,
		r.ReportedPlayerID.String(), nullString(r.ReportedUserID), r.ReportedName,
		r.Reason, string(excerpt))
	return err
}

func (s *SQLStore) Reports(ctx context.Context) ([]ReportRecord, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, created_at, game_id, game_name, reporter_player_id, reporter_user_id, reporter_name,
		        reported_player_id, reported_user_id, reported_name, reason, excerpt
		 FROM player_reports ORDER BY created_at DESC, id`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var out []ReportRecord
	for rows.Next() {
		var (
			id, gid, rpid, tpid, excerpt string
			at                           int64
			r                            ReportRecord
			ruser, tuser                 sql.NullString
		)
		if err := rows.Scan(&id, &at, &gid, &r.GameName, &rpid, &ruser, &r.ReporterName,
			&tpid, &tuser, &r.ReportedName, &r.Reason, &excerpt); err != nil {
			return nil, err
		}
		r.ID, r.GameID = parseUUIDOrNil(id), parseUUIDOrNil(gid)
		r.ReporterPlayerID, r.ReportedPlayerID = parseUUIDOrNil(rpid), parseUUIDOrNil(tpid)
		r.ReporterUserID, r.ReportedUserID = ruser.String, tuser.String
		r.CreatedAt = fromMillis(at)
		if err := json.Unmarshal([]byte(excerpt), &r.Excerpt); err != nil {
			return nil, fmt.Errorf("report %s excerpt: %w", id, err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLStore) DeleteReport(ctx context.Context, id uuid.UUID) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM player_reports WHERE id = ?`, id.String())
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err == nil && n == 0 {
		return ErrStoreNotFound
	}
	return nil
}

func (s *SQLStore) PruneReports(ctx context.Context, before time.Time, keep int) (int, error) {
	res, err := s.db.ExecContext(ctx, `DELETE FROM player_reports WHERE created_at < ?`, toMillis(before))
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	res, err = s.db.ExecContext(ctx,
		`DELETE FROM player_reports WHERE id IN
		   (SELECT id FROM player_reports ORDER BY created_at DESC, id LIMIT -1 OFFSET ?)`, keep)
	if err != nil {
		return int(n), err
	}
	m, _ := res.RowsAffected()
	return int(n + m), nil
}

// --- memory -------------------------------------------------------------

func (s *memoryStore) AddBlock(_ context.Context, gameID uuid.UUID, b BlockRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, have := range s.blocks[gameID] {
		if have.Kind == b.Kind && have.Subject == b.Subject {
			return nil
		}
	}
	s.blocks[gameID] = append(s.blocks[gameID], b)
	return nil
}

func (s *memoryStore) Blocks(_ context.Context, gameID uuid.UUID) ([]BlockRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]BlockRecord(nil), s.blocks[gameID]...), nil
}

func (s *memoryStore) AddReport(_ context.Context, r ReportRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r.Excerpt = append([]ReportLine(nil), r.Excerpt...)
	s.reports = append(s.reports, r)
	return nil
}

func (s *memoryStore) Reports(_ context.Context) ([]ReportRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := append([]ReportRecord(nil), s.reports...)
	sort.SliceStable(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}

func (s *memoryStore) DeleteReport(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, r := range s.reports {
		if r.ID == id {
			s.reports = append(s.reports[:i], s.reports[i+1:]...)
			return nil
		}
	}
	return ErrStoreNotFound
}

func (s *memoryStore) PruneReports(_ context.Context, before time.Time, keep int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	kept := s.reports[:0:0]
	for _, r := range s.reports {
		if !r.CreatedAt.Before(before) {
			kept = append(kept, r)
		}
	}
	sort.SliceStable(kept, func(i, j int) bool { return kept[i].CreatedAt.After(kept[j].CreatedAt) })
	if len(kept) > keep {
		kept = kept[:keep]
	}
	n := len(s.reports) - len(kept)
	s.reports = kept
	return n, nil
}
