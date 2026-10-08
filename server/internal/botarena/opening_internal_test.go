package botarena

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// opening_internal_test.go pins #2693's arena numbers on a hand-built
// observer feed: mulligans, the kept hand's size, and the land drops on
// the seat's own turns 2–4.

func TestOpeningCountsMulligansKeepsAndMissedLandDrops(t *testing.T) {
	me, them := uuid.New(), uuid.New()
	view := func(seq, active, hand int) protocol.GameView {
		v := protocol.GameView{
			Seats: []protocol.PlayerView{{ID: me.String()}, {ID: them.String()}},
			Turn:  protocol.TurnView{Seq: seq, ActiveSeat: active, Step: "precombat_main"},
		}
		for i := 0; i < hand; i++ {
			v.Seats[0].Hand.Cards = append(v.Seats[0].Hand.Cards, protocol.CardView{InstanceID: uuid.NewString()})
		}
		return v
	}
	ev := func(seat uuid.UUID, v protocol.GameView, m legal.Move) aiseat.DecisionEvent {
		return aiseat.DecisionEvent{Seat: seat, Applied: true, Index: 0, Input: aiseat.Input{View: v, Moves: []legal.Move{m}}}
	}
	mull := legal.Move{Type: legal.TypeMulligan, Kind: legal.KindMulligan}
	keep := legal.Move{Type: legal.TypeKeepHand, Kind: legal.KindMulligan}
	land := legal.Move{Type: legal.TypeCastSpell, Kind: legal.KindLand}
	pass := legal.Move{Type: legal.TypePassPriority, Kind: legal.KindPass}

	w := newOpeningWatch(me)
	w.Observe(ev(me, view(1, 0, 7), mull))
	w.Observe(ev(me, view(1, 0, 6), keep))
	// Another seat's events are not this seat's.
	w.Observe(ev(them, view(1, 0, 7), mull))
	// Own turns 1..5 at seqs 1, 3, 5, 7, 9: lands on 1, 3 and 9, so
	// turns 3 and 4 (seqs 5 and 7) miss. Turn 5 is not counted.
	for _, seq := range []int{1, 3, 5, 7, 9} {
		if seq == 1 || seq == 3 || seq == 9 {
			w.Observe(ev(me, view(seq, 0, 5), land))
		}
		w.Observe(ev(me, view(seq, 0, 5), pass))
		// The opponent's turn in between is not an own turn.
		w.Observe(ev(me, view(seq+1, 1, 5), pass))
	}
	got := w.result()
	if got.Seats != 1 || got.Mulligans != 1 || got.Kept[6] != 1 || got.DropTurns != 3 || got.Missed != 2 {
		t.Fatalf("result = %+v, want 1 keep of 6 after 1 mulligan, 3 drop turns, 2 missed", got)
	}
}
