package tablesetups

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

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

func mustUser(t *testing.T, d *db.DB, name string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, ?, 1, 1)`, id.String(), name); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func TestGetWithNoRowIsNotFound(t *testing.T) {
	s, d := openStore(t)
	if _, err := s.Get(context.Background(), mustUser(t, d, "A")); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get = %v, want ErrNotFound", err)
	}
}

func TestPutThenGetRoundTrips(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u, mate := mustUser(t, d, "A"), uuid.New()
	game := uuid.New()
	at := time.UnixMilli(1_700_000_000_456).UTC()
	s.now = func() time.Time { return at }

	want := Setup{
		Settings:   json.RawMessage(`{"starting_life":40,"bot_pace":"normal"}`),
		Bots:       []Bot{{Tier: "heuristic", DeckID: "izzet-aggro", Name: "Bot 1"}, {Tier: "random", DeckID: "tutorial", Name: "Bot 2"}},
		Tablemates: []uuid.UUID{mate},
	}
	if err := s.Put(ctx, u, game, want); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, err := s.Get(ctx, u)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.GameID != game || !got.UpdatedAt.Equal(at) {
		t.Errorf("record = %+v", got)
	}
	if string(got.Setup.Settings) != string(want.Settings) || !reflect.DeepEqual(got.Setup.Bots, want.Bots) || !reflect.DeepEqual(got.Setup.Tablemates, want.Tablemates) {
		t.Errorf("setup = %+v, want %+v", got.Setup, want)
	}
}

func TestPutReplacesTheRowAndKeepsOnePerUser(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")
	if err := s.Put(ctx, u, uuid.New(), Setup{Bots: []Bot{{Tier: "random", Name: "x"}}}); err != nil {
		t.Fatal(err)
	}
	second := uuid.New()
	if err := s.Put(ctx, u, second, Setup{Settings: json.RawMessage(`{"a":1}`)}); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	if got.GameID != second || len(got.Setup.Bots) != 0 || string(got.Setup.Settings) != `{"a":1}` {
		t.Errorf("record = %+v", got)
	}
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM table_setups`).Scan(&n); err != nil || n != 1 {
		t.Errorf("table_setups has %d rows (%v), want 1", n, err)
	}
}

// A zero body stores arrays and an object, never null, and a nil game
// reads back as the nil UUID.
func TestPutNormalisesAnEmptySetup(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")
	if err := s.Put(ctx, u, uuid.Nil, Setup{}); err != nil {
		t.Fatal(err)
	}
	var body string
	var game *string
	if err := d.QueryRow(`SELECT body, game_id FROM table_setups`).Scan(&body, &game); err != nil {
		t.Fatal(err)
	}
	if body != `{"settings":{},"bots":[],"tablemates":[]}` || game != nil {
		t.Errorf("body = %s, game_id = %v", body, game)
	}
	got, err := s.Get(ctx, u)
	if err != nil || got.GameID != uuid.Nil {
		t.Errorf("Get = %+v, %v", got, err)
	}
}

func TestSetupsAreScopedToTheUserAndNeedARealOne(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	a, b := mustUser(t, d, "A"), mustUser(t, d, "B")
	if err := s.Put(ctx, a, uuid.Nil, Setup{Bots: []Bot{{Tier: "random"}}}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, b); !errors.Is(err, ErrNotFound) {
		t.Errorf("B reads A's setup: %v", err)
	}
	if err := s.Put(ctx, uuid.New(), uuid.Nil, Setup{}); err == nil {
		t.Error("Put for a user that does not exist succeeded")
	}
	if err := s.Put(ctx, uuid.Nil, uuid.Nil, Setup{}); err == nil {
		t.Error("Put for the nil user succeeded")
	}
}

func TestSetupSurvivesItsGame(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")
	game := uuid.New()
	if _, err := d.Exec(`INSERT INTO games (id, name, state, created_at) VALUES (?, 'G', 'lobby', 1)`, game.String()); err != nil {
		t.Fatal(err)
	}
	if err := s.Put(ctx, u, game, Setup{}); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`DELETE FROM games WHERE id = ?`, game.String()); err != nil {
		t.Fatal(err)
	}
	if got, err := s.Get(ctx, u); err != nil || got.GameID != game {
		t.Errorf("after deleting the game: %+v, %v", got, err)
	}
}

func TestNoStore(t *testing.T) {
	var s Store = NoStore{}
	if _, err := s.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get = %v", err)
	}
	if err := s.Put(context.Background(), uuid.New(), uuid.Nil, Setup{}); !errors.Is(err, ErrNoStore) {
		t.Errorf("Put = %v", err)
	}
}
