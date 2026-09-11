package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// blockingView: seat 1 attacks seat 0 with `atk`; seat 0 has `blk`.
func blockingView(t *testing.T, myLife int, atk, blk protocol.CardView) protocol.GameView {
	t.Helper()
	return newView(
		[]protocol.PlayerView{newSeat(0, withLife(myLife)), newSeat(1)},
		withBattlefield(atk, blk),
		withTurn(5, 1, "declare_blockers"),
	)
}

// The sprint's phrasing is "blocks minimise incoming damage subject
// to not trading up". Both halves are here: the bot blocks when the
// exchange is good for it, and does not throw a real creature under
// an attacker that eats it for free.
func TestBlocksTradeUpNotDown(t *testing.T) {
	// Every case offers two blockers: the one under test and a spare
	// 1/1, so the planner has to CHOOSE rather than take the only
	// move on offer. "" means the bot should decline to block at all.
	const spare = 11
	for _, tc := range []struct {
		name     string
		life     int
		attacker protocol.CardView
		blocker  protocol.CardView
		want     string
	}{
		{
			name:     "big blocker eats a small attacker",
			life:     40,
			attacker: creature(cardID(20), 1, "Bear", 2, 2, attacking(0)),
			blocker:  creature(cardID(10), 0, "Giant", 5, 5),
			want:     cardID(10),
		},
		{
			name:     "an even trade is worth taking",
			life:     40,
			attacker: creature(cardID(20), 1, "Bear", 2, 2, attacking(0)),
			blocker:  creature(cardID(10), 0, "Bear", 2, 2),
			want:     cardID(10),
		},
		{
			name:     "do not chump a big attacker at a healthy life total",
			life:     40,
			attacker: creature(cardID(20), 1, "Giant", 5, 5, attacking(0)),
			blocker:  creature(cardID(10), 0, "Bear", 2, 2),
			want:     "",
		},
		{
			name:     "trade the 1/1 into a deathtoucher, never the 6/6",
			life:     40,
			attacker: creature(cardID(20), 1, "Adder", 1, 1, keywords("deathtouch"), attacking(0)),
			blocker:  creature(cardID(10), 0, "Colossus", 6, 6),
			want:     cardID(spare),
		},
		{
			name:     "chump with the cheapest body when the alternative is dying",
			life:     6,
			attacker: creature(cardID(20), 1, "Giant", 5, 5, attacking(0)),
			blocker:  creature(cardID(10), 0, "Bear", 2, 2),
			want:     cardID(spare),
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// A defender is offered blocks BEFORE priority reaches
			// it, so there is deliberately no pass in this window.
			v := blockingView(t, tc.life, tc.attacker, tc.blocker)
			v.Battlefield.Cards = append(v.Battlefield.Cards, creature(cardID(spare), 0, "Spare", 1, 1))
			v.Battlefield.Count = len(v.Battlefield.Cards)
			in := input(0, v,
				blockMove(t, 0, cardID(10), cardID(20)),
				blockMove(t, 0, cardID(spare), cardID(20)),
			)
			d := decide(t, heuristic.New(), in)
			if tc.want == "" {
				if d.Index != aiseat.Decline {
					t.Fatalf("blocked with %s; the bot should have declined", chose(t, in, d))
				}
				return
			}
			if d.Index == aiseat.Decline {
				t.Fatalf("declined; expected a block with %s", tc.want)
			}
			if got := chose(t, in, d); got != "Block with "+tc.want {
				t.Fatalf("chose %q, want a block with %s", got, tc.want)
			}
		})
	}
}

// Declining is the mechanism that lets a defender stop blocking at
// all: there is no pass on offer in a blocks-only window, so without
// it the bot would have to keep declaring until it ran out of
// creatures.
func TestBlocksOnlyWindowDeclinesRatherThanChumping(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(20), 1, "Giant", 6, 6, attacking(0)),
			creature(cardID(10), 0, "Bear", 2, 2),
			creature(cardID(11), 0, "Cub", 1, 1),
		),
		withTurn(5, 1, "declare_blockers"),
	)
	in := input(0, v,
		blockMove(t, 0, cardID(10), cardID(20)),
		blockMove(t, 0, cardID(11), cardID(20)),
	)
	if d := decide(t, heuristic.New(), in); d.Index != aiseat.Decline {
		t.Fatalf("chose %q; at 40 life neither body should go under a 6/6", chose(t, in, d))
	}
}

func TestAttacksAvoidUnprofitableSwings(t *testing.T) {
	// A lone 2/2 into an untapped 5/5 is a gift.
	v := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(
			creature(cardID(10), 0, "Bear", 2, 2),
			creature(cardID(20), 1, "Giant", 5, 5),
		),
		withTurn(5, 0, "declare_attackers"),
	)
	in := input(0, v, passMove(0), attackMove(t, 0, cardID(10), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q, want no attack", got)
	}

	// The same 2/2 into an open board is free damage.
	open := newView(
		[]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(creature(cardID(10), 0, "Bear", 2, 2)),
		withTurn(5, 0, "declare_attackers"),
	)
	in = input(0, open, passMove(0), attackMove(t, 0, cardID(10), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == "Pass priority" {
		t.Fatal("an undefended opponent must be attacked")
	}
}

func TestLethalSwingsThroughAnOpenBoard(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(12)), newSeat(1, withLife(2))},
		withBattlefield(creature(cardID(10), 0, "Bear", 2, 2)),
		withTurn(9, 0, "declare_attackers"),
	)
	in := input(0, v, passMove(0), attackMove(t, 0, cardID(10), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == "Pass priority" {
		t.Fatal("a lethal swing must be taken even with the bot's own life low")
	}
}

// A bot that alpha-strikes into three untapped opponents dies to the
// crack-back. It keeps a body home once its own life is in range.
func TestAttackReserveKeepsABlockerHome(t *testing.T) {
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(18)), newSeat(1)},
		withBattlefield(
			creature(cardID(10), 0, "Bear", 2, 2),
			creature(cardID(20), 1, "Their Bear", 2, 2),
		),
		withTurn(6, 0, "declare_attackers"),
	)
	// The bot's last untapped creature, with a live opposing board:
	// it stays home.
	in := input(0, v, passMove(0), attackMove(t, 0, cardID(10), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
		t.Fatalf("chose %q; the bot's last blocker should stay home at 18 life", got)
	}

	// With two bodies it can afford to send one.
	v2 := newView(
		[]protocol.PlayerView{newSeat(0, withLife(18)), newSeat(1)},
		withBattlefield(
			creature(cardID(10), 0, "Bear", 2, 2),
			creature(cardID(11), 0, "Cub", 2, 2),
			creature(cardID(20), 1, "Their Bear", 2, 2),
		),
		withTurn(6, 0, "declare_attackers"),
	)
	in = input(0, v2, passMove(0), attackMove(t, 0, cardID(10), 1), attackMove(t, 0, cardID(11), 1))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == "Pass priority" {
		t.Fatal("with two untapped creatures one of them should attack")
	}
}
