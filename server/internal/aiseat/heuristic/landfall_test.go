package heuristic_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// landfall_test.go — #3072: Burnished Hart's activation was offered 346
// times on simic-ramp and taken once (#2480), with Tatyova on the
// battlefield in most of those windows. A land the bot puts onto the
// battlefield is now worth what its own landfall payoffs pay for it.

const hartLabel = "{3}, Sacrifice Burnished Hart: Search your library for up to two basic land cards, put them onto the battlefield tapped, then shuffle."

func burnishedHart(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Burnished Hart", 2, 2)
	c.TypeLine, c.ManaCost = "Artifact Creature — Elk", "{3}"
	c.ActivatedAbilities = []protocol.ActivatedAbilityView{{
		Index: 0, Ref: "own:0", Label: hartLabel, SacrificeSelf: true,
		Purpose: &protocol.PurposeView{Lands: 2, Answers: &[]string{"value"}},
	}}
	return c
}

// withLandfallPayoff gives a permanent a triggered row with that payoff.
func withLandfallPayoff(c protocol.CardView, l protocol.LandfallPayoffView) protocol.CardView {
	c.AbilityRows = append(c.AbilityRows, protocol.AbilityRowView{
		Kind: "triggered", Label: "landfall",
		Purpose: &protocol.PurposeView{LandfallPayoff: &l},
	})
	return c
}

func tatyova(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Tatyova, Benthic Druid", 3, 3)
	c.TypeLine, c.ManaCost = "Legendary Creature — Merfolk Druid", "{3}{G}{U}"
	return withLandfallPayoff(c, protocol.LandfallPayoffView{Draws: 1, LifeGain: 1})
}

func courser(id string, controller int) protocol.CardView {
	c := creature(id, controller, "Courser of Kruphix", 2, 4)
	c.TypeLine, c.ManaCost = "Enchantment Creature — Centaur", "{1}{G}{G}"
	return withLandfallPayoff(c, protocol.LandfallPayoffView{LifeGain: 1})
}

func hartMove(t *testing.T, seat int, src string) legal.Move {
	return legal.Move{
		Type: legal.TypeActivateAbility, Player: seatID(seat), Kind: legal.KindActivate,
		Label: "Burnished Hart: " + hartLabel, Source: uuid.MustParse(src),
		Params: mustJSON(t, map[string]any{"source_card_id": src, "ability_index": 0}),
	}
}

// hartRank prices the Hart's activation and passing for seat 0 in its
// own precombat main phase on turn 8, with seven lands (the mana
// deficit closed) and `others` on the battlefield beside the Hart.
func hartRank(t *testing.T, cfg heuristic.Config, others ...protocol.CardView) (act, pass float64) {
	t.Helper()
	bf := append([]protocol.CardView{burnishedHart(cardID(1), 0)}, others...)
	for i := range 7 {
		bf = append(bf, land(cardID(40+i), 0))
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(8, 0, "precombat_main"))
	in := input(0, v, passMove(0), hartMove(t, 0, cardID(1)))
	vals := map[int]float64{}
	for _, c := range heuristic.NewWithConfig(cfg).Rank(context.Background(), in) {
		vals[c.Index] = c.Value
	}
	return vals[1], vals[0]
}

// With Tatyova out, the Hart's two lands are two cards and two life:
// the bot activates it.
func TestHartIsActivatedWithTatyovaOut(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	act, pass := hartRank(t, cfg, tatyova(cardID(2), 0))
	if act <= cfg.PassThreshold || act <= pass {
		t.Errorf("Hart with Tatyova out: %.2f against pass %.2f — want it activated", act, pass)
	}
	without, _ := hartRank(t, cfg)
	t.Logf("Hart: %.2f with Tatyova, %.2f without, pass %.2f", act, without, pass)
	if got, want := act-without, 2*(1.2+0.0); got < want-1e-9 {
		t.Errorf("Tatyova adds %.2f to the Hart's two lands, want at least two cards' worth (%.2f)", got, want)
	}

	// The decision, not only the price.
	bf := []protocol.CardView{burnishedHart(cardID(1), 0), tatyova(cardID(2), 0)}
	for i := range 7 {
		bf = append(bf, land(cardID(40+i), 0))
	}
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2), newSeat(3)},
		withBattlefield(bf...), withTurn(8, 0, "precombat_main"))
	in := input(0, v, passMove(0), hartMove(t, 0, cardID(1)))
	if d := decide(t, heuristic.New(), in); d.Index != 1 {
		t.Errorf("decided move %d (%s), want the Hart's activation", d.Index, d.Reason)
	}
}

