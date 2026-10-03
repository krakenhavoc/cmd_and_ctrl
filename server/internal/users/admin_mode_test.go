package users

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/discord"
)

// Tests for ADR 0112 §2 item 1: admin mode is per-person server state,
// off by default, written through, lapsing after 12 hours, and off on
// every failure.

type flakyModeStore struct {
	AdminModeStore
	loads      atomic.Int32
	failLoad   bool
	failWrites bool
}

func (f *flakyModeStore) AdminModesOn(ctx context.Context) (map[uuid.UUID]time.Time, error) {
	f.loads.Add(1)
	if f.failLoad {
		return nil, errors.New("disk on fire")
	}
	return f.AdminModeStore.AdminModesOn(ctx)
}

func (f *flakyModeStore) SetAdminMode(ctx context.Context, id uuid.UUID, at time.Time) error {
	if f.failWrites {
		return errors.New("disk full")
	}
	return f.AdminModeStore.SetAdminMode(ctx, id, at)
}

func (f *flakyModeStore) ClearAdminModeIfAt(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	if f.failWrites {
		return false, errors.New("disk full")
	}
	return f.AdminModeStore.ClearAdminModeIfAt(ctx, id, at)
}

func storedAdminModeAt(t *testing.T, s *SQLStore, id uuid.UUID) int64 {
	t.Helper()
	var ms int64
	if err := s.db.QueryRow(`SELECT admin_mode_at FROM users WHERE id = ?`, id.String()).Scan(&ms); err != nil {
		t.Fatalf("read admin_mode_at: %v", err)
	}
	return ms
}

func TestAdminModeIsOffByDefault(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, err := s.UpsertFromDiscord(ctx, alice, "", "identify")
	if err != nil {
		t.Fatal(err)
	}
	m, err := NewAdminModes(ctx, s)
	if err != nil {
		t.Fatalf("NewAdminModes: %v", err)
	}
	if m.On(u.ID, time.Now()) {
		t.Error("a new person is in admin mode")
	}
	var none *AdminModes
	if none.On(u.ID, time.Now()) {
		t.Error("a nil set is in admin mode")
	}
	if m.On(uuid.Nil, time.Now()) {
		t.Error("uuid.Nil is in admin mode")
	}
}

func TestAdminModeSetWritesThroughAndSurvivesARestart(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, err := s.UpsertFromDiscord(ctx, alice, "", "identify")
	if err != nil {
		t.Fatal(err)
	}
	fs := &flakyModeStore{AdminModeStore: s}
	m, _ := NewAdminModes(ctx, fs)
	at := time.Date(2026, 10, 2, 20, 0, 0, 123_456_789, time.UTC)
	m.SetClock(func() time.Time { return at })

	ends, err := m.Set(ctx, u.ID, true)
	if err != nil {
		t.Fatalf("Set on: %v", err)
	}
	if want := at.Truncate(time.Millisecond).Add(12 * time.Hour); !ends.Equal(want) {
		t.Errorf("ends %v, want %v", ends, want)
	}
	if !m.On(u.ID, at) {
		t.Error("not on after Set(true)")
	}
	if got := storedAdminModeAt(t, s, u.ID); got != at.UnixMilli() {
		t.Errorf("admin_mode_at = %d, want %d", got, at.UnixMilli())
	}
	// On is answered from memory: no reload.
	for i := 0; i < 50; i++ {
		m.On(u.ID, at)
	}
	if n := fs.loads.Load(); n != 1 {
		t.Errorf("AdminModesOn called %d times, want once at construction", n)
	}

	// A restart loads it back.
	again, err := NewAdminModes(ctx, s)
	if err != nil || !again.On(u.ID, at.Add(time.Hour)) {
		t.Errorf("after a restart: on=%v err=%v, want on", again.On(u.ID, at.Add(time.Hour)), err)
	}

	if _, err := m.Set(ctx, u.ID, false); err != nil {
		t.Fatalf("Set off: %v", err)
	}
	if m.On(u.ID, at) || storedAdminModeAt(t, s, u.ID) != 0 {
		t.Error("Set(false) left admin mode on")
	}
	// Off while off is a no-op, not an error.
	if _, err := m.Set(ctx, u.ID, false); err != nil {
		t.Errorf("Set off twice: %v", err)
	}
	if _, err := m.Set(ctx, uuid.New(), true); !errors.Is(err, ErrNotFound) {
		t.Errorf("Set for an unknown user: %v, want ErrNotFound", err)
	}
}

