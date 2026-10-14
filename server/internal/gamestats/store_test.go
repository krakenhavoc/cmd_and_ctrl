package gamestats

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openStore(t *testing.T) (*SQLStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d), d
}

func mustGame(t *testing.T, d *db.DB, id uuid.UUID) {
	t.Helper()
	if _, err := d.Exec(`INSERT INTO games (id, name, state, created_at) VALUES (?, 'Friday', 'ended', 1)`, id.String()); err != nil {
		t.Fatalf("insert game: %v", err)
	}
}

func TestStoreRoundTrip(t *testing.T) {
	s, d := openStore(t)
	g := Compute(fixture())
	mustGame(t, d, g.GameID)
	ctx := context.Background()
	if err := s.Put(ctx, g); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := s.Get(ctx, g.GameID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !reflect.DeepEqual(got, g) {
		t.Errorf("round trip changed the record:\n got %+v\nwant %+v", got, g)
	}
}

func TestStorePutReplaces(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	g := Compute(fixture())
	mustGame(t, d, g.GameID)
	if err := s.Put(ctx, g); err != nil {
		t.Fatal(err)
	}
	in := fixture()
	in.Log = nil
	again := Compute(in)
	if err := s.Put(ctx, again); err != nil {
		t.Fatalf("second Put: %v", err)
	}
	got, err := s.Get(ctx, g.GameID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Turns != 0 || len(got.Seats[2].Cards) != 0 {
		t.Errorf("Put did not replace the old record: turns %d, cards %+v", got.Turns, got.Seats[2].Cards)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM seat_stats WHERE game_id = ?`, g.GameID.String()).Scan(&n); err != nil || n != 4 {
		t.Errorf("seat_stats rows %d (%v), want 4", n, err)
	}
}

func TestStoreGoesWithItsGame(t *testing.T) {
	s, d := openStore(t)
	ctx := context.Background()
	g := Compute(fixture())
	mustGame(t, d, g.GameID)
	if err := s.Put(ctx, g); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`DELETE FROM games WHERE id = ?`, g.GameID.String()); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, g.GameID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after the game was deleted: %v, want ErrNotFound", err)
	}
	for _, table := range []string{"seat_stats", "seat_card_stats"} {
		var n int
		if err := d.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n); err != nil || n != 0 {
			t.Errorf("%s has %d rows (%v) after its game was deleted", table, n, err)
		}
	}
}

func TestStoreRefusesAGameWithNoRow(t *testing.T) {
	s, _ := openStore(t)
	if err := s.Put(context.Background(), Compute(fixture())); err == nil {
		t.Fatal("Put for a game with no games row succeeded")
	}
}

func TestNoStore(t *testing.T) {
	var s Store = NoStore{}
	if err := s.Put(context.Background(), Compute(fixture())); err != nil {
		t.Fatalf("NoStore.Put: %v", err)
	}
	if _, err := s.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("NoStore.Get: %v", err)
	}
}
