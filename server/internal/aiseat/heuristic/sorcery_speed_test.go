package heuristic

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestSorcerySpeedReadsAbilitiesOnTheStack is #1352's view-side half.
// The policy cannot ask the engine's sorcery-timing read — it sees a
// GameView — so it derives "is this my sorcery window" from the wire,
// and an activated or triggered ability reaches the wire only in
// stack_items: it has no card in stack.cards. A trigger on the stack
// is an instant-speed window (CR 307.1), and the policy has to price
// its moves against the instant threshold there.
func TestSorcerySpeedReadsAbilitiesOnTheStack(t *testing.T) {
	me := uuid.New()
	view := func(items []protocol.StackItemView) protocol.GameView {
		var v protocol.GameView
		v.Seats = []protocol.PlayerView{{ID: me.String()}}
		v.Turn.ActiveSeat = 0
		v.Turn.Step = "precombat_main"
		v.StackItems = items
		return v
	}
	p := New()

	if st := p.newState(aiseat.Input{Seat: me, View: view(nil)}); !st.sorcerySpeed {
		t.Fatal("setup: my main phase with an empty stack is not a sorcery window")
	}
	trigger := []protocol.StackItemView{{ID: uuid.NewString(), Kind: "triggered", Controller: me.String()}}
	if st := p.newState(aiseat.Input{Seat: me, View: view(trigger)}); st.sorcerySpeed {
		t.Error("a trigger on the stack left the sorcery window open")
	}
}
