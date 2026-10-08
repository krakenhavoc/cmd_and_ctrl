package playmat

import (
	"context"
	"errors"
	"image/color"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func newService(t *testing.T, f *Fetcher) (*Service, string, uuid.UUID) {
	t.Helper()
	d, err := db.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { _ = d.Close() })
	user := uuid.New()
	if _, err := d.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'A', 1, 1)`, user.String()); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(t.TempDir(), "playmats")
	return NewService(NewFileStore(dir), d.DB, f), dir, user
}

func filesIn(t *testing.T, dir string) []string {
	t.Helper()
	ents, err := os.ReadDir(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range ents {
		out = append(out, e.Name())
	}
	return out
}

// checkInvariant is ADR 0128 §11's rule, asked of the database itself:
// no user's active pointer names an image that is not saved in one of
// that user's own slots. Every write test ends here.
func checkInvariant(t *testing.T, s *Service) {
	t.Helper()
	var n int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM users u WHERE u.playmat_id IS NOT NULL AND NOT EXISTS
		(SELECT 1 FROM user_playmats p WHERE p.user_id = u.id AND p.playmat_id = u.playmat_id)`).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("%d users have an active playmat that is not one of their saved slots", n)
	}
}

func mat(t *testing.T, w, h int, c color.Color) []byte { return pngBytes(t, solid(w, h, c)) }

func TestSetStoresAJPEGInTheSlotAndMakesTheFirstMatActive(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	if got := s.URL(user); got != "" {
		t.Fatalf("a new user has URL %q", got)
	}
	sl, err := s.SetFromBytes(ctx, user, 2, mat(t, 64, 48, color.White))
	if err != nil {
		t.Fatal(err)
	}
	if sl.Slot != 2 || sl.Width != 64 || sl.Height != 48 || !strings.HasPrefix(sl.URL, "/playmats/") {
		t.Errorf("slot = %+v", sl)
	}
	if s.URL(user) != sl.URL {
		t.Errorf("URL() = %q, want the first mat saved to be active (%q)", s.URL(user), sl.URL)
	}
	st, err := s.State(ctx, user)
	if err != nil || st.Active != 2 || len(st.Slots) != 1 || st.Slots[0].ID != sl.ID {
		t.Errorf("State = %+v, %v", st, err)
	}
	if fs := filesIn(t, dir); len(fs) != 1 || fs[0] != sl.ID+".jpg" {
		t.Errorf("files = %v", fs)
	}
	checkInvariant(t, s)
}

func TestASecondMatDoesNotChangeTheActiveOne(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	one, err := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White))
	if err != nil {
		t.Fatal(err)
	}
	two, err := s.SetFromBytes(ctx, user, 3, mat(t, 30, 30, color.Black))
	if err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != one.URL {
		t.Errorf("URL() = %q, want slot 1's %q: saving a second mat must not switch", s.URL(user), one.URL)
	}
	st, _ := s.State(ctx, user)
	if len(st.Slots) != 2 || st.Active != 1 || st.Slots[1].ID != two.ID {
		t.Errorf("State = %+v", st)
	}
	if fs := filesIn(t, dir); len(fs) != 2 {
		t.Errorf("files = %v, want both", fs)
	}
	checkInvariant(t, s)
}

func TestSavingIntoAnEmptySlotWhileShowingNoneShowsIt(t *testing.T) {
	// The rule is "an empty slot while none is active": saving a NEW
	// mat shows it, and replacing a mat that was already saved does not
	// turn a deliberately empty table back on.
	s, _, user := newService(t, nil)
	ctx := context.Background()
	if _, err := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White)); err != nil {
		t.Fatal(err)
	}
	if err := s.Activate(ctx, user, 0); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != "" {
		t.Fatal("Activate(0) left a mat on show")
	}
	two, err := s.SetFromBytes(ctx, user, 2, mat(t, 20, 20, color.Black))
	if err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != two.URL {
		t.Errorf("URL() = %q, want the new mat shown (%q)", s.URL(user), two.URL)
	}
	// Replacing a FILLED slot while none is active changes nothing shown.
	if err := s.Activate(ctx, user, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetFromBytes(ctx, user, 1, mat(t, 22, 22, color.White)); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != "" {
		t.Errorf("replacing a filled slot turned a mat on: %q", s.URL(user))
	}
	checkInvariant(t, s)
}

