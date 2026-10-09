package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// rocknow_test.go is ADR 0136's amendment of 2026-10-09 (a rock against
// a spell over two turns, Config.PlanRockTwoTurns) and ADR 0126 §2's
// amendment of the same day (an idle late rock, Config.IdleLateRocks).
// Each test fails with its knob off.

func decideWith(t *testing.T, in aiseat.Input, tweak func(*heuristic.Config)) aiseat.Decision {
	t.Helper()
	cfg := heuristic.DefaultConfig()
	if tweak != nil {
		tweak(&cfg)
	}
	d, err := heuristic.NewWithConfig(cfg).Decide(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// rockOrBearBoard is the (d) class of PR 4b's diagnosis: two lands, so
// the turn pays for Arcane Signet or a two-drop, not both. The hand
// holds a land for next turn and a four-drop.
func rockOrBearBoard(t *testing.T, fourDrop bool) aiseat.Input {
	t.Helper()
	signet := arcaneSignet(cardID(10), 0)
	bear := creature(cardID(11), 0, "Grizzly Bears", 2, 2)
	bear.TypeLine, bear.ManaCost = "Creature — Bear", "{1}{G}"
	hand := []protocol.CardView{signet, bear, planForest(cardID(13))}
	if fourDrop {
		big := creature(cardID(12), 0, "Siege Wurm", 5, 5, keywords("trample"))
		big.TypeLine, big.ManaCost = "Creature — Wurm", "{3}{G}"
		hand = append(hand, big)
	} else {
		other := creature(cardID(12), 0, "Runeclaw Bear", 2, 2)
		other.TypeLine, other.ManaCost = "Creature — Bear", "{1}{G}"
		hand = append(hand, other)
	}
	v := newView(
		[]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)},
		withTurn(3, 0, "precombat_main"),
		withBattlefield(planForest(cardID(1)), planForest(cardID(2))),
	)
	return input(0, v,
		passMove(0),
		stampedCast(t, cardID(11), "Cast Grizzly Bears", "{1}{G}"),
		stampedCast(t, cardID(10), "Cast Arcane Signet", "{2}"),
	)
}

// TestRockNowWhenItReachesTheFourDrop: with the Signet now, next turn's
// four mana casts the four-drop; with the bear now, next turn's three
// mana does not. The two turns are worth more with the Signet first.
func TestRockNowWhenItReachesTheFourDrop(t *testing.T) {
	in := rockOrBearBoard(t, true)
	off := decideWith(t, in, func(c *heuristic.Config) { c.PlanRockTwoTurns = false })
	if got := moveLabel(in, off); got != "Cast Grizzly Bears" {
		t.Fatalf("with the two-turn comparison off the bot casts %q (%s); this board no longer shows the problem", got, off.Reason)
	}
	on := decideWith(t, in, nil)
	if got := moveLabel(in, on); got != "Cast Arcane Signet" {
		t.Errorf("the bot casts %q (%s), want Arcane Signet", got, on.Reason)
	}
	if !strings.HasPrefix(on.Reason, "two turns: Arcane Signet now, then Siege Wurm (+") {
		t.Errorf("reason = %q", on.Reason)
	}
	// With no discount at all next turn is worth nothing, and the bear
	// wins on this turn alone.
	if d := decideWith(t, in, func(c *heuristic.Config) { c.PlanNextTurnDiscount = 0 }); moveLabel(in, d) != "Cast Grizzly Bears" {
		t.Errorf("at discount 0 the bot casts %q (%s), want the bear", moveLabel(in, d), d.Reason)
	}
	// The comparison is the plan's: with the plan off it does not run.
	if d := decideWith(t, in, func(c *heuristic.Config) { c.PlanTurnMana = false }); moveLabel(in, d) != "Cast Grizzly Bears" {
		t.Errorf("with the plan off the bot casts %q (%s), want the bear", moveLabel(in, d), d.Reason)
	}
}

// TestRockWaitsWhenTheSameCardsAreCastEitherWay: with a second two-drop
// in place of the four-drop, both orders cast the same three cards over
// the two turns, so the bear, worth more now, goes first.
func TestRockWaitsWhenTheSameCardsAreCastEitherWay(t *testing.T) {
	in := rockOrBearBoard(t, false)
	if d := decideWith(t, in, nil); moveLabel(in, d) != "Cast Grizzly Bears" {
		t.Errorf("the bot casts %q (%s), want the bear", moveLabel(in, d), d.Reason)
	}
}

// idleRockBoard is the second main phase of a late turn: seven lands
// untapped, Arcane Signet in hand, nothing else to cast and no open
// deficit, so §2 prices the Signet below zero.
func idleRockBoard(t *testing.T, step string, extra ...protocol.CardView) aiseat.Input {
	t.Helper()
	bf := []protocol.CardView{}
	for i := 1; i <= 7; i++ {
		bf = append(bf, planForest(cardID(i)))
	}
	hand := append([]protocol.CardView{arcaneSignet(cardID(10), 0)}, extra...)
	v := newView(
		[]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1)},
		withTurn(9, 0, step),
		withBattlefield(bf...),
	)
	return input(0, v, passMove(0), stampedCast(t, cardID(10), "Cast Arcane Signet", "{2}"))
}

// TestIdleLateRockIsCastInTheLastMainPhase is ADR 0126 §2's "Arcane
// Signet, turn 9" row as amended: in the turn's last main phase, with
// nothing else to spend the mana on, the Signet is cast.
func TestIdleLateRockIsCastInTheLastMainPhase(t *testing.T) {
	in := idleRockBoard(t, "postcombat_main")
	off := decideWith(t, in, func(c *heuristic.Config) { c.IdleLateRocks = false })
	if in.Moves[off.Index].Kind != "pass" {
		t.Fatalf("with the knob off the bot makes %q (%s); this board no longer shows the problem", moveLabel(in, off), off.Reason)
	}
	on := decideWith(t, in, nil)
	if got := moveLabel(in, on); got != "Cast Arcane Signet" {
		t.Errorf("the bot makes %q (%s), want Arcane Signet", got, on.Reason)
	}
	// The first main phase is not the turn's last window: the bot may
	// still draw into something, or attack and then cast.
	in = idleRockBoard(t, "precombat_main")
	if d := decideWith(t, in, nil); in.Moves[d.Index].Kind != "pass" {
		t.Errorf("in the first main phase the bot makes %q (%s), want a pass", moveLabel(in, d), d.Reason)
	}
}

// TestIdleLateRockWaitsForAPricedCast: a cast priced above zero takes
// the mana, so the Signet is not idle.
func TestIdleLateRockWaitsForAPricedCast(t *testing.T) {
	bear := creature(cardID(11), 0, "Grizzly Bears", 2, 2)
	bear.TypeLine, bear.ManaCost = "Creature — Bear", "{1}{G}"
	in := idleRockBoard(t, "postcombat_main", bear)
	in.Moves = append(in.Moves, stampedCast(t, cardID(11), "Cast Grizzly Bears", "{1}{G}"))
	if d := decideWith(t, in, nil); moveLabel(in, d) != "Cast Grizzly Bears" {
		t.Errorf("the bot makes %q (%s), want the bear", moveLabel(in, d), d.Reason)
	}
}
