package heuristic_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// ramp_curve_test.go is #5084: Burnished Hart's "{3}, Sacrifice Burnished
// Hart: Search your library for up to two basic land cards" was taken 8
// and 9 times in 16 thopter-battle games. Its two lands were priced by
// the hand's deficit alone, so a hand of cheap cards made them two mana
// against a 2/2. With PriceRampCurve on, the lands are ramp up to the
// mana a deck on curve wants two turns from now while the round is below
// RampWantCap, and the creature is not charged the flat tapped-blocker
// price its sacrifice has already paid. The switch ships off: in the
// arena the curve seats won no more (docs/bot-testing.md).

// hartBoard is seat 0's own precombat main phase in round `round`, with
// the Hart, `lands` untapped lands, the hand `hand` and an opponent with
// `theirs` on the battlefield: the Hart's activation and a pass.
func hartBoard(t *testing.T, round, lands int, hand []protocol.CardView, theirs ...protocol.CardView) aiseat.Input {
	t.Helper()
	bf := []protocol.CardView{burnishedHart(cardID(1), 0)}
	for i := range lands {
		bf = append(bf, land(cardID(40+i), 0))
	}
	bf = append(bf, theirs...)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(hand...)), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(round, 0, "precombat_main"))
	return input(0, v, passMove(0), hartMove(t, 0, cardID(1)))
}

// priceOf is move 1's price under cfg (the Hart's activation, or the
// cast on offer), and the pass's.
func priceOf(cfg heuristic.Config, in aiseat.Input) (move, pass float64) {
	vals := map[int]float64{}
	for _, c := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		vals[c.Index] = c.Value
	}
	return vals[1], vals[0]
}

// cheapHand is a thopter-battle hand the bot can already cast: a
// zero-drop, a three-drop and a four-drop, so the hand's deficit is
// closed from four lands on.
func cheapHand() []protocol.CardView {
	ornithopter := creature(cardID(70), 0, "Ornithopter", 0, 2, keywords("flying"))
	ornithopter.TypeLine, ornithopter.ManaCost = "Artifact Creature — Thopter", "{0}"
	drake := creature(cardID(71), 0, "Drake", 3, 3, keywords("flying"))
	drake.ManaCost = "{2}{U}"
	ogre := creature(cardID(72), 0, "Ogre", 4, 4)
	ogre.ManaCost = "{3}{U}"
	return []protocol.CardView{ornithopter, drake, ogre}
}

// curve is today's heuristic with the curve on, as the arena's
// `heuristic-rampcurve` seats it.
func curve() heuristic.Config {
	cfg := heuristic.DefaultConfig()
	cfg.PriceRampCurve = true
	return cfg
}

// noCurve is today's heuristic: the switch ships off.
func noCurve() heuristic.Config { return heuristic.DefaultConfig() }

// The issue's position: early in the game, with spare mana and a hand
// the bot can cast, the Hart is cracked for two lands with the curve
// on. With it off (the default, and the baseline) it is kept, as the
// arena saw.
func TestHartIsCrackedEarlyOnTheCurve(t *testing.T) {
	for _, tc := range []struct {
		name         string
		round, lands int
	}{
		{"round 3 with three lands", 3, 3},
		{"round 4 with four lands", 4, 4},
		{"round 5 with five lands", 5, 5},
		{"round 6 a land behind", 6, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := hartBoard(t, tc.round, tc.lands, cheapHand())
			cfg := curve()
			act, pass := priceOf(cfg, in)
			off, _ := priceOf(noCurve(), in)
			t.Logf("Hart: %.2f with the curve, %.2f without, pass %.2f", act, off, pass)
			if got := chose(t, in, decide(t, heuristic.NewWithConfig(curve()), in)); got != "Burnished Hart: "+hartLabel {
				t.Errorf("the heuristic chose %q (Hart %.2f, pass %.2f), want the Hart's activation", got, act, pass)
			}
			if off >= cfg.PassThreshold {
				t.Errorf("with the curve off the Hart prices %.2f, want it under PassThreshold as before #5084", off)
			}
			if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Pass priority" {
				t.Errorf("with the curve off the heuristic chose %q, want the pass it made before #5084", got)
			}
			if got := chose(t, in, decide(t, baseline(), in)); got != "Pass priority" {
				t.Errorf("the baseline chose %q, want a pass", got)
			}
		})
	}
}