func TestReplacingASlotDeletesTheOldFileAndFollowsTheActivePointer(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	first, err := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White))
	if err != nil {
		t.Fatal(err)
	}
	second, err := s.SetFromBytes(ctx, user, 1, mat(t, 30, 30, color.Black))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID == second.ID {
		t.Fatal("a replacement kept the id")
	}
	fs := filesIn(t, dir)
	if len(fs) != 1 || fs[0] != second.ID+".jpg" {
		t.Errorf("files = %v, want only the new one", fs)
	}
	if _, _, err := s.Open(first.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("the old id still opens: %v", err)
	}
	if s.URL(user) != second.URL {
		t.Errorf("URL() = %q, want the replacement %q (slot 1 was the active one)", s.URL(user), second.URL)
	}
	checkInvariant(t, s)
}

func TestARejectedReplacementKeepsTheOldPlaymat(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	first, err := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.SetFromBytes(ctx, user, 1, []byte("not an image")); !errors.Is(err, ErrNotImage) {
		t.Fatalf("err = %v", err)
	}
	if s.URL(user) != first.URL {
		t.Error("a refused upload changed the playmat")
	}
	if fs := filesIn(t, dir); len(fs) != 1 {
		t.Errorf("files = %v", fs)
	}
	checkInvariant(t, s)
}

func TestSlotsOutsideOneToThreeAreRefused(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	for _, n := range []int{-1, 0, 4, 99} {
		if _, err := s.SetFromBytes(ctx, user, n, mat(t, 8, 8, color.White)); !errors.Is(err, ErrBadSlot) {
			t.Errorf("SetFromBytes slot %d = %v, want ErrBadSlot", n, err)
		}
		if _, err := s.SetFromURL(ctx, user, n, "https://example.com/x.png"); !errors.Is(err, ErrBadSlot) {
			t.Errorf("SetFromURL slot %d = %v, want ErrBadSlot", n, err)
		}
		if err := s.Remove(ctx, user, n); !errors.Is(err, ErrBadSlot) {
			t.Errorf("Remove slot %d = %v, want ErrBadSlot", n, err)
		}
		if _, err := s.Fit(ctx, user, n, 0, 0); !errors.Is(err, ErrBadSlot) {
			t.Errorf("Fit slot %d = %v, want ErrBadSlot", n, err)
		}
	}
	if err := s.Activate(ctx, user, 4); !errors.Is(err, ErrBadSlot) {
		t.Errorf("Activate 4 = %v, want ErrBadSlot", err)
	}
	if fs := filesIn(t, dir); len(fs) != 0 {
		t.Errorf("a refused slot left %v behind", fs)
	}
	// The schema says the same, as a backstop.
	if _, err := s.db.Exec(`INSERT INTO user_playmats (user_id, slot, playmat_id, created_at) VALUES (?, 4, 'x', 1)`, user.String()); err == nil {
		t.Error("the table accepted slot 4")
	}
}

func TestActivateSwitchesWithoutTouchingFiles(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	one, _ := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White))
	two, _ := s.SetFromBytes(ctx, user, 2, mat(t, 30, 30, color.Black))
	before := filesIn(t, dir)

	if err := s.Activate(ctx, user, 2); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != two.URL {
		t.Errorf("URL() = %q, want %q", s.URL(user), two.URL)
	}
	if err := s.Activate(ctx, user, 0); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != "" {
		t.Errorf("URL() = %q, want none", s.URL(user))
	}
	if err := s.Activate(ctx, user, 1); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != one.URL {
		t.Errorf("URL() = %q, want %q", s.URL(user), one.URL)
	}
	if after := filesIn(t, dir); strings.Join(after, ",") != strings.Join(before, ",") {
		t.Errorf("switching changed the files: %v -> %v", before, after)
	}
	if err := s.Activate(ctx, user, 3); !errors.Is(err, ErrNoSlot) {
		t.Errorf("activating an empty slot = %v, want ErrNoSlot", err)
	}
	if s.URL(user) != one.URL {
		t.Error("a refused activation changed the active mat")
	}
	// A fresh service (no cache) reads the same thing from the database.
	fresh := NewService(s.files, s.db, nil)
	if fresh.URL(user) != one.URL {
		t.Errorf("a fresh service shows %q, want %q", fresh.URL(user), one.URL)
	}
	checkInvariant(t, s)
}

