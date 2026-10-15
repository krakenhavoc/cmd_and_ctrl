package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// temporary_body_test.go — #3802 and #3776 on hand-built views; the same
// shapes through the real engine are in internal/aiseat.
//
// #3802: a creature cast for evoke, warp, dash or blitz is a body that
// leaves this turn. Its price drops the lasting body; a haste attack that
// is still ahead and the card it returns or draws are added back.
//
// #3776: the pump of a turn-face-up trigger is priced by the attack it
// enables this turn.

const tbHandID = 1

// altCast is the cast move that claims alternative cost `key`, as the
// enumerator labels it ("" is the hard cast).
func altCast(t *testing.T, id, label, key string) legal.Move {
	m := castMove(t, 0, id, label)
	m.Params = mustJSON(t, map[string]any{"instance_id": id, "from_zone": "hand", "alternative_cost": key})
	return m
}

// tbView is a four-seat table at `step` on seat 0's turn with `card` in
// hand, the opponents' life set to oppLife, and `extra` on the battlefield.
func tbView(card protocol.CardView, step string, oppLife int, extra ...protocol.CardView) protocol.GameView {
	seats := []protocol.PlayerView{
		newSeat(0, withHand(card)),
		newSeat(1, withLife(oppLife)), newSeat(2, withLife(oppLife)), newSeat(3, withLife(oppLife)),
	}
	return newView(seats, withBattlefield(append(manaLands(8, 0, 500), extra...)...), withTurn(6, 0, step))
}

func withAltCosts(c protocol.CardView, costs ...protocol.AlternativeCostView) protocol.CardView {
	c.AlternativeCosts = costs
	return c
}

// A pair of values for the same hand card: the hard cast and the cast
// that claims `key`.
func tbPrice(t *testing.T, pol *heuristic.Policy, c protocol.CardView, key, step string, oppLife int) (hard, alt float64) {
	id := cardID(tbHandID)
	v := tbView(c, step, oppLife)
	h := altCast(t, id, "Cast "+c.Name, "")
	a := altCast(t, id, "Cast "+c.Name+" ("+key+")", key)
	in := input(0, v, passMove(0), h, a)
	return rankValue(t, pol, in, h.Label), rankValue(t, pol, in, a.Label)
}

func noTemporaryBodies() *heuristic.Policy {
	cfg := heuristic.DefaultConfig()
	cfg.PriceTemporaryBodies = false
	return heuristic.NewWithConfig(cfg)
}

// Evoke: the body is sacrificed, so the evoked cast is worth the enters
// effect and the hard cast is worth that and the body.
func TestEvokedCreatureIsWorthLessThanTheHardCast(t *testing.T) {
	c := creature(cardID(tbHandID), 0, "Mulldrifter", 2, 2)
	c.ManaCost = "{4}{U}"
	c.Purpose = &protocol.PurposeView{Draws: 2}
	c = withAltCosts(c, protocol.AlternativeCostView{Key: "evoke", Label: "Evoke {2}{U}", ManaCost: "{2}{U}"})

	hard, evoke := tbPrice(t, heuristic.New(), c, "evoke", "precombat_main", 40)
	if evoke >= hard {
		t.Errorf("evoked %.3f is no worse than the hard cast %.3f: the body is sacrificed", evoke, hard)
	}
	// Off, the two are priced as the same permanent.
	hard, evoke = tbPrice(t, noTemporaryBodies(), c, "evoke", "precombat_main", 40)
	if !nearly(evoke, hard) {
		t.Errorf("with PriceTemporaryBodies off, evoked %.3f and hard cast %.3f differ", evoke, hard)
	}
}

