package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// waitmain_test.go is the owner's gate on #2457 (ADR 0126's amendment
// of 2026-10-12): before its first main phase, with an empty stack, the
// bot holds an instant for the main phase, where the same cast is still
// on offer beside every sorcery-speed use of the same mana.

func noWaitForMain() heuristic.Config {
	c := heuristic.DefaultConfig()
	c.WaitForMain = false
	return c
}

// Unexpected Windfall pitching a spare land, late, prices well above
// InstantThreshold. In the bot's own upkeep or draw step it waits; in
// its main phase it is cast.
func TestAnInstantWaitsForTheMainPhase(t *testing.T) {
	windfall, cast := windfallFixture(t)
	spare := land(cardID(2), 0)
	at := func(step string, opts ...func(*protocol.GameView)) func(*heuristic.Policy) string {
		return func(pol *heuristic.Policy) string {
			v := newView([]protocol.PlayerView{newSeat(0, withHand(windfall, spare)), newSeat(1)},
				withBattlefield(manaLands(7, 0, 100)...), withTurn(9, 0, step))
			for _, o := range opts {
				o(&v)
			}
			in := input(0, v, passMove(0), cast(spare.InstanceID))
			if v := rankValue(t, pol, in, in.Moves[1].Label); v <= heuristic.DefaultConfig().InstantThreshold {
				t.Fatalf("the cast prices %.2f, not above InstantThreshold: the test shows nothing", v)
			}
			d := decide(t, pol, in)
			if d.Index == 0 && step != "precombat_main" && pol.Config().WaitForMain && !strings.HasPrefix(d.Reason, "wait for the main phase") {
				t.Errorf("%s: passed with reason %q, want it to say it waits for the main phase", step, d.Reason)
			}
			return chose(t, in, d)
		}
	}
	const castIt, pass = "Cast Unexpected Windfall", "Pass priority"
	for _, tc := range []struct {
		name string
		run  func(*heuristic.Policy) string
		want string
		off  string
	}{
		{"own draw step", at("draw"), pass, castIt},
		{"own upkeep", at("upkeep"), pass, castIt},
		{"own main phase", at("precombat_main"), castIt, castIt},
		{"own draw step, something on the stack", at("draw", withStack(spell(cardID(9), 1, "Divination", "{2}{U}"))), castIt, castIt},
		{"an opponent's draw step", at("draw", func(v *protocol.GameView) { v.Turn.ActiveSeat, v.Turn.PriorityHolder = 1, 0 }), castIt, castIt},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.run(heuristic.New()); !strings.HasPrefix(got, tc.want) {
				t.Errorf("chose %q, want %q", got, tc.want)
			}
			if got := tc.run(heuristic.NewWithConfig(noWaitForMain())); !strings.HasPrefix(got, tc.off) {
				t.Errorf("with the gate off chose %q, want %q", got, tc.off)
			}
		})
	}
}

func TestBaselineDoesNotWaitForMain(t *testing.T) {
	if heuristic.BaselineConfig().WaitForMain {
		t.Error("BaselineConfig holds instants for the main phase; it is the pre-S66 heuristic")
	}
}
