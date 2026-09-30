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