// The enters effect is still counted on an evoke: the evoked Mulldrifter
// beats the same creature evoked with nothing to enter.
func TestEvokedCreatureKeepsItsEntersEffect(t *testing.T) {
	plain := creature(cardID(tbHandID), 0, "Mulldrifter", 2, 2)
	plain.ManaCost = "{4}{U}"
	plain = withAltCosts(plain, protocol.AlternativeCostView{Key: "evoke", ManaCost: "{2}{U}"})
	draws := plain
	draws.Purpose = &protocol.PurposeView{Draws: 2}
	_, withDraw := tbPrice(t, heuristic.New(), draws, "evoke", "precombat_main", 40)
	_, without := tbPrice(t, heuristic.New(), plain, "evoke", "precombat_main", 40)
	if withDraw <= without {
		t.Errorf("an evoked creature that draws two (%.3f) is worth no more than one that does nothing (%.3f)", withDraw, without)
	}
}

// Blitz and dash give a haste attack, so a lethal swing outweighs the
// body: the opponents are at 5, the creature has 5 power and nothing
// blocks. Warp gives nothing of the sort without haste.
func TestHasteAlternativeCostsBeatTheHardCastOnlyForALethalSwing(t *testing.T) {
	for _, tc := range []struct{ key, name string }{{"blitz", "Workshop Warchief"}, {"dash", "Ragavan"}} {
		t.Run(tc.key, func(t *testing.T) {
			c := creature(cardID(tbHandID), 0, tc.name, 5, 3)
			c.ManaCost = "{3}{G}{G}"
			c = withAltCosts(c, protocol.AlternativeCostView{Key: tc.key, ManaCost: "{4}{G}{G}"})

			hard, alt := tbPrice(t, heuristic.New(), c, tc.key, "precombat_main", 40)
			if alt >= hard {
				t.Errorf("%s: with a healthy table the alternative %.3f is no worse than the hard cast %.3f", tc.key, alt, hard)
			}
			hard, alt = tbPrice(t, heuristic.New(), c, tc.key, "precombat_main", 5)
			if alt <= hard {
				t.Errorf("%s: for lethal the alternative %.3f is no better than the hard cast %.3f", tc.key, alt, hard)
			}
			// The attack is gone in the second main phase.
			_, late := tbPrice(t, heuristic.New(), c, tc.key, "postcombat_main", 5)
			if late >= alt {
				t.Errorf("%s: after combat the alternative still prices the swing: %.3f vs %.3f", tc.key, late, alt)
			}
			// Off, the two are the same permanent.
			hard, alt = tbPrice(t, noTemporaryBodies(), c, tc.key, "precombat_main", 5)
			if !nearly(alt, hard) {
				t.Errorf("%s: with PriceTemporaryBodies off, alternative %.3f and hard cast %.3f differ", tc.key, alt, hard)
			}
		})
	}
}

// Warp: a creature with haste of its own swings when warped; one without
// does not, so the same lethal board is worth more to the hasty one.
func TestWarpedCreatureSwingsOnlyWithHaste(t *testing.T) {
	warp := protocol.AlternativeCostView{Key: "warp", ManaCost: "{B}"}
	slow := creature(cardID(tbHandID), 0, "Slow Warper", 5, 3)
	slow.ManaCost = "{3}{B}"
	slow = withAltCosts(slow, warp)
	fast := slow
	fast.Name = "Fast Warper"
	fast.Abilities = []string{"haste"}

	_, slowAlt := tbPrice(t, heuristic.New(), slow, "warp", "precombat_main", 5)
	_, fastAlt := tbPrice(t, heuristic.New(), fast, "warp", "precombat_main", 5)
	if fastAlt <= slowAlt {
		t.Errorf("a warped creature with haste (%.3f) is worth no more than one without (%.3f) with a lethal swing on", fastAlt, slowAlt)
	}
	hard, alt := tbPrice(t, heuristic.New(), slow, "warp", "precombat_main", 40)
	if alt >= hard {
		t.Errorf("a warped creature %.3f is no worse than the hard cast %.3f", alt, hard)
	}
}

