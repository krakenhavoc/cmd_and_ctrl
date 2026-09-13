package lobby

// bugreport_gamelog_test.go — the S31 sub-PR 0 artifact: the public
// game log pinned alongside the replay.

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seedGameLog puts a couple of table-visible events on the room's game
// so the projection has something to render.
func seedGameLog(t *testing.T, l *Lobby, gameID uuid.UUID) {
	t.Helper()
	room := l.RoomOf(gameID)
	if room == nil || room.Game == nil {
		t.Fatalf("no room for game %s", gameID)
	}
	g := room.Game
	p, err := g.AddPlayer("Aang", []game.Card{game.NewCard("Filler", uuid.Nil)})
	if err != nil {
		t.Fatalf("AddPlayer: %v", err)
	}
	boltID := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: boltID,
			Name:       "Lightning Bolt",
			Owner:      p.ID,
			Controller: p.ID,
		})
		g.EmitEvent(game.Event{
			Kind: game.EventStepBegan, Actor: p.ID, Amount: 2, Label: string(game.StepPrecombatMain),
		})
		g.EmitEvent(game.Event{
			Kind: game.EventCast, Actor: p.ID, Source: boltID, CardID: boltID,
			OldZone: game.ZoneHand, NewZone: game.ZoneStack,
		})
	})
}

func TestBugReportPinsGameLogAndServesItToAdminsOnly(t *testing.T) {
	rep := &recordingReporter{}
	srv, l, a, _, _ := attachStack(t, rep)
	meta, err := l.Create("FNM")
	if err != nil {
		t.Fatalf("create game: %v", err)
	}
	gameID := meta.ID
	seedGameLog(t, l, gameID)
	tok := playerSession(t, a, gameID)
	admin := adminSession(t, a)

	resp := postMultipartReport(t, srv, tok, map[string]any{
		"title":   "the bolt resolved twice",
		"context": map[string]any{"game_id": gameID.String(), "turn": 2},
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body = %s", resp.StatusCode, b)
	}
	var out struct {
		ReportID string `json:"report_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if out.ReportID == "" {
		t.Fatal("no report_id — nothing was pinned")
	}

	body := rep.bodies[0]
	if !strings.Contains(body, "Pinned game log: `GET /bugreport/"+out.ReportID+"/gamelog` (admin)") {
		t.Errorf("issue body missing the pinned game-log row\n----\n%s", body)
	}
	// ADR 0017 §4: game state is referenced, never inlined.
	if strings.Contains(body, "Lightning Bolt") {
		t.Errorf("game log content leaked into the issue body\n----\n%s", body)
	}

	got := doAuthGet(t, srv, "/bugreport/"+out.ReportID+"/gamelog", admin)
	if got.code != http.StatusOK {
		t.Fatalf("admin game-log status = %d, want 200", got.code)
	}
	if !strings.Contains(got.body, "Aang cast Lightning Bolt") {
		t.Errorf("pinned game log missing the cast line:\n%s", got.body)
	}
	if !strings.Contains(got.body, "Turn 2 — Aang · precombat main") {
		t.Errorf("pinned game log missing the step line:\n%s", got.body)
	}

	if p := doAuthGet(t, srv, "/bugreport/"+out.ReportID+"/gamelog", tok); p.code != http.StatusForbidden {
		t.Errorf("player game-log status = %d, want 403", p.code)
	}
	if u := doAuthGet(t, srv, "/bugreport/"+out.ReportID+"/gamelog", ""); u.code != http.StatusUnauthorized {
		t.Errorf("unauthenticated game-log status = %d, want 401", u.code)
	}
}

// A session bound to game A must not be able to make the server render
// game B's log by naming it in the context — the same rule the replay
// pin follows.
func TestBugReportDoesNotPinAnotherGamesLog(t *testing.T) {
	rep := &recordingReporter{}
	srv, l, a, _, _ := attachStack(t, rep)
	mine, err := l.Create("mine")
	if err != nil {
		t.Fatalf("create mine: %v", err)
	}
	theirs, err := l.Create("theirs")
	if err != nil {
		t.Fatalf("create theirs: %v", err)
	}
	seedGameLog(t, l, theirs.ID)
	tok := playerSession(t, a, mine.ID)

	resp := postMultipartReport(t, srv, tok, map[string]any{
		"title":   "peeking",
		"context": map[string]any{"game_id": theirs.ID.String()},
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body = %s", resp.StatusCode, b)
	}
	if strings.Contains(rep.bodies[0], "Pinned game log") {
		t.Errorf("pinned another game's log\n----\n%s", rep.bodies[0])
	}
}

func TestBugReportWithoutGameLogSaysNothing(t *testing.T) {
	rep := &recordingReporter{}
	srv, _, a, _, _ := attachStack(t, rep)
	resp := postMultipartReport(t, srv, adminSession(t, a), map[string]any{
		"title": "no game at all",
	})
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("status = %d body = %s", resp.StatusCode, b)
	}
	if strings.Contains(rep.bodies[0], "Pinned game log") {
		t.Errorf("reported a game log for a report with no game\n----\n%s", rep.bodies[0])
	}
}
