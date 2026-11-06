package heuristic_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// pump_x_test.go — #4976: a benefit until end of turn whose power or
// toughness is the announced X (Tyvar's Stand's +X/+X, Hatred's +X/+0)
// is priced at the X the move names, under Config.PriceXPumps.

// xTrick is a seat-0 instant whose one target entry is a benefit with
// pump: the printed part, and the sides that are X.
func xTrick(id, name, cost string, pump *protocol.PumpView, powerX, toughnessX bool) protocol.CardView {
	c := spell(id, 0, name, cost)
	c.Purpose = targetEntries(protocol.TargetPurposeView{Slot: 0, UntilEndOfTurn: "benefit", Pump: pump,
		PumpPowerIsX: powerX, PumpToughnessIsX: toughnessX})
	return c
}

func tyvarsStand(id string) protocol.CardView {
	return xTrick(id, "Tyvar's Stand", "{X}{G}", &protocol.PumpView{Keywords: []string{"hexproof", "indestructible"}}, true, true)
}

// hatred is Hatred: "As an additional cost to cast this spell, pay X
// life. Target creature gets +X/+0 until end of turn."
func hatred(id string) protocol.CardView {
	c := xTrick(id, "Hatred", "{3}{B}{B}", nil, true, false)
	c.AdditionalCost = &protocol.AdditionalCostView{DemandsX: true, Label: "Pay X life"}
	return c
}

func castXMove(t *testing.T, id string, x int, target string) legal.Move {
	t.Helper()
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(0), Kind: legal.KindCast,
		Label: fmt.Sprintf("Cast at X=%d", x), Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "x_value": x,
			"targets": []map[string]string{cardTarget(target)}}),
	}
}

func xRanks(t *testing.T, pol *heuristic.Policy, v protocol.GameView, moves ...legal.Move) []float64 {
	t.Helper()
	out := make([]float64, len(moves))
	for _, c := range pol.Rank(context.Background(), input(0, v, moves...)) {
		out[c.Index] = c.Value
	}
	return out
}

// xPumps is today's heuristic with PriceXPumps on; DefaultConfig ships it
// off (the arena found no gain), so heuristic.New() is the price off.
func xPumps() *heuristic.Policy {
	return withConfig(func(c *heuristic.Config) { c.PriceXPumps = true })
}

func noXPumps() *heuristic.Policy { return heuristic.New() }

// attackBoard is the bot's 2/2 attacking seat 1 before blocks, seat 1 on
// `life` with an untapped 3/3 that can block it, and four lands.
func attackBoard(hand protocol.CardView, life int) protocol.GameView {
	return newView([]protocol.PlayerView{newSeat(0, withHand(hand)), newSeat(1, withLife(life))},
		withBattlefield(land(cardID(91), 0), land(cardID(92), 0), land(cardID(93), 0), land(cardID(94), 0),
			creature(cardID(3), 0, "Attacker", 2, 2, attacking(1)),
			creature(cardID(2), 1, "Their Bear", 3, 3)),
		withTurn(5, 0, "declare_attackers"))
}

// TestTyvarsStandAtXIsPricedAsItsPump is the issue's acceptance test:
// Tyvar's Stand cast for X=3 on an attacker before blocks is priced as
// +3/+3 with hexproof and indestructible, exactly as a spell that prints
// that amount is. With PriceXPumps off it is priced as before, a benefit
// with no amount, and the X changes nothing.
func TestTyvarsStandAtXIsPricedAsItsPump(t *testing.T) {
	ts := tyvarsStand(cardID(90))
	printed := xTrick(cardID(90), "Tyvar's Stand", "{X}{G}",
		&protocol.PumpView{Power: 3, Toughness: 3, Keywords: []string{"hexproof", "indestructible"}}, false, false)
	noAmount := xTrick(cardID(90), "Tyvar's Stand", "{X}{G}",
		&protocol.PumpView{Keywords: []string{"hexproof", "indestructible"}}, false, false)
	at := func(x int) legal.Move { return castXMove(t, cardID(90), x, cardID(3)) }

	x3 := xRanks(t, xPumps(), attackBoard(ts, 40), passMove(0), at(3))[1]
	want := xRanks(t, xPumps(), attackBoard(printed, 40), passMove(0), at(3))[1]
	if x3 != want {
		t.Errorf("Tyvar's Stand at X=3 is priced %.3f, want %.3f, the price of a printed +3/+3", x3, want)
	}
	if x1 := xRanks(t, xPumps(), attackBoard(ts, 40), passMove(0), at(1))[1]; x1 >= x3 {
		t.Errorf("X=1 is priced %.3f, not below X=3's %.3f: the X is not read", x1, x3)
	}

	off := xRanks(t, noXPumps(), attackBoard(ts, 40), passMove(0), at(3))[1]
	old := xRanks(t, noXPumps(), attackBoard(noAmount, 40), passMove(0), at(3))[1]
	if off != old {
		t.Errorf("with PriceXPumps off X=3 is priced %.3f, want %.3f, a benefit with no amount", off, old)
	}
	if off >= x3 {
		t.Errorf("with PriceXPumps off X=3 is priced %.3f, not below %.3f with it on", off, x3)
	}
}

