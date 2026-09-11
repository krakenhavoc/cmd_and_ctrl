package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// hopelessAt builds the S31 concede-path fixture: the bot at 1 life
// with an empty hand and an empty board, and an opponent with a
// creature that will finish it.
func hopelessAt(turn int, opts ...seatOpt) aiseat.Input {
	me := append([]seatOpt{withLife(1), withHandCount(0)}, opts...)
	v := newView(
		[]protocol.PlayerView{newSeat(0, me...), newSeat(1)},
		withBattlefield(creature(cardID(20), 1, "Killer", 4, 4)),
		withTurn(turn, 1, "upkeep"),
	)
	return input(0, v, passMove(0))
}

func TestConcedeOnlyAfterThreeHopelessTurns(t *testing.T) {
	p := heuristic.New()
	for turn := 1; turn <= 2; turn++ {
		if p.ShouldConcede(hopelessAt(turn)) {
			t.Fatalf("conceded on turn %d after %d hopeless turns; the bar is three", turn, turn)
		}
	}
	if !p.ShouldConcede(hopelessAt(3)) {
		t.Fatal("three consecutive hopeless turns must end in a concede")
	}
}

func TestConcedeCounterIsPerTurnNotPerDecision(t *testing.T) {
	p := heuristic.New()
	// Twenty decisions in one turn is one hopeless turn, not twenty.
	for i := 0; i < 20; i++ {
		if p.ShouldConcede(hopelessAt(1)) {
			t.Fatalf("conceded after %d decisions inside a single turn", i+1)
		}
	}
	if got := p.HopelessTurns(); got != 1 {
		t.Fatalf("HopelessTurns = %d after one turn, want 1", got)
	}
}

// The heuristic is tuned conservative on purpose (S31: "default
// conservative so humans get the finish"). Each of these is a
// position a bot must keep playing.
func TestConcedeIsConservative(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   func(turn int) aiseat.Input
	}{
		{
			name: "a card in hand is an upswing",
			in:   func(turn int) aiseat.Input { return hopelessAt(turn, withHandCount(1)) },
		},
		{
			name: "life above the floor",
			in:   func(turn int) aiseat.Input { return hopelessAt(turn, withLife(9)) },
		},
		{
			name: "a creature on the board",
			in: func(turn int) aiseat.Input {
				v := newView(
					[]protocol.PlayerView{newSeat(0, withLife(1), withHandCount(0)), newSeat(1)},
					withBattlefield(
						creature(cardID(10), 0, "Mine", 1, 1),
						creature(cardID(20), 1, "Killer", 4, 4),
					),
					withTurn(turn, 1, "upkeep"),
				)
				return input(0, v, passMove(0))
			},
		},
		{
			name: "nobody left who can finish it",
			in: func(turn int) aiseat.Input {
				v := newView(
					[]protocol.PlayerView{newSeat(0, withLife(1), withHandCount(0)), newSeat(1, withHandCount(0))},
					withTurn(turn, 1, "upkeep"),
				)
				return input(0, v, passMove(0))
			},
		},
		{
			name: "last seat standing",
			in: func(turn int) aiseat.Input {
				v := newView(
					[]protocol.PlayerView{newSeat(0, withLife(1), withHandCount(0)), newSeat(1, withEliminated())},
					withTurn(turn, 0, "upkeep"),
				)
				return input(0, v, passMove(0))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := heuristic.New()
			for turn := 1; turn <= 8; turn++ {
				if p.ShouldConcede(tc.in(turn)) {
					t.Fatalf("conceded on turn %d", turn)
				}
			}
		})
	}
}

// One turn of hope resets the count: a bot that drew an answer and
// then lost it again starts the clock over.
func TestConcedeCounterResetsOnAnyHope(t *testing.T) {
	p := heuristic.New()
	p.ShouldConcede(hopelessAt(1))
	p.ShouldConcede(hopelessAt(2))
	if p.ShouldConcede(hopelessAt(3, withHandCount(2))) {
		t.Fatal("conceded on the turn it drew into a hand")
	}
	if got := p.HopelessTurns(); got != 0 {
		t.Fatalf("HopelessTurns = %d, want the counter reset", got)
	}
	if p.ShouldConcede(hopelessAt(4)) || p.ShouldConcede(hopelessAt(5)) {
		t.Fatal("the count must restart from zero")
	}
	if !p.ShouldConcede(hopelessAt(6)) {
		t.Fatal("three fresh hopeless turns must end in a concede")
	}
}

func TestConcedeCanBeDisabled(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	cfg.Concede = false
	p := heuristic.NewWithConfig(cfg)
	for turn := 1; turn <= 10; turn++ {
		if p.ShouldConcede(hopelessAt(turn)) {
			t.Fatalf("conceded on turn %d with Concede disabled", turn)
		}
	}
}
