package lobby

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// This file is the dedicated home for GET /games/{id}/replay's
// authorization and redaction behavior. The replay log is the single
// richest secret the server holds — a full unfiltered snapshot stream
// with every hand and every library order — so the rules governing who
// may read it and in what form get named, greppable tests rather than
// living as incidental assertions inside a wider integration test.
//
// Two properties are pinned here:
//
//  1. The game-state gate: non-admins are refused until the game ends.
//  2. The redaction: non-admins receive the stream through
//     protocol.FilterViewFor, so a replay never shows a viewer more
//     than the live WS connection already showed them.
//
// Both are security properties. If either regresses, these fail.

// replayFixture is the world a replay test runs against: a live HTTP
// stack backed by a real dump dir, a started two-seat game, and a
// replay file on disk holding the UNFILTERED snapshot stream that
// ws.Room actually writes.
type replayFixture struct {
	srv    *httptest.Server
	lobby  *Lobby
	auth   auth.Authenticator
	gameID uuid.UUID
	alice  uuid.UUID
	bob    uuid.UUID
	// view is the unfiltered view every line of the replay carries.
	view protocol.GameView
	// lines is how many JSONL records the replay file holds.
	lines int
}

// newReplayFixture builds the default fixture: two seats, small decks.
func newReplayFixture(t *testing.T) *replayFixture {
	t.Helper()
	return newReplayFixtureWithDeckSize(t, 24)
}

// newReplayFixtureWithDeckSize builds the fixture with a configurable
// deck size. Bigger decks make each snapshot line bigger, which is how
// the oversized-line test gets a record past 64 KiB.
func newReplayFixtureWithDeckSize(t *testing.T, deckSize int) *replayFixture {
	t.Helper()

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	dumpDir := t.TempDir()
	mgr := ws.NewRoomManager(log, dumpDir)
	l := NewLobby(mgr)
	a := auth.NewMemoryAuthenticator()
	srv := httptest.NewServer(Handler(Config{
		Lobby:      l,
		Auth:       a,
		AdminToken: "shared-admin-token",
		Log:        log,
	}))
	t.Cleanup(srv.Close)

	meta, err := l.Create("Replay Gate")
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	_, alice, err := l.Join(meta.ID, meta.InviteToken, "Alice")
	if err != nil {
		t.Fatalf("join Alice: %v", err)
	}
	_, bob, err := l.Join(meta.ID, meta.InviteToken, "Bob")
	if err != nil {
		t.Fatalf("join Bob: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, alice, "dummy", fillerDeck("A", deckSize)); err != nil {
		t.Fatalf("SetDeck Alice: %v", err)
	}
	if _, err := l.SetDeck(meta.ID, bob, "dummy", fillerDeck("B", deckSize)); err != nil {
		t.Fatalf("SetDeck Bob: %v", err)
	}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}

	g, err := l.LookupGame(meta.ID)
	if err != nil {
		t.Fatalf("LookupGame: %v", err)
	}
	view := protocol.ViewOfGame(g)

	// The lobby-phase mutations above already appended real replay
	// lines. Overwrite them with a deterministic two-line stream of
	// the same unfiltered view so assertions about line count and
	// per-line content don't drift when setup changes shape.
	var b strings.Builder
	const lines = 2
	for seq := 1; seq <= lines; seq++ {
		raw, err := json.Marshal(protocol.SnapshotPayload{Seq: uint64(seq), Game: view})
		if err != nil {
			t.Fatalf("marshal snapshot: %v", err)
		}
		b.Write(raw)
		b.WriteByte('\n')
	}
	writeReplay(t, dumpDir, meta.ID, b.String())

	return &replayFixture{
		srv: srv, lobby: l, auth: a,
		gameID: meta.ID, alice: alice, bob: bob,
		view: view, lines: lines,
	}
}

func fillerDeck(prefix string, n int) []game.Card {
	d := []game.Card{game.NewCommander(prefix+"-cmdr", uuid.Nil)}
	for i := 0; i < n; i++ {
		d = append(d, game.NewCard(fmt.Sprintf("%s-filler-%d", prefix, i), uuid.Nil))
	}
	return d
}

