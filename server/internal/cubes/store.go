package cubes

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

// Store keeps cubes. Every read and write names the owner: a cube is
// its owner's alone, and another person's id is ErrNotFound, the same
// answer as a cube that does not exist. Safe for concurrent use.
type Store interface {
	// Create saves a new cube. ErrNameTaken when the owner has one by
	// that name, ErrTooMany at MaxCubes.
	Create(ctx context.Context, c Cube) (Cube, error)
	// Update replaces a cube's name, list and singleton setting.
	Update(ctx context.Context, c Cube) (Cube, error)
	// List is a person's cubes, by name.
	List(ctx context.Context, owner uuid.UUID) ([]Cube, error)
	Get(ctx context.Context, owner, id uuid.UUID) (Cube, error)
	Delete(ctx context.Context, owner, id uuid.UUID) error
}

// SQLStore is migration 0039's table.
type SQLStore struct {
	db  *sql.DB
	now func() time.Time
}

// NewSQLStore wraps an opened, migrated database. The caller owns d.
func NewSQLStore(d *db.DB) *SQLStore {
	return &SQLStore{db: d.DB, now: func() time.Time { return time.Now().UTC() }}
}

var _ Store = (*SQLStore)(nil)

const cubeColumns = `id, user_id, name, singleton, list, card_count, created_at, updated_at`

func scanCube(scan func(...any) error) (Cube, error) {
	var (
		c                Cube
		id, owner        string
		singleton        int
		created, updated int64
	)
	if err := scan(&id, &owner, &c.Name, &singleton, &c.List, &c.CardCount, &created, &updated); err != nil {
		return Cube{}, err
	}
	var err error
	if c.ID, err = uuid.Parse(id); err != nil {
		return Cube{}, err
	}
	if c.OwnerID, err = uuid.Parse(owner); err != nil {
		return Cube{}, err
	}
	c.Singleton = singleton != 0
	c.CreatedAt, c.UpdatedAt = time.UnixMilli(created).UTC(), time.UnixMilli(updated).UTC()
	return c, nil
}

func flag(b bool) int {
	if b {
		return 1
	}
	return 0
}

// nameTaken reports whether err is the (user_id, name) index refusing a
// second cube of one name.
func nameTaken(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE") && strings.Contains(err.Error(), "cubes.")
}

// Create implements Store.
func (s *SQLStore) Create(ctx context.Context, c Cube) (Cube, error) {
	if c.OwnerID == uuid.Nil {
		return Cube{}, errors.New("cubes: a cube needs an owner")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Cube{}, err
	}
	defer func() { _ = tx.Rollback() }()
	var n int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM cubes WHERE user_id = ?`, c.OwnerID.String()).Scan(&n); err != nil {
		return Cube{}, err
	}
	if n >= MaxCubes {
		return Cube{}, ErrTooMany
	}
	c.ID = uuid.New()
	now := s.now()
	c.CreatedAt, c.UpdatedAt = now, now
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO cubes (`+cubeColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		c.ID.String(), c.OwnerID.String(), c.Name, flag(c.Singleton), c.List, c.CardCount,
		now.UnixMilli(), now.UnixMilli()); err != nil {
		if nameTaken(err) {
			return Cube{}, ErrNameTaken
		}
		return Cube{}, err
	}
	if err := tx.Commit(); err != nil {
		return Cube{}, err
	}
	c.CreatedAt, c.UpdatedAt = time.UnixMilli(now.UnixMilli()).UTC(), time.UnixMilli(now.UnixMilli()).UTC()
	return c, nil
}

// Update implements Store.
func (s *SQLStore) Update(ctx context.Context, c Cube) (Cube, error) {
	res, err := s.db.ExecContext(ctx,
		`UPDATE cubes SET name = ?, singleton = ?, list = ?, card_count = ?, updated_at = ? WHERE id = ? AND user_id = ?`,
		c.Name, flag(c.Singleton), c.List, c.CardCount, s.now().UnixMilli(), c.ID.String(), c.OwnerID.String())
	if err != nil {
		if nameTaken(err) {
			return Cube{}, ErrNameTaken
		}
		return Cube{}, err
	}
	if n, err := res.RowsAffected(); err != nil {
		return Cube{}, err
	} else if n == 0 {
		return Cube{}, ErrNotFound
	}
	return s.Get(ctx, c.OwnerID, c.ID)
}

// List implements Store.
func (s *SQLStore) List(ctx context.Context, owner uuid.UUID) ([]Cube, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT `+cubeColumns+` FROM cubes WHERE user_id = ? ORDER BY name COLLATE NOCASE, id`, owner.String())
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	out := []Cube{}
	for rows.Next() {
		c, err := scanCube(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// Get implements Store.
func (s *SQLStore) Get(ctx context.Context, owner, id uuid.UUID) (Cube, error) {
	c, err := scanCube(s.db.QueryRowContext(ctx,
		`SELECT `+cubeColumns+` FROM cubes WHERE id = ? AND user_id = ?`, id.String(), owner.String()).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return Cube{}, ErrNotFound
	}
	return c, err
}

// Delete implements Store.
func (s *SQLStore) Delete(ctx context.Context, owner, id uuid.UUID) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM cubes WHERE id = ? AND user_id = ?`, id.String(), owner.String())
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil {
		return err
	} else if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ForgetAccount is account deletion's part (ADR 0149, ADR 0222 §8),
// inside the deletion's own transaction: every cube of the person's is
// deleted. It returns how many there were.
func ForgetAccount(ctx context.Context, tx *sql.Tx, user string) (int, error) {
	res, err := tx.ExecContext(ctx, `DELETE FROM cubes WHERE user_id = ?`, user)
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}
