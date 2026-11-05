package gamestats

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func openPrefs(t *testing.T) (*SQLPrefsStore, *db.DB) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	return NewSQLPrefsStore(d), d
}

func mustUser(t *testing.T, d *db.DB, id string) {
	t.Helper()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'A', 1, 1)`, id); err != nil {
		t.Fatalf("insert user: %v", err)
	}
}

func TestAdvancedDefaultIsOff(t *testing.T) {
	if AdvancedDefault {
		t.Fatal("AdvancedDefault is true; ADR 0217 owner answer 1 keeps advanced stats opt-in")
	}
}

func TestPrefsEffectiveFollowsTheDefaultOnlyWhenUnset(t *testing.T) {
	on, off := true, false
	cases := []struct {
		p    Prefs
		def  bool
		want bool
	}{
		{Prefs{}, false, false},
		{Prefs{}, true, true},
		{Prefs{Advanced: &on, SetBy: SetBySelf}, false, true},
		// Anyone who switched it off themselves stays off when the
		// default flips.
		{Prefs{Advanced: &off, SetBy: SetBySelf}, true, false},
		{Prefs{Advanced: &on, SetBy: SetByAdmin}, false, true},
	}
	for i, c := range cases {
		if got := c.p.Effective(c.def); got != c.want {
			t.Errorf("case %d: Effective(%v) = %v, want %v", i, c.def, got, c.want)
		}
	}
}

func TestPrefsStoreRoundTrip(t *testing.T) {
	s, d := openPrefs(t)
	ctx := context.Background()
	mustUser(t, d, "u1")
	p, err := s.Prefs(ctx, "u1")
	if err != nil || p.Advanced != nil || p.SetBy != "" {
		t.Fatalf("no row: %+v %v; want the default", p, err)
	}
	now := time.UnixMilli(5000)
	if _, err := s.SetAdvanced(ctx, "u1", true, SetByAdmin, now); err != nil {
		t.Fatal(err)
	}
	p, _ = s.Prefs(ctx, "u1")
	if p.Advanced == nil || !*p.Advanced || p.SetBy != SetByAdmin || p.UpdatedAt.UnixMilli() != 5000 {
		t.Fatalf("after the admin's switch: %+v", p)
	}
	if _, err := s.SetAdvanced(ctx, "u1", false, SetBySelf, now.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	p, _ = s.Prefs(ctx, "u1")
	if p.Advanced == nil || *p.Advanced || p.SetBy != SetBySelf {
		t.Fatalf("after the person's own switch: %+v", p)
	}
	list, err := s.Accounts(ctx, 0)
	if err != nil || len(list) != 1 || list[0].UserID != "u1" {
		t.Fatalf("Accounts = %+v, %v", list, err)
	}
}

func TestPrefsStoreRefusesAMissingOrDeletedAccountAndABadSetter(t *testing.T) {
	s, d := openPrefs(t)
	ctx := context.Background()
	if _, err := s.SetAdvanced(ctx, "nobody", true, SetBySelf, time.Now()); !errors.Is(err, ErrNoUser) {
		t.Errorf("missing account: %v, want ErrNoUser", err)
	}
	mustUser(t, d, "gone")
	if _, err := d.Exec(`UPDATE users SET deleted_at = 9 WHERE id = 'gone'`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetAdvanced(ctx, "gone", true, SetByAdmin, time.Now()); !errors.Is(err, ErrNoUser) {
		t.Errorf("deleted account: %v, want ErrNoUser", err)
	}
	mustUser(t, d, "u1")
	if _, err := s.SetAdvanced(ctx, "u1", true, SetBy("bot"), time.Now()); err == nil {
		t.Error("a setter other than self or admin was accepted")
	}
}

func TestPrefsStoreSwitchKeepsTheSharingColumns(t *testing.T) {
	s, d := openPrefs(t)
	ctx := context.Background()
	mustUser(t, d, "u1")
	if _, err := d.Exec(`INSERT INTO stats_prefs (user_id, share_pods, share_link_hash, share_link_at, updated_at) VALUES ('u1', 1, 'abc', 7, 7)`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetAdvanced(ctx, "u1", true, SetBySelf, time.UnixMilli(9)); err != nil {
		t.Fatal(err)
	}
	var pods int
	var hash string
	if err := d.QueryRow(`SELECT share_pods, share_link_hash FROM stats_prefs WHERE user_id = 'u1'`).Scan(&pods, &hash); err != nil {
		t.Fatal(err)
	}
	if pods != 1 || hash != "abc" {
		t.Errorf("the switch changed the sharing columns: share_pods %d, hash %q", pods, hash)
	}
}

func TestPrefsStoreDefault(t *testing.T) {
	s, _ := openPrefs(t)
	ctx := context.Background()
	def, err := s.Default(ctx)
	if err != nil || def.Stored || def.Advanced != AdvancedDefault {
		t.Fatalf("unset default: %+v %v; want AdvancedDefault, not stored", def, err)
	}
	if _, err := s.SetDefault(ctx, true, time.UnixMilli(10)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetDefault(ctx, true, time.UnixMilli(11)); err != nil {
		t.Fatal("second SetDefault:", err)
	}
	def, _ = s.Default(ctx)
	if !def.Advanced || !def.Stored || def.UpdatedAt.UnixMilli() != 11 {
		t.Fatalf("stored default: %+v", def)
	}
}