// TestTyvarsStandForTheKill: the defender is on 4 with no untapped
// blocker. A 2/2 attacker needs +2 to kill, so X=2 is the kill and the bot
// casts it; with the price off the kill is not seen.
func TestTyvarsStandForTheKill(t *testing.T) {
	ts := tyvarsStand(cardID(90))
	v := newView([]protocol.PlayerView{newSeat(0, withHand(ts)), newSeat(1, withLife(4))},
		withBattlefield(land(cardID(91), 0), land(cardID(92), 0), land(cardID(93), 0),
			creature(cardID(3), 0, "Attacker", 2, 2, attacking(1)),
			creature(cardID(2), 1, "Tapped Bear", 3, 3, tapped())),
		withTurn(5, 0, "declare_attackers"))
	moves := []legal.Move{passMove(0), castXMove(t, ts.InstanceID, 2, cardID(3))}
	on := xRanks(t, xPumps(), v, moves...)[1]
	off := xRanks(t, noXPumps(), v, moves...)[1]
	if on-off < heuristic.DefaultConfig().LethalBonus {
		t.Errorf("X=2 for the kill is priced %.3f on and %.3f off; want LethalBonus between them", on, off)
	}
	if i, why := decideOn(t, xPumps(), v, moves...); i != 1 {
		t.Errorf("picked %d (%s), want Tyvar's Stand for the kill", i, why)
	}
}

// TestHatredIsChargedTheLifeItsXPays: Hatred's X is paid in life and is
// the +X/+0, and the move's declared cost does not carry that life. The
// enumerator offers the largest X the life allows. At 20 life, X=19 on
// an unblocked attacker, with the defender on 40, is a swing that does
// not kill for nineteen life, and the bot passes; the same
// Hatred at X=4 for the kill on a defender at 6 is cast.
func TestHatredIsChargedTheLifeItsXPays(t *testing.T) {
	h := hatred(cardID(90))
	mine := func(life, theirs int) protocol.GameView {
		return newView([]protocol.PlayerView{newSeat(0, withHand(h), withLife(life)), newSeat(1, withLife(theirs))},
			withBattlefield(land(cardID(91), 0), land(cardID(92), 0), land(cardID(93), 0), land(cardID(94), 0), land(cardID(95), 0),
				creature(cardID(3), 0, "Attacker", 2, 2, attacking(1)),
				creature(cardID(2), 1, "Their Bear", 3, 3, tapped())),
			withTurn(5, 0, "declare_attackers"))
	}
	big := []legal.Move{passMove(0), castXMove(t, h.InstanceID, 19, cardID(3))}
	if i, why := decideOn(t, xPumps(), mine(20, 40), big...); i != 0 {
		t.Errorf("paid 19 of 20 life for a swing that does not kill: %s", why)
	}
	on := xRanks(t, xPumps(), mine(20, 40), big...)[1]
	if on >= 0 {
		t.Errorf("Hatred at X=19 from 20 life is priced %.3f, want below passing", on)
	}
	kill := []legal.Move{passMove(0), castXMove(t, h.InstanceID, 4, cardID(3))}
	if i, why := decideOn(t, xPumps(), mine(20, 6), kill...); i != 1 {
		t.Errorf("picked %d (%s), want Hatred at X=4 for the kill", i, why)
	}
	// The charge is the life of the X the move names, and only under the
	// switch: priced at X=4, it is the printed +4/+0 less that life.
	printed := xTrick(cardID(90), "Hatred", "{3}{B}{B}", &protocol.PumpView{Power: 4}, false, false)
	if got, plain := xRanks(t, xPumps(), mine(20, 6), kill...)[1],
		xRanks(t, xPumps(), newView([]protocol.PlayerView{newSeat(0, withHand(printed), withLife(20)), newSeat(1, withLife(6))},
			withBattlefield(land(cardID(91), 0), land(cardID(92), 0), land(cardID(93), 0), land(cardID(94), 0), land(cardID(95), 0),
				creature(cardID(3), 0, "Attacker", 2, 2, attacking(1)),
				creature(cardID(2), 1, "Their Bear", 3, 3, tapped())),
			withTurn(5, 0, "declare_attackers")), kill...)[1]; got >= plain {
		t.Errorf("Hatred at X=4 is priced %.3f, not below the %.3f of a +4/+0 that costs no life", got, plain)
	}
}

// TestHatredOffKeepsTodaysPrice: with PriceXPumps off, Hatred's X is
// neither read nor charged, as before #4976.
func TestHatredOffKeepsTodaysPrice(t *testing.T) {
	h := hatred(cardID(90))
	noAmount := xTrick(cardID(90), "Hatred", "{3}{B}{B}", nil, false, false)
	noAmount.AdditionalCost = h.AdditionalCost
	at := func(c protocol.CardView) float64 {
		v := attackBoard(c, 40)
		return xRanks(t, noXPumps(), v, passMove(0), castXMove(t, cardID(90), 3, cardID(3)))[1]
	}
	if got, want := at(h), at(noAmount); got != want {
		t.Errorf("with PriceXPumps off Hatred at X=3 is priced %.3f, want %.3f", got, want)
	}
}
