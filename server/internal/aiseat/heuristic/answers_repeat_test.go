package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// answers_repeat_test.go is #2939 gap 2: a pump row that needs several
// activations to keep a creature alive.

// shade is a 2/2 whose only row is "{R}: +1/+1 until end of turn", the
// pump that needs two activations to live through a Lightning Bolt.
func shade() protocol.CardView {
	c := creature(cardID(1), 0, "Shade", 2, 2)
	a := []string{"pump"}
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Label: "{R}: This creature gets +1/+1 until end of turn.", ManaCost: "{R}",
		Purpose: &protocol.PurposeView{Answers: &a, Pump: &protocol.PumpView{Power: 1, Toughness: 1}},
	}}
	return c
}

// noRepeatedPumps is the heuristic with only #2939 gap 2's switch off.
func noRepeatedPumps() *heuristic.Policy {
	c := heuristic.DefaultConfig()
	c.PriceRepeatedPumps = false
	return heuristic.NewWithConfig(c)
}

// A Bolt for 3 at a 2/2 Shade takes two {R} pumps. The bot pumps when
// its mana pays both, counts a pump already on the stack as one of them,
// and neither starts a pump it cannot finish nor stacks a pump past what
// the Bolt needs. Before, one activation that did not cover the damage
// was never priced, so the Shade died with its mana up.
func TestPumpNeedingTwoActivations(t *testing.T) {
	b, bi := bolt(1, cardID(9), cardID(1), 3)
	pumpOnStack := protocol.StackItemView{ID: "pump", Kind: "ability", Controller: seatID(0).String(), SourceCardID: cardID(1)}
	for _, tc := range []struct {
		name  string
		lands int
		onTop int
		want  string
	}{
		{"two lands pay both pumps", 2, 0, "Pump"},
		{"one land cannot pay both", 1, 0, "Pass priority"},
		{"one pump up and one land left", 1, 1, "Pump"},
		{"two pumps already cover it", 2, 2, "Pass priority"},
		{"four lands, one pump up", 4, 1, "Pump"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			items := make([]protocol.StackItemView, tc.onTop)
			for i := range items {
				items[i] = pumpOnStack
			}
			cards := append([]protocol.CardView{shade()}, manaLands(tc.lands, 0, 100)...)
			in := input(0, respond(cards, b, bi, items...), passMove(0), activate(t, 0, cardID(1), "Pump", nil))
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != tc.want {
				t.Errorf("the heuristic chose %q, want %q", got, tc.want)
			}
			// Switched off, the bot prices only a pump that covers the
			// damage in one activation, which this one never does.
			if got := chose(t, in, decide(t, noRepeatedPumps(), in)); got != "Pass priority" {
				t.Errorf("without PriceRepeatedPumps it chose %q, want a pass", got)
			}
			if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
				t.Errorf("the baseline chose %q, want a pass", got)
			}
		})
	}
}

// A pump that takes three activations costs a flat ActivateBase for each
// one after the first, so it is priced below one that takes a single
// activation, and a pump the mana cannot reach is not priced at all.
func TestRepeatedPumpPriceFallsWithTheSteps(t *testing.T) {
	pol := heuristic.New()
	price := func(damage, lands int) float64 {
		b, bi := bolt(1, cardID(9), cardID(1), damage)
		cards := append([]protocol.CardView{shade()}, manaLands(lands, 0, 100)...)
		return rankValue(t, pol, input(0, respond(cards, b, bi), passMove(0), activate(t, 0, cardID(1), "Pump", nil)), "Pump")
	}
	one, two, three := price(2, 4), price(3, 4), price(4, 4)
	if !nearly(one-two, 0.5) || !nearly(two-three, 0.5) {
		t.Errorf("1, 2 and 3 activations priced %.3f, %.3f, %.3f; want each 0.5 (ActivateBase) below the last", one, two, three)
	}
	if out := price(4, 2); out >= three {
		t.Errorf("a three-pump save with two lands priced %.3f, want below the reachable %.3f", out, three)
	}
}