func TestAdminModeLapsesAfterTwelveHours(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	m, _ := NewAdminModes(ctx, s)
	at := time.Date(2026, 10, 2, 20, 0, 0, 0, time.UTC)
	m.SetClock(func() time.Time { return at })
	if _, err := m.Set(ctx, u.ID, true); err != nil {
		t.Fatal(err)
	}
	if !m.On(u.ID, at.Add(12*time.Hour-time.Millisecond)) {
		t.Error("off before 12 hours")
	}
	// HTTP sees the lapse at the next request, with no write.
	if m.On(u.ID, at.Add(12*time.Hour)) {
		t.Error("still on at exactly 12 hours")
	}
	if storedAdminModeAt(t, s, u.ID) == 0 {
		t.Fatal("the lapse wrote before the sweep")
	}

	// Switching on again restarts the 12 hours.
	m.SetClock(func() time.Time { return at.Add(11 * time.Hour) })
	if _, err := m.Set(ctx, u.ID, true); err != nil {
		t.Fatal(err)
	}
	if !m.On(u.ID, at.Add(22*time.Hour)) {
		t.Error("switching on again did not restart the 12 hours")
	}

	// The sweep ends a lapsed mode in the database and the map.
	ended, err := m.Sweep(ctx, at.Add(11*time.Hour+12*time.Hour))
	if err != nil {
		t.Fatalf("Sweep: %v", err)
	}
	if len(ended) != 1 || ended[0] != u.ID {
		t.Errorf("Sweep ended %v, want [%s]", ended, u.ID)
	}
	if storedAdminModeAt(t, s, u.ID) != 0 {
		t.Error("the sweep left admin_mode_at set")
	}
	if ended, _ := m.Sweep(ctx, at.Add(48*time.Hour)); len(ended) != 0 {
		t.Errorf("a second sweep ended %v, want nothing", ended)
	}
}

