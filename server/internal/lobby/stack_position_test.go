package lobby

// stack_position_test.go: #3429, the movable stack panel (ADR 0176). A
// seat saves where it put the panel, reads it back on its own view only,
// resets it, and nobody else can place it for them.

import (
	"encoding/json"
	"math"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// activeTable seats two humans and starts the game.
func activeTable(t *testing.T) (*rematchTable, string) {
	t.Helper()
	s := newSeatRemoveStack(t)
	meta, err := s.l.CreateBy("Friday night", uuid.Nil)
	if err != nil {
		t.Fatal(err)
	}
	rt := &rematchTable{s: s, meta: meta, seats: map[string]sessionResponse{}}
	for _, n := range []string{"Alice", "Bob"} {
		rt.seats[n] = joinAs(t, s.srv, meta, n)
		if _, err := s.l.SetDeck(meta.ID, rt.seats[n].PlayerID, n+"'s deck", botDeck(30)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.l.Start(meta.ID); err != nil {
		t.Fatalf("Start: %v", err)
	}
	return rt, meta.ID.String()
}

func (rt *rematchTable) stackPos(t *testing.T, method, token string, body any) int {
	t.Helper()
	resp := doJSON(t, rt.s.srv, method, "/games/"+rt.meta.ID.String()+"/stack-position", token, body)
	defer resp.Body.Close()
	return resp.StatusCode
}

// viewPos is what a seat's own view says about its stack position.
func (rt *rematchTable) viewPos(name string) map[string]protocol.StackPositionView {
	return rt.s.l.RoomOf(rt.meta.ID).ViewFor(rt.seats[name].PlayerID.String()).StackPositions
}

func TestStackPositionIsSavedReadBackAndResetOnTheSeatsOwnView(t *testing.T) {
	rt, _ := activeTable(t)
	alice, bob := rt.seats["Alice"], rt.seats["Bob"]

	if got := rt.viewPos("Alice"); got != nil {
		t.Fatalf("a fresh table has a saved position: %+v", got)
	}
	if code := rt.stackPos(t, http.MethodPut, alice.Token, map[string]float64{"x": 0.6, "y": 0.25}); code != http.StatusNoContent {
		t.Fatalf("save: status %d, want 204", code)
	}
	got := rt.viewPos("Alice")
	want := protocol.StackPositionView{X: 0.6, Y: 0.25}
	if len(got) != 1 || got[alice.PlayerID.String()] != want {
		t.Errorf("alice's view = %+v, want only her own %+v", got, want)
	}
	// Another seat, and a spectator, never see it.
	if got := rt.viewPos("Bob"); got != nil {
		t.Errorf("bob sees alice's placement: %+v", got)
	}
	room := rt.s.l.RoomOf(rt.meta.ID)
	if got := room.ViewFor(protocol.SpectatorViewerID).StackPositions; got != nil {
		t.Errorf("a spectator sees a placement: %+v", got)
	}
	// It is the seat's own, so bob saving does not move alice's.
	if code := rt.stackPos(t, http.MethodPut, bob.Token, map[string]float64{"x": 0, "y": 1}); code != http.StatusNoContent {
		t.Fatalf("bob save: status %d", code)
	}
	if got := rt.viewPos("Alice")[alice.PlayerID.String()]; got != want {
		t.Errorf("alice's spot moved to %+v", got)
	}
	if got := rt.viewPos("Bob")[bob.PlayerID.String()]; got != (protocol.StackPositionView{X: 0, Y: 1}) {
		t.Errorf("bob's spot = %+v", got)
	}

	if code := rt.stackPos(t, http.MethodDelete, alice.Token, nil); code != http.StatusNoContent {
		t.Fatalf("reset: status %d, want 204", code)
	}
	if got := rt.viewPos("Alice"); got != nil {
		t.Errorf("after reset alice still has %+v", got)
	}
	if got := rt.viewPos("Bob"); len(got) != 1 {
		t.Errorf("alice's reset cleared bob's spot: %+v", got)
	}
	// Resetting a seat with nothing saved is fine.
	if code := rt.stackPos(t, http.MethodDelete, alice.Token, nil); code != http.StatusNoContent {
		t.Errorf("second reset: status %d, want 204", code)
	}
}

func TestStackPositionRefusesWhoIsNotASeatAndWhatIsNotAFraction(t *testing.T) {
	rt, _ := activeTable(t)
	alice := rt.seats["Alice"]

	// A spectator.
	resp := postJSON(t, rt.s.srv, "/games/"+rt.meta.ID.String()+"/spectate", "",
		spectateRequest{InviteToken: rt.meta.SpectatorInvite, Name: "watcher"})
	var spec sessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&spec); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if code := rt.stackPos(t, http.MethodPut, spec.Token, map[string]float64{"x": 0.5, "y": 0.5}); code != http.StatusForbidden {
		t.Errorf("spectator save: status %d, want 403", code)
	}
	if code := rt.stackPos(t, http.MethodDelete, spec.Token, nil); code != http.StatusForbidden {
		t.Errorf("spectator reset: status %d, want 403", code)
	}
	// The admin token binds no seat.
	if code := rt.stackPos(t, http.MethodPut, "shared-admin-token", map[string]float64{"x": 0.5, "y": 0.5}); code != http.StatusForbidden && code != http.StatusUnauthorized {
		t.Errorf("admin save: status %d, want 401 or 403", code)
	}
	// No session at all.
	if code := rt.stackPos(t, http.MethodPut, "", map[string]float64{"x": 0.5, "y": 0.5}); code != http.StatusUnauthorized {
		t.Errorf("anonymous save: status %d, want 401", code)
	}

	for name, body := range map[string]any{
		"negative":  map[string]float64{"x": -0.1, "y": 0.5},
		"past one":  map[string]float64{"x": 0.5, "y": 1.01},
		"missing y": map[string]float64{"x": 0.5},
		"unknown":   map[string]float64{"x": 0.5, "y": 0.5, "z": 1},
	} {
		if code := rt.stackPos(t, http.MethodPut, alice.Token, body); code != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", name, code)
		}
	}
	if got := rt.viewPos("Alice"); got != nil {
		t.Errorf("a refused save left %+v", got)
	}
	// Direct callers: NaN and infinities are not fractions either.
	for _, f := range []float64{math.NaN(), math.Inf(1)} {
		if err := rt.s.l.SetStackPosition(rt.meta.ID, alice.PlayerID, &protocol.StackPositionView{X: f, Y: 0}); err != ErrBadStackPosition {
			t.Errorf("SetStackPosition(%v) = %v, want ErrBadStackPosition", f, err)
		}
	}
	// A seat id that is not at the table.
	if err := rt.s.l.SetStackPosition(rt.meta.ID, uuid.New(), &protocol.StackPositionView{}); err != ErrStackPositionNotSeated {
		t.Errorf("a stranger's save = %v, want ErrStackPositionNotSeated", err)
	}
}

func TestStackPositionIsGoneOnceTheGameHasEnded(t *testing.T) {
	rt, _ := activeTable(t)
	alice, bob := rt.seats["Alice"], rt.seats["Bob"]
	if code := rt.stackPos(t, http.MethodPut, alice.Token, map[string]float64{"x": 0.5, "y": 0.5}); code != http.StatusNoContent {
		t.Fatalf("save: status %d", code)
	}
	room := rt.s.l.RoomOf(rt.meta.ID)
	if _, _, err := room.Apply(bob.PlayerID, func() error { return room.Game.Concede(bob.PlayerID) }); err != nil {
		t.Fatal(err)
	}
	if got := rt.viewPos("Alice"); got != nil {
		t.Errorf("an ended game still carries %+v", got)
	}
	if code := rt.stackPos(t, http.MethodPut, alice.Token, map[string]float64{"x": 0.1, "y": 0.1}); code != http.StatusConflict {
		t.Errorf("save after the end: status %d, want 409", code)
	}
}