// The card back: with the attack out of the picture (second main phase),
// a blitz (its death draws) and a dash or a warp (the card is kept) are
// each worth one card more than an evoke, which spends it for nothing.
func TestBlitzDashAndWarpGiveTheCardBack(t *testing.T) {
	pol := heuristic.New()
	hand := pol.Config().Weights.Hand
	c := creature(cardID(tbHandID), 0, "Test Creature", 3, 3)
	c.ManaCost = "{2}{R}"
	c = withAltCosts(c,
		protocol.AlternativeCostView{Key: "evoke", ManaCost: "{R}"},
		protocol.AlternativeCostView{Key: "blitz", ManaCost: "{R}"},
		protocol.AlternativeCostView{Key: "dash", ManaCost: "{R}"},
		protocol.AlternativeCostView{Key: "warp", ManaCost: "{R}"})
	_, evoke := tbPrice(t, pol, c, "evoke", "postcombat_main", 40)
	for _, key := range []string{"blitz", "dash", "warp"} {
		_, got := tbPrice(t, pol, c, key, "postcombat_main", 40)
		if !nearly(got-evoke, hand) {
			t.Errorf("%s is worth %.3f more than an evoke, want one card (%.3f)", key, got-evoke, hand)
		}
	}
}

// A bound offer ("blitz-graveyard") is the same keyword.
func TestBoundBlitzOfferIsPricedAsBlitz(t *testing.T) {
	c := creature(cardID(tbHandID), 0, "Tenacious Underdog", 3, 2)
	c = withAltCosts(c,
		protocol.AlternativeCostView{Key: "blitz", ManaCost: "{2}{B}"},
		protocol.AlternativeCostView{Key: "blitz-graveyard", ManaCost: "{2}{B}"})
	_, a := tbPrice(t, heuristic.New(), c, "blitz", "precombat_main", 40)
	_, b := tbPrice(t, heuristic.New(), c, "blitz-graveyard", "precombat_main", 40)
	if !nearly(a, b) {
		t.Errorf("blitz %.3f and blitz-graveyard %.3f are priced differently", a, b)
	}
}

// --- turn face up with a pump -------------------------------------------

func hulkPurpose(pump bool) *protocol.PurposeView {
	p := &protocol.PurposeView{SkipsTurns: 1}
	if pump {
		p.Pump = &protocol.PumpView{Power: 6, Toughness: 6, Keywords: []string{"trample"}}
	}
	return &protocol.PurposeView{TurnedFaceUp: p}
}

func hulkBoard(oppLife int, step string, hulk protocol.CardView, extra ...protocol.CardView) protocol.GameView {
	seats := []protocol.PlayerView{newSeat(0), newSeat(1, withLife(oppLife)), newSeat(2, withLife(oppLife)), newSeat(3, withLife(oppLife))}
	bf := append(manaLands(5, 0, 100), hulk)
	return newView(seats, withBattlefield(append(bf, extra...)...), withTurn(6, 0, step))
}

// A flip for lethal is taken: every opponent at 8 with an empty board, the
// face-down 2/2 deals 2 and the 8/8 deals 8.
func TestTurningAHulkUpForLethalOutweighsTheSkippedTurn(t *testing.T) {
	id := cardID(1)
	pol := heuristic.New()
	flat := pol.Config().SpecialActionValue
	turn := pol.Config().Weights.Turn
	m := turnUpMove(id)
	hulk := facedown(id, 0, hulkPurpose(true))

	lethal := rankValue(t, pol, input(0, hulkBoard(8, "precombat_main", hulk), passMove(0), m), m.Label)
	if lethal <= pol.Config().PassThreshold {
		t.Errorf("a flip for lethal priced %.3f, want it above passing", lethal)
	}
	if lethal <= flat-turn+pol.Config().LethalBonus/2 {
		t.Errorf("a flip for lethal priced %.3f, want the lethal bonus on top of the %.3f a flip costs", lethal, flat-turn)
	}
	// 40 life: six more damage is not worth a turn.
	healthy := rankValue(t, pol, input(0, hulkBoard(40, "precombat_main", hulk), passMove(0), m), m.Label)
	if healthy >= pol.Config().PassThreshold {
		t.Errorf("a flip that kills nobody priced %.3f, want it below passing", healthy)
	}
	// The same lethal board after combat is the plain flat-minus-turn.
	if got := rankValue(t, pol, input(0, hulkBoard(8, "postcombat_main", hulk), passMove(0), m), m.Label); !nearly(got, flat-turn) {
		t.Errorf("after combat the flip priced %.3f, want the flat %.3f less a turn", got, flat-turn)
	}
}