// A lapse found at boot (the server was down when it lapsed) is swept
// on the first tick, and reads as off before that.
func TestAdminModeLapsedAtBootReadsOffAndIsSwept(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	long := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	if err := s.SetAdminMode(ctx, u.ID, long); err != nil {
		t.Fatal(err)
	}
	m, err := NewAdminModes(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	now := long.Add(13 * time.Hour)
	if m.On(u.ID, now) {
		t.Error("a mode that lapsed while the server was down reads on")
	}
	if ended, err := m.Sweep(ctx, now); err != nil || len(ended) != 1 {
		t.Errorf("first sweep: %v %v", ended, err)
	}
	if storedAdminModeAt(t, s, u.ID) != 0 {
		t.Error("not cleared")
	}
}

// The sweep never undoes a switch made after it read the row.
func TestAdminModeSweepDoesNotClearANewerSwitch(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	old := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	newer := old.Add(20 * time.Hour)
	if err := s.SetAdminMode(ctx, u.ID, newer); err != nil {
		t.Fatal(err)
	}
	if cleared, err := s.ClearAdminModeIfAt(ctx, u.ID, old); err != nil || cleared {
		t.Errorf("ClearAdminModeIfAt(old) = %v, %v; want false: the row moved on", cleared, err)
	}
	if storedAdminModeAt(t, s, u.ID) != newer.UnixMilli() {
		t.Error("the newer switch was cleared")
	}
}

// A failed boot load starts with nobody in admin mode (ADR 0112 §2
// item 1): the safe failure is fewer admins, not a refused boot.
func TestAdminModeBootLoadFailureIsPlayerModeForEveryone(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	now := time.Now().UTC()
	if err := s.SetAdminMode(ctx, u.ID, now); err != nil {
		t.Fatal(err)
	}
	m, err := NewAdminModes(ctx, &flakyModeStore{AdminModeStore: s, failLoad: true})
	if err == nil {
		t.Fatal("a failed load reported no error")
	}
	if m == nil {
		t.Fatal("a failed load returned no set; the server needs one to boot")
	}
	if m.On(u.ID, now) {
		t.Error("a failed load left someone in admin mode")
	}
	if m2, err := NewAdminModes(ctx, nil); err == nil || m2 == nil || m2.On(u.ID, now) {
		t.Errorf("no store: %v %v", m2, err)
	}
}

// A write that fails: switching on leaves the person off; switching off
// drops them anyway. Both report the error.
func TestAdminModeFailedWritesFailClosed(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	fs := &flakyModeStore{AdminModeStore: s}
	m, _ := NewAdminModes(ctx, fs)
	now := time.Now().UTC()

	fs.failWrites = true
	if _, err := m.Set(ctx, u.ID, true); err == nil {
		t.Error("a failed write switched on without an error")
	}
	if m.On(u.ID, now) {
		t.Error("a failed write switched admin mode on")
	}

	fs.failWrites = false
	if _, err := m.Set(ctx, u.ID, true); err != nil {
		t.Fatal(err)
	}
	fs.failWrites = true
	if _, err := m.Set(ctx, u.ID, false); err == nil {
		t.Error("a failed write switched off without an error")
	}
	if m.On(u.ID, now) {
		t.Error("a failed switch-off left admin mode on in memory")
	}

	// A sweep whose clear fails still drops the person.
	fs.failWrites = false
	m.SetClock(func() time.Time { return now })
	if _, err := m.Set(ctx, u.ID, true); err != nil {
		t.Fatal(err)
	}
	fs.failWrites = true
	ended, err := m.Sweep(ctx, now.Add(13*time.Hour))
	if err == nil || len(ended) != 1 {
		t.Errorf("sweep with a failing clear: %v %v; want the user ended and the error", ended, err)
	}
	if m.On(u.ID, now) {
		t.Error("a failed sweep left the person in the map")
	}
}

// A row from the future (the clock went back) is not trusted.
func TestAdminModeFromTheFutureIsOff(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	now := time.Date(2026, 10, 2, 0, 0, 0, 0, time.UTC)
	if err := s.SetAdminMode(ctx, u.ID, now.Add(6*time.Hour)); err != nil {
		t.Fatal(err)
	}
	m, _ := NewAdminModes(ctx, s)
	if m.On(u.ID, now) {
		t.Error("a switch six hours in the future reads on")
	}
}

// Signing out everywhere ends admin mode in the same statement, and in
// the cache (ADR 0112 §2 item 7).
func TestRevokeAllEndsAdminMode(t *testing.T) {
	ctx := context.Background()
	s, _ := openStore(t, nil)
	u, _ := s.UpsertFromDiscord(ctx, alice, "", "identify")
	other, _ := s.UpsertFromDiscord(ctx, discord.User{ID: "222", Username: "bob"}, "", "identify")
	m, _ := NewAdminModes(ctx, s)
	r, err := NewRevocations(ctx, s)
	if err != nil {
		t.Fatal(err)
	}
	r.EndAdminModeOnRevoke(m)
	now := time.Now().UTC()
	for _, id := range []uuid.UUID{u.ID, other.ID} {
		if _, err := m.Set(ctx, id, true); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := r.RevokeAll(ctx, u.ID); err != nil {
		t.Fatalf("RevokeAll: %v", err)
	}
	if m.On(u.ID, now) {
		t.Error("admin mode survived RevokeAll in the cache")
	}
	if storedAdminModeAt(t, s, u.ID) != 0 {
		t.Error("admin mode survived RevokeAll in the database")
	}
	if !m.On(other.ID, now) || storedAdminModeAt(t, s, other.ID) == 0 {
		t.Error("RevokeAll ended someone else's admin mode")
	}
	// A restart agrees.
	again, _ := NewAdminModes(ctx, s)
	if again.On(u.ID, now) {
		t.Error("admin mode came back after a restart")
	}
}