func (f *replayFixture) path() string {
	return "/games/" + f.gameID.String() + "/replay"
}

// end drives the fixture's game to StateEnded.
func (f *replayFixture) end(t *testing.T) {
	t.Helper()
	g, err := f.lobby.LookupGame(f.gameID)
	if err != nil {
		t.Fatalf("LookupGame: %v", err)
	}
	g.End()
	if st := g.CurrentState(); st != game.StateEnded {
		t.Fatalf("game state after End(): got %q, want %q", st, game.StateEnded)
	}
}

// session mints a token for an arbitrary principal. The package's
// playerSession helper invents a random PlayerID; replay filtering
// tests need the token's PlayerID to be a real seat in the fixture,
// so they mint principals directly.
func (f *replayFixture) session(t *testing.T, p auth.Principal) string {
	t.Helper()
	tok, _, err := f.auth.Issue(context.Background(), p, time.Hour)
	if err != nil {
		t.Fatalf("issue session: %v", err)
	}
	return tok
}

// TestReplayGateRefusesNonAdminsUntilTheGameEnds pins the game-state
// check in downloadReplay. The replay stream is live hidden
// information while the game runs: admins may pull it at any time,
// everyone else only once the game has ended.
//
// Every row is asserted in both game states, so inverting the
// condition cannot pass — flipping `!=` to `==` turns the in-progress
// player row from 403 into 200 and the ended player row from 200 into
// 403, and both halves of the table fail.
func TestReplayGateRefusesNonAdminsUntilTheGameEnds(t *testing.T) {
	f := newReplayFixture(t)

	admin := f.session(t, auth.Principal{Role: auth.RoleAdmin, Name: "root"})
	alice := f.session(t, auth.Principal{
		Role: auth.RolePlayer, GameID: f.gameID, PlayerID: f.alice, Name: "Alice",
	})
	spectator := f.session(t, auth.Principal{
		Role: auth.RoleSpectator, GameID: f.gameID, Name: "Watcher",
	})
	// Seated in some *other* game — the seat check is a separate
	// condition and must hold in both game states.
	outsider := f.session(t, auth.Principal{
		Role: auth.RolePlayer, GameID: uuid.New(), PlayerID: uuid.New(), Name: "Intruder",
	})

	type row struct {
		who   string
		token string
		want  int
	}

	// Guard against a vacuous pass: if setup ever stopped starting the
	// game, "in progress" would be false and the first table would be
	// asserting the wrong branch.
	g, err := f.lobby.LookupGame(f.gameID)
	if err != nil {
		t.Fatalf("LookupGame: %v", err)
	}
	if st := g.CurrentState(); st == game.StateEnded {
		t.Fatalf("fixture game is already %q — the in-progress table would prove nothing", st)
	}

	inProgress := []row{
		{"admin", admin, http.StatusOK},
		{"seated player", alice, http.StatusForbidden},
		{"spectator", spectator, http.StatusForbidden},
		{"other game's player", outsider, http.StatusForbidden},
		{"unauthenticated", "", http.StatusUnauthorized},
	}
	for _, tc := range inProgress {
		got := doAuthGet(t, f.srv, f.path(), tc.token)
		if got.code != tc.want {
			t.Errorf("in-progress %s: status = %d, want %d (body %s)",
				tc.who, got.code, tc.want, snip(got.body))
		}
	}

	f.end(t)

	ended := []row{
		{"admin", admin, http.StatusOK},
		{"seated player", alice, http.StatusOK},
		{"spectator", spectator, http.StatusOK},
		// Ending the game does not widen the audience beyond the table.
		{"other game's player", outsider, http.StatusForbidden},
		{"unauthenticated", "", http.StatusUnauthorized},
	}
	for _, tc := range ended {
		got := doAuthGet(t, f.srv, f.path(), tc.token)
		if got.code != tc.want {
			t.Errorf("ended %s: status = %d, want %d (body %s)",
				tc.who, got.code, tc.want, snip(got.body))
		}
	}
}

