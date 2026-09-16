package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// counter_cost_test.go — #625: the counter-removal component of
// legal.MoveCost reaching the policy. The failure it prevents is the
// #74 shape one cost component over: Heart of Kiran's alternative crew
// looks free from its params, so a bot with a 1-loyalty planeswalker
// would crew a Vehicle it has no plan for by killing the walker.

func walkerView(id string, controller, loyalty int) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "Planeswalker", TypeLine: "Legendary Planeswalker — Test",
		Owner: seatID(controller).String(), Controller: seatID(controller).String(), KnownByYou: true,
		Counters: map[string]int{"loyalty": loyalty},
	}
}

func heartView(id string, controller int) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "Heart of Kiran", TypeLine: "Legendary Artifact — Vehicle",
		Owner: seatID(controller).String(), Controller: seatID(controller).String(), KnownByYou: true,
		Power: 4, Toughness: 4,
	}
}

func counterCost(card, kind string, n int) *legal.MoveCost {
	return &legal.MoveCost{Counters: []legal.CounterPrice{{CardID: uuid.MustParse(card), Counter: kind, N: n}}}
}

const heartCrew = "Heart of Kiran: Crew — remove a loyalty counter from a planeswalker you control"

// The refusal: a low-payoff crew is not worth a planeswalker's last
// loyalty counter.
func TestBotWillNotKillAOneLoyaltyWalkerToCrew(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(heartView(cardID(1), 0), walkerView(cardID(2), 0, 1), land(cardID(10), 0)))
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), heartCrew, counterCost(cardID(2), "loyalty", 1)),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == heartCrew {
		t.Fatal("the bot crewed Heart of Kiran by killing its 1-loyalty planeswalker")
	}
	for _, c := range heuristic.New().Rank(context.Background(), in) {
		if c.Index == 1 && c.Value >= 0 {
			t.Errorf("the walker-killing crew is valued %.2f, want below passing", c.Value)
		}
	}

	// The control, so the refusal is the price and not a bot that never
	// crews: the same move with the cost unpriced is taken.
	free := input(0, v, passMove(0), activateMove(t, 0, cardID(1), heartCrew, nil))
	if got := chose(t, free, decide(t, heuristic.New(), free)); got != heartCrew {
		t.Errorf("control: the bot does not take the same crew for free (chose %q), so the refusal above proves nothing about the price", got)
	}
}

// And it is a price, not a ban: the same crew paid from a walker with
// loyalty to spare ranks above the one that kills a walker.
func TestCounterCostIsPricedAgainstThePermanentItComesFrom(t *testing.T) {
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(heartView(cardID(1), 0), walkerView(cardID(2), 0, 1), walkerView(cardID(3), 0, 6), land(cardID(10), 0)))
	fromSmall := heartCrew + " (from the 1-loyalty walker)"
	fromBig := heartCrew + " (from the 6-loyalty walker)"
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), fromSmall, counterCost(cardID(2), "loyalty", 1)),
		activateMove(t, 0, cardID(1), fromBig, counterCost(cardID(3), "loyalty", 1)),
	)
	ranked := heuristic.New().Rank(context.Background(), in)
	pos := map[int]int{}
	val := map[int]float64{}
	for i, c := range ranked {
		pos[c.Index] = i
		val[c.Index] = c.Value
	}
	if pos[2] > pos[1] {
		t.Errorf("ranked the walker-killing payment (%.2f) above the cheap one (%.2f)", val[1], val[2])
	}
	if got := chose(t, in, decide(t, heuristic.New(), in)); got == fromSmall {
		t.Error("the bot chose to pay from the 1-loyalty walker with a 6-loyalty one on offer")
	}
}

// A counter the policy has no weight for (Dragon's Hoard's gold) is
// cheap but not free, and removing a -1/-1 counter is a gain rather
// than a cost.
func TestCounterKindsArePricedInTheEvaluationsUnits(t *testing.T) {
	hoard := protocol.CardView{
		InstanceID: cardID(1), Name: "Dragon's Hoard", TypeLine: "Artifact",
		Owner: seatID(0).String(), Controller: seatID(0).String(), KnownByYou: true,
		Counters: map[string]int{"gold": 2},
	}
	shrunk := creature(cardID(2), 0, "Shrunk Bear", 1, 1)
	shrunk.Counters = map[string]int{"-1/-1": 1}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(hoard, shrunk, land(cardID(10), 0)))
	const gold, minus, free = "Hoard: spend gold", "Broker: remove the -1/-1", "free activation"
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), gold, counterCost(cardID(1), "gold", 1)),
		activateMove(t, 0, cardID(2), minus, counterCost(cardID(2), "-1/-1", 1)),
		activateMove(t, 0, cardID(1), free, nil),
	)
	val := map[int]float64{}
	for _, c := range heuristic.New().Rank(context.Background(), in) {
		val[c.Index] = c.Value
	}
	if !(val[1] < val[3]) {
		t.Errorf("spending a gold counter (%.2f) priced as free (%.2f)", val[1], val[3])
	}
	if !(val[2] > val[3]) {
		t.Errorf("removing a -1/-1 counter (%.2f) should be worth more than a free activation (%.2f)", val[2], val[3])
	}
}
