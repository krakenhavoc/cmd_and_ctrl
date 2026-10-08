package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// land_swap_test.go — #2469's re-price (land_swap.go): a land swap is
// priced by what it nets. Late in a game, with nothing in hand the bot
// cannot cast, Harrow nets one land: SpellFloor under that, the
// sacrificed land given back untapped, less the card. Priced gross it
// was two lands of ramp, floored before the sacrifice was charged, and
// came to about −0.20, below the Rampant Growth beside it.

func lateHarrowTable(t *testing.T, sacrificeTapped bool) aiseat.Input {
	t.Helper()
	h := harrow(cardID(1), 2)
	rg := spell(cardID(2), 2, "Rampant Growth", "{1}{G}")
	rg.TypeLine = "Sorcery"
	rg.Purpose = &protocol.PurposeView{Lands: 1}
	lands := manaLands(7, 2, 100)
	if sacrificeTapped {
		lands[0].Tapped = true
	}
	return input(2, newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withHand(h, rg)), newSeat(3)},
		withBattlefield(lands...), withTurn(9, 2, "postcombat_main")),
		passMove(2), sacrificeCast(t, 2, h, lands[0].InstanceID), castMove(t, 2, rg.InstanceID, "Cast Rampant Growth"))
}

func TestALateHarrowIsPricedByTheLandItNets(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	w := cfg.Weights
	for name, tc := range map[string]struct {
		tapped bool
		want   float64
	}{
		// The floor, plus the replacement land, less the untapped land
		// sacrificed, less the card.
		"untapped land": {false, cfg.SpellFloor + w.ManaSource - w.ManaSource - w.Hand},
		// A tapped land comes back untapped.
		"tapped land": {true, cfg.SpellFloor + w.ManaSource - w.TappedManaSource - w.Hand},
	} {
		in := lateHarrowTable(t, tc.tapped)
		got := rankValue(t, heuristic.New(), in, "Cast Harrow")
		if !nearly(got, tc.want) {
			t.Errorf("%s: Harrow priced %+.2f, want %+.2f", name, got, tc.want)
		}
		if rg := rankValue(t, heuristic.New(), in, "Cast Rampant Growth"); got < rg-1e-9 {
			t.Errorf("%s: Harrow %+.2f below Rampant Growth %+.2f: it nets the same land, untapped, at instant speed", name, got, rg)
		}
		if got <= cfg.LeftoverThreshold {
			t.Errorf("%s: Harrow priced %+.2f, want above LeftoverThreshold", name, got)
		}
	}
}

// In the end step before its turn, with nothing left to ramp toward,
// the bot casts Harrow. Priced gross (NetLandSwaps off) it passed.
func TestTheBotCastsALateHarrowInTheEndStepBeforeItsTurn(t *testing.T) {
	h := harrow(cardID(1), 2)
	lands := manaLands(7, 2, 100)
	in := input(2, newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withHand(h)), newSeat(3)},
		withBattlefield(lands...), withTurn(9, 1, "end")),
		passMove(2), sacrificeCast(t, 2, h, lands[0].InstanceID))
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Harrow" {
		t.Errorf("chose %q, want Harrow: it nets a land at a price above LeftoverThreshold (#2469)", got)
	}
	gross := heuristic.DefaultConfig()
	gross.NetLandSwaps = false
	if got := chose(t, in, decide(t, heuristic.NewWithConfig(gross), in)); got != "Pass priority" {
		t.Errorf("NetLandSwaps off chose %q, want the pass (the gross price is below zero)", got)
	}
}

// The ramp premium is paid on the lands the swap ADDS, not on the ones
// it replaces: early, with a deficit of three, Harrow nets one more
// source a turn, as a Rampant Growth does.
func TestHarrowsRampPremiumIsOnTheLandItAdds(t *testing.T) {
	h := harrow(cardID(1), 2)
	lands := manaLands(2, 2, 100)
	in := input(2, newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withHand(h, rampFiveDrop(cardID(9), 2))), newSeat(3)},
		withBattlefield(lands...), withTurn(3, 2, "precombat_main")),
		passMove(2), sacrificeCast(t, 2, h, lands[0].InstanceID))
	cfg := heuristic.DefaultConfig()
	w := cfg.Weights
	// One net land, ManaSource plus one RampPerMana; the replacement at
	// ManaSource; less the sacrificed land and the card.
	want := w.ManaSource + cfg.RampPerMana + w.ManaSource - w.ManaSource - w.Hand
	if got := rankValue(t, heuristic.New(), in, "Cast Harrow"); !nearly(got, want) {
		t.Errorf("Harrow priced %+.2f, want %+.2f", got, want)
	}
}

// A swap the purpose does not more than replace, or a sacrifice that is
// not a land, is priced as before: NetLandSwaps does not move it.
func TestOnlyANetLandSwapIsRepriced(t *testing.T) {
	gross := heuristic.DefaultConfig()
	gross.NetLandSwaps = false
	for name, p := range map[string]*protocol.PurposeView{
		"one for one": {Lands: 1, Draws: 1},
		"no lands":    {Draws: 2},
	} {
		c := harrow(cardID(1), 2)
		c.Name, c.Purpose = "Look-alike", p
		in, label := harrowTable(t, c)
		if a, b := rankValue(t, heuristic.New(), in, label), rankValue(t, heuristic.NewWithConfig(gross), in, label); !nearly(a, b) {
			t.Errorf("%s: priced %+.2f, %+.2f with NetLandSwaps off; want no change", name, a, b)
		}
	}
	c := harrow(cardID(1), 2)
	imp := creature(cardID(30), 2, "Imp", 0, 1)
	lands := manaLands(4, 2, 100)
	in := input(2, newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2, withHand(c, rampFiveDrop(cardID(9), 2))), newSeat(3)},
		withBattlefield(append(lands, imp)...), withTurn(5, 1, "end")),
		passMove(2), sacrificeCast(t, 2, c, imp.InstanceID))
	if a, b := rankValue(t, heuristic.New(), in, "Cast Harrow"), rankValue(t, heuristic.NewWithConfig(gross), in, "Cast Harrow"); !nearly(a, b) {
		t.Errorf("sacrificed creature: priced %+.2f, %+.2f with NetLandSwaps off; want no change", a, b)
	}
}
