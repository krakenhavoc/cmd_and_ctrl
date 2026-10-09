package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// holdinstants_test.go is ADR 0136 §5 (PR 5, #2668): in a turn plan an
// instant-speed member is held for the end step before the bot's turn
// unless a later member needs its mana, or it draws with mana left after
// it. Each test that holds fails with Config.PlanHoldInstants off, which
// casts instants like sorceries and leaves Harrow out of the plan.

// greenHarrow is Harrow with the lands it puts onto the battlefield
// entering untapped, as the catalog declares it (ADR 0136 owner
// answer 4).
func greenHarrow(id string) protocol.CardView {
	h := harrow(id, 0)
	h.Purpose = &protocol.PurposeView{Lands: 2, LandsUntapped: 2}
	return h
}

func forests(n, from int) []protocol.CardView {
	out := make([]protocol.CardView, n)
	for i := range out {
		out[i] = planForest(cardID(from + i))
	}
	return out
}

func greenCreature(id, name, cost string, power, tough int, opts ...cardOpt) protocol.CardView {
	c := creature(id, 0, name, power, tough, opts...)
	c.ManaCost = cost
	return c
}

func decideHold(t *testing.T, cfg heuristic.Config, in aiseat.Input) (aiseat.Decision, []aiseat.PlanMember) {
	t.Helper()
	d, tr, err := heuristic.NewWithConfig(cfg).DecideTraced(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	return d, tr.Plan
}

func holdOff() heuristic.Config {
	cfg := heuristic.DefaultConfig()
	cfg.PlanHoldInstants = false
	return cfg
}

// #2668's main-phase case: five lands, Harrow and a two-drop. Both fit
// this turn, and nothing this turn needs Harrow's lands, so the two-drop
// is cast now and Harrow is kept for the end step before the bot's turn,
// its three mana reserved.
func TestPlanHoldsHarrowForTheEndStep(t *testing.T) {
	h := greenHarrow(cardID(1))
	bear := greenCreature(cardID(2), "Bear", "{1}{G}", 2, 2)
	lands := forests(5, 100)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(h, bear, rampFiveDrop(cardID(9), 0))), newSeat(1)},
		withTurn(5, 0, "precombat_main"), withBattlefield(lands...))
	in := input(0, v,
		passMove(0),
		sacrificeCast(t, 0, h, lands[0].InstanceID),
		stampedCast(t, bear.InstanceID, "Cast Bear", "{1}{G}"),
	)

	d, plan := decideHold(t, heuristic.DefaultConfig(), in)
	if got := moveLabel(in, d); got != "Cast Bear" {
		t.Errorf("chose %q (%s), want the two-drop now and Harrow held", got, d.Reason)
	}
	if len(plan) != 2 || plan[0].Label != "Cast Bear" || plan[0].Held || plan[1].Label != "Cast Harrow" || !plan[1].Held {
		t.Errorf("trace plan = %+v, want the Bear, then Harrow held", plan)
	}
	if !strings.Contains(d.Reason, "Harrow at the end step") {
		t.Errorf("reason = %q, want Harrow named as held", d.Reason)
	}

	if _, plan := decideHold(t, holdOff(), in); len(plan) != 0 {
		t.Errorf("with PlanHoldInstants off the trace carries a plan %+v: Harrow is not a member", plan)
	}

	// Four lands: Harrow's lands pay for the two-drop, so Harrow is cast
	// now, first, as a mana member (§4 item 1).
	v.Battlefield.Cards = lands[:4]
	in.View = v
	d, plan = decideHold(t, heuristic.DefaultConfig(), in)
	if got := moveLabel(in, d); got != "Cast Harrow" {
		t.Errorf("four lands: chose %q (%s), want Harrow now for the two-drop", got, d.Reason)
	}
	for _, m := range plan {
		if m.Held {
			t.Errorf("four lands: trace plan %+v holds a member a later one needs", plan)
		}
	}
}

// When every member is held the bot passes, its mana kept up: Harrow and
// a second ramp instant, six lands, nothing cast at sorcery speed. With
// the switch off the second one is cast in the main phase.
func TestPlanPassesWhenEveryMemberIsHeld(t *testing.T) {
	h := greenHarrow(cardID(1))
	grow := spell(cardID(2), 0, "Grow Wild", "{2}{G}")
	grow.Purpose = &protocol.PurposeView{Lands: 2, LandsUntapped: 2}
	lands := forests(7, 100)
	lands[6].Tapped = true
	v := newView([]protocol.PlayerView{newSeat(0, withHand(h, grow, rampFiveDrop(cardID(9), 0))), newSeat(1)},
		withTurn(5, 0, "precombat_main"), withBattlefield(lands...))
	in := input(0, v,
		passMove(0),
		sacrificeCast(t, 0, h, lands[6].InstanceID),
		stampedCast(t, grow.InstanceID, "Cast Grow Wild", "{2}{G}"),
	)

	d, plan := decideHold(t, heuristic.DefaultConfig(), in)
	if got := moveLabel(in, d); got != "Pass priority" {
		t.Errorf("chose %q (%s), want the pass: both members wait for the end step", got, d.Reason)
	}
	if len(plan) != 2 || !plan[0].Held || !plan[1].Held {
		t.Errorf("trace plan = %+v, want both members held", plan)
	}
	if d, _ := decideHold(t, holdOff(), in); moveLabel(in, d) == "Pass priority" {
		t.Errorf("with PlanHoldInstants off the bot passes (%s); this board no longer shows the rule", d.Reason)
	}
}

