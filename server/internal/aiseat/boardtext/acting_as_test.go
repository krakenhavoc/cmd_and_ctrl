package boardtext_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/boardtext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ADR 0167 §7: a move the seat makes for a player whose turn it plays
// says whose it is; the seat's own moves say nothing.
func TestActingAsNoteNamesTheControlledSeat(t *testing.T) {
	me, bob := uuid.New(), uuid.New()
	v := &protocol.GameView{Seats: []protocol.PlayerView{
		{ID: me.String(), Name: "Alice"},
		{ID: bob.String(), Name: "Bob"},
	}}
	if got := boardtext.ActingAsNote(v, legal.Move{Player: bob}, me.String()); got != "as Bob" {
		t.Errorf("a move for Bob: %q, want %q", got, "as Bob")
	}
	if got := boardtext.ActingAsNote(v, legal.Move{Player: me}, me.String()); got != "" {
		t.Errorf("the seat's own move: %q, want none", got)
	}
	if got := boardtext.ActingAsNote(v, legal.Move{}, me.String()); got != "" {
		t.Errorf("a move with no player: %q, want none", got)
	}
}