func TestRemovingTheActiveSlotClearsThePointer(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	one, _ := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White))
	two, _ := s.SetFromBytes(ctx, user, 2, mat(t, 30, 30, color.Black))
	if err := s.Remove(ctx, user, 1); err != nil { // slot 1 is the active one
		t.Fatal(err)
	}
	if s.URL(user) != "" {
		t.Errorf("URL() = %q after removing the active slot, want none", s.URL(user))
	}
	st, _ := s.State(ctx, user)
	if st.Active != 0 || len(st.Slots) != 1 || st.Slots[0].ID != two.ID {
		t.Errorf("State = %+v: the other saved mat must survive and none be active", st)
	}
	if fs := filesIn(t, dir); len(fs) != 1 || fs[0] != two.ID+".jpg" {
		t.Errorf("files = %v, want only slot 2's", fs)
	}
	if _, _, err := s.Open(one.ID); !errors.Is(err, ErrNotFound) {
		t.Errorf("the removed image still opens: %v", err)
	}
	checkInvariant(t, s)
}

func TestRemovingAnInactiveSlotLeavesTheActiveMatAlone(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	one, _ := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White))
	_, _ = s.SetFromBytes(ctx, user, 2, mat(t, 30, 30, color.Black))
	if err := s.Remove(ctx, user, 2); err != nil {
		t.Fatal(err)
	}
	if s.URL(user) != one.URL {
		t.Errorf("URL() = %q, want %q", s.URL(user), one.URL)
	}
	if fs := filesIn(t, dir); len(fs) != 1 {
		t.Errorf("files = %v", fs)
	}
	if err := s.Remove(ctx, user, 3); err != nil {
		t.Errorf("removing an empty slot: %v", err)
	}
	checkInvariant(t, s)
}

