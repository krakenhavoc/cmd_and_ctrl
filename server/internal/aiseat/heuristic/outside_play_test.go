package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// outside_play_test.go — #3955, Wish: the action that names a card to play
// is a move a bot takes rather than passing over.
func TestABotNamesACardUnderTheWishRight(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(land(cardID(30), 0)))
	name := legal.Move{
		Type: legal.TypeOutsidePlay, Player: seatID(0), Kind: legal.KindSpecialAction,
		Label:  "Wish: name a card you own from outside the game to play this turn",
		Source: uuid.Nil,
		Params: mustJSON(t, map[string]any{"kind": "outside_play"}),
	}
	in := input(0, v, passMove(0), name)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != name.Label {
		t.Fatalf("chose %q, want the naming action", got)
	}
}
