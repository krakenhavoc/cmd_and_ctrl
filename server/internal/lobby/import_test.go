package lobby

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// import_test.go covers the one-time move of ADR 0041's
// lobby/<id>.json files into the database (ADR 0051 "Migration").

// oldBinaryGame plays the part of the binary before this change: it
// runs a game with a data dir (so the engine writes its restore
// point) and leaves lobby/<id>.json behind, exactly as the old
// persistMetaLocked wrote it.
func oldBinaryGame(t *testing.T, dir, name string, seat func(l *Lobby, meta GameMeta)) GameMeta {
	t.Helper()
	mgr := ws.NewRoomManager(quietLogger(), dir)
	l := NewLobby(mgr)
	meta, err := l.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if seat != nil {
		seat(l, meta)
	}
	full, err := l.Get(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	writeLegacyFile(t, dir, full.ID, full)
	return full
}

func writeLegacyFile(t *testing.T, dir string, id uuid.UUID, v any) string {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	path := legacyMetaPath(dir, id)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, raw, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// recordingHandler captures log lines so a test can assert one was
// written.
type recordingHandler struct {
	mu    sync.Mutex
	lines []string
}

func (h *recordingHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *recordingHandler) Handle(_ context.Context, r slog.Record) error {
	var b strings.Builder
	b.WriteString(r.Message)
	r.Attrs(func(a slog.Attr) bool {
		b.WriteString(" " + a.Key + "=" + a.Value.String())
		return true
	})
	h.mu.Lock()
	h.lines = append(h.lines, b.String())
	h.mu.Unlock()
	return nil
}
func (h *recordingHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *recordingHandler) WithGroup(string) slog.Handler      { return h }

func (h *recordingHandler) contains(sub string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, l := range h.lines {
		if strings.Contains(l, sub) {
			return true
		}
	}
	return false
}

// The first boot of the new binary over the old binary's data dir:
// the game comes back, with its seats, and the old links work.
func TestImporterHappyPathAndRestore(t *testing.T) {
	dir := t.TempDir()
	meta := oldBinaryGame(t, dir, "Before the database", func(l *Lobby, meta GameMeta) {
		if _, _, err := l.JoinWithIdentity(meta.ID, meta.InviteToken, "", DiscordIdentity{
			ID: "123456789012345678", Username: "alice", GlobalName: "Alice", AvatarHash: "abc",
		}); err != nil {
			t.Fatal(err)
		}
		if _, _, err := l.Join(meta.ID, meta.InviteToken, "Bob"); err != nil {
			t.Fatal(err)
		}
	})

	l, _ := newDurableLobby(t, dir)
	if n := l.RestoreFromDisk(quietLogger()); n != 1 {
		t.Fatalf("restored %d games, want 1", n)
	}

	// The file was renamed, not deleted.
	legacy := legacyMetaPath(dir, meta.ID)
	if _, err := os.Stat(legacy); !os.IsNotExist(err) {
		t.Errorf("lobby/<id>.json still in place after import (err=%v)", err)
	}
	if _, err := os.Stat(legacy + importedSuffix); err != nil {
		t.Errorf("lobby/<id>.json.imported missing: %v", err)
	}

	// Rows.
	d := openTestDB(t, dir)
	var name, state string
	if err := d.QueryRow(`SELECT name, state FROM games WHERE id = ?`, meta.ID.String()).Scan(&name, &state); err != nil {
		t.Fatal(err)
	}
	if name != "Before the database" || state != "lobby" {
		t.Errorf("games row = %q/%q", name, state)
	}
	var userID, pending *string
	if err := d.QueryRow(`SELECT user_id, pending_discord_id FROM seats WHERE game_id = ? AND seat = 0`,
		meta.ID.String()).Scan(&userID, &pending); err != nil {
		t.Fatal(err)
	}
	if userID != nil || pending == nil || *pending != "123456789012345678" {
		t.Errorf("discord seat: user_id=%v pending_discord_id=%v; want NULL and the snowflake", userID, pending)
	}
	var invites int
	_ = d.QueryRow(`SELECT COUNT(*) FROM invites WHERE game_id = ?`, meta.ID.String()).Scan(&invites)
	if invites != 2 {
		t.Errorf("invites rows = %d, want 2", invites)
	}
	raw, _ := os.ReadFile(filepath.Join(dir, "db", "cmdctrl.sqlite-wal"))
	main, _ := os.ReadFile(filepath.Join(dir, "db", "cmdctrl.sqlite"))
	if bytes.Contains(raw, []byte(meta.InviteToken)) || bytes.Contains(main, []byte(meta.InviteToken)) {
		t.Error("the importer stored a plaintext invite token")
	}

	// The restored lobby entry, built from rows plus the engine.
	back, err := l.Get(meta.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(back.Players) != 2 || back.Players[0].Name != "Alice" || back.Players[1].Name != "Bob" {
		t.Fatalf("seats after import: %+v", back.Players)
	}
	if back.Players[0].DiscordID != "123456789012345678" || back.Players[0].DiscordAvatarHash != "abc" {
		t.Errorf("discord identity lost on import: %+v", back.Players[0])
	}
	if _, _, err := l.Join(meta.ID, meta.InviteToken, "Carol"); err != nil {
		t.Errorf("old player link rejected after import: %v", err)
	}
	if _, err := l.Spectate(meta.ID, meta.SpectatorInvite); err != nil {
		t.Errorf("old spectator link rejected after import: %v", err)
	}
}

// A second boot finds nothing new to do, and a boot whose renames
// were interrupted finishes them without writing the game twice.
func TestImporterIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	meta := oldBinaryGame(t, dir, "Twice", nil)

	store := NewSQLStore(openTestDB(t, dir))
	ctx := context.Background()
	res, err := store.ImportLegacyLobby(ctx, dir, quietLogger())
	if err != nil || res.Imported != 1 {
		t.Fatalf("first import = %+v, %v; want 1 imported", res, err)
	}

	res, err = store.ImportLegacyLobby(ctx, dir, quietLogger())
	if err != nil || res != (ImportResult{}) {
		t.Fatalf("second import = %+v, %v; want nothing done", res, err)
	}

	// Simulate a crash between the commit and the rename: the file is
	// back under its old name, the row is already there.
	legacy := legacyMetaPath(dir, meta.ID)
	if err := os.Rename(legacy+importedSuffix, legacy); err != nil {
		t.Fatal(err)
	}
	res, err = store.ImportLegacyLobby(ctx, dir, quietLogger())
	if err != nil || res.Imported != 0 || res.AlreadyPresent != 1 {
		t.Fatalf("resumed import = %+v, %v; want 1 already present", res, err)
	}
	if _, err := os.Stat(legacy + importedSuffix); err != nil {
		t.Errorf("resumed import did not finish the rename: %v", err)
	}
	var games int
	_ = store.db.QueryRow(`SELECT COUNT(*) FROM games`).Scan(&games)
	if games != 1 {
		t.Errorf("games rows = %d, want 1", games)
	}
}

// A corrupt file is skipped with a log line and left exactly where it
// was; the good file beside it still goes in.
func TestImporterSkipsACorruptFile(t *testing.T) {
	dir := t.TempDir()
	good := oldBinaryGame(t, dir, "Fine", nil)

	badID := uuid.New()
	badPath := legacyMetaPath(dir, badID)
	if err := os.WriteFile(badPath, []byte(`{"name": "half a fi`), 0o600); err != nil {
		t.Fatal(err)
	}
	// Valid JSON but no usable invite: nothing to import it as.
	noInviteID := uuid.New()
	noInvitePath := writeLegacyFile(t, dir, noInviteID, map[string]any{"name": "tokenless"})

	rec := &recordingHandler{}
	store := NewSQLStore(openTestDB(t, dir))
	res, err := store.ImportLegacyLobby(context.Background(), dir, slog.New(rec))
	if err != nil {
		t.Fatal(err)
	}
	if res.Imported != 1 || res.Skipped != 2 {
		t.Errorf("import = %+v; want 1 imported, 2 skipped", res)
	}
	for _, p := range []string{badPath, noInvitePath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("skipped file %s was moved: %v", filepath.Base(p), err)
		}
		if _, err := os.Stat(p + importedSuffix); !os.IsNotExist(err) {
			t.Errorf("skipped file %s was renamed", filepath.Base(p))
		}
		if !rec.contains(p) {
			t.Errorf("no log line names %s", filepath.Base(p))
		}
	}
	if _, err := os.Stat(legacyMetaPath(dir, good.ID) + importedSuffix); err != nil {
		t.Errorf("the good file was not imported alongside: %v", err)
	}
}

// Invites are looked up by hash, and a revoked or expired invite is
// refused on every path that takes one.
func TestRevokedAndExpiredInvitesAreRefused(t *testing.T) {
	dir := t.TempDir()
	l, _ := newDurableLobby(t, dir)
	d := openTestDB(t, dir)
	ctx := context.Background()

	revoked, err := l.Create("Revoked")
	if err != nil {
		t.Fatal(err)
	}
	// Live first: the lookup by hash finds it.
	if id, err := l.FindByInvite(revoked.InviteToken); err != nil || id != revoked.ID {
		t.Fatalf("FindByInvite before revoke = %v, %v", id, err)
	}
	h, _ := hashInvite(revoked.InviteToken)
	if err := l.store.RevokeInvite(ctx, h, time.Now()); err != nil {
		t.Fatal(err)
	}
	sh, _ := hashInvite(revoked.SpectatorInvite)
	if err := l.store.RevokeInvite(ctx, sh, time.Now()); err != nil {
		t.Fatal(err)
	}

	expired, err := l.Create("Expired")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE invites SET expires_at = ? WHERE game_id = ?`,
		time.Now().Add(-time.Minute).UnixMilli(), expired.ID.String()); err != nil {
		t.Fatal(err)
	}

	for _, m := range []GameMeta{revoked, expired} {
		if _, err := l.FindByInvite(m.InviteToken); err != ErrInvalidInvite {
			t.Errorf("%s: FindByInvite = %v, want ErrInvalidInvite", m.Name, err)
		}
		if _, _, err := l.Join(m.ID, m.InviteToken, "Mallory"); err != ErrInvalidInvite {
			t.Errorf("%s: Join = %v, want ErrInvalidInvite", m.Name, err)
		}
		if _, err := l.Spectate(m.ID, m.SpectatorInvite); err != ErrInvalidInvite {
			t.Errorf("%s: Spectate = %v, want ErrInvalidInvite", m.Name, err)
		}
		if _, _, err := l.Preview(m.ID, m.InviteToken); err != ErrInvalidInvite {
			t.Errorf("%s: Preview = %v, want ErrInvalidInvite", m.Name, err)
		}
	}

	// A future expiry is still good.
	later, err := l.Create("Later")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`UPDATE invites SET expires_at = ? WHERE game_id = ?`,
		time.Now().Add(time.Hour).UnixMilli(), later.ID.String()); err != nil {
		t.Fatal(err)
	}
	if _, _, err := l.Join(later.ID, later.InviteToken, "Alice"); err != nil {
		t.Errorf("an invite expiring in an hour was refused: %v", err)
	}
}

// A malformed token never reaches the store, and a token for one game
// does not open another.
func TestInviteLookupIsScopedAndStrict(t *testing.T) {
	l := newTestLobby(t)
	a, _ := l.Create("A")
	b, _ := l.Create("B")
	for _, bad := range []string{"", "not base64 !!", a.InviteToken + "A", strings.ToUpper(a.InviteToken)} {
		if bad == a.InviteToken {
			continue
		}
		if _, err := l.FindByInvite(bad); err != ErrInvalidInvite {
			t.Errorf("FindByInvite(%q) = %v, want ErrInvalidInvite", bad, err)
		}
	}
	if _, _, err := l.Join(b.ID, a.InviteToken, "Mallory"); err != ErrInvalidInvite {
		t.Errorf("A's invite joined B: %v", err)
	}
	if _, _, err := l.Join(a.ID, a.SpectatorInvite, "Mallory"); err != ErrInvalidInvite {
		t.Errorf("a spectator invite seated a player: %v", err)
	}
}
