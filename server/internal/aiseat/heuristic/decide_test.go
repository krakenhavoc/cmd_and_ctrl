package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func decide(t *testing.T, p *heuristic.Policy, in aiseat.Input) aiseat.Decision {
	t.Helper()
	d, err := p.Decide(context.Background(), in)
	if err != nil {
		t.Fatalf("Decide: %v", err)
	}
	if d.Index != aiseat.Decline && (d.Index < 0 || d.Index >= len(in.Moves)) {
		t.Fatalf("Decide returned index %d out of %d moves", d.Index, len(in.Moves))
	}
	return d
}

func chose(t *testing.T, in aiseat.Input, d aiseat.Decision) string {
	t.Helper()
	if d.Index == aiseat.Decline {
		return "<decline>"
	}
	return in.Moves[d.Index].Label
}

func TestNoMovesIsAnError(t *testing.T) {
	if _, err := heuristic.New().Decide(context.Background(), aiseat.Input{}); err != aiseat.ErrNoMoves {
		t.Fatalf("want ErrNoMoves, got %v", err)
	}
}

func TestSingleMoveIsTakenWithoutThinking(t *testing.T) {
	in := input(0, newView([]protocol.PlayerView{newSeat(0), newSeat(1)}), passMove(0))
	d := decide(t, heuristic.New(), in)
	if d.Index != 0 {
		t.Fatalf("index %d", d.Index)
	}
}

// The one sequencing rule that matters at this level: land, then
// spend. A bot that casts first and plays its land afterwards has
// thrown away a mana every turn it did it.
func TestLandDropComesBeforeAnyCast(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0, withHand(
		land(cardID(1), 0),
		creature(cardID(2), 0, "Bear", 2, 2),
		creature(cardID(3), 0, "Dragon", 6, 6),
	)), newSeat(1)}, withBattlefield(land(cardID(10), 0), land(cardID(11), 0)))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, cardID(2), "Cast Bear"),
		castMove(t, 0, cardID(3), "Cast Dragon"),
		landMove(t, 0, cardID(1)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Play Mountain" {
		t.Fatalf("chose %q, want the land drop", got)
	}
}

func TestBiggerThreatGetsCast(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0, withHand(
		creature(cardID(2), 0, "Bear", 2, 2),
		creature(cardID(3), 0, "Dragon", 6, 6),
	)), newSeat(1)}, withBattlefield(land(cardID(10), 0), land(cardID(11), 0)))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, cardID(2), "Cast Bear"),
		castMove(t, 0, cardID(3), "Cast Dragon"),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Dragon" {
		t.Fatalf("chose %q, want the bigger creature", got)
	}
}

// Removal goes at the table's biggest problem, not at whoever happens
// to be enumerated first. This is the four-player-first requirement
// from ADR 0033 §3 / the S31 issue's "threat ranking across three
// opponents", and it is the behaviour a duel-shaped heuristic gets
// wrong.
func TestRemovalGoesAtTheLeader(t *testing.T) {
	bolt := spell(cardID(1), 0, "Lightning Bolt", "{R}")
	v := newView(
		[]protocol.PlayerView{newSeat(0, withHand(bolt)), newSeat(1, withHandCount(1)), newSeat(2, withHandCount(7)), newSeat(3)},
		withBattlefield(
			land(cardID(10), 0),
			creature(cardID(20), 1, "Small", 1, 1),
			creature(cardID(21), 2, "Huge", 7, 7),
			creature(cardID(22), 3, "Middling", 3, 3),
		))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, cardID(1), "Bolt the small one", cardTarget(cardID(20))),
		castMove(t, 0, cardID(1), "Bolt the huge one", cardTarget(cardID(21))),
		castMove(t, 0, cardID(1), "Bolt the middling one", cardTarget(cardID(22))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Bolt the huge one" {
		t.Fatalf("chose %q, want the leader's creature", got)
	}
}

func TestSpellsAreNotPointedAtOurselves(t *testing.T) {
	bolt := spell(cardID(1), 0, "Lightning Bolt", "{R}")
	v := newView([]protocol.PlayerView{newSeat(0, withHand(bolt)), newSeat(1)},
		withBattlefield(land(cardID(10), 0)))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, cardID(1), "Bolt myself", playerTarget(0)),
		castMove(t, 0, cardID(1), "Bolt them", playerTarget(1)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Bolt them" {
		t.Fatalf("chose %q", got)
	}
}

func TestOurOwnSpellIsNotCountered(t *testing.T) {
	counter := spell(cardID(1), 0, "Counterspell", "{U}{U}")
	mine := spell(cardID(30), 0, "My Spell", "{2}")
	theirs := spell(cardID(31), 1, "Their Spell", "{2}")
	v := newView([]protocol.PlayerView{newSeat(0, withHand(counter)), newSeat(1)},
		withBattlefield(land(cardID(10), 0), land(cardID(11), 0)),
		withStack(mine, theirs),
		withTurn(3, 1, "precombat_main"))
	in := input(0, v,
		passMove(0),
		castMove(t, 0, cardID(1), "Counter mine", cardTarget(cardID(30))),
		castMove(t, 0, cardID(1), "Counter theirs", cardTarget(cardID(31))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Counter theirs" {
		t.Fatalf("chose %q", got)
	}
}

// Floating mana buys a bot nothing — casts auto-tap — and a policy
// that reaches for it every priority window produces a table that
// never advances.
func TestManaAbilitiesAreNotActivatedForNothing(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(land(cardID(10), 0)))
	in := input(0, v, passMove(0), manaMove(t, 0, cardID(10)))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q, want the pass", got)
	}
}