// A Hulk that cannot attack this turn gains nothing from the pump.
func TestTurningASummoningSickHulkUpIsNotAttackPriced(t *testing.T) {
	id := cardID(1)
	pol := heuristic.New()
	m := turnUpMove(id)
	hulk := facedown(id, 0, hulkPurpose(true))
	hulk.SummoningSick = true
	want := pol.Config().SpecialActionValue - pol.Config().Weights.Turn
	if got := rankValue(t, pol, input(0, hulkBoard(8, "precombat_main", hulk), passMove(0), m), m.Label); !nearly(got, want) {
		t.Errorf("a summoning-sick Hulk's flip priced %.3f, want %.3f", got, want)
	}
	hulk.SummoningSick, hulk.Tapped = false, true
	if got := rankValue(t, pol, input(0, hulkBoard(8, "precombat_main", hulk), passMove(0), m), m.Label); !nearly(got, want) {
		t.Errorf("a tapped Hulk's flip priced %.3f, want %.3f", got, want)
	}
}

// Trample is read: a 2/2 chump blocker stops an unpumped 2/2 but an 8/8
// trampler carries six past it, which is lethal at 6.
func TestTheHulksTrampleCarriesThePumpPastABlocker(t *testing.T) {
	id := cardID(1)
	pol := heuristic.New()
	m := turnUpMove(id)
	chump := func(controller int) protocol.CardView {
		return creature(cardID(900+controller), controller, "Squire", 2, 2)
	}
	board := func(p *protocol.PurposeView) protocol.GameView {
		return hulkBoard(6, "precombat_main", facedown(id, 0, p), chump(1), chump(2), chump(3))
	}
	with := rankValue(t, pol, input(0, board(hulkPurpose(true)), passMove(0), m), m.Label)

	noTrample := hulkPurpose(true)
	noTrample.TurnedFaceUp.Pump.Keywords = nil
	without := rankValue(t, pol, input(0, board(noTrample), passMove(0), m), m.Label)
	if with <= without+pol.Config().LethalBonus/2 {
		t.Errorf("flip with trample %.3f vs without %.3f: the trample should be what makes it lethal", with, without)
	}
}

// Off, or declaring no pump, the flip is the flat value less the turn.
func TestAHulkFlipWithoutAPumpOrWithTheSwitchOffIsFlatLessATurn(t *testing.T) {
	id := cardID(1)
	m := turnUpMove(id)
	pol := heuristic.New()
	want := pol.Config().SpecialActionValue - pol.Config().Weights.Turn

	noPump := facedown(id, 0, hulkPurpose(false))
	if got := rankValue(t, pol, input(0, hulkBoard(8, "precombat_main", noPump), passMove(0), m), m.Label); !nearly(got, want) {
		t.Errorf("a flip with no declared pump priced %.3f, want %.3f", got, want)
	}
	cfg := heuristic.DefaultConfig()
	cfg.PriceFaceUpPump = false
	off := heuristic.NewWithConfig(cfg)
	hulk := facedown(id, 0, hulkPurpose(true))
	if got := rankValue(t, off, input(0, hulkBoard(8, "precombat_main", hulk), passMove(0), m), m.Label); !nearly(got, want) {
		t.Errorf("with PriceFaceUpPump off the flip priced %.3f, want %.3f", got, want)
	}
}
