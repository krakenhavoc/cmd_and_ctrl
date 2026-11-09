package lobby

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planechase_test.go — ADR 0216 §1: the Planechase option on POST
// /games and PATCH /games/{id}/settings, locked once the game starts,
// and a Planechase table the server cannot build a planar deck for
// refuses to start.

type lobbyPlanes struct{ n int }

func (p lobbyPlanes) PlanarCards(game.PlanarKinds) []game.Card {
	out := make([]game.Card, p.n)
	for i := range out {
		c := game.NewCard(fmt.Sprintf("Lobby Plane %02d", i), uuid.Nil)
		c.OracleID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(c.Name)).String()
		c.TypeLine = "Plane — Lobbyworld"
		c.Layout = "planar"
		out[i] = c
	}
	return out
}

func TestCreatingATableWithPlanechaseTurnsTheOptionOn(t *testing.T) {
	s := newTableStack(t)
	s.lobby.mgr.SetPlanechaseOffered(true)
	alice, _ := s.signIn(t, "discord-4870", "Alice")
	code, out, _, raw := s.create(t, alice, createGameRequest{Name: "Planar table", Planechase: true})
	if code != http.StatusCreated {
		t.Fatalf("create: %d %s", code, raw)
	}
	if !settingsOf(t, s.lobby, out.ID).Planechase {
		t.Error("POST /games with planechase left the option off")
	}
	code, plain, _, raw := s.create(t, alice, createGameRequest{Name: "Plain table"})
	if code != http.StatusCreated {
		t.Fatalf("create: %d %s", code, raw)
	}
	if settingsOf(t, s.lobby, plain.ID).Planechase {
		t.Error("a table created without planechase has it on")
	}
}

func TestPlanechaseIsSetInTheLobbyAndLockedAtStart(t *testing.T) {
	srv, l, a := newTestHTTPStack(t)
	l.mgr.SetPlanechaseOffered(true)
	meta, _, aliceTok, _ := settingsTable(t, l, a)
	path := "/games/" + meta.ID.String() + "/settings"

	resp := patchJSON(t, srv, path, aliceTok, map[string]any{"planechase": true})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("planechase in the lobby: got %d, want 200", resp.StatusCode)
	}
	if !settingsOf(t, l, meta.ID).Planechase {
		t.Fatal("the patch did not turn Planechase on")
	}
	for _, p := range mustGet(t, l, meta.ID).Players {
		uploadDummyDeck(t, l, meta.ID, p.PlayerID)
	}

	// No card database: the table does not start, and the HTTP answer
	// is the 409 a refused start gets.
	prev := game.PlanarCards
	game.PlanarCards = nil
	t.Cleanup(func() { game.PlanarCards = prev })
	_, err := l.Start(meta.ID)
	if !errors.Is(err, game.ErrNoPlanarCards) {
		t.Fatalf("Start with no planar source: %v, want ErrNoPlanarCards", err)
	}
	rec := httptest.NewRecorder()
	writeLobbyError(rec, err)
	if rec.Code != http.StatusConflict {
		t.Errorf("a refused Planechase start is %d, want 409", rec.Code)
	}
	game.PlanarCards = lobbyPlanes{n: 5}
	if _, err := l.Start(meta.ID); !errors.Is(err, game.ErrPlanarDeckTooSmall) {
		t.Fatalf("Start with 5 planes: %v, want ErrPlanarDeckTooSmall", err)
	}

	game.PlanarCards = lobbyPlanes{n: 20}
	if _, err := l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	resp = patchJSON(t, srv, path, aliceTok, map[string]any{"planechase": false})
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("planechase after Start: got %d, want 422", resp.StatusCode)
	}
	if !settingsOf(t, l, meta.ID).Planechase {
		t.Error("the refused patch turned Planechase off")
	}
}

// ADR 0216 owner answer 3: a server that does not offer Planechase (prod,
// until the plane batches land) refuses every way of turning it on, and
// still lets a table turn it off.
func TestPlanechaseIsRefusedWhereTheServerDoesNotOfferIt(t *testing.T) {
	s := newTableStack(t)
	alice, _ := s.signIn(t, "discord-4870-off", "Alice")
	code, _, _, raw := s.create(t, alice, createGameRequest{Name: "Planar table", Planechase: true})
	if code != http.StatusBadRequest {
		t.Fatalf("create with planechase on a server that does not offer it: %d %s, want 400", code, raw)
	}

	srv, l, a := newTestHTTPStack(t)
	meta, _, aliceTok, _ := settingsTable(t, l, a)
	path := "/games/" + meta.ID.String() + "/settings"
	resp := patchJSON(t, srv, path, aliceTok, map[string]any{"planechase": true})
	resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("PATCH planechase on: got %d, want 400", resp.StatusCode)
	}
	if settingsOf(t, l, meta.ID).Planechase {
		t.Fatal("the refused patch turned Planechase on")
	}
	on := true
	if _, err := l.UpdateSettings(meta.ID, uuid.Nil, game.SettingsPatch{Planechase: &on}); !errors.Is(err, game.ErrPlanechaseNotOffered) {
		t.Fatalf("UpdateSettings: %v, want ErrPlanechaseNotOffered", err)
	}
	resp = patchJSON(t, srv, path, aliceTok, map[string]any{"planechase": false})
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("PATCH planechase off: got %d, want 200", resp.StatusCode)
	}
}
