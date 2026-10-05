package model

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestStackOwnershipNamesAStolenSpellsCaster — ADR 0104: a spell taken
// on the stack acts for its new controller, and the prompt says both
// who controls it and who cast it.
func TestStackOwnershipNamesAStolenSpellsCaster(t *testing.T) {
	v := &protocol.GameView{Seats: []protocol.PlayerView{
		{ID: "a", Name: "Alice", Seat: 0},
		{ID: "b", Name: "Bob", Seat: 1},
	}}
	plain := &protocol.StackItemView{Controller: "a"}
	if got, want := stackOwnership(plain, v, "b"), "cast by Alice"; got != want {
		t.Errorf("untouched spell: %q, want %q", got, want)
	}
	stolen := &protocol.StackItemView{Controller: "b", DefaultController: "a"}
	if got, want := stackOwnership(stolen, v, "b"), "controlled by Bob (YOU) (cast by Alice)"; got != want {
		t.Errorf("stolen spell: %q, want %q", got, want)
	}
}

// #2279: an ability is activated or triggered, never cast (CR 601-603).
func TestStackOwnershipDoesNotSayCastForAbilities(t *testing.T) {
	v := &protocol.GameView{Seats: []protocol.PlayerView{
		{ID: "a", Name: "Alice", Seat: 0},
		{ID: "b", Name: "Bob", Seat: 1},
	}}
	for _, c := range []struct{ kind, want string }{
		{"spell", "cast by Alice"},
		{"activated", "activated by Alice"},
		{"triggered", "triggered, controlled by Alice"},
	} {
		it := &protocol.StackItemView{Kind: c.kind, Controller: "a"}
		if got := stackOwnership(it, v, "b"); got != c.want {
			t.Errorf("%s: %q, want %q", c.kind, got, c.want)
		}
	}
}