// snip keeps a failed-gate message readable. A wrongly-granted replay
// body is megabytes of snapshot JSON; the first line of it is plenty
// to tell an error response from a leak.
func snip(body string) string {
	const max = 200
	if len(body) > max {
		return fmt.Sprintf("%q… (%d bytes total)", body[:max], len(body))
	}
	return fmt.Sprintf("%q", body)
}

// decodeReplay parses a JSONL replay body back into snapshot payloads.
func decodeReplay(t *testing.T, body string) []protocol.SnapshotPayload {
	t.Helper()
	var out []protocol.SnapshotPayload
	dec := json.NewDecoder(strings.NewReader(body))
	for {
		var p protocol.SnapshotPayload
		err := dec.Decode(&p)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("decode replay line %d: %v", len(out)+1, err)
		}
		out = append(out, p)
	}
	return out
}

// seatIndex finds the view index of a seat by player ID.
func seatIndex(t *testing.T, v protocol.GameView, id uuid.UUID) int {
	t.Helper()
	for i, s := range v.Seats {
		if s.ID == id.String() {
			return i
		}
	}
	t.Fatalf("seat %s not present in view", id)
	return -1
}

// TestReplayIsFilteredPerViewer pins the redaction half of the
// contract. The replay on disk is unfiltered — that is what admins
// get, and what bug reports pin. Everyone else reads it through
// protocol.FilterViewFor, so the replay can never show a viewer
// something the live WS connection didn't already show them.
//
// Without this, a player downloading a finished game's replay reads
// every opponent's hand and library order for the whole game.
func TestReplayIsFilteredPerViewer(t *testing.T) {
	f := newReplayFixture(t)
	f.end(t)

	aliceSeat := seatIndex(t, f.view, f.alice)
	bobSeat := seatIndex(t, f.view, f.bob)

	// The fixture must actually contain the secret we claim to be
	// protecting, or every assertion below passes for the wrong reason.
	if len(f.view.Seats[aliceSeat].Hand.Cards) == 0 || len(f.view.Seats[bobSeat].Hand.Cards) == 0 {
		t.Fatalf("fixture view has empty hands (alice %d, bob %d) — nothing to redact",
			len(f.view.Seats[aliceSeat].Hand.Cards), len(f.view.Seats[bobSeat].Hand.Cards))
	}

	t.Run("admin gets the unfiltered stream", func(t *testing.T) {
		tok := f.session(t, auth.Principal{Role: auth.RoleAdmin, Name: "root"})
		got := doAuthGet(t, f.srv, f.path(), tok)
		if got.code != http.StatusOK {
			t.Fatalf("status = %d, want 200", got.code)
		}
		payloads := decodeReplay(t, got.body)
		if len(payloads) != f.lines {
			t.Fatalf("line count = %d, want %d", len(payloads), f.lines)
		}
		for i, p := range payloads {
			if len(p.Game.Seats[aliceSeat].Hand.Cards) == 0 {
				t.Errorf("line %d: admin should see Alice's hand, got 0 cards", i)
			}
			if len(p.Game.Seats[bobSeat].Hand.Cards) == 0 {
				t.Errorf("line %d: admin should see Bob's hand, got 0 cards", i)
			}
		}
	})

	t.Run("seated player sees only their own hand", func(t *testing.T) {
		tok := f.session(t, auth.Principal{
			Role: auth.RolePlayer, GameID: f.gameID, PlayerID: f.alice, Name: "Alice",
		})
		got := doAuthGet(t, f.srv, f.path(), tok)
		if got.code != http.StatusOK {
			t.Fatalf("status = %d, want 200", got.code)
		}
		payloads := decodeReplay(t, got.body)
		if len(payloads) != f.lines {
			t.Fatalf("line count = %d, want %d (filtering must not drop records)",
				len(payloads), f.lines)
		}
		for i, p := range payloads {
			// Sequence numbers must survive the transform — the
			// scrubber orders frames by them.
			if p.Seq != uint64(i+1) {
				t.Errorf("line %d: seq = %d, want %d", i, p.Seq, i+1)
			}
			if len(p.Game.Seats[aliceSeat].Hand.Cards) == 0 {
				t.Errorf("line %d: Alice must still see her own hand", i)
			}
			if n := len(p.Game.Seats[bobSeat].Hand.Cards); n != 0 {
				t.Errorf("line %d: Bob's hand leaked to Alice — %d cards", i, n)
			}
			// The count survives redaction; only the identities go.
			if p.Game.Seats[bobSeat].Hand.Count == 0 {
				t.Errorf("line %d: Bob's hand COUNT should survive filtering", i)
			}
			if n := len(p.Game.Seats[bobSeat].Library.Cards); n != 0 {
				t.Errorf("line %d: Bob's library order leaked to Alice — %d cards", i, n)
			}
		}
		// Belt and braces: the raw bytes must not carry an opponent's
		// card names at all, however the view is shaped.
		for _, c := range f.view.Seats[bobSeat].Hand.Cards {
			if c.Name != "" && strings.Contains(got.body, c.Name) {
				t.Errorf("opponent hand card %q appears in Alice's replay bytes", c.Name)
			}
		}
	})

	t.Run("spectator sees no hands at all", func(t *testing.T) {
		tok := f.session(t, auth.Principal{
			Role: auth.RoleSpectator, GameID: f.gameID, Name: "Watcher",
		})
		got := doAuthGet(t, f.srv, f.path(), tok)
		if got.code != http.StatusOK {
			t.Fatalf("status = %d, want 200", got.code)
		}
		payloads := decodeReplay(t, got.body)
		if len(payloads) != f.lines {
			t.Fatalf("line count = %d, want %d", len(payloads), f.lines)
		}
		for i, p := range payloads {
			for s, seat := range p.Game.Seats {
				if n := len(seat.Hand.Cards); n != 0 {
					t.Errorf("line %d seat %d: hand visible to spectator — %d cards", i, s, n)
				}
			}
		}
	})
}

