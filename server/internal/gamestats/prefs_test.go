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

// The share link (ADR 0217 §6, slice 6): minting stores only the hash,
// minting again replaces it, revoking clears it, and neither touches the
// advanced switch.
func TestPrefsStoreShareLink(t *testing.T) {
	s, d := openPrefs(t)
	ctx := context.Background()
	mustUser(t, d, "u1")
	if _, err := s.SetShareLink(ctx, "nobody", "h0", time.UnixMilli(1)); !errors.Is(err, ErrNoUser) {
		t.Errorf("a missing account: %v, want ErrNoUser", err)
	}
	if _, err := s.SetAdvanced(ctx, "u1", true, SetBySelf, time.UnixMilli(5)); err != nil {
		t.Fatal(err)
	}
	p, err := s.SetShareLink(ctx, "u1", "h1", time.UnixMilli(10))
	if err != nil || p.ShareLinkAt.UnixMilli() != 10 || p.Advanced == nil || !*p.Advanced {
		t.Fatalf("after minting: %+v %v; want the link's time and the switch kept", p, err)
	}
	if id, err := s.ShareLinkOwner(ctx, "h1"); err != nil || id != "u1" {
		t.Fatalf("ShareLinkOwner(h1) = %q, %v", id, err)
	}
	if _, err := s.SetShareLink(ctx, "u1", "h2", time.UnixMilli(20)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ShareLinkOwner(ctx, "h1"); !errors.Is(err, ErrNoShareLink) {
		t.Errorf("the rotated-out link still works: %v", err)
	}
	if id, _ := s.ShareLinkOwner(ctx, "h2"); id != "u1" {
		t.Errorf("the new link does not resolve: %q", id)
	}
	if err := s.ClearShareLink(ctx, "u1", time.UnixMilli(30)); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ShareLinkOwner(ctx, "h2"); !errors.Is(err, ErrNoShareLink) {
		t.Errorf("a revoked link still works: %v", err)
	}
	p, _ = s.Prefs(ctx, "u1")
	if !p.ShareLinkAt.IsZero() || p.Advanced == nil || !*p.Advanced {
		t.Errorf("after revoking: %+v; want no link and the switch kept", p)
	}
	if err := s.ClearShareLink(ctx, "nobody", time.UnixMilli(40)); err != nil {
		t.Errorf("revoking with no link: %v", err)
	}
	if _, err := s.ShareLinkOwner(ctx, ""); !errors.Is(err, ErrNoShareLink) {
		t.Errorf("an empty hash: %v", err)
	}
}

// A share link on a row with no switch keeps advanced NULL (the
// default), which the table's CHECK requires along with set_by.
func TestPrefsStoreShareLinkWithoutASwitch(t *testing.T) {
	s, d := openPrefs(t)
	ctx := context.Background()
	mustUser(t, d, "u1")
	p, err := s.SetShareLink(ctx, "u1", "h1", time.UnixMilli(10))
	if err != nil || p.Advanced != nil || p.SetBy != "" {
		t.Fatalf("SetShareLink: %+v %v; want the default kept", p, err)
	}
	if list, _ := s.Accounts(ctx, 0); len(list) != 0 {
		t.Errorf("a link alone listed the account as switched: %+v", list)
	}
	if _, err := d.Exec(`UPDATE users SET deleted_at = 9 WHERE id = 'u1'`); err != nil {
		t.Fatal(err)
	}
	if _, err := s.ShareLinkOwner(ctx, "h1"); !errors.Is(err, ErrNoShareLink) {
		t.Errorf("a deleted account's link still works: %v", err)
	}
}

func TestSetSharePodsLeavesTheOtherColumnsAlone(t *testing.T) {
	s, d := openPrefs(t)
	mustUser(t, d, "u1")
	mustUser(t, d, "u2")
	ctx := context.Background()
	now := time.UnixMilli(1000)
	if _, err := s.SetAdvanced(ctx, "u1", true, SetByAdmin, now); err != nil {
		t.Fatal(err)
	}
	got, err := s.SetSharePods(ctx, "u1", true, now)
	if err != nil {
		t.Fatal(err)
	}
	if !got.SharePods || got.Advanced == nil || !*got.Advanced || got.SetBy != SetByAdmin {
		t.Errorf("after sharing: %+v, want share_pods and the admin's advanced switch both kept", got)
	}
	got, err = s.SetSharePods(ctx, "u1", false, now)
	if err != nil || got.SharePods {
		t.Errorf("switching off: %+v %v", got, err)
	}
	if _, err := s.SetSharePods(ctx, "nobody", true, now); !errors.Is(err, ErrNoUser) {
		t.Errorf("an unknown account: %v, want ErrNoUser", err)
	}
	// A person with no row at all can switch it on, and keeps following the default.
	got, err = s.SetSharePods(ctx, "u2", true, now)
	if err != nil || !got.SharePods || got.Advanced != nil {
		t.Errorf("a new row: %+v %v, want sharing and advanced following the default", got, err)
	}
}
