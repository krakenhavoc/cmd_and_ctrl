package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func entryCounterBody(c protocol.CardView, kind string, amount int, x bool) protocol.CardView {
	entries := []protocol.EntryCounterView{{Kind: kind, Amount: amount, AmountIsX: x}}
	c.Purpose = &protocol.PurposeView{EntryCounters: &entries}
	return c
}

func entryCastPrice(t *testing.T, cfg heuristic.Config, c protocol.CardView, x int) float64 {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0, withHand(c)), newSeat(1)}, withTurn(5, 0, "precombat_main"))
	move := castMove(t, 0, c.InstanceID, "Cast a creature")
	move.Params = mustJSON(t, map[string]any{"instance_id": c.InstanceID, "from_zone": "hand", "x_value": x})
	in := input(0, v, passMove(0), move)
	for _, r := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		if r.Index == 1 {
			held := in.View.Seats[0].Hand.Cards[0]
			if held.Power != c.Power || held.Toughness != c.Toughness {
				t.Fatal("cast pricing mutated the input view")
			}
			return r.Value
		}
	}
	t.Fatal("cast was not ranked")
	return 0
}

func TestEntryCountersPriceTheResolvedCreatureBody(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	hydra := entryCounterBody(creature(cardID(1), 0, "Different name", 0, 0), "+1/+1", 4, false)
	hydra.ManaCost, hydra.Abilities = "{3}{G}{G}", []string{"trample"}
	vanilla := hydra
	vanilla.Power, vanilla.Toughness, vanilla.Purpose = 4, 4, nil
	price := entryCastPrice(t, cfg, hydra, 0)
	if want := entryCastPrice(t, cfg, vanilla, 0); price < want {
		t.Fatalf("Hydra %.2f below vanilla 4/4 trample %.2f", price, want)
	}
	off := cfg
	off.PriceEntryCounters = false
	if old := entryCastPrice(t, off, hydra, 0); old >= price {
		t.Fatalf("backout %.2f, enabled %.2f", old, price)
	}
	for _, tc := range []struct {
		cost     string
		keywords []string
		xs       []int
	}{
		{"{X}{W}", nil, []int{0, 1, 6, 11}},
		{"{X}{G}{U}", []string{"flying", "trample"}, []int{0, 2, 11}},
	} {
		c := entryCounterBody(creature(cardID(2), 0, "Arbitrary X creature", 0, 0), "+1/+1", 0, true)
		c.ManaCost, c.Abilities = tc.cost, tc.keywords
		last, old := -1e9, entryCastPrice(t, off, c, 0)
		for _, x := range tc.xs {
			got := entryCastPrice(t, cfg, c, x)
			if got <= last {
				t.Fatalf("%s: X=%d price %.2f did not exceed %.2f", tc.cost, x, got, last)
			}
			last = got
			if legacy := entryCastPrice(t, off, c, x); legacy != old {
				t.Fatalf("backout depends on X: %.2f vs %.2f", legacy, old)
			}
		}
		if got := entryCastPrice(t, cfg, c, 0); got != old {
			t.Fatalf("X=0 invented body value: %.2f vs %.2f", got, old)
		}
	}
	if !cfg.PriceEntryCounters || heuristic.BaselineConfig().PriceEntryCounters {
		t.Fatal("entry-counter default/baseline switches disagree")
	}
}

func TestEntryCountersIgnoreNonBodyKindsAndDoNotRepriceBattlefield(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	c := creature(cardID(3), 0, "Counter body", 4, 4)
	c.ManaCost = "{3}"
	plain := entryCastPrice(t, cfg, c, 0)
	charge := entryCounterBody(c, "charge", 9, false)
	if got := entryCastPrice(t, cfg, charge, 0); got != plain {
		t.Fatalf("charge changed body price %.2f to %.2f", plain, got)
	}
	minus := entryCounterBody(c, "-1/-1", 2, false)
	vanilla := c
	vanilla.Power, vanilla.Toughness = 2, 2
	if got, want := entryCastPrice(t, cfg, minus, 0), entryCastPrice(t, cfg, vanilla, 0); got != want {
		t.Fatalf("minus counters %.2f, vanilla2/2 %.2f", got, want)
	}
	c = entryCounterBody(c, "+1/+1", 4, false)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(c))
	on := heuristic.NewWithConfig(cfg).CostFuelPrice(input(0, v))
	cfg.PriceEntryCounters = false
	off := heuristic.NewWithConfig(cfg).CostFuelPrice(input(0, v))
	candidate := legal.TargetCandidate{ID: mustUUID(t, c.InstanceID)}
	if on(candidate) != off(candidate) {
		t.Fatal("battlefield fuel counted entry counters again")
	}
}
