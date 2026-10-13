package budget

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openSQL(t *testing.T, dir string) (*SQLStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), dir)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLStore(d), d
}

func rowCount(t *testing.T, d *db.DB) int {
	t.Helper()
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM bot_model_budget`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

// The counter and the switch survive closing and reopening the real
// database file, which is what a deploy does.
func TestSQLStorePersistsAcrossReopeningTheDatabase(t *testing.T) {
	dir := t.TempDir()
	clk := newClock("2026-10-12T09:00:00Z")

	s1, d1 := openSQL(t, dir)
	b1 := New(Limits{DailyTokens: 150, MonthlyTokens: 1000}, s1, WithClock(clk.now), WithLogger(quiet()))
	c1 := b1.Wrap(usageClient(100, 0))
	_ = call(t, c1)
	_ = call(t, c1)
	if err := b1.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	if err := d1.Close(); err != nil {
		t.Fatal(err)
	}

	s2, _ := openSQL(t, dir)
	b2 := New(Limits{DailyTokens: 150, MonthlyTokens: 1000}, s2, WithClock(clk.now), WithLogger(quiet()))
	st := b2.Status()
	if st.DailyTokens != 200 || st.MonthlyTokens != 200 || st.Enabled {
		t.Fatalf("after reopening: %+v, want 200/200 and the switch off", st)
	}
	_ = b2.SetEnabled(true)
	if err := call(t, b2.Wrap(usageClient(1, 1))); err != model.ErrBudgetExceeded {
		t.Fatalf("daily cap must hold across the restart: err = %v", err)
	}
}

// Old periods are pruned on every save, so the table never grows past
// one day, one month and the switch.
func TestSQLStoreRollsOverAndPrunesOldPeriods(t *testing.T) {
	s, d := openSQL(t, t.TempDir())
	clk := newClock("2026-10-31T23:00:00Z")
	b := New(Limits{DailyTokens: 1000, MonthlyTokens: 1000}, s, WithClock(clk.now), WithLogger(quiet()))
	c := b.Wrap(usageClient(10, 0))
	_ = call(t, c)
	_ = b.SetEnabled(true)
	if n := rowCount(t, d); n != 3 {
		t.Fatalf("rows = %d, want day + month + switch", n)
	}

	clk.set("2026-11-01T01:00:00Z") // new day AND new month
	_ = call(t, c)
	if n := rowCount(t, d); n != 3 {
		t.Fatalf("rows = %d after the roll-over, want still 3 (old rows pruned)", n)
	}
	snap, err := s.Load(context.Background(), "2026-11-01", "2026-11")
	if err != nil || snap.DayTokens != 10 || snap.MonthTokens != 10 {
		t.Fatalf("Load = %+v, %v, want 10/10 for the new periods", snap, err)
	}
	old, _ := s.Load(context.Background(), "2026-10-31", "2026-10")
	if old.DayTokens != 0 || old.MonthTokens != 0 {
		t.Fatalf("old periods = %+v, want gone", old)
	}
}

func TestSQLStoreSwitchRoundTrips(t *testing.T) {
	s, _ := openSQL(t, t.TempDir())
	ctx := context.Background()
	snap, err := s.Load(ctx, "2026-10-12", "2026-10")
	if err != nil || snap != (Snapshot{}) {
		t.Fatalf("empty store: %+v, %v", snap, err)
	}
	if err := s.SaveDisabled(ctx, true); err != nil {
		t.Fatal(err)
	}
	if snap, _ = s.Load(ctx, "2026-10-12", "2026-10"); !snap.Disabled {
		t.Fatal("switch did not persist")
	}
	if err := s.SaveDisabled(ctx, false); err != nil {
		t.Fatal(err)
	}
	if snap, _ = s.Load(ctx, "2026-10-12", "2026-10"); snap.Disabled {
		t.Fatal("switch did not clear")
	}
}