// Without a payoff, or with one that is not worth the Hart's body, the
// two lands are still mana alone once the deficit has closed, and the
// bot keeps the Hart, as before #3072.
func TestHartIsKeptWhenTheLandsPayLittle(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	for _, tc := range []struct {
		name   string
		others []protocol.CardView
	}{
		{"no payoff", nil},
		// Two life is not worth a 2/2.
		{"Courser of Kruphix only", []protocol.CardView{courser(cardID(2), 0)}},
		// An opponent's Tatyova pays the opponent.
		{"an opponent's Tatyova", []protocol.CardView{tatyova(cardID(2), 1)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			act, pass := hartRank(t, cfg, tc.others...)
			t.Logf("Hart: %.2f, pass %.2f", act, pass)
			if act >= cfg.PassThreshold || act >= pass {
				t.Errorf("Hart: %.2f against pass %.2f — want it kept", act, pass)
			}
		})
	}
}

// The switch: off, Tatyova adds nothing and the Hart is kept.
func TestHartWithTatyovaNeedsTheSwitch(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	cfg.PriceLandfallPayoffs = false
	if act, pass := hartRank(t, cfg, tatyova(cardID(2), 0)); act >= cfg.PassThreshold || act >= pass {
		t.Errorf("PriceLandfallPayoffs off: Hart with Tatyova %.2f against pass %.2f — want it kept", act, pass)
	}
	if heuristic.BaselineConfig().PriceLandfallPayoffs {
		t.Error("BaselineConfig prices landfall payoffs; want it off")
	}
	if !heuristic.DefaultConfig().PriceLandfallPayoffs {
		t.Error("DefaultConfig does not price landfall payoffs; want it on")
	}
}

// A ramp spell is worth more with Tatyova out: Harrow's two lands are
// two cards.
func TestARampSpellIsWorthMoreWithTatyovaOut(t *testing.T) {
	harrow := spell(cardID(3), 0, "Harrow", "{2}{G}")
	harrow.Purpose = &protocol.PurposeView{Lands: 2, LandsUntapped: 2}
	rank := func(others ...protocol.CardView) float64 {
		bf := append([]protocol.CardView{}, others...)
		for i := range 7 {
			bf = append(bf, land(cardID(40+i), 0))
		}
		v := newView([]protocol.PlayerView{newSeat(0, withHand(harrow)), newSeat(1)},
			withBattlefield(bf...), withTurn(8, 0, "precombat_main"))
		in := input(0, v, passMove(0), castMove(t, 0, harrow.InstanceID, "Cast Harrow"))
		vals := map[int]float64{}
		for _, c := range heuristic.New().Rank(context.Background(), in) {
			vals[c.Index] = c.Value
		}
		return vals[1]
	}
	without, with := rank(), rank(tatyova(cardID(2), 0))
	if with-without < 2*1.2-1e-9 {
		t.Errorf("Harrow is %.2f without Tatyova and %.2f with her, want at least two cards more", without, with)
	}
}

// Each kind of payoff is priced: a token, a counter, energy, damage to
// each opponent and mana all make the Hart's two lands worth more.
func TestEveryLandfallAmountIsPriced(t *testing.T) {
	cfg := heuristic.DefaultConfig()
	base, _ := hartRank(t, cfg)
	for _, tc := range []struct {
		name string
		pay  protocol.LandfallPayoffView
	}{
		{"draws", protocol.LandfallPayoffView{Draws: 1}},
		{"life", protocol.LandfallPayoffView{LifeGain: 1}},
		{"tokens", protocol.LandfallPayoffView{Tokens: 1}},
		{"mana", protocol.LandfallPayoffView{Mana: 1}},
		{"counters", protocol.LandfallPayoffView{Counters: 1}},
		{"energy", protocol.LandfallPayoffView{Energy: 1}},
		{"damage", protocol.LandfallPayoffView{DamageEachOpponent: 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payoff := withLandfallPayoff(creature(cardID(2), 0, "Payoff", 1, 1), tc.pay)
			if act, _ := hartRank(t, cfg, payoff); act <= base {
				t.Errorf("Hart with a %s payoff %.2f, without %.2f — want more", tc.name, act, base)
			}
		})
	}
}
