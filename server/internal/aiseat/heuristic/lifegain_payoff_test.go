package heuristic_test

import (
	"context"
	"math"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// lifegain_payoff_test.go — #4979: Heliod, Sun-Crowned puts a +1/+1
// counter on a creature whenever its controller gains life, and its own
// lifelink grant on an attacker was priced by the life alone. The
// whole-engine checks that the grant still waits for combat and for
// spare mana are aiseat/heliod_grant_bot_test.go's.

// withLifeGainPayoff gives a permanent a "whenever you gain life" row
// with that payoff.
func withLifeGainPayoff(c protocol.CardView, l protocol.LifeGainPayoffView) protocol.CardView {
	c.AbilityRows = append(c.AbilityRows, protocol.AbilityRowView{
		Kind: "triggered", Label: "whenever you gain life",
		Purpose: &protocol.PurposeView{LifeGainPayoff: &l},
	})
	return c
}

// heliodGranter is lifelinkGranter with Heliod's lifegain trigger.
func heliodGranter(id string) protocol.CardView {
	return withLifeGainPayoff(lifelinkGranter(id), protocol.LifeGainPayoffView{Counters: 1})
}

// grantPrice ranks the lifelink grant on the bot's 2/2 in `step` of its
// own turn, attacking an open board when the step is in combat, with
// `src` as the granter and `others` on the battlefield.
func grantPrice(t *testing.T, cfg heuristic.Config, step string, src protocol.CardView, others ...protocol.CardView) float64 {
	t.Helper()
	attacker := cardID(3)
	var opts []cardOpt
	if step != "precombat_main" {
		opts = append(opts, attacking(1))
	}
	board := append(twoPlains(), src, creature(attacker, 0, "Attacker", 2, 2, opts...))
	board = append(board, others...)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, step))
	moves := []legal.Move{passMove(0), grantMove(t, src.InstanceID, attacker)}
	for _, c := range heuristic.NewWithConfig(cfg).Rank(context.Background(), input(0, v, moves...)) {
		if c.Index == 1 {
			return c.Value
		}
	}
	t.Fatal("the grant was not ranked")
	return 0
}

// TestHeliodsLifelinkGrantIsPricedWithItsCounter: the grant on the bot's
// own attacker, in its declare blockers step with spare mana, is worth
// the +1/+1 counter its life-gain event puts on a creature
// (Weights.Power + Weights.Toughness) beyond the life, once per event
// and per payoff (CR 119.9). An opponent's payoff pays nothing; before
// attacks the grant is priced as before, so it still waits for combat;
// and with the switch off the grant is priced by its life alone.
func TestHeliodsLifelinkGrantIsPricedWithItsCounter(t *testing.T) {
	w := heuristic.DefaultWeights()
	counter := w.Power + w.Toughness
	src := cardID(1)
	def := heuristic.DefaultConfig()
	plain := grantPrice(t, def, "declare_blockers", lifelinkGranter(src))
	if got := grantPrice(t, def, "declare_blockers", heliodGranter(src)) - plain; math.Abs(got-counter) > 1e-9 {
		t.Errorf("Heliod's grant is priced %.2f over a plain lifelink grant, want the counter's %.2f", got, counter)
	}
	pridemate := withLifeGainPayoff(creature(cardID(5), 0, "Ajani's Pridemate", 2, 2), protocol.LifeGainPayoffView{Counters: 1})
	if got := grantPrice(t, def, "declare_blockers", heliodGranter(src), pridemate) -
		grantPrice(t, def, "declare_blockers", lifelinkGranter(src), creature(cardID(5), 0, "Bear", 2, 2)); math.Abs(got-2*counter) > 1e-9 {
		t.Errorf("with a Pridemate out too the grant is priced %.2f over a plain one, want two counters' %.2f", got, 2*counter)
	}
	theirs := withLifeGainPayoff(creature(cardID(6), 1, "Opposing Pridemate", 2, 2, tapped()), protocol.LifeGainPayoffView{Counters: 1})
	base := grantPrice(t, def, "declare_blockers", lifelinkGranter(src), creature(cardID(6), 1, "Opposing Bear", 2, 2, tapped()))
	if got := grantPrice(t, def, "declare_blockers", lifelinkGranter(src), theirs); got != base {
		t.Errorf("an opponent's lifegain payoff moved the grant from %.2f to %.2f", base, got)
	}
	if a, b := grantPrice(t, def, "precombat_main", heliodGranter(src)), grantPrice(t, def, "precombat_main", lifelinkGranter(src)); a != b {
		t.Errorf("before attacks Heliod's grant is %.2f against a plain one's %.2f, want the same: it waits for combat", a, b)
	}
	off := def
	off.PriceLifeGainPayoffs = false
	if a, b := grantPrice(t, off, "declare_blockers", heliodGranter(src)), grantPrice(t, off, "declare_blockers", lifelinkGranter(src)); a != b {
		t.Errorf("with PriceLifeGainPayoffs off Heliod's grant is %.2f against %.2f", a, b)
	}
}

// TestHeliodsGrantBeatsAPumpForItsCounter: the bot's 2/2 attacking an
// open board, two Plains, and two {1}{W} rows on offer: Heliod's lifelink
// and another permanent's +1/+0. The extra point of damage outbids two
// life, so without Heliod's trigger the bot pumps; with it, lifelink is
// also a +1/+1 counter that stays, and the bot gives lifelink.
func TestHeliodsGrantBeatsAPumpForItsCounter(t *testing.T) {
	heliod, pumper, attacker := cardID(1), cardID(2), cardID(3)
	pump := granter(pumper, "{1}{W}: Target creature gets +1/+0 until end of turn.", "benefit", &protocol.PumpView{Power: 1})
	decide := func(cfg heuristic.Config) int {
		t.Helper()
		board := append(twoPlains(), heliodGranter(heliod), pump, creature(attacker, 0, "Attacker", 2, 2, attacking(1)))
		v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)}, withBattlefield(board...), withTurn(3, 0, "declare_blockers"))
		i, _ := decideOn(t, heuristic.NewWithConfig(cfg), v, passMove(0), grantMove(t, heliod, attacker), grantMove(t, pumper, attacker))
		return i
	}
	if i := decide(heuristic.DefaultConfig()); i != 1 {
		t.Errorf("picked %d, want Heliod's lifelink on the attacker for its counter", i)
	}
	off := heuristic.DefaultConfig()
	off.PriceLifeGainPayoffs = false
	if i := decide(off); i != 2 {
		t.Errorf("with PriceLifeGainPayoffs off picked %d, want the pump: the position is meant to turn on the counter", i)
	}
}