// TestReplayFilterHandlesOversizedSnapshotLines guards the streaming
// implementation against a line-length cap. A single snapshot record
// is a whole GameView and comfortably exceeds bufio.Scanner's 64 KiB
// default token size in a real Commander game, so a scanner-based
// filter would silently truncate the replay mid-stream — data loss
// that looks like a short game rather than an error.
func TestReplayFilterHandlesOversizedSnapshotLines(t *testing.T) {
	const bufioDefaultMax = 64 * 1024

	f := newReplayFixtureWithDeckSize(t, 600)
	f.end(t)

	// Prove the fixture actually produces an oversized record,
	// otherwise this test cannot detect the bug it exists for.
	raw, err := json.Marshal(protocol.SnapshotPayload{Seq: 1, Game: f.view})
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	if len(raw) <= bufioDefaultMax {
		t.Fatalf("fixture snapshot line is only %d bytes; need > %d to exercise the cap",
			len(raw), bufioDefaultMax)
	}

	tok := f.session(t, auth.Principal{
		Role: auth.RolePlayer, GameID: f.gameID, PlayerID: f.alice, Name: "Alice",
	})
	got := doAuthGet(t, f.srv, f.path(), tok)
	if got.code != http.StatusOK {
		t.Fatalf("status = %d, want 200", got.code)
	}
	payloads := decodeReplay(t, got.body)
	if len(payloads) != f.lines {
		t.Fatalf("line count = %d, want %d — oversized records were dropped or truncated",
			len(payloads), f.lines)
	}
	aliceSeat := seatIndex(t, f.view, f.alice)
	for i, p := range payloads {
		if len(p.Game.Seats[aliceSeat].Hand.Cards) == 0 {
			t.Errorf("line %d: own hand missing from an oversized record", i)
		}
	}
}