// Late in the game, on a stalled board, the Hart is a blocker the bot
// keeps: from round RampWantCap on the curve adds nothing, so the lands
// are worth what the hand asks of them, and a hand it can cast asks
// nothing. That holds a land or two behind the curve, too.
func TestHartIsKeptLateOnAStalledBoard(t *testing.T) {
	wurm := func(i int) protocol.CardView { return creature(cardID(80+i), 1, "Wurm", 7, 7, keywords("trample")) }
	for _, tc := range []struct {
		name         string
		round, lands int
	}{
		{"round 7 with seven lands", 7, 7},
		{"round 10 with eight lands", 10, 8},
		{"round 9 with five lands", 9, 5},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in := hartBoard(t, tc.round, tc.lands, cheapHand(), wurm(0), wurm(1))
			act, pass := priceOf(curve(), in)
			t.Logf("Hart: %.2f, pass %.2f", act, pass)
			if got := chose(t, in, decide(t, heuristic.NewWithConfig(curve()), in)); got != "Pass priority" {
				t.Errorf("the heuristic chose %q (Hart %.2f), want the Hart kept", got, act)
			}
		})
	}
}

// The curve and the hand are not added up: a hand that already asks for
// two more mana is paid two, with the curve on or off. The only
// difference the switch makes there is the tapped-blocker charge it
// stops making twice.
func TestRampCurveTakesTheLargerDeficit(t *testing.T) {
	big := creature(cardID(73), 0, "Wurm", 7, 7, keywords("trample"))
	big.ManaCost = "{5}{U}"
	for _, round := range []int{4, 9} {
		in := hartBoard(t, round, 4, []protocol.CardView{big})
		on, _ := priceOf(curve(), in)
		off, _ := priceOf(noCurve(), in)
		if got, want := on-off, 0.30; got < want-1e-9 || got > want+1e-9 {
			t.Errorf("round %d, a six-drop in hand: the curve moves the Hart from %.2f to %.2f, want only the %.2f blocker charge", round, off, on, want)
		}
	}
	// Late, with no curve to add, the switch only drops that charge.
	in := hartBoard(t, 10, 8, cheapHand())
	on, _ := priceOf(curve(), in)
	off, _ := priceOf(noCurve(), in)
	if got, want := on-off, 0.30; got < want-1e-9 || got > want+1e-9 {
		t.Errorf("round 10: the switch moves the Hart from %.2f to %.2f, want %.2f (the blocker charge)", off, on, want)
	}
}

// The curve prices a row that gives up its source for lands, and nothing
// else: a ramp sorcery that puts two lands onto the battlefield keeps
// ADR 0126 §2's hand deficit.
func TestRampCurveLeavesRampSpellsAlone(t *testing.T) {
	cultivate := spell(cardID(60), 0, "Cultivate", "{2}{G}")
	cultivate.TypeLine = "Sorcery"
	cultivate.Purpose = &protocol.PurposeView{Lands: 2}
	var bf []protocol.CardView
	for i := range 4 {
		bf = append(bf, land(cardID(40+i), 0))
	}
	v := newView([]protocol.PlayerView{newSeat(0, withHand(append(cheapHand(), cultivate)...)), newSeat(1)},
		withBattlefield(bf...), withTurn(4, 0, "precombat_main"))
	in := input(0, v, passMove(0), castMove(t, 0, cardID(60), "Cast Cultivate"))
	on, _ := priceOf(curve(), in)
	off, _ := priceOf(noCurve(), in)
	if on != off {
		t.Errorf("the curve moves Cultivate's cast from %.2f to %.2f, want it unchanged", off, on)
	}
}

// A row that sacrifices its source but declares no lands is not on the
// curve, and keeps every charge it had.
func TestRampCurveNeedsDeclaredLands(t *testing.T) {
	in := hartBoard(t, 4, 4, cheapHand())
	in.View.Battlefield.Cards[0].ActivatedAbilities[0].Purpose = &protocol.PurposeView{Draws: 1, Answers: &[]string{"value"}}
	on, _ := priceOf(curve(), in)
	off, _ := priceOf(noCurve(), in)
	if on != off {
		t.Errorf("a self-sacrificing draw row moves from %.2f to %.2f with the curve on, want it unchanged", off, on)
	}
}

// The switch ships off (it won no more in the arena, docs/bot-testing.md),
// leading the round by two when it is turned on, and is off in the
// baseline.
func TestPriceRampCurveDefaults(t *testing.T) {
	if cfg := heuristic.DefaultConfig(); cfg.PriceRampCurve || cfg.RampCurveLead != 2 {
		t.Errorf("DefaultConfig: PriceRampCurve %v, RampCurveLead %d; want off, two", cfg.PriceRampCurve, cfg.RampCurveLead)
	}
	if cfg := heuristic.BaselineConfig(); cfg.PriceRampCurve || cfg.RampCurveLead != 0 {
		t.Errorf("BaselineConfig: PriceRampCurve %v, RampCurveLead %d; want both zero", cfg.PriceRampCurve, cfg.RampCurveLead)
	}
}