// Only a ramp or draw instant is held (the owner's decision of
// 2026-10-09, question 2): a flash creature beside Harrow is cast now,
// and Harrow waits.
func TestPlanHoldsOnlyRampAndDrawInstants(t *testing.T) {
	h := greenHarrow(cardID(1))
	ambusher := greenCreature(cardID(2), "Ambusher", "{2}{G}", 3, 3, keywords("flash"))
	lands := forests(6, 100)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(h, ambusher, rampFiveDrop(cardID(9), 0))), newSeat(1)},
		withTurn(5, 0, "precombat_main"), withBattlefield(lands...))
	in := input(0, v,
		passMove(0),
		sacrificeCast(t, 0, h, lands[0].InstanceID),
		stampedCast(t, ambusher.InstanceID, "Cast Ambusher", "{2}{G}"),
	)
	d, plan := decideHold(t, heuristic.DefaultConfig(), in)
	if got := moveLabel(in, d); got != "Cast Ambusher" {
		t.Errorf("chose %q (%s), want the flash creature now", got, d.Reason)
	}
	if len(plan) != 2 || plan[0].Held || plan[0].Label != "Cast Ambusher" || !plan[1].Held {
		t.Errorf("trace plan = %+v, want the Ambusher now and Harrow held", plan)
	}
}

// What remains of this turn's plan stays held (the owner's decision of
// 2026-10-09, question 1). The plan casts the Bear and holds Harrow; in
// the next window Harrow is alone, a plan of one, and the policy that
// chose the plan passes. A policy that chose no plan this turn casts it,
// as the gated position harrow-sacrifices-the-tapped-land asks.
func TestHeldRemainderStaysHeldThisTurn(t *testing.T) {
	h := greenHarrow(cardID(1))
	bear := greenCreature(cardID(2), "Bear", "{1}{G}", 2, 2)
	lands := forests(5, 100)
	big := rampFiveDrop(cardID(9), 0)
	first := input(0, newView([]protocol.PlayerView{newSeat(0, withHand(h, bear, big)), newSeat(1)},
		withTurn(5, 0, "precombat_main"), withBattlefield(lands...)),
		passMove(0),
		sacrificeCast(t, 0, h, lands[0].InstanceID),
		stampedCast(t, bear.InstanceID, "Cast Bear", "{1}{G}"),
	)
	after := append([]protocol.CardView(nil), lands...)
	after[3].Tapped, after[4].Tapped = true, true
	cast := bear
	cast.SummoningSick = true
	after = append(after, cast)
	second := input(0, newView([]protocol.PlayerView{newSeat(0, withHand(h, big)), newSeat(1)},
		withTurn(5, 0, "precombat_main"), withBattlefield(after...)),
		passMove(0),
		sacrificeCast(t, 0, h, lands[3].InstanceID),
	)

	for name, cfg := range map[string]heuristic.Config{"on": heuristic.DefaultConfig(), "off": holdOff()} {
		pol := heuristic.NewWithConfig(cfg)
		if _, _, err := pol.DecideTraced(context.Background(), first); err != nil {
			t.Fatal(err)
		}
		d, tr, err := pol.DecideTraced(context.Background(), second)
		if err != nil {
			t.Fatal(err)
		}
		got := moveLabel(second, d)
		switch name {
		case "on":
			if got != "Pass priority" {
				t.Errorf("on: chose %q (%s), want Harrow kept for the end step", got, d.Reason)
			}
			if len(tr.Plan) != 1 || !tr.Plan[0].Held || tr.Plan[0].Label != "Cast Harrow" {
				t.Errorf("on: trace plan = %+v, want Harrow marked held", tr.Plan)
			}
		case "off":
			if got != "Cast Harrow" {
				t.Errorf("off: chose %q (%s), want Harrow", got, d.Reason)
			}
		}
	}
	// A fresh policy has no plan from this turn: a lone Harrow is cast.
	if got := moveLabel(second, decide(t, heuristic.New(), second)); got != "Cast Harrow" {
		t.Errorf("fresh policy chose %q, want Harrow: a plan of one is today's choice", got)
	}
}

// A draw spell with mana left after it is cast now, before the rest, so
// the card it finds can still be cast this turn (§4, §5): the switch
// changes nothing here.
func TestPlanCastsAnInstantDrawNowWithManaLeft(t *testing.T) {
	draw := spell(cardID(1), 0, "Think Again", "{1}{U}")
	draw.Purpose = &protocol.PurposeView{Draws: 2}
	bear := greenCreature(cardID(2), "Bear", "{1}{G}", 2, 2)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(draw, bear)), newSeat(1)},
		withTurn(5, 0, "precombat_main"),
		withBattlefield(planIsland(cardID(100)), planIsland(cardID(101)), planForest(cardID(102)), planForest(cardID(103)), planForest(cardID(104))))
	in := input(0, v,
		passMove(0),
		stampedCast(t, draw.InstanceID, "Cast Think Again", "{1}{U}"),
		stampedCast(t, bear.InstanceID, "Cast Bear", "{1}{G}"),
	)
	for name, cfg := range map[string]heuristic.Config{"on": heuristic.DefaultConfig(), "off": holdOff()} {
		d, plan := decideHold(t, cfg, in)
		if got := moveLabel(in, d); got != "Cast Think Again" {
			t.Errorf("%s: chose %q (%s), want the draw first", name, got, d.Reason)
		}
		for _, m := range plan {
			if m.Held {
				t.Errorf("%s: trace plan %+v holds the draw", name, plan)
			}
		}
	}
}
