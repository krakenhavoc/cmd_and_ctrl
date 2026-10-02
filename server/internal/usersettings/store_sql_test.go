package usersettings

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
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

func TestFirstPutNeedsRevisionZeroAndSetsRevisionOne(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")
	at := time.UnixMilli(1_700_000_000_123).UTC()
	s.now = func() time.Time { return at }

	if _, err := s.Put(ctx, u, 15, json.RawMessage(`{"a":1}`), 3); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("first Put with If-Match 3 = %v, want a conflict", err)
	}
	got, err := s.Put(ctx, u, 15, json.RawMessage(`{"a":1}`), 0)
	if err != nil {
		t.Fatalf("Put: %v", err)
	}
	if got.Revision != 1 || got.Version != 15 || string(got.Body) != `{"a":1}` || !got.UpdatedAt.Equal(at) {
		t.Errorf("got %+v", got)
	}
	read, err := s.Get(ctx, u)
	if err != nil || read.Revision != 1 || string(read.Body) != `{"a":1}` {
		t.Errorf("Get = %+v, %v", read, err)
	}
	// A second "first" write is a conflict that returns what is there.
	cur, err := s.Put(ctx, u, 15, json.RawMessage(`{"a":2}`), 0)
	if !errors.Is(err, ErrRevisionConflict) || cur.Revision != 1 || string(cur.Body) != `{"a":1}` {
		t.Errorf("Put at revision 0 over a row = %+v, %v", cur, err)
	}
}

func TestPutBumpsRevisionAndChecksIfMatch(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")
	if _, err := s.Put(ctx, u, 15, json.RawMessage(`{"a":1}`), 0); err != nil {
		t.Fatal(err)
	}
	got, err := s.Put(ctx, u, 16, json.RawMessage(`{"a":2}`), 1)
	if err != nil || got.Revision != 2 || got.Version != 16 || string(got.Body) != `{"a":2}` {
		t.Fatalf("Put = %+v, %v", got, err)
	}

	cur, err := s.Put(ctx, u, 16, json.RawMessage(`{"a":3}`), 1)
	if !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("stale If-Match = %v, want a conflict", err)
	}
	if cur.Revision != 2 || string(cur.Body) != `{"a":2}` {
		t.Errorf("conflict copy = %+v, want the current one", cur)
	}
	if read, _ := s.Get(ctx, u); string(read.Body) != `{"a":2}` {
		t.Errorf("a refused write changed the row: %s", read.Body)
	}
}

func TestPutRefusesAnOlderClientVersion(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")
	if _, err := s.Put(ctx, u, 16, json.RawMessage(`{"a":1}`), 0); err != nil {
		t.Fatal(err)
	}
	cur, err := s.Put(ctx, u, 15, json.RawMessage(`{"a":2}`), 1)
	if !errors.Is(err, ErrVersionTooOld) || cur.Version != 16 || string(cur.Body) != `{"a":1}` {
		t.Fatalf("Put = %+v, %v; want ErrVersionTooOld and the current copy", cur, err)
	}
	// The same version is fine.
	if _, err := s.Put(ctx, u, 16, json.RawMessage(`{"a":2}`), 1); err != nil {
		t.Errorf("same version: %v", err)
	}
	// A stale revision outranks an old version.
	if _, err := s.Put(ctx, u, 1, json.RawMessage(`{}`), 1); !errors.Is(err, ErrRevisionConflict) {
		t.Errorf("stale revision + old version = %v, want the conflict", err)
	}
}

func TestPutEnforcesTheSizeCap(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")

	exact := json.RawMessage(`{"a":"` + strings.Repeat("x", MaxBodyBytes-8) + `"}`)
	if len(exact) != MaxBodyBytes {
		t.Fatalf("test body is %d bytes", len(exact))
	}
	if _, err := s.Put(ctx, u, 1, exact, 0); err != nil {
		t.Fatalf("a body of exactly 32 KiB: %v", err)
	}
	over := json.RawMessage(`{"a":"` + strings.Repeat("x", MaxBodyBytes-7) + `"}`)
	if _, err := s.Put(ctx, u, 1, over, 1); !errors.Is(err, ErrInvalid) {
		t.Fatalf("a body of 32 KiB + 1 = %v, want ErrInvalid", err)
	}
	if read, _ := s.Get(ctx, u); read.Revision != 1 {
		t.Errorf("the refused write bumped the revision to %d", read.Revision)
	}
}