func TestRemoveAllDeletesEveryFileAndClearsThePointer(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	for slot := 1; slot <= 3; slot++ {
		if _, err := s.SetFromBytes(ctx, user, slot, mat(t, 20+slot, 20, color.White)); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.RemoveAll(ctx, user); err != nil {
		t.Fatal(err)
	}
	st, _ := s.State(ctx, user)
	if st.Active != 0 || len(st.Slots) != 0 || s.URL(user) != "" {
		t.Errorf("State = %+v, URL %q", st, s.URL(user))
	}
	if fs := filesIn(t, dir); len(fs) != 0 {
		t.Errorf("files = %v, want none", fs)
	}
	if err := s.RemoveAll(ctx, user); err != nil {
		t.Errorf("removing none: %v", err)
	}
	checkInvariant(t, s)
}

func TestStateSkipsASlotWhoseFileIsGoneButItsSlotStillWorks(t *testing.T) {
	s, dir, user := newService(t, nil)
	ctx := context.Background()
	sl, _ := s.SetFromBytes(ctx, user, 1, mat(t, 20, 20, color.White))
	if err := os.Remove(filepath.Join(dir, sl.ID+".jpg")); err != nil {
		t.Fatal(err)
	}
	st, err := s.State(ctx, user)
	if err != nil || len(st.Slots) != 0 || st.Active != 0 {
		t.Errorf("State = %+v, %v; a wiped file must read as no playmat, not a dead link", st, err)
	}
	if _, err := s.SetFromBytes(ctx, user, 1, mat(t, 24, 24, color.Black)); err != nil {
		t.Errorf("replacing the orphaned slot: %v", err)
	}
	if st, _ := s.State(ctx, user); len(st.Slots) != 1 {
		t.Errorf("State after replacing = %+v", st)
	}
	checkInvariant(t, s)
}

func TestAnUnknownUserHasNoPlaymatAndCannotSetOne(t *testing.T) {
	s, dir, _ := newService(t, nil)
	ghost := uuid.New()
	if _, err := s.SetFromBytes(context.Background(), ghost, 1, mat(t, 8, 8, color.White)); !errors.Is(err, ErrNoUser) {
		t.Errorf("err = %v, want ErrNoUser", err)
	}
	if fs := filesIn(t, dir); len(fs) != 0 {
		t.Errorf("a failed set left %v behind", fs)
	}
	if st, err := s.State(context.Background(), ghost); err != nil || len(st.Slots) != 0 {
		t.Errorf("State for nobody = %+v, %v", st, err)
	}
	if err := s.Activate(context.Background(), ghost, 0); !errors.Is(err, ErrNoUser) {
		t.Errorf("Activate for nobody = %v, want ErrNoUser", err)
	}
}

func TestOnePersonsMatsAreNotAnothersSlots(t *testing.T) {
	s, _, alice := newService(t, nil)
	ctx := context.Background()
	bob := uuid.New()
	if _, err := s.db.Exec(`INSERT INTO users (id, display_name, created_at, last_seen_at) VALUES (?, 'B', 1, 1)`, bob.String()); err != nil {
		t.Fatal(err)
	}
	a, _ := s.SetFromBytes(ctx, alice, 1, mat(t, 20, 20, color.White))
	b, _ := s.SetFromBytes(ctx, bob, 1, mat(t, 30, 30, color.Black))
	if s.URL(alice) != a.URL || s.URL(bob) != b.URL {
		t.Errorf("alice %q bob %q", s.URL(alice), s.URL(bob))
	}
	if err := s.Remove(ctx, alice, 1); err != nil {
		t.Fatal(err)
	}
	if s.URL(bob) != b.URL {
		t.Error("removing alice's slot 1 touched bob's")
	}
	checkInvariant(t, s)
}

func TestDisabledServices(t *testing.T) {
	var nilSvc *Service
	if nilSvc.Enabled() || nilSvc.URL(uuid.New()) != "" {
		t.Error("a nil service is enabled")
	}
	noDir := NewService(NewFileStore(""), nil, nil)
	if noDir.Enabled() {
		t.Error("a service with no directory and no database is enabled")
	}
	if _, err := noDir.SetFromBytes(context.Background(), uuid.New(), 1, nil); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
	if _, err := noDir.SetFromURL(context.Background(), uuid.New(), 1, "https://example.com/"); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
	if _, err := noDir.State(context.Background(), uuid.New()); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
	if err := noDir.Activate(context.Background(), uuid.New(), 1); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
	if err := noDir.Remove(context.Background(), uuid.New(), 1); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
	if _, err := noDir.Fit(context.Background(), uuid.New(), 1, 0, 0); !errors.Is(err, ErrDisabled) {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
}

func TestSetFromURLStoresTheBytesNotTheLink(t *testing.T) {
	img := pngBytes(t, solid(24, 16, color.RGBA{9, 9, 200, 255}))
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write(img) }))
	defer srv.Close()
	s, dir, user := newService(t, testFetcher(srv, fetcherHooks{allow: guard}))
	info, err := s.SetFromURL(context.Background(), user, 2, srv.URL+"/mat.png")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(info.URL, "127.0.0.1") || !strings.HasPrefix(info.URL, "/playmats/") {
		t.Errorf("the stored playmat points at %q; it must be our own route", info.URL)
	}
	if fs := filesIn(t, dir); len(fs) != 1 {
		t.Errorf("files = %v", fs)
	}
	// And a link that fails stores nothing and keeps what was there.
	if _, err := s.SetFromURL(context.Background(), user, 2, "https://10.0.0.1/x.png"); !errors.Is(err, ErrFetch) {
		t.Errorf("err = %v, want ErrFetch", err)
	}
	if s.URL(user) != info.URL {
		t.Error("a refused link changed the playmat")
	}
}

func TestOpenRefusesAnythingThatIsNotAMintedID(t *testing.T) {
	s, _, _ := newService(t, nil)
	for _, id := range []string{"", "..", "../x", "/etc/passwd", "a/b", "A0000000-0000-4000-8000-000000000000", "x.jpg", strings.Repeat("a", 300)} {
		if _, _, err := s.Open(id); !errors.Is(err, ErrNotFound) {
			t.Errorf("Open(%q) = %v, want ErrNotFound", id, err)
		}
	}
}
