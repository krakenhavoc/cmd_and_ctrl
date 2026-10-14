package notify

import (
	"context"
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

func mustUser(t *testing.T, d *db.DB) uuid.UUID {
	t.Helper()
	id := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'A', 1, 1)`, id.String()); err != nil {
		t.Fatalf("insert user: %v", err)
	}
	return id
}

func on() *bool  { v := true; return &v }
func off() *bool { v := false; return &v }

// ADR 0162 §1: off until the player opts in. No row is every switch off,
// and nobody without a switch on is ever Active.
func TestStoreIsOffByDefault(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d)
	p, err := s.Get(ctx, u)
	if err != nil {
		t.Fatal(err)
	}
	if p != (Prefs{}) {
		t.Errorf("Get with no row = %+v, want every switch off", p)
	}
	for _, k := range []Kind{KindTurn, KindTableReady, KindTableFilled} {
		if p.Wants(k) {
			t.Errorf("the default wants %s", k)
		}
	}
	if _, err := s.Put(ctx, u, Update{Turn: off()}); err != nil {
		t.Fatal(err)
	}
	active, err := s.Active(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(active) != 0 {
		t.Errorf("Active = %v, want nobody: a saved row with every switch off is still off", active)
	}
}

func TestStorePutAppliesOnlyTheNamedSwitches(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d)
	if _, err := s.Put(ctx, u, Update{Turn: on(), TableFilled: on()}); err != nil {
		t.Fatal(err)
	}
	p, err := s.Put(ctx, u, Update{TableFilled: off()})
	if err != nil {
		t.Fatal(err)
	}
	if !p.Turn || p.TableReady || p.TableFilled {
		t.Errorf("after two puts: %+v, want turn on and the rest off", p)
	}
	if got, _ := s.Get(ctx, u); got != p {
		t.Errorf("Get = %+v, Put returned %+v", got, p)
	}
	active, _ := s.Active(ctx)
	if got, ok := active[u]; !ok || !got.Turn {
		t.Errorf("Active = %v, want %s with turn on", active, u)
	}
}

// ADR 0162 §4: a refusal stops every DM, and turning a switch off and on
// again clears it. Turning one off, or re-sending one that is already
// on, does not.
func TestStoreRefusalStopsAndASwitchTurnedOnClearsIt(t *testing.T) {
	ctx := context.Background()
	s, d := openStore(t)
	u := mustUser(t, d)
	if _, err := s.Put(ctx, u, Update{Turn: on(), TableReady: on()}); err != nil {
		t.Fatal(err)
	}
	at := time.UnixMilli(1_800_000_000_000).UTC()
	if err := s.MarkFailed(ctx, u, at); err != nil {
		t.Fatal(err)
	}
	p, _ := s.Get(ctx, u)
	if !p.DMFailedAt.Equal(at) || p.Wants(KindTurn) {
		t.Errorf("after MarkFailed: %+v, want the refusal recorded and nothing wanted", p)
	}
	if active, _ := s.Active(ctx); len(active) != 0 {
		t.Errorf("a refused person is still Active: %v", active)
	}
	for _, u2 := range []Update{{Turn: on()}, {TableReady: off()}} {
		if p, _ = s.Put(ctx, u, u2); p.DMFailedAt.IsZero() {
			t.Fatalf("Put %+v cleared the refusal; only turning a switch on may", u2)
		}
	}
	if p, _ = s.Put(ctx, u, Update{TableReady: on()}); !p.DMFailedAt.IsZero() {
		t.Errorf("turning table_ready back on left the refusal: %+v", p)
	}
	if active, _ := s.Active(ctx); len(active) != 1 {
		t.Errorf("after clearing, Active = %v, want the person back", active)
	}
}