// Holding an instant is a real option. On an opponent's turn the bar
// for spending one is higher than it is in our own main phase.
func TestInstantsAreHeldOnAnOpponentsTurnUnlessTheyAreWorthIt(t *testing.T) {
	bolt := spell(cardID(1), 0, "Lightning Bolt", "{R}")
	seats := []protocol.PlayerView{newSeat(0, withHand(bolt)), newSeat(1)}
	oppTurn := withTurn(4, 1, "upkeep")

	weak := newView(seats, withBattlefield(land(cardID(10), 0), creature(cardID(20), 1, "Mouse", 1, 1)), oppTurn)
	in := input(0, weak, passMove(0), castMove(t, 0, cardID(1), "Bolt the mouse", cardTarget(cardID(20))))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q on a 1/1; the bot should hold", got)
	}

	strong := newView(seats, withBattlefield(land(cardID(10), 0), creature(cardID(20), 1, "Dragon", 6, 6, keywords("flying"))), oppTurn)
	in = input(0, strong, passMove(0), castMove(t, 0, cardID(1), "Bolt the dragon", cardTarget(cardID(20))))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Bolt the dragon" {
		t.Fatalf("chose %q on a 6/6 flier; the bot should answer it", got)
	}
}

func TestMulliganKeepsACastableHand(t *testing.T) {
	keep := legal.Move{Type: legal.TypeKeepHand, Player: seatID(0), Kind: legal.KindMulligan, Label: "Keep hand"}
	mull := legal.Move{
		Type: legal.TypeMulligan, Player: seatID(0), Kind: legal.KindMulligan, Label: "Mulligan to 7",
		Params: mustJSON(t, map[string]int{"hand_size": 7}),
	}
	hand := func(lands int) []protocol.CardView {
		var out []protocol.CardView
		for i := 0; i < lands; i++ {
			out = append(out, land(cardID(100+i), 0))
		}
		for i := lands; i < 7; i++ {
			out = append(out, creature(cardID(100+i), 0, "Bear", 2, 2))
		}
		return out
	}
	for _, tc := range []struct {
		lands int
		want  string
	}{
		{0, "Mulligan to 7"},
		{1, "Mulligan to 7"},
		{2, "Keep hand"},
		{3, "Keep hand"},
		{4, "Keep hand"},
		{5, "Keep hand"},
		{6, "Mulligan to 7"},
		{7, "Mulligan to 7"},
	} {
		v := newView([]protocol.PlayerView{newSeat(0, withHand(hand(tc.lands)...)), newSeat(1)})
		in := input(0, v, keep, mull)
		if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
			t.Errorf("%d lands: chose %q, want %q", tc.lands, got, tc.want)
		}
	}
}

func TestMulliganStopsDigging(t *testing.T) {
	keep := legal.Move{Type: legal.TypeKeepHand, Player: seatID(0), Kind: legal.KindMulligan, Label: "Keep hand"}
	mull := legal.Move{
		Type: legal.TypeMulligan, Player: seatID(0), Kind: legal.KindMulligan, Label: "Mulligan to 5",
		Params: mustJSON(t, map[string]int{"hand_size": 5}),
	}
	seat := newSeat(0)
	seat.MulligansTaken = 2
	seat.Hand = protocol.ZoneView{Kind: "hand", Count: 0}
	in := input(0, newView([]protocol.PlayerView{seat, newSeat(1)}), keep, mull)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Keep hand" {
		t.Fatalf("chose %q on a landless hand after two mulligans; digging to five costs more than it buys", got)
	}
}

// A cancelled context must not produce an invalid answer or a long
// one: ADR 0033 §10 is explicit that the table never waits on a bot.
func TestDecideRespectsAHardDeadline(t *testing.T) {
	var moves []legal.Move
	moves = append(moves, passMove(0))
	hand := []protocol.CardView{}
	for i := 0; i < 200; i++ {
		hand = append(hand, creature(cardID(200+i), 0, "Bear", 2, 2))
		moves = append(moves, castMove(t, 0, cardID(200+i), "Cast Bear"))
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	d, err := heuristic.New().Decide(ctx, input(0, v, moves...))
	if err != nil {
		t.Fatalf("a cancelled context must still yield a move: %v", err)
	}
	if d.Index != aiseat.Decline && (d.Index < 0 || d.Index >= len(moves)) {
		t.Fatalf("index %d out of %d", d.Index, len(moves))
	}
}

func TestReasonIsAlwaysPopulated(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0, withHand(creature(cardID(2), 0, "Bear", 2, 2))), newSeat(1)},
		withBattlefield(land(cardID(10), 0), land(cardID(11), 0)))
	in := input(0, v, passMove(0), castMove(t, 0, cardID(2), "Cast Bear"))
	d := decide(t, heuristic.New(), in)
	if strings.TrimSpace(d.Reason) == "" {
		t.Fatal("every decision must carry a reason for the log")
	}
}