func TestValidate(t *testing.T) {
	deep4 := json.RawMessage(`{"a":{"b":{"c":{}}}}`)
	deep5 := json.RawMessage(`{"a":{"b":{"c":{"d":{}}}}}`)
	cases := []struct {
		name    string
		version int
		body    json.RawMessage
		ok      bool
	}{
		{"object", 1, json.RawMessage(`{}`), true},
		{"max version", 1000, json.RawMessage(`{}`), true},
		{"version 0", 0, json.RawMessage(`{}`), false},
		{"version 1001", 1001, json.RawMessage(`{}`), false},
		{"negative version", -1, json.RawMessage(`{}`), false},
		{"empty", 1, nil, false},
		{"array", 1, json.RawMessage(`[]`), false},
		{"scalar", 1, json.RawMessage(`"x"`), false},
		{"null", 1, json.RawMessage(`null`), false},
		{"truncated", 1, json.RawMessage(`{"a":`), false},
		{"depth 4", 1, deep4, true},
		{"depth 5", 1, deep5, false},
		{"arrays count", 1, json.RawMessage(`{"a":[[[[1]]]]}`), false},
		{"brackets in strings do not", 1, json.RawMessage(`{"a":"{{{{{{[[[[\"}"}`), true},
		{"leading space", 1, json.RawMessage(` {"a":1}`), true},
	}
	for _, c := range cases {
		err := Validate(c.version, c.body)
		if (err == nil) != c.ok {
			t.Errorf("%s: Validate = %v, want ok=%v", c.name, err, c.ok)
		}
		if err != nil && !errors.Is(err, ErrInvalid) {
			t.Errorf("%s: %v does not wrap ErrInvalid", c.name, err)
		}
	}
}

func TestSettingsAreScopedToTheUser(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	a, b := mustUser(t, d, "A"), mustUser(t, d, "B")
	if _, err := s.Put(ctx, a, 1, json.RawMessage(`{"who":"a"}`), 0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, b); !errors.Is(err, ErrNotFound) {
		t.Errorf("B reads A's settings: %v", err)
	}
	if _, err := s.Put(ctx, b, 1, json.RawMessage(`{"who":"b"}`), 0); err != nil {
		t.Errorf("B's first write: %v", err)
	}
}

func TestPutRefusesAnUnknownUser(t *testing.T) {
	s, _ := openStore(t)
	if _, err := s.Put(context.Background(), uuid.New(), 1, json.RawMessage(`{}`), 0); err == nil {
		t.Error("Put for a user that does not exist succeeded")
	}
	if _, err := s.Put(context.Background(), uuid.Nil, 1, json.RawMessage(`{}`), 0); err == nil {
		t.Error("Put for the nil user succeeded")
	}
}

// Of N writers holding the same revision, exactly one wins.
func TestConcurrentPutsOnOneRevisionHaveOneWinner(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d, "A")
	if _, err := s.Put(ctx, u, 1, json.RawMessage(`{"n":0}`), 0); err != nil {
		t.Fatal(err)
	}
	const n = 8
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		wins int
	)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := s.Put(ctx, u, 1, json.RawMessage(`{"n":1}`), 1)
			if err == nil {
				mu.Lock()
				wins++
				mu.Unlock()
			} else if !errors.Is(err, ErrRevisionConflict) && !isBusy(err) {
				t.Errorf("Put: %v", err)
			}
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Errorf("%d writers won, want exactly 1", wins)
	}
	if read, _ := s.Get(ctx, u); read.Revision != 2 {
		t.Errorf("revision = %d, want 2", read.Revision)
	}
}

func TestNoStore(t *testing.T) {
	var s Store = NoStore{}
	if _, err := s.Get(context.Background(), uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get = %v", err)
	}
	if _, err := s.Put(context.Background(), uuid.New(), 1, json.RawMessage(`{}`), 0); !errors.Is(err, ErrNoStore) {
		t.Errorf("Put = %v", err)
	}
}
